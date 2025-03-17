package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"tg_otp_auth_svc/config"
	"tg_otp_auth_svc/internal/delivery/grpc"

	"tg_otp_auth_svc/internal/infra/repository"
	"tg_otp_auth_svc/internal/infra/usecase"
	"tg_otp_auth_svc/pkg/database"
	"tg_otp_auth_svc/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "tg_otp_auth_svc/proto"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
	}

	// Определяем окружение
	env := cfg.App.Env
	fmt.Println("Окружение из конфига:", env)

	// Инициализируем логгер
	logger.InitLogger(env)
	defer logger.Sync()

	// Подключаемся к базе данных
	database.ConnectDB(cfg)
	defer database.CloseDB()

	// Инициализируем репозитории
	userRepo := &repository.PostgresUserRepository{}
	authRepo := &repository.PostgresAuthAttemptRepository{}

	// Инициализируем Use Case
	userUC := usecase.NewUserUseCase(userRepo)
	authUC := usecase.NewAuthUseCase(authRepo)

	// Инициализируем gRPC-хендлер
	authHandler := delivery.NewAuthHandler(authUC)

	// Запускаем gRPC-сервер
	go startGRPCServer(cfg, authHandler)

	// Логируем успешный запуск
	logger.Logger.Info("Приложение запущено", zap.String("env", env))
	fmt.Println("✅ Приложение успешно запущено!")

	// Ожидание завершения работы приложения
	WaitForShutdown()
}

// startGRPCServer запускает gRPC-сервер
func startGRPCServer(cfg *config.Config, authHandler *delivery.AuthHandler) {
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
