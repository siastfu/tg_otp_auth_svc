package database

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"tg_otp_auth_svc/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB - глобальный объект подключения к БД
var DB *pgxpool.Pool

// ConnectDB устанавливает соединение с PostgreSQL
func ConnectDB(cfg *config.Config) {
	// Кодируем пароль для корректного URL-формата
	encodedPassword := url.QueryEscape(cfg.Database.Password)

	// Формируем строку подключения к БД
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User, encodedPassword, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode,
	)

	// Настройка пула соединений
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Ошибка парсинга конфигурации БД: %v", err)
	}

	// Устанавливаем параметры пула соединений
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConns = int32(cfg.Database.MaxConnections) // ✅ Теперь берём `max_connections` из конфига
	poolConfig.MinConns = 2
	// Подключаемся к БД
	DB, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	log.Println("✅ Подключение к PostgreSQL установлено")
}

// CloseDB закрывает соединение с БД
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("✅ Соединение с PostgreSQL закрыто")
	}
}
