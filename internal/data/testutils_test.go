package data

import (
	"database/sql"
	"flag"
	"os"
	"testing"
)

var testDBdsn = flag.String("db-dsn", os.Getenv("RELOHELPER_TEST_DB_DSN"), "PostgreSQL DSN for testing")

// Stub flag to allow passing cmd flag during testing.
var _ = flag.String("env", "", "Environment flag for testing")

func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", *testDBdsn)
	if err != nil {
		t.Fatal(err)
	}

	return db
}
