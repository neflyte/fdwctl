package util

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
)

const (
	sqlGetUsermaps = `WITH remoteuser AS (
	SELECT authorization_identifier, foreign_server_name, option_value AS remoteuser FROM information_schema.user_mapping_options WHERE option_name = 'user'
), remotepassword AS (
	SELECT authorization_identifier, foreign_server_name, option_value AS remotepassword FROM information_schema.user_mapping_options WHERE option_name = 'password'
)
SELECT ru.authorization_identifier, ru.remoteuser, rp.remotepassword, ru.foreign_server_name
FROM remoteuser ru
JOIN remotepassword rp ON ru.authorization_identifier = rp.authorization_identifier AND ru.foreign_server_name = rp.foreign_server_name`

	sqlDropUsermap   = `DROP USER MAPPING IF EXISTS FOR "%s" SERVER "%s"`
	sqlCreateUsermap = `CREATE USER MAPPING FOR "%s" SERVER "%s" OPTIONS (user '%s', password '%s')`
	sqlUpdateUsermap = `ALTER USER MAPPING FOR "%s" SERVER "%s" OPTIONS (%s)`
)

func FindUserMap(usermaps []model.UserMap, localuser string) *model.UserMap {
	for _, usermap := range usermaps {
		if usermap.LocalUser == localuser {
			return &usermap
		}
	}
	return nil
}

func GetUserMapsForServer(ctx context.Context, dbConnection *sql.DB, foreignServer string) ([]model.UserMap, error) {
	log := logger.Log(ctx).
		WithField("function", "GetUserMapsForServer")
	query := sqlGetUsermaps
	qArgs := make([]interface{}, 1)
	if foreignServer != "" {
		query += " WHERE ru.foreign_server_name = $1"
		qArgs[0] = foreignServer
	}
	log.Tracef("query: %s, args: %#v", query, qArgs)
	userRows, err := dbConnection.Query(query, qArgs...)
	if err != nil {
		log.Errorf("error getting users for server: %s", err)
		return nil, err
	}
	defer database.CloseRows(ctx, userRows)
	users := make([]model.UserMap, 0)
	for userRows.Next() {
		user := new(model.UserMap)
		user.RemoteSecret = model.Secret{}
		err = userRows.Scan(&user.LocalUser, &user.RemoteUser, &user.RemoteSecret.Value, &user.ServerName)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		users = append(users, *user)
	}
	if userRows.Err() != nil {
		log.Errorf("error iterating result rows: %s", userRows.Err())
		return nil, userRows.Err()
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

func DropUserMap(ctx context.Context, dbConnection *sql.DB, usermap model.UserMap, dropLocalUser bool) error {
	log := logger.Log(ctx).
		WithField("function", "DropUserMap")
	if usermap.ServerName == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	query := fmt.Sprintf(sqlDropUsermap, usermap.LocalUser, usermap.ServerName)
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(query)
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

func CreateUserMap(ctx context.Context, dbConnection *sql.DB, usermap model.UserMap) error {
	var secretValue string
	var err error

	log := logger.Log(ctx).
		WithField("function", "CreateUserMap")
	if usermap.ServerName == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	// Check if the secret is defined before resolving it
	if usermap.RemoteSecret.IsDefined() {
		secretValue, err = GetSecret(ctx, usermap.RemoteSecret)
		if err != nil {
			return logger.ErrorfAsError(log, "error getting secret value: %s", err)
		}
	} else {
		secretValue = ""
	}
	// FIXME: There could be no password at all; check for a password before using it in the SQL statement
	query := fmt.Sprintf(sqlCreateUsermap, usermap.LocalUser, usermap.ServerName, usermap.RemoteUser, secretValue)
	log.Tracef("query: %s", query)
	_, err = dbConnection.Exec(query)
	if err != nil {
		log.Errorf("error creating user mapping: %s", err)
		return err
	}
	return nil
}

func UpdateUserMap(ctx context.Context, dbConnection *sql.DB, usermap model.UserMap) error {
	log := logger.Log(ctx).
		WithField("function", "UpdateUserMap")
	if usermap.ServerName == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	optArgs := make([]string, 0)
	if usermap.RemoteUser != "" {
		optArgs = append(optArgs, fmt.Sprintf("SET user '%s'", usermap.RemoteUser))
	}
	if usermap.RemoteSecret.IsDefined() {
		secretValue, err := GetSecret(ctx, usermap.RemoteSecret)
		if err != nil {
			return logger.ErrorfAsError(log, "error getting secret value: %s", err)
		}
		optArgs = append(optArgs, fmt.Sprintf("SET password '%s'", secretValue))
	}
	query := fmt.Sprintf(sqlUpdateUsermap, usermap.LocalUser, usermap.ServerName, strings.Join(optArgs, ", "))
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(query)
	if err != nil {
		log.Errorf("error editing user mapping: %s", err)
		return err
	}
	return nil
}
