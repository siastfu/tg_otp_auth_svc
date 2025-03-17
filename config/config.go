package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Env  string `mapstructure:"env"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"app"`

	Database struct {
		Host           string `mapstructure:"host"`
		Port           int    `mapstructure:"port"`
		User           string `mapstructure:"user"`
		Password       string `mapstructure:"password"`
		DBName         string `mapstructure:"dbname"`
		SSLMode        string `mapstructure:"sslmode"`
		MaxConnections int    `mapstructure:"max_connections"` // ✅ Теперь загружаем `max_connections`
	} `mapstructure:"database"`

	GRPC struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"grpc"`
}

// LoadConfig загружает конфигурацию
func LoadConfig() (*Config, error) {
	viper.SetConfigType("yaml")

	// Определяем окружение (по умолчанию local)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	fmt.Println("Загружается конфиг для окружения:", env) // Отладочный вывод

	// Загружаем конфиг по окружению
	viper.SetConfigName("config." + env)
	viper.AddConfigPath("./config")

	// Читаем конфиг
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Ошибка загрузки конфига: %v", err)
		return nil, err
	}

	// Маппинг значений в структуру
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Ошибка парсинга конфига: %v", err)
		return nil, err
	}

	fmt.Println("Конфиг успешно загружен для окружения:", config.App.Env)
	return &config, nil
}
