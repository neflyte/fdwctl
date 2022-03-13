package util

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neflyte/fdwctl/internal/model"
	"github.com/stretchr/testify/require"
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
	createSchemaSQL := fmt.Sprintf(`CREATE SCHEMA "%s"`, schemaName)

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
	createSchemaSQL := fmt.Sprintf(`CREATE SCHEMA "%s"`, schemaName)

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

	mock.ExpectQuery(regexp.QuoteMeta(sqlGetSchemaEnums)).
		WillReturnRows(
			sqlmock.NewRows([]string{"schema", "name"}).
				AddRow("my-schema", "my-enum"),
		).
		RowsWillBeClosed()
	mock.ExpectClose()

	expected := []*model.SchemaEnum{
		&model.SchemaEnum{
			Name:   "my-enum",
			Schema: "my-schema",
		},
	}
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

	mock.ExpectQuery(regexp.QuoteMeta(sqlSchemaEnumsInTables)).
		WithArgs(schemaName).
		WillReturnRows(
			sqlmock.NewRows([]string{"cuu.udt_name", "cuu.udt_schema"}).
				AddRow("my-enum", schemaName),
		).
		RowsWillBeClosed()
	mock.ExpectClose()

	expected := []*model.SchemaEnum{
		&model.SchemaEnum{
			Name:   "my-enum",
			Schema: schemaName,
		},
	}
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

	mock.ExpectQuery(regexp.QuoteMeta(sqlEnumStrings)).
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

func TestUnit_GetSchemasForServer_Nominal_NoServerName(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	queryResults := sqlmock.NewRows([]string{"ft.foreign_table_schema", "ft.foreign_server_name", "remote_schema"}).
		AddRow("public", "my-server", "public").
		AddRow("my-schema", "other-server", "public")

	mock.ExpectQuery(
		"SELECT DISTINCT ft.foreign_table_schema, ft.foreign_server_name, ftos.option_value AS remote_schema " +
			"FROM information_schema.foreign_tables ft JOIN information_schema.foreign_table_options ftos ON " +
			"ftos.foreign_table_schema = ft.foreign_table_schema " +
			"AND ftos.foreign_table_catalog = ft.foreign_table_catalog " +
			"AND ftos.foreign_table_name = ft.foreign_table_name " +
			"AND ftos.option_name = 'schema_name'",
	).
		WillReturnRows(queryResults).
		RowsWillBeClosed()
	mock.ExpectClose()

	expected := []model.Schema{
		{ServerName: "my-server", LocalSchema: "public", RemoteSchema: "public"},
		{ServerName: "other-server", LocalSchema: "my-schema", RemoteSchema: "public"},
	}

	actual, err := GetSchemasForServer(context.Background(), db, "")
	require.NotNil(t, actual)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_GetSchemasForServer_Nominal_WithServerName(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	queryResults := sqlmock.NewRows([]string{"ft.foreign_table_schema", "ft.foreign_server_name", "remote_schema"}).
		AddRow("my-schema", "other-server", "public")

	mock.ExpectQuery(
		"SELECT DISTINCT ft.foreign_table_schema, ft.foreign_server_name, ftos.option_value AS remote_schema " +
			"FROM information_schema.foreign_tables ft JOIN information_schema.foreign_table_options ftos ON " +
			"ftos.foreign_table_schema = ft.foreign_table_schema " +
			"AND ftos.foreign_table_catalog = ft.foreign_table_catalog " +
			"AND ftos.foreign_table_name = ft.foreign_table_name " +
			"AND ftos.option_name = 'schema_name' WHERE ft.foreign_server_name = \\$1",
	).
		WithArgs("other-server").
		WillReturnRows(queryResults).
		RowsWillBeClosed()
	mock.ExpectClose()

	expected := []model.Schema{
		{ServerName: "other-server", LocalSchema: "my-schema", RemoteSchema: "public"},
	}

	actual, err := GetSchemasForServer(context.Background(), db, "other-server")
	require.NotNil(t, actual)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_DropSchema_NoCascade(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schema := model.Schema{
		ServerName:   "my-server",
		LocalSchema:  "local-schema",
		RemoteSchema: "remote-schema",
	}

	expectedSQL := fmt.Sprintf(`DROP SCHEMA "%s"`, schema.LocalSchema)
	mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectClose()

	err := DropSchema(context.Background(), db, schema, false)
	require.Nil(t, err)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_DropSchema_WithCascade(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	schema := model.Schema{
		ServerName:   "my-server",
		LocalSchema:  "local-schema",
		RemoteSchema: "remote-schema",
	}

	expectedSQL := fmt.Sprintf(`DROP SCHEMA "%s" CASCADE`, schema.LocalSchema)
	mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectClose()

	err := DropSchema(context.Background(), db, schema, true)
	require.Nil(t, err)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}
