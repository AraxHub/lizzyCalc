package app

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"lizzyCalc/internal/api/grpc"
	"lizzyCalc/internal/api/http"
	"lizzyCalc/internal/infrastructure/click"
	"lizzyCalc/internal/infrastructure/kafka"
	"lizzyCalc/internal/infrastructure/mongo"
	"lizzyCalc/internal/infrastructure/pg"
	"lizzyCalc/internal/infrastructure/redis"
	"log"
)

const AppName = "CALCULATOR"

// FeatureFlags — фича-флаги. Переменные: CALCULATOR_FEATURE_FLAGS_*.
type FeatureFlags struct {
	UsePGStorage bool `envconfig:"PG"`
}

// Config — конфиг приложения. Заполняется через envconfig с префиксом CALCULATOR.
type Config struct {
	Server       http.ServerConfig `envconfig:"SERVER"`
	Grpc         grpc.Config       `envconfig:"GRPC"`
	FeatureFlags FeatureFlags      `envconfig:"FEATURE"`
	DB           pg.Config         `envconfig:"DB"`
	Redis        redis.Config      `envconfig:"REDIS"`
	Kafka        kafka.Config      `envconfig:"KAFKA"`
	ClickHouse   click.Config      `envconfig:"CLICKHOUSE"`
	Mongo        mongo.Config      `envconfig:"MONGO"`
}

// LoadCfg загружает конфиг: подтягивает .env (godotenv), затем заполняет структуру из окружения (envconfig).
func LoadCfg() (Config, error) {
	if err := godotenv.Load("deployment/localCalc/.env"); err != nil {
		log.Printf("config: .env не найден, используем окружение: %v", err)
	}

	var cfg Config
	if err := envconfig.Process(AppName, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
