package util

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neflyte/fdwctl/internal/logger"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	logger.SetFormat(logger.TextFormat)
	logger.SetLevel(logger.TraceLevel)
	os.Exit(m.Run())
}

func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating new sqlmock: %s", err)
	}
	return db, mock
}

func closeSQLMock(t *testing.T, db *sql.DB) {
	if db != nil {
		err := db.Close()
		if err != nil {
			t.Errorf("error closing sqlmock db: %s", err)
			t.Fail()
		}
	}
}
