package data

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

var testDBdsn = flag.String("db-dsn", os.Getenv("RELOHELPER_TEST_DB_DSN"), "PostgreSQL DSN for testing")

// Stub flag to allow passing cmd flag during testing.
var _ = flag.String("env", "", "Environment flag for testing")

func newTestDB(t *testing.T) *pgxpool.Pool {
	db, err := pgxpool.New(context.Background(), *testDBdsn)
	if err != nil {
		t.Fatal(err)
	}

	return db
}
