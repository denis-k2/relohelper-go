package main

import (
	"context"
	"expvar"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/denis-k2/relohelper-go/internal/data"
	"github.com/denis-k2/relohelper-go/internal/exchangerates"
	"github.com/denis-k2/relohelper-go/internal/mailer"
	"github.com/denis-k2/relohelper-go/internal/vcs"
)

var (
	version  = vcs.Version()
	revision = vcs.Revision()
)

type application struct {
	config        config
	logger        *slog.Logger
	db            *pgxpool.Pool
	models        data.Models
	mailer        mailer.Mailer
	exchangeRates *exchangerates.Service
	wg            sync.WaitGroup
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "server startup failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, displayVersion, err := parseConfig(os.Args[1:], os.LookupEnv)
	if err != nil {
		return err
	}
	if displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Revision:\t%s\n", revision)
		return nil
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info(
		"configuration loaded",
		"env", cfg.env,
		"api_port", cfg.port,
		"metrics_port", cfg.metrics.port,
		"auth_enabled", cfg.auth.enabled,
		"limiter_enabled", cfg.limiter.enabled,
		"db_max_open_conns", cfg.db.maxOpenConns,
	)

	db, err := openDB(cfg)
	if err != nil {
		return err
	}
	defer func() {
		db.Close()
	}()
	logger.Info("database connection pool established")

	registerMetrics(version, revision, db)
	setDBStatsProvider(db)

	app := &application{
		config:        cfg,
		logger:        logger,
		db:            db,
		models:        data.NewModels(db),
		mailer:        mailer.NewSMTPMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
		exchangeRates: exchangerates.NewService(logger, cfg.exchangeRates.appID),
	}

	err = app.serve()
	if err != nil {
		return err
	}

	return nil
}

func openDB(cfg config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = int32(cfg.db.maxOpenConns)
	poolConfig.MaxConnIdleTime = cfg.db.maxIdleTime

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func registerMetrics(version, revision string, db *pgxpool.Pool) {
	expvar.NewString("version").Set(version)
	expvar.NewString("revision").Set(revision)

	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("database", expvar.Func(func() any {
		stats := db.Stat()
		return map[string]any{
			"acquired_conns":           stats.AcquiredConns(),
			"canceled_acquire_count":   stats.CanceledAcquireCount(),
			"constructing_conns":       stats.ConstructingConns(),
			"empty_acquire_count":      stats.EmptyAcquireCount(),
			"empty_acquire_wait_time":  stats.EmptyAcquireWaitTime().String(),
			"idle_conns":               stats.IdleConns(),
			"max_conns":                stats.MaxConns(),
			"total_conns":              stats.TotalConns(),
			"new_conns_count":          stats.NewConnsCount(),
			"max_lifetime_destroy_cnt": stats.MaxLifetimeDestroyCount(),
			"max_idle_destroy_count":   stats.MaxIdleDestroyCount(),
		}
	}))

	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
}
