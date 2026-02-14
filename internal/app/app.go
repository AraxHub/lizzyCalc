package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	apihttp "lizzyCalc/internal/api/http"
	"lizzyCalc/internal/api/http/controllers/calculator"
	"lizzyCalc/internal/api/http/controllers/system"
	"lizzyCalc/internal/infrastructure/pg"
	"lizzyCalc/internal/pkg/logger"
	calclUsecase "lizzy
	calclUsecase "lizzyCalc/internal/usecase/calculator"
)

// App — приложение, хранит только конфиг.
type App struct {
	cfg Config
}

// New создаёт приложение с конфигом (БД подключается в Run).
func New(cfg Config) *App {
	return &App{cfg: cfg}
}

// Run подключается к БД, линейно инициализирует зависимости и запускает HTTP-сервер (блокирующий вызов).
func (a *App) Run() error {
	db, err := pg.New(&a.cfg.DB)
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}
	defer db.Close()

	if err := pg.Migrate(context.Background(), db); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	log := logger.New()
	slog.SetDefault(log)

	repo := pg.NewOperationRepo(db, log)

	uc := calclUsecase.New(repo, log)

	srv := apihttp.NewServer(a.cfg.Server)
	srv.AddController(
		system.New(repo, log),
		calculator.New(uc, log))

	addr := a.cfg.Server.Host + ":" + a.cfg.Server.Port
	slog.Info("application started", "addr", addr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return srv.Start(ctx)
}
