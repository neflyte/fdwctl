package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestUnit_ensureSchema_SchemaExists(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schemaName := "my-schema"

	mock.ExpectQuery("SELECT 1 FROM information_schema.schemata WHERE schema_name = \\$1").
		WithArgs(schemaName).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1)).
		RowsWillBeClosed()
	mock.ExpectClose()

	err := ensureSchema(context.Background(), db, schemaName)
	require.Nil(t, err)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_ensureSchema_SchemaDoesNotExist(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schemaName := "my-schema"
	createSchemaSQL := fmt.Sprintf("CREATE SCHEMA %s", schemaName)

	mock.ExpectQuery("SELECT 1 FROM information_schema.schemata WHERE schema_name = \\$1").
		WithArgs(schemaName).
		WillReturnRows(sqlmock.NewRows([]string{"1"})).
		RowsWillBeClosed()
	mock.ExpectExec(createSchemaSQL).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectClose()

	err := ensureSchema(context.Background(), db, schemaName)
	require.Nil(t, err)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_ensureSchema_ExistenceCheckFailed(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schemaName := "my-schema"

	mock.ExpectQuery("SELECT 1 FROM information_schema.schemata WHERE schema_name = \\$1").
		WithArgs(schemaName).
		WillReturnError(errors.New("QUERY ERROR"))
	mock.ExpectClose()

	err := ensureSchema(context.Background(), db, schemaName)
	require.NotNil(t, err)
	require.Equal(t, "QUERY ERROR", err.Error())
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_ensureSchema_ExistenceCheckRowError(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schemaName := "my-schema"

	mock.ExpectQuery("SELECT 1 FROM information_schema.schemata WHERE schema_name = \\$1").
		WithArgs(schemaName).
		WillReturnRows(
			sqlmock.NewRows([]string{"1"}).
				AddRow(1).
				RowError(0, errors.New("ROW ERROR")),
		).
		RowsWillBeClosed()
	mock.ExpectClose()

	err := ensureSchema(context.Background(), db, schemaName)
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "ROW ERROR"))
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_ensureSchema_ExistenceCheckScanError(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schemaName := "my-schema"

	mock.ExpectQuery("SELECT 1 FROM information_schema.schemata WHERE schema_name = \\$1").
		WithArgs(schemaName).
		WillReturnRows(
			sqlmock.NewRows([]string{"1"}).
				AddRow("one"),
		).
		RowsWillBeClosed()
	mock.ExpectClose()

	err := ensureSchema(context.Background(), db, schemaName)
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "Scan error"))
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_ensureSchema_CreateSchemaError(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schemaName := "my-schema"
	createSchemaSQL := fmt.Sprintf("CREATE SCHEMA %s", schemaName)

	mock.ExpectQuery("SELECT 1 FROM information_schema.schemata WHERE schema_name = \\$1").
		WithArgs(schemaName).
		WillReturnRows(sqlmock.NewRows([]string{"1"})).
		RowsWillBeClosed()
	mock.ExpectExec(createSchemaSQL).
		WillReturnError(errors.New("QUERY ERROR"))
	mock.ExpectClose()

	err := ensureSchema(context.Background(), db, schemaName)
	require.NotNil(t, err)
	require.Equal(t, "QUERY ERROR", err.Error())
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}
