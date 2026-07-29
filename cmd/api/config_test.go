package main

import (
	"strings"
	"testing"
	"time"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := loadConfig(mapLookup(map[string]string{
		"RELOHELPER_DB_DSN": "postgres://test",
	}))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.port != 4000 {
		t.Errorf("port = %d, want 4000", cfg.port)
	}
	if cfg.env != "development" {
		t.Errorf("env = %q, want development", cfg.env)
	}
	if cfg.db.maxOpenConns != 25 {
		t.Errorf("max open connections = %d, want 25", cfg.db.maxOpenConns)
	}
	if cfg.db.maxIdleTime != 15*time.Minute {
		t.Errorf("max idle time = %s, want 15m", cfg.db.maxIdleTime)
	}
	if !cfg.auth.enabled || !cfg.limiter.enabled {
		t.Error("authentication and rate limiter should be enabled by default")
	}
	if cfg.metrics.port != 0 {
		t.Errorf("metrics port = %d, want 0", cfg.metrics.port)
	}
}

func TestLoadConfigOverrides(t *testing.T) {
	cfg, err := loadConfig(mapLookup(map[string]string{
		"RELOHELPER_DB_DSN":                 "postgres://configured",
		"RELOHELPER_ENV":                    "production",
		"RELOHELPER_PORT":                   "4100",
		"RELOHELPER_DB_MAX_OPEN_CONNS":      "12",
		"RELOHELPER_DB_MAX_IDLE_TIME":       "30s",
		"RELOHELPER_LIMITER_RPS":            "2.5",
		"RELOHELPER_LIMITER_BURST":          "7",
		"RELOHELPER_LIMITER_ENABLED":        "false",
		"RELOHELPER_AUTH_ENABLED":           "false",
		"RELOHELPER_METRICS_PORT":           "4101",
		"RELOHELPER_BATCH_MAX_IDS":          "50",
		"RELOHELPER_BATCH_MAX_DETAILED_IDS": "10",
		"RELOHELPER_EXCHANGE_RATES_APP_ID":  "exchange-key",
		"RELOHELPER_SMTP_HOST":              "smtp.example.com",
		"RELOHELPER_SMTP_PORT":              "2525",
		"RELOHELPER_SMTP_USERNAME":          "mailer",
		"RELOHELPER_SMTP_PASSWORD":          "secret",
		"RELOHELPER_SMTP_SENDER":            "Relohelper <test@example.com>",
	}))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.port != 4100 || cfg.metrics.port != 4101 {
		t.Errorf("ports = (%d, %d), want (4100, 4101)", cfg.port, cfg.metrics.port)
	}
	if cfg.db.dsn != "postgres://configured" || cfg.db.maxOpenConns != 12 || cfg.db.maxIdleTime != 30*time.Second {
		t.Errorf("unexpected database config: %+v", cfg.db)
	}
	if cfg.limiter.rps != 2.5 || cfg.limiter.burst != 7 || cfg.limiter.enabled {
		t.Errorf("unexpected limiter config: %+v", cfg.limiter)
	}
	if cfg.auth.enabled {
		t.Error("authentication should be disabled")
	}
	if cfg.batch.maxIDs != 50 || cfg.batch.maxDetailedIDs != 10 {
		t.Errorf("unexpected batch config: %+v", cfg.batch)
	}
	if cfg.smtp.host != "smtp.example.com" || cfg.smtp.port != 2525 ||
		cfg.smtp.username != "mailer" || cfg.smtp.password != "secret" ||
		cfg.smtp.sender != "Relohelper <test@example.com>" {
		t.Errorf("unexpected SMTP config: %+v", cfg.smtp)
	}
	if cfg.exchangeRates.appID != "exchange-key" {
		t.Errorf("exchange rates app ID = %q, want exchange-key", cfg.exchangeRates.appID)
	}
}

func TestLoadConfigRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr string
	}{
		{name: "missing DSN", key: "RELOHELPER_DB_DSN", value: "", wantErr: "RELOHELPER_DB_DSN is required"},
		{name: "invalid integer", key: "RELOHELPER_PORT", value: "http", wantErr: "RELOHELPER_PORT must be an integer"},
		{name: "invalid environment", key: "RELOHELPER_ENV", value: "prod", wantErr: "RELOHELPER_ENV"},
		{name: "invalid duration", key: "RELOHELPER_DB_MAX_IDLE_TIME", value: "15", wantErr: "RELOHELPER_DB_MAX_IDLE_TIME"},
		{name: "invalid boolean", key: "RELOHELPER_AUTH_ENABLED", value: "sometimes", wantErr: "RELOHELPER_AUTH_ENABLED"},
		{name: "shared ports", key: "RELOHELPER_METRICS_PORT", value: "4000", wantErr: "must differ"},
		{name: "detailed batch exceeds batch", key: "RELOHELPER_BATCH_MAX_DETAILED_IDS", value: "101", wantErr: "RELOHELPER_BATCH_MAX_DETAILED_IDS"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := map[string]string{
				"RELOHELPER_DB_DSN": "postgres://test",
			}
			values[tt.key] = tt.value

			_, err := loadConfig(mapLookup(values))
			if err == nil {
				t.Fatal("loadConfig() error = nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("loadConfig() error = %q, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestParseConfigVersionDoesNotRequireEnvironment(t *testing.T) {
	_, displayVersion, err := parseConfig([]string{"-version"}, mapLookup(nil))
	if err != nil {
		t.Fatal(err)
	}
	if !displayVersion {
		t.Error("displayVersion = false, want true")
	}
}

func TestParseConfigCLIOverridesEnvironment(t *testing.T) {
	cfg, displayVersion, err := parseConfig(
		[]string{"-auth-enabled=true", "-db-max-open-conns=8"},
		mapLookup(map[string]string{
			"RELOHELPER_DB_DSN":                 "postgres://test",
			"RELOHELPER_AUTH_ENABLED":           "false",
			"RELOHELPER_DB_MAX_OPEN_CONNS":      "4",
			"RELOHELPER_LIMITER_ENABLED":        "false",
			"RELOHELPER_DB_MAX_IDLE_TIME":       "1m",
			"RELOHELPER_BATCH_MAX_IDS":          "50",
			"RELOHELPER_BATCH_MAX_DETAILED_IDS": "10",
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if displayVersion {
		t.Error("displayVersion = true, want false")
	}
	if !cfg.auth.enabled {
		t.Error("CLI auth override was not applied")
	}
	if cfg.db.maxOpenConns != 8 {
		t.Errorf("max open connections = %d, want 8", cfg.db.maxOpenConns)
	}
	if cfg.limiter.enabled {
		t.Error("unmodified limiter environment value was not preserved")
	}
}

func mapLookup(values map[string]string) envLookup {
	return func(name string) (string, bool) {
		value, ok := values[name]
		return value, ok
	}
}
