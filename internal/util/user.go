package util

import (
	"context"
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
)

type ForeignServerUserMapping struct {
	ServerName     string
	LocalUser      string
	RemoteUser     string
	RemotePassword string
}

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
	log.Tracef("query: %s, args: %#v", query, args)
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

func GetUsersForServer(ctx context.Context, dbConnection *pgx.Conn, foreignServer string) ([]ForeignServerUserMapping, error) {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "GetUsersForServer")
	query, qArgs, err := sqrl.
		Select("u.authorization_identifier", "ou.option_value", "op.option_value", "s.srvname").
		From("information_schema.user_mappings u").
		Join("information_schema.user_mapping_options ou ON ou.authorization_identifier = u.authorization_identifier AND ou.option_name = 'user'").
		Join("information_schema.user_mapping_options op ON op.authorization_identifier = u.authorization_identifier AND op.option_name = 'password'").
		Join("pg_user_mappings s ON s.usename = u.authorization_identifier").
		Where(sqrl.Eq{"s.srvname": foreignServer}).
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return nil, err
	}
	log.Tracef("query: %s, args: %#v", query, qArgs)
	userRows, err := dbConnection.Query(ctx, query, qArgs...)
	if err != nil {
		log.Errorf("error getting users for server: %s", err)
		return nil, err
	}
	defer userRows.Close()
	users := make([]ForeignServerUserMapping, 0)
	for userRows.Next() {
		user := new(ForeignServerUserMapping)
		err = userRows.Scan(&user.LocalUser, &user.RemoteUser, &user.RemotePassword, &user.ServerName)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		users = append(users, *user)
	}
	return users, nil
}
