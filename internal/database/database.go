/*
Package database handles database connection operations
*/
package database

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
)

// GetConnection returns an established connection to a database using the supplied connection string
func GetConnection(ctx context.Context, connectionString string) (*pgx.Conn, error) {
	log := logger.Log(ctx).
		WithField("function", "GetConnection")
	if connectionString == "" {
		return nil, logger.ErrorfAsError(log, "database connection string is required")
	}
	log.Debugf("opening database connection to %s", logger.SanitizedURLString(connectionString))
	conn, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return nil, logger.ErrorfAsError(log, "error connecting to database: %s", err)
	}
	return conn, err
}

// CloseConnection closes a database connection and logs any resulting errors
func CloseConnection(ctx context.Context, conn *pgx.Conn) {
	log := logger.Log(ctx).
		WithField("function", "CloseConnection")
	if conn != nil && !conn.IsClosed() {
		log.Debug("closing database connection")
		err := conn.Close(ctx)
		if err != nil {
			log.Errorf("error closing database connection: %s", err)
		}
	}
}
