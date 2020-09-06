/*
Package database handles database connection operations
*/
package database

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/neflyte/fdwctl/internal/logger"
)

// GetConnection returns an established connection to a database using the supplied connection string
func GetConnection(ctx context.Context, connectionString string) (*sqlx.DB, error) {
	log := logger.Log(ctx).
		WithField("function", "GetConnection")
	if connectionString == "" {
		return nil, logger.ErrorfAsError(log, "database connection string is required")
	}
	log.Debugf("opening database connection to %s", logger.SanitizedURLString(connectionString))
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, logger.ErrorfAsError(log, "error connecting to database: %s", err)
	}
	return sqlx.NewDb(db, "postgres"), nil
}

// CloseConnection closes a database connection and logs any resulting errors
func CloseConnection(ctx context.Context, conn *sqlx.DB) {
	log := logger.Log(ctx).
		WithField("function", "CloseConnection")
	if conn != nil {
		log.Debug("closing database connection")
		err := conn.Close()
		if err != nil {
			log.Errorf("error closing database connection: %s", err)
		}
	}
}

// CloseRows closes a Rows object and logs any resulting errors
func CloseRows(ctx context.Context, rows *sqlx.Rows) {
	log := logger.Log(ctx).
		WithField("function", "CloseRows")
	if rows != nil {
		log.Debug("closing rows")
		err := rows.Close()
		if err != nil {
			log.Errorf("error closing rows: %s", err)
		}
	}
}
