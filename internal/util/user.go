package util

import (
	"context"
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
)

func EnsureUser(ctx context.Context, dbConnection *pgx.Conn, userName string, userPassword string) error {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "EnsureUser")
	query, args, err := sqrl.Select("1").
		From("pg_user").
		Where(sqrl.Eq{"usename": userName}).
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("error creating query: %s", err)
	}
	rows, err := dbConnection.Query(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error verifying user: %s", err)
	}
	defer rows.Close()
	userExists := false
	if rows.Next() {
		var foo int
		err = rows.Scan(&foo)
		if err != nil {
			return fmt.Errorf("error scanning result row: %s", err)
		}
		if foo == 1 {
			userExists = true
		}
	}
	if !userExists {
		log.Debugf("user does not exist; creating")
		query = fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", userName, userPassword)
		log.Tracef("query: %s", query)
		_, err = dbConnection.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("error creating user: %s", err)
		}
		log.Infof("user %s created", userName)
		return nil
	}
	log.Infof("user %s exists", userName)
	return nil
}
