package logger

import (
	"fmt"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger - глобальный объект логгера
var Logger *zap.Logger

// InitLogger инициализирует логгер в зависимости от окружения
func InitLogger(env string) {
	var cfg zap.Config

	if env == "local" {
		fmt.Println("Выбран режим логирования: JSON (local)")
		cfg = zap.NewDevelopmentConfig()
		cfg.Encoding = "json" // Принудительно используем JSON
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		fmt.Println("Выбран режим логирования: строковый (dev/prod)")
		cfg = zap.NewProductionConfig()
		cfg.Encoding = "console" // Принудительно используем строковой формат
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Создаем логгер
	var err error
	Logger, err = cfg.Build()
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}

	zap.ReplaceGlobals(Logger)
	Logger.Info("Логгер инициализирован", zap.String("env", env))
}

// Sync закрывает логгер перед завершением работы
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
