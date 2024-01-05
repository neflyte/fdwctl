/*
Package database handles database connection operations
*/
package database

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"

	"github.com/neflyte/fdwctl/lib/logger"
)

const (
	// driverName is the name of the database driver
	driverName = "pgx"
)

// GetConnection returns an established connection to a database using the supplied connection string
func GetConnection(ctx context.Context, connectionString string) (*sql.DB, error) {
	log := logger.Log(ctx).
		WithField("function", "GetConnection")
	if connectionString == "" {
		return nil, logger.ErrorfAsError(log, "database connection string is required")
	}
	log.Debugf("opening database connection to %s", logger.SanitizedURLString(connectionString))
	conn, err := sql.Open(driverName, connectionString)
	if err != nil {
		return nil, logger.ErrorfAsError(log, "error connecting to database: %s", err)
	}
	return conn, err
}

// CloseConnection closes a database connection and logs any resulting errors
func CloseConnection(ctx context.Context, conn *sql.DB) {
	log := logger.Log(ctx).
		WithField("function", "CloseConnection")
	if conn != nil {
		log.Trace("closing database connection")
		err := conn.Close()
		if err != nil {
			log.Errorf("error closing database connection: %s", err)
		}
	}
}

// CloseRows closes a database Rows object and logs any resulting errors
func CloseRows(ctx context.Context, rows *sql.Rows) {
	log := logger.Log(ctx).
		WithField("function", "CloseRows")
	if rows != nil {
		log.Trace("closing result rows")
		err := rows.Close()
		if err != nil {
			log.Errorf("error closing result rows: %s", err)
		}
	}
}
