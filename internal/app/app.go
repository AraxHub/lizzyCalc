package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	apigrpc "lizzyCalc/internal/api/grpc"
	apihttp "lizzyCalc/internal/api/http"
	"lizzyCalc/internal/api/http/controllers/calculator"
	"lizzyCalc/internal/api/http/controllers/system"
	"lizzyCalc/internal/infrastructure/click"
	"lizzyCalc/internal/infrastructure/kafka"
	"lizzyCalc/internal/infrastructure/mongo"
	"lizzyCalc/internal/infrastructure/pg"
	"lizzyCalc/internal/infrastructure/redis"
	"lizzyCalc/internal/pkg/logger"
	"lizzyCalc/internal/ports"
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

// Run подключается к БД и Redis, инициализирует зависимости и запускает HTTP-сервер (блокирующий вызов).
func (a *App) Run() error {
	log := logger.New()
	slog.SetDefault(log)

	var repo ports.IOperationRepository
	if a.cfg.FeatureFlags.UsePGStorage {
		db, err := pg.New(&a.cfg.DB)
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
		defer db.Close()
		if err := pg.Migrate(context.Background(), db); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
		repo = pg.NewOperationRepo(db, log)
		log.Info("storage: PostgreSQL")
	} else {
		mongoClient, err := mongo.New(context.Background(), &a.cfg.Mongo)
		if err != nil {
			return fmt.Errorf("mongo: %w", err)
		}
		defer func() { _ = mongoClient.Disconnect(context.Background()) }()
		repo = mongo.NewOperationRepo(mongoClient, log)
		log.Info("storage: MongoDB")
	}

	rdb, err := redis.New(&a.cfg.Redis)
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	defer rdb.Close()

	cache := redis.NewCache(rdb, log)

	prod := kafka.NewProducer(&a.cfg.Kafka)
	defer prod.Close()

	ch, err := click.New(&a.cfg.ClickHouse)
	if err != nil {
		return fmt.Errorf("clickhouse: %w", err)
	}
	defer ch.Close()

	analyticsWriter := click.NewOperationWriter(ch)
	if err := analyticsWriter.EnsureTable(context.Background()); err != nil {
		return fmt.Errorf("clickhouse ensure table: %w", err)
	}
	log.Info("clickhouse table default.operations_analytics ensured")

	uc := calclUsecase.New(repo, cache, prod, analyticsWriter, log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	consumer := kafka.NewConsumer(&a.cfg.Kafka, uc, log)
	defer consumer.Close()
	go consumer.Run(ctx)

	grpcAddr := a.cfg.Grpc.Host + ":" + a.cfg.Grpc.Port
	grpcSrv := apigrpc.NewServer(grpcAddr, uc, log)
	go func() {
		if err := grpcSrv.Start(); err != nil {
			slog.Error("grpc server failed", "error", err)
		}
	}()

	srv := apihttp.NewServer(a.cfg.Server)
	srv.AddController(
		system.New(repo, log),
		calculator.New(uc, log))

	httpAddr := a.cfg.Server.Host + ":" + a.cfg.Server.Port
	slog.Info("application started", "http", httpAddr, "grpc", grpcAddr)

	if err := srv.Start(ctx); err != nil {
		return err
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return grpcSrv.Stop(shutdownCtx)
}
