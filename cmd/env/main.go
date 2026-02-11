// Обучающий пример: godotenv + envconfig в одном файле.
//
// godotenv — загружает переменные из файла .env в os.Environ (локальная разработка).
// envconfig — заполняет структуру из переменных окружения по тегам env.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// ServerConfig — настройки HTTP-сервера (вложенная структура).
// При префиксе APP и теге SERVER переменные: APP_SERVER_HOST, APP_SERVER_PORT.
type ServerConfig struct {
	Host string `env:"HOST" default:"0.0.0.0"`
	Port string `env:"PORT" default:"8080"`
}

// DBConfig — настройки подключения к БД (вложенная структура).
// При префиксе APP и теге DB переменные: APP_DB_HOST, APP_DB_PORT и т.д.
type DBConfig struct {
	Host     string `env:"HOST" default:"localhost"`
	Port     string `env:"PORT" default:"5433"`
	User     string `env:"USER" default:"postgres"`
	Password string `env:"PASSWORD" default:"postgres"`
	DBName   string `env:"NAME" default:"lizzycalc"`
	SSLMode  string `env:"SSLMODE" default:"disable"`
}

// Config — конфиг приложения. Префикс "APP" задаётся в Process("APP", &cfg).
// Все переменные: APP_LOG_LEVEL, APP_SERVER_HOST, APP_SERVER_PORT, APP_DB_HOST, ...
type Config struct {
	LogLevel string       `env:"LOG_LEVEL" default:"info"`
	Server   ServerConfig `env:"SERVER"`
	DB       DBConfig     `env:"DB"`
}

func main() {
	// --- godotenv: загрузка .env в окружение ---
	// Load читает файл .env и добавляет пары KEY=VALUE в os.Environ.
	// Если файла нет — ошибка. Игнорируем её: в прод обычно .env не используют.
	if err := godotenv.Load(); err != nil {
		log.Printf("файл .env не найден (игнорируем): %v", err)
	}
	// После Load() переменные из .env доступны через os.Getenv("APP_PORT") и т.д.

	// --- envconfig: заполнение структуры из окружения ---
	// Process("APP", &cfg) — префикс APP, переменные: APP_SERVER_HOST, APP_DB_PORT, APP_LOG_LEVEL и т.д.
	var cfg Config
	if err := envconfig.Process("APP", &cfg); err != nil {
		log.Fatalf("ошибка конфига: %v", err)
	}

	// Используем конфиг
	fmt.Println("Конфиг из env (префикс APP):")
	fmt.Printf("  LogLevel: %s\n", cfg.LogLevel)
	fmt.Printf("  Server:   %s:%s\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("  DB:       host=%s port=%s user=%s dbname=%s sslmode=%s\n",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.DBName, cfg.DB.SSLMode)

	if v := os.Getenv("APP_SERVER_PORT"); v != "" {
		fmt.Printf("  os.Getenv(\"APP_SERVER_PORT\") = %q\n", v)
	}
}
