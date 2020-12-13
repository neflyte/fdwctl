package util

import (
	"context"
	"errors"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/neflyte/fdwctl/internal/model"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestUnit_GetExtensions_Nominal(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	extensionRows := sqlmock.
		NewRows([]string{"extname", "extversion"}).
		AddRow("postgres_fdw", "1.0.0")

	expected := []model.Extension{
		{Name: "postgres_fdw", Version: "1.0.0"},
	}

	mock.
		ExpectQuery("SELECT extname, extversion FROM pg_extension").
		WillReturnRows(extensionRows).
		RowsWillBeClosed()
	mock.ExpectClose()

	actual, err := GetExtensions(context.Background(), db)
	require.Nil(t, err)
	require.Greater(t, len(actual), 0)
	require.Equal(t, expected, actual)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_GetExtensions_QueryError(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	mock.
		ExpectQuery("SELECT extname, extversion FROM pg_extension").
		WillReturnError(errors.New("QUERY ERROR"))
	mock.ExpectClose()

	actual, err := GetExtensions(context.Background(), db)
	require.NotNil(t, err)
	require.Nil(t, actual)
	require.Equal(t, "QUERY ERROR", err.Error())
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_GetExtensions_RowError(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	extensionRows := sqlmock.
		NewRows([]string{"extname", "extversion"}).
		AddRow("postgres_fdw", "1.0.0").
		RowError(0, errors.New("ROW ERROR"))

	mock.
		ExpectQuery("SELECT extname, extversion FROM pg_extension").
		WillReturnRows(extensionRows).
		RowsWillBeClosed()
	mock.ExpectClose()

	actual, err := GetExtensions(context.Background(), db)
	require.NotNil(t, err)
	require.Nil(t, actual)
	require.Equal(t, "ROW ERROR", err.Error())
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_CreateExtension_Nominal(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	extension := model.Extension{
		Name:    "postgres_fdw",
		Version: "1.0.0",
	}

	mock.ExpectExec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", extension.Name)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectClose()

	err := CreateExtension(context.Background(), db, extension)
	require.Nil(t, err)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_CreateExtension_ExecError(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	extension := model.Extension{
		Name:    "postgres_fdw",
		Version: "1.0.0",
	}

	mock.ExpectExec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", extension.Name)).
		WillReturnError(errors.New("QUERY ERROR"))
	mock.ExpectClose()

	err := CreateExtension(context.Background(), db, extension)
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "QUERY ERROR"))
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_DropExtension_Nominal(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	extension := model.Extension{
		Name:    "postgres_fdw",
		Version: "1.0.0",
	}

	mock.ExpectExec(fmt.Sprintf("DROP EXTENSION IF EXISTS %s", extension.Name)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectClose()

	err := DropExtension(context.Background(), db, extension)
	require.Nil(t, err)
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}

func TestUnit_DropExtension_ExecError(t *testing.T) {
	db, mock := newSQLMock(t)
	defer closeSQLMock(t, db)

	extension := model.Extension{
		Name:    "postgres_fdw",
		Version: "1.0.0",
	}

	mock.ExpectExec(fmt.Sprintf("DROP EXTENSION IF EXISTS %s", extension.Name)).
		WillReturnError(errors.New("QUERY ERROR"))
	mock.ExpectClose()

	err := DropExtension(context.Background(), db, extension)
	require.NotNil(t, err)
	require.True(t, strings.Contains(err.Error(), "QUERY ERROR"))
	closeSQLMock(t, db)
	require.Nil(t, mock.ExpectationsWereMet())
}
