package util

import (
	"context"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
)

func ExtensionIsCreated(ctx context.Context, dbconn *pgx.Conn) bool {
	log := logger.
		Root().
		WithContext(ctx).
		WithField("function", "extensionIsCreated")
	if dbconn == nil {
		log.Error("nil db connection")
		return false
	}
	query, args, _ := sqrl.
		Select("1").
		From("pg_extension").
		Where(sqrl.Eq{"extname": "postgres_fdw"}).
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	log.Tracef("query: %s, args: %#v", query, args)
	_, err := dbconn.Exec(ctx, query, args...)
	if err != nil {
		log.Errorf("error querying for extension: %s", err)
		return false
	}
	return true
}
