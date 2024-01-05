package util

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/neflyte/fdwctl/lib/database"
	"github.com/neflyte/fdwctl/lib/logger"
	"github.com/neflyte/fdwctl/lib/model"
)

const (
	sqlCreateServer      = `CREATE SERVER "%s" FOREIGN DATA WRAPPER postgres_fdw OPTIONS (host '%s', port '%d', dbname '%s')`
	sqlDropServer        = `DROP SERVER "%s"`
	sqlUpdateServer      = `ALTER SERVER "%s" OPTIONS (%s)`
	sqlRenameServer      = `ALTER SERVER "%s" RENAME TO "%s"`
	sqlForeignServerInfo = `SELECT fs.foreign_server_name, fs.foreign_data_wrapper_name, fs.authorization_identifier, fsoh.option_value AS hostname, fsop.option_value::int AS port, fsod.option_value AS dbname
	FROM information_schema.foreign_servers fs
	JOIN information_schema.foreign_server_options fsoh ON fsoh.foreign_server_name = fs.foreign_server_name AND fsoh.option_name = 'host'
	JOIN information_schema.foreign_server_options fsop ON fsop.foreign_server_name = fs.foreign_server_name AND fsop.option_name = 'port'
	JOIN information_schema.foreign_server_options fsod ON fsod.foreign_server_name = fs.foreign_server_name AND fsod.option_name = 'dbname'`
)

func GetServers(ctx context.Context, dbConnection *sql.DB) ([]model.ForeignServer, error) {
	log := logger.Log(ctx).
		WithField("function", "GetServers")
	log.Tracef("query: %s", sqlForeignServerInfo)
	rows, err := dbConnection.Query(sqlForeignServerInfo)
	if err != nil {
		log.Errorf("error querying for servers: %s", err)
		return nil, err
	}
	defer database.CloseRows(ctx, rows)
	servers := make([]model.ForeignServer, 0)
	for rows.Next() {
		server := new(model.ForeignServer)
		err = rows.Scan(&server.Name, &server.Wrapper, &server.Owner, &server.Host, &server.Port, &server.DB)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		servers = append(servers, *server)
	}
	if rows.Err() != nil {
		log.Errorf("error iterating result rows: %s", rows.Err())
		return nil, rows.Err()
	}
	return servers, nil
}

func FindForeignServer(foreignServers []model.ForeignServer, serverName string) *model.ForeignServer {
	for _, server := range foreignServers {
		if server.Name == serverName {
			return &server
		}
	}
	return nil
}

func DropServer(ctx context.Context, dbConnection *sql.DB, servername string, cascade bool) error {
	log := logger.Log(ctx).
		WithField("function", "DropServer")
	if servername == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	query := fmt.Sprintf(sqlDropServer, servername)
	if cascade {
		query = fmt.Sprintf("%s CASCADE", query)
	}
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(query)
	if err != nil {
		log.Errorf("error dropping server: %s", err)
		return err
	}
	return nil
}

func CreateServer(ctx context.Context, dbConnection *sql.DB, server model.ForeignServer) error {
	log := logger.Log(ctx).
		WithField("function", "CreateServer")
	query := fmt.Sprintf(sqlCreateServer, server.Name, server.Host, server.Port, server.DB)
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(query)
	if err != nil {
		log.Errorf("error creating server: %s", err)
		return err
	}
	return nil
}

func UpdateServer(ctx context.Context, dbConnection *sql.DB, server model.ForeignServer) error {
	log := logger.Log(ctx).
		WithField("function", "UpdateServer")
	// Edit server hostname, port, and dbname
	opts := make([]string, 0)
	if server.Host != "" {
		opts = append(opts, fmt.Sprintf("SET host '%s'", server.Host))
	}
	if server.Port > 0 {
		opts = append(opts, fmt.Sprintf("SET port '%d'", server.Port))
	}
	if server.DB != "" {
		opts = append(opts, fmt.Sprintf("SET dbname '%s'", server.DB))
	}
	query := fmt.Sprintf(sqlUpdateServer, server.Name, strings.Join(opts, ","))
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(query)
	if err != nil {
		log.Errorf("error updating server: %s", err)
		return err
	}

	return nil
}

func UpdateServerName(ctx context.Context, dbConnection *sql.DB, server model.ForeignServer, newServerName string) error {
	log := logger.Log(ctx).
		WithField("function", "UpdateServerName")
	query := fmt.Sprintf(sqlRenameServer, server.Name, newServerName)
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(query)
	if err != nil {
		log.Errorf("error renaming server object: %s", err)
		return err
	}
	return nil
}

// DiffForeignServers produces a list of `ForeignServers` to remove, add, and modify to bring `dbServers` in line with `dStateServers`
func DiffForeignServers(dStateServers []model.ForeignServer, dbServers []model.ForeignServer) (fsRemove []model.ForeignServer, fsAdd []model.ForeignServer, fsModify []model.ForeignServer) {
	// Init return variables
	fsRemove = make([]model.ForeignServer, 0)
	fsAdd = make([]model.ForeignServer, 0)
	fsModify = make([]model.ForeignServer, 0)
	// fsRemove
	for _, dbServer := range dbServers {
		if FindForeignServer(dStateServers, dbServer.Name) == nil {
			fsRemove = append(fsRemove, dbServer)
		}
	}
	// fsAdd + fsModify
	for _, dStateServer := range dStateServers {
		if FindForeignServer(dbServers, dStateServer.Name) == nil {
			fsAdd = append(fsAdd, dStateServer)
		} else {
			fsModify = append(fsModify, dStateServer)
		}
	}
	return
}
