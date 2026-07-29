package data

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	testDBDSN := os.Getenv("RELOHELPER_TEST_DB_DSN")
	if testDBDSN == "" {
		t.Fatal("RELOHELPER_TEST_DB_DSN is required")
	}

	db, err := pgxpool.New(context.Background(), testDBDSN)
	if err != nil {
		t.Fatal(err)
	}

	return db
}
