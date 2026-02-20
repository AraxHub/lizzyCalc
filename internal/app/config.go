package app

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"lizzyCalc/internal/api/http"
	"lizzyCalc/internal/infrastructure/click"
	"lizzyCalc/internal/infrastructure/kafka"
	"lizzyCalc/internal/infrastructure/pg"
	"lizzyCalc/internal/infrastructure/redis"
)

const AppName = "CALCULATOR"

// GrpcConfig — настройки gRPC-сервера. Переменные: CALCULATOR_GRPC_HOST, CALCULATOR_GRPC_PORT.
type GrpcConfig struct {
	Host string `env:"HOST" default:"0.0.0.0"`
	Port string `env:"PORT" default:"9090"`
}

// Config — конфиг приложения. Заполняется через envconfig с префиксом CALCULATOR.
type Config struct {
	Server http.ServerConfig `env:"SERVER"`
	Grpc   GrpcConfig        `env:"GRPC"`
	DB     pg.Config         `env:"DB"`
	Redis      redis.Config      `env:"REDIS"`
	Kafka      kafka.Config      `env:"KAFKA"`
	ClickHouse click.Config      `env:"CLICKHOUSE"`
}

// LoadCfg загружает конфиг: подтягивает .env (godotenv), затем заполняет структуру из окружения (envconfig).
func LoadCfg() (Config, error) {
	if err := godotenv.Load("/Users/admin/liz education/lizzyCalc/deployment/localCalc/.env"); err != nil {
		log.Printf("config: .env не найден, используем окружение: %v", err)
	}

	var cfg Config
	if err := envconfig.Process(AppName, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
