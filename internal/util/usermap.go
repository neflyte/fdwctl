package util

import (
	"context"
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"strings"
)

func FindUserMap(usermaps []model.UserMap, localuser string) *model.UserMap {
	for _, usermap := range usermaps {
		if usermap.LocalUser == localuser {
			return &usermap
		}
	}
	return nil
}

func GetUserMapsForServer(ctx context.Context, dbConnection *pgx.Conn, foreignServer string) ([]model.UserMap, error) {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "GetUserMapsForServer")
	qbuilder := sqrl.
		Select("u.authorization_identifier", "ou.option_value", "op.option_value", "s.srvname").
		From("information_schema.user_mappings u").
		Join("information_schema.user_mapping_options ou ON ou.authorization_identifier = u.authorization_identifier AND ou.option_name = 'user'").
		Join("information_schema.user_mapping_options op ON op.authorization_identifier = u.authorization_identifier AND op.option_name = 'password'").
		Join("pg_user_mappings s ON s.usename = u.authorization_identifier")
	if foreignServer != "" {
		qbuilder = qbuilder.Where(sqrl.Eq{"s.srvname": foreignServer})
	}
	query, qArgs, err := qbuilder.
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
	users := make([]model.UserMap, 0)
	for userRows.Next() {
		user := new(model.UserMap)
		err = userRows.Scan(&user.LocalUser, &user.RemoteUser, &user.RemoteSecret.Value, &user.ServerName)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		users = append(users, *user)
	}
	return users, nil
}

func DiffUserMaps(dStateUserMaps []model.UserMap, dbUserMaps []model.UserMap) (umRemove []model.UserMap, umAdd []model.UserMap, umModify []model.UserMap) {
	// Init return variables
	umRemove = make([]model.UserMap, 0)
	umAdd = make([]model.UserMap, 0)
	umModify = make([]model.UserMap, 0)
	// umRemove
	for _, dbUserMap := range dbUserMaps {
		if FindUserMap(dStateUserMaps, dbUserMap.LocalUser) == nil {
			umRemove = append(umRemove, dbUserMap)
		}
	}
	// umAdd + umModify
	for _, dStateUserMap := range dStateUserMaps {
		if FindUserMap(dbUserMaps, dStateUserMap.LocalUser) == nil {
			umAdd = append(umAdd, dStateUserMap)
		} else {
			umModify = append(umModify, dStateUserMap)
		}
	}
	return
}

func DropUserMap(ctx context.Context, dbConnection *pgx.Conn, usermap model.UserMap, dropLocalUser bool) error {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "DropUserMap")
	if usermap.ServerName == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	query := fmt.Sprintf("DROP USER MAPPING IF EXISTS FOR %s SERVER %s", usermap.LocalUser, usermap.ServerName)
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(ctx, query)
	if err != nil {
		log.Errorf("error dropping user mapping: %s", err)
		return err
	}
	if dropLocalUser {
		err = DropUser(ctx, dbConnection, usermap.LocalUser)
		if err != nil {
			log.Errorf("error dropping local user %s: %s", usermap.LocalUser, err)
			return err
		}
		log.Infof("user %s dropped", usermap.LocalUser)
	}
	return nil
}

func CreateUserMap(ctx context.Context, dbConnection *pgx.Conn, usermap model.UserMap) error {
	var secretValue string
	var err error
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "CreateUserMap")
	if usermap.ServerName == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	// Check if the secret is defined before resolving it
	if SecretIsDefined(usermap.RemoteSecret) {
		secretValue, err = GetSecret(ctx, usermap.RemoteSecret)
		if err != nil {
			return logger.ErrorfAsError(log, "error getting secret value: %s", err)
		}
	} else {
		secretValue = ""
	}
	// FIXME: There could be no password at all; check for a password before using it in the SQL statement
	query := fmt.Sprintf("CREATE USER MAPPING FOR %s SERVER %s OPTIONS (user '%s', password '%s')", usermap.LocalUser, usermap.ServerName, usermap.RemoteUser, secretValue)
	log.Tracef("query: %s", query)
	_, err = dbConnection.Exec(ctx, query)
	if err != nil {
		log.Errorf("error creating user mapping: %s", err)
		return err
	}
	return nil
}

func UpdateUserMap(ctx context.Context, dbConnection *pgx.Conn, usermap model.UserMap) error {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "UpdateUserMap")
	if usermap.ServerName == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	optArgs := make([]string, 0)
	query := fmt.Sprintf("ALTER USER MAPPING FOR %s SERVER %s OPTIONS (", usermap.LocalUser, usermap.ServerName)
	if usermap.RemoteUser != "" {
		optArgs = append(optArgs, fmt.Sprintf("SET user '%s'", usermap.RemoteUser))
	}
	if SecretIsDefined(usermap.RemoteSecret) {
		secretValue, err := GetSecret(ctx, usermap.RemoteSecret)
		if err != nil {
			return logger.ErrorfAsError(log, "error getting secret value: %s", err)
		}
		optArgs = append(optArgs, fmt.Sprintf("SET password '%s'", secretValue))
	}
	query = fmt.Sprintf("%s %s )", query, strings.Join(optArgs, ", "))
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(ctx, query)
	if err != nil {
		log.Errorf("error editing user mapping: %s", err)
		return err
	}
	return nil
}
