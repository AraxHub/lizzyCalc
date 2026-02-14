package main

import (
	"log/slog"
	"os"

	"lizzyCalc/internal/app"
)

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
