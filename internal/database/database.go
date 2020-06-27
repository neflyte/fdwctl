package database

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
)

func GetConnection(ctx context.Context, connectionString string) (*pgx.Conn, error) {
	return pgx.Connect(ctx, connectionString)
}

func CloseConnection(ctx context.Context, conn *pgx.Conn) {
	log := logger.
		Root().
		WithContext(ctx).
		WithField("function", "CloseConnection")
	if conn != nil {
		err := conn.Close(ctx)
		if err != nil {
			log.Errorf("error closing database connection: %s", err)
		}
	}
}
