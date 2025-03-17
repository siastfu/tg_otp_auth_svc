package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"tg_otp_auth_svc/internal/domain"
	"time"

	"tg_otp_auth_svc/config"
	grpcHandler "tg_otp_auth_svc/internal/delivery/grpc"
	"tg_otp_auth_svc/internal/infra/repository"
	"tg_otp_auth_svc/internal/infra/usecase"
	"tg_otp_auth_svc/pkg/database"
	"tg_otp_auth_svc/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "tg_otp_auth_svc/proto"

	"github.com/joho/godotenv"
	"tg_otp_auth_svc/internal/delivery/telegram" // Добавляем поддержку Telegram-бота
)

func main() {
	// Загружаем .env файл (если он есть)
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ Не удалось загрузить .env файл, используются переменные окружения")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	logger.InitLogger(cfg.App.Env)
	defer logger.Sync()

	database.ConnectDB(cfg)
	defer database.CloseDB()

	authRepo := &repository.PostgresAuthAttemptRepository{}
	userRepo := &repository.PostgresUserRepository{}     // ✅ Добавляем UserRepo
	authUC := usecase.NewAuthUseCase(authRepo, userRepo) // ✅ Теперь передаем оба аргумента

	authHandler := grpcHandler.NewAuthHandler(authUC)

	// Получаем токен Telegram-бота
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("❌ Ошибка: переменная TELEGRAM_BOT_TOKEN не задана!")
	}

	authUCImpl, ok := authUC.(*usecase.AuthUseCaseImpl)
	if !ok {
		log.Fatal("Ошибка приведения authUC к *usecase.AuthUseCaseImpl")
	}

	// Создаём и запускаем Telegram-бота
	tgHandler, err := telegram.NewTelegramHandler(authUCImpl, botToken)
	if err != nil {
		log.Fatalf("Ошибка запуска Telegram-бота: %v", err)
	}

	go tgHandler.Run()

	// Запускаем gRPC-сервер
	go startGRPCServer(cfg, authHandler)

	// Запускаем воркер для проверки истекших ссылок
	go startWorker(authUC)

	logger.Logger.Info("Приложение запущено", zap.String("env", cfg.App.Env))
	fmt.Println("✅ Приложение успешно запущено!")

	WaitForShutdown()
}

// startWorker запускает фоновую задачу для обработки истекших ссылок
func startWorker(authUC domain.AuthUseCase) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		fmt.Println("🔄 Проверяем истекшие ссылки...")
		err := authUC.MarkExpiredAuthAttempts()
		if err != nil {
			logger.Logger.Error("Ошибка при обработке истекших ссылок", zap.Error(err))
		}
	}
}

// startGRPCServer запускает gRPC-сервер
func startGRPCServer(cfg *config.Config, authHandler *grpcHandler.AuthHandler) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Fatalf("Ошибка запуска gRPC-сервера: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, authHandler)
	reflection.Register(grpcServer)

	fmt.Println("✅ gRPC-сервер запущен на порту", cfg.GRPC.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Ошибка работы gRPC-сервера: %v", err)
	}
}

// WaitForShutdown - корректное завершение работы приложения
func WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	logger.Logger.Info("Получен сигнал завершения", zap.String("signal", sig.String()))

	fmt.Println("🛑 Завершаем работу приложения...")
	logger.Logger.Sync()
}
