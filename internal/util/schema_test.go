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

func TestUnit_getEnums_Nominal(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	mock.ExpectQuery("SELECT typname FROM pg_type WHERE typtype = \\$1").
		WithArgs("e").
		WillReturnRows(
			sqlmock.NewRows([]string{"typname"}).
				AddRow("my-enum"),
		).
		RowsWillBeClosed()
	mock.ExpectClose()

	expected := []string{"my-enum"}
	actual, err := getEnums(context.Background(), db)

	require.Nil(t, err)
	require.NotNil(t, actual)
	require.Equal(t, expected, actual)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_getSchemaEnumsUsedInTables_Nominal(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schemaName := "my-schema"

	mock.ExpectQuery(
		"SELECT DISTINCT cuu.udt_name FROM information_schema.column_udt_usage cuu "+
			"JOIN pg_type t ON t.typname = cuu.udt_name WHERE \\(t.typtype = \\$1 AND cuu.table_schema = \\$2\\)",
	).
		WithArgs("e", schemaName).
		WillReturnRows(
			sqlmock.NewRows([]string{"cuu.udt_name"}).
				AddRow("my-enum"),
		).
		RowsWillBeClosed()
	mock.ExpectClose()

	expected := []string{"my-enum"}
	actual, err := getSchemaEnumsUsedInTables(context.Background(), db, schemaName)
	require.Nil(t, err)
	require.NotNil(t, actual)
	require.Equal(t, expected, actual)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_getEnumStrings_Nominal(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	enumType := "e"

	mock.ExpectQuery(
		"SELECT e.enumlabel FROM pg_type t JOIN pg_enum e ON t.oid = e.enumtypid " +
			"WHERE t.typname = \\$1 ORDER BY e.enumsortorder ASC",
	).
		WithArgs(enumType).
		WillReturnRows(
			sqlmock.NewRows([]string{"e.enumlabel"}).
				AddRow("valueOne").
				AddRow("valueTwo"),
		).
		RowsWillBeClosed()
	mock.ExpectClose()

	expected := []string{"valueOne", "valueTwo"}
	actual, err := getEnumStrings(context.Background(), db, enumType)
	require.Nil(t, err)
	require.NotNil(t, actual)
	require.Equal(t, expected, actual)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}
