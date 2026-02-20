package main

import (
	"log/slog"
	"os"

	"lizzyCalc/internal/app"

	_ "lizzyCalc/docs"
)

// @title LizzyCalc API
// @version 1.0
// @description API калькулятора с кэшированием, хранением истории и аналитикой.
// @description Поддерживает операции: сложение (+), вычитание (-), умножение (*), деление (/).

// @host localhost:8080
// @BasePath /

// @tag.name calculator
// @tag.description Операции калькулятора: вычисления и история

// @tag.name system
// @tag.description Системные эндпоинты: liveness, readiness

func main() {
	cfg, err := app.LoadCfg()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	a := app.New(cfg)
	if err := a.Run(); err != nil {
		slog.Error("run failed", "error", err)
		os.Exit(1)
	}
}
