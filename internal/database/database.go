package database

import (
	"context"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/sirupsen/logrus"
)

var (
	pgNoticeLogger logrus.FieldLogger
)

func init() {
	pgNoticeLogger = logger.Root().WithField("pgx", "NOTICE")
}

func PGNoticeHandler(_ *pgconn.PgConn, notice *pgconn.Notice) {
	pgNoticeLogger.Infof("[%s] %s", notice.Code, notice.Message)
}

func GetConnection(ctx context.Context, connectionString string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, connectionString)
	if err == nil {
		conn.Config().OnNotice = PGNoticeHandler
	}
	return conn, err
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
