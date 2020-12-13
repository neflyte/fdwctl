package util

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elgris/sqrl"

	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
)

func EnsureUser(ctx context.Context, dbConnection *sql.DB, userName string, userPassword string) error {
	log := logger.Log(ctx).
		WithField("function", "EnsureUser")
	query, args, err := sqrl.Select("1").
		From("pg_user").
		Where(sqrl.Eq{"usename": userName}).
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("error creating query: %s", err)
	}
	log.Tracef("query: %s, args: %#v", query, args)
	rows, err := dbConnection.Query(query, args...)
	if err != nil {
		return fmt.Errorf("error verifying user: %s", err)
	}
	defer database.CloseRows(ctx, rows)
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
	if rows.Err() != nil {
		log.Errorf("error iterating result rows: %s", rows.Err())
		return rows.Err()
	}
	if !userExists {
		log.Debugf("user does not exist; creating")
		query = fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", userName, userPassword)
		log.Tracef("query: %s", query)
		_, err = dbConnection.Exec(query)
		if err != nil {
			return fmt.Errorf("error creating user: %s", err)
		}
		log.Infof("user %s created", userName)
		return nil
	}
	log.Infof("user %s exists", userName)
	return nil
}

func DropUser(ctx context.Context, dbConnection *sql.DB, username string) error {
	log := logger.Log(ctx).
		WithField("function", "DropUser")
	if username == "" {
		return logger.ErrorfAsError(log, "user name is required")
	}
	query := fmt.Sprintf("DROP USER IF EXISTS %s", username)
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(query)
	if err != nil {
		log.Errorf("error dropping user: %s", err)
		return err
	}
	return nil
}
