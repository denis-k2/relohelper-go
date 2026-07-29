package main

import (
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleTime  time.Duration
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	auth struct {
		enabled bool
	}
	metrics struct {
		port int
	}
	batch struct {
		maxIDs         int
		maxDetailedIDs int
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	exchangeRates struct {
		appID string
	}
}

type envLookup func(string) (string, bool)

func loadConfig(lookup envLookup) (config, error) {
	cfg, err := readConfig(lookup)
	if err != nil {
		return config{}, err
	}
	if err := cfg.validate(); err != nil {
		return config{}, err
	}

	return cfg, nil
}

func readConfig(lookup envLookup) (config, error) {
	var cfg config
	var err error

	cfg.port, err = envInt(lookup, "RELOHELPER_PORT", 4000)
	if err != nil {
		return config{}, err
	}
	cfg.env = envString(lookup, "RELOHELPER_ENV", "development")

	cfg.db.dsn = envString(lookup, "RELOHELPER_DB_DSN", "")
	cfg.db.maxOpenConns, err = envInt(lookup, "RELOHELPER_DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return config{}, err
	}
	cfg.db.maxIdleTime, err = envDuration(lookup, "RELOHELPER_DB_MAX_IDLE_TIME", 15*time.Minute)
	if err != nil {
		return config{}, err
	}

	cfg.limiter.rps, err = envFloat64(lookup, "RELOHELPER_LIMITER_RPS", 10)
	if err != nil {
		return config{}, err
	}
	cfg.limiter.burst, err = envInt(lookup, "RELOHELPER_LIMITER_BURST", 20)
	if err != nil {
		return config{}, err
	}
	cfg.limiter.enabled, err = envBool(lookup, "RELOHELPER_LIMITER_ENABLED", true)
	if err != nil {
		return config{}, err
	}

	cfg.auth.enabled, err = envBool(lookup, "RELOHELPER_AUTH_ENABLED", true)
	if err != nil {
		return config{}, err
	}

	cfg.metrics.port, err = envInt(lookup, "RELOHELPER_METRICS_PORT", 0)
	if err != nil {
		return config{}, err
	}

	cfg.batch.maxIDs, err = envInt(lookup, "RELOHELPER_BATCH_MAX_IDS", 100)
	if err != nil {
		return config{}, err
	}
	cfg.batch.maxDetailedIDs, err = envInt(lookup, "RELOHELPER_BATCH_MAX_DETAILED_IDS", 20)
	if err != nil {
		return config{}, err
	}

	cfg.smtp.host = envString(lookup, "RELOHELPER_SMTP_HOST", "")
	cfg.smtp.port, err = envInt(lookup, "RELOHELPER_SMTP_PORT", 25)
	if err != nil {
		return config{}, err
	}
	cfg.smtp.username = envString(lookup, "RELOHELPER_SMTP_USERNAME", "")
	cfg.smtp.password = envString(lookup, "RELOHELPER_SMTP_PASSWORD", "")
	cfg.smtp.sender = envString(lookup, "RELOHELPER_SMTP_SENDER", "Relohelper <no-reply@relohelper.local>")

	cfg.exchangeRates.appID = envString(lookup, "RELOHELPER_EXCHANGE_RATES_APP_ID", "")

	return cfg, nil
}

func (cfg config) validate() error {
	if cfg.port < 1 || cfg.port > 65535 {
		return fmt.Errorf("RELOHELPER_PORT must be between 1 and 65535")
	}
	switch cfg.env {
	case "development", "staging", "production", "testLogs":
	default:
		return fmt.Errorf("RELOHELPER_ENV must be development, staging, or production")
	}
	if strings.TrimSpace(cfg.db.dsn) == "" {
		return fmt.Errorf("RELOHELPER_DB_DSN is required")
	}
	if cfg.db.maxOpenConns < 1 {
		return fmt.Errorf("RELOHELPER_DB_MAX_OPEN_CONNS must be greater than 0")
	}
	if cfg.db.maxIdleTime < 0 {
		return fmt.Errorf("RELOHELPER_DB_MAX_IDLE_TIME must not be negative")
	}
	if cfg.limiter.rps <= 0 {
		return fmt.Errorf("RELOHELPER_LIMITER_RPS must be greater than 0")
	}
	if cfg.limiter.burst < 1 {
		return fmt.Errorf("RELOHELPER_LIMITER_BURST must be greater than 0")
	}
	if cfg.metrics.port < 0 || cfg.metrics.port > 65535 {
		return fmt.Errorf("RELOHELPER_METRICS_PORT must be 0 or between 1 and 65535")
	}
	if cfg.metrics.port == cfg.port {
		return fmt.Errorf("RELOHELPER_METRICS_PORT must differ from RELOHELPER_PORT")
	}
	if cfg.batch.maxIDs < 1 {
		return fmt.Errorf("RELOHELPER_BATCH_MAX_IDS must be greater than 0")
	}
	if cfg.batch.maxDetailedIDs < 1 || cfg.batch.maxDetailedIDs > cfg.batch.maxIDs {
		return fmt.Errorf("RELOHELPER_BATCH_MAX_DETAILED_IDS must be between 1 and RELOHELPER_BATCH_MAX_IDS")
	}
	if cfg.smtp.port < 1 || cfg.smtp.port > 65535 {
		return fmt.Errorf("RELOHELPER_SMTP_PORT must be between 1 and 65535")
	}

	return nil
}

func parseConfig(args []string, lookup envLookup) (config, bool, error) {
	if len(args) == 1 && (args[0] == "-version" || args[0] == "--version") {
		return config{}, true, nil
	}

	cfg, err := readConfig(lookup)
	if err != nil {
		return config{}, false, err
	}

	flags := flag.NewFlagSet("api", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.IntVar(&cfg.port, "port", cfg.port, "API server port")
	flags.StringVar(&cfg.env, "env", cfg.env, "Runtime environment")
	flags.StringVar(&cfg.db.dsn, "db-dsn", cfg.db.dsn, "PostgreSQL DSN")
	flags.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", cfg.db.maxOpenConns, "PostgreSQL max open connections")
	flags.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", cfg.db.maxIdleTime, "PostgreSQL max connection idle time")
	flags.Float64Var(&cfg.limiter.rps, "limiter-rps", cfg.limiter.rps, "Rate limiter maximum requests per second")
	flags.IntVar(&cfg.limiter.burst, "limiter-burst", cfg.limiter.burst, "Rate limiter maximum burst")
	flags.BoolVar(&cfg.limiter.enabled, "limiter-enabled", cfg.limiter.enabled, "Enable rate limiter")
	flags.BoolVar(&cfg.auth.enabled, "auth-enabled", cfg.auth.enabled, "Enable authentication")
	flags.IntVar(&cfg.metrics.port, "metrics-port", cfg.metrics.port, "Dedicated metrics port")
	flags.IntVar(&cfg.batch.maxIDs, "batch-max-ids", cfg.batch.maxIDs, "Maximum batch IDs")
	flags.IntVar(&cfg.batch.maxDetailedIDs, "batch-max-detailed-ids", cfg.batch.maxDetailedIDs, "Maximum detailed batch IDs")
	flags.StringVar(&cfg.smtp.host, "smtp-host", cfg.smtp.host, "SMTP host")
	flags.IntVar(&cfg.smtp.port, "smtp-port", cfg.smtp.port, "SMTP port")
	flags.StringVar(&cfg.smtp.username, "smtp-username", cfg.smtp.username, "SMTP username")
	flags.StringVar(&cfg.smtp.password, "smtp-password", cfg.smtp.password, "SMTP password")
	flags.StringVar(&cfg.smtp.sender, "smtp-sender", cfg.smtp.sender, "SMTP sender")
	flags.StringVar(&cfg.exchangeRates.appID, "exchange-rates-app-id", cfg.exchangeRates.appID, "Open Exchange Rates app ID")
	displayVersion := flags.Bool("version", false, "Display version and exit")

	if err := flags.Parse(args); err != nil {
		return config{}, false, err
	}
	if flags.NArg() != 0 {
		return config{}, false, fmt.Errorf("unexpected arguments: %s", strings.Join(flags.Args(), " "))
	}
	if *displayVersion {
		return config{}, true, nil
	}
	if err := cfg.validate(); err != nil {
		return config{}, false, err
	}

	return cfg, false, nil
}

func envString(lookup envLookup, name, fallback string) string {
	if value, ok := lookup(name); ok {
		return value
	}
	return fallback
}

func envInt(lookup envLookup, name string, fallback int) (int, error) {
	raw, ok := lookup(name)
	if !ok {
		return fallback, nil
	}

	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, err)
	}
	return value, nil
}

func envFloat64(lookup envLookup, name string, fallback float64) (float64, error) {
	raw, ok := lookup(name)
	if !ok {
		return fallback, nil
	}

	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number: %w", name, err)
	}
	return value, nil
}

func envBool(lookup envLookup, name string, fallback bool) (bool, error) {
	raw, ok := lookup(name)
	if !ok {
		return fallback, nil
	}

	value, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", name, err)
	}
	return value, nil
}

func envDuration(lookup envLookup, name string, fallback time.Duration) (time.Duration, error) {
	raw, ok := lookup(name)
	if !ok {
		return fallback, nil
	}

	value, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", name, err)
	}
	return value, nil
}
