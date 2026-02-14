package app

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"lizzyCalc/internal/api/http"
	"lizzyCalc/internal/infrastructure/pg"
)

const AppName = "CALCULATOR"

// Config — конфиг приложения. Заполняется через envconfig с префиксом APP.
type Config struct {
	Server http.ServerConfig `env:"SERVER"`
	DB     pg.Config         `env:"DB"`
}

// LoadCfg загружает конфиг: подтягивает .env (godotenv), затем заполняет структуру из окружения (envconfig).
func LoadCfg() (Config, error) {
	if err := godotenv.Load("/Users/admin/lizzyCalc/deployment/localCalc/.env"); err != nil {
		log.Printf("config: .env не найден, используем окружение: %v", err)
	}

	var cfg Config
	if err := envconfig.Process(AppName, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
