package database

import (
	"context"
	"errors"
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

func pgNoticeHandler(_ *pgconn.PgConn, notice *pgconn.Notice) {
	pgNoticeLogger.Warnf("[%s] %s", notice.Code, notice.Message)
}

func GetConnection(ctx context.Context, connectionString string) (*pgx.Conn, error) {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "GetConnection")
	if connectionString == "" {
		return nil, errors.New("database connection string is required")
	}
	log.Debugf("opening database connection to %s", connectionString)
	conn, err := pgx.Connect(ctx, connectionString)
	if err == nil {
		conn.Config().OnNotice = pgNoticeHandler
	} else {
		log.Debugf("error connecting to database: %s", err)
	}
	return conn, err
}

func CloseConnection(ctx context.Context, conn *pgx.Conn) {
	log := logger.
		Root().
		WithContext(ctx).
		WithField("function", "CloseConnection")
	if conn != nil {
		log.Debug("closing database connection")
		err := conn.Close(ctx)
		if err != nil {
			log.Errorf("error closing database connection: %s", err)
		}
	}
}
