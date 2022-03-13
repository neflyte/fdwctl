package util

import (
	"database/sql"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/stretchr/testify/require"
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

func TestUnit_StringCoalesce_Nominal(t *testing.T) {
	expected := " foo "
	actual := StringCoalesce("", " ", " foo ", "bar")
	require.Equal(t, expected, actual)
}

func TestUnit_StringCoalesce_SpacesAndEmptyStrings(t *testing.T) {
	expected := ""
	actual := StringCoalesce(" ", "", "   ", "", "", " ")
	require.Equal(t, expected, actual)
}

func TestUnit_StartsWithNumber(t *testing.T) {
	require.True(t, StartsWithNumber("1NightInRio"))
	require.False(t, StartsWithNumber("TwoDaysInLA"))
}
