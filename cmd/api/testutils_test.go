package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/denis-k2/relohelper-go/internal/data"
)

var (
	logger  *slog.Logger
	testApp *application
	testDB  *sql.DB
)

// configureTestLogger configures a logger for testing.
// Logs are printed to os.Stdout at Debug level when env flag is "testLogs",
// otherwise output is discarded.
func configureTestLogger(env string) {
	if env == "testLogs" {
		// Verbose logger for debugging: output to os.Stdout.
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	} else {
		// Silent logger to keep test output clean
		logger = slog.New(slog.DiscardHandler)
	}
}

func TestMain(m *testing.M) {
	testCfg, err := parseFlags()
	if err != nil {
		logger.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	configureTestLogger(testCfg.env)

	testApp, testDB, err = newTestApplication(testCfg)
	if err != nil {
		logger.Error("failed to initialize application", "error", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testDB.Close(); err != nil {
		logger.Error("failed to close DB", "error", err)
	}

	os.Exit(code)
}

func newTestApplication(cfg config) (*application, *sql.DB, error) {
	db, err := openDB(cfg)
	if err != nil {
		logger.Error("database connection error", "error", err)
		return nil, nil, err
	}
	logger.Info("database connection pool established")

	return &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}, db, nil
}

type testServer struct {
	*httptest.Server
}

func newTestServer(h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)
	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, []byte) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, body
}

func unmarshalJSON(t *testing.T, body []byte, gotPtr any) {
	err := json.Unmarshal(body, gotPtr)
	if err != nil {
		t.Fatalf("Unable to parse %q: %v", body, err)
	}
}

type QueryParams struct {
	costEnabled    bool
	indicesEnabled bool
	climateEnabled bool
}

func cityFildsToBool(c data.City) QueryParams {
	return QueryParams{
		costEnabled:    c.NumbeoCost != nil,
		indicesEnabled: c.NumbeoIndices != nil,
		climateEnabled: c.AvgClimate != nil,
	}
}

type QueryParamsCountry struct {
	numbeoIndicesEnabled  bool
	legatumIndicesEnabled bool
}

func countryFildsToBool(c data.Country) QueryParamsCountry {
	return QueryParamsCountry{
		numbeoIndicesEnabled:  c.NumbeoIndices != nil,
		legatumIndicesEnabled: c.LegatumIndices != nil,
	}
}
