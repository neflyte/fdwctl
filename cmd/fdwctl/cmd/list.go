package cmd

import (
	"errors"
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
)

var (
	listCmd = &cobra.Command{
		Use:               "list <object type>",
		Short:             "List objects",
		PersistentPreRunE: preDoList,
		PersistentPostRun: postDoList,
	}
	listServerCmd = &cobra.Command{
		Use:   "server",
		Short: "List foreign server objects",
		Run:   listServers,
	}
	listExtensionCmd = &cobra.Command{
		Use:   "extension",
		Short: "List extensions",
		Run:   listExtension,
	}
	dbConnection *pgx.Conn
)

type ServerObject struct {
	Name    string
	Wrapper string
	Owner   string
}

func init() {
	listCmd.AddCommand(listServerCmd)
	listCmd.AddCommand(listExtensionCmd)
}

func preDoList(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoList")
	if connectionString == "" {
		return errors.New("fdw database connection string is required")
	}
	dbConnection, err = database.GetConnection(cmd.Context(), connectionString)
	if err != nil {
		log.Errorf("error getting database connection: %s", err)
		return err
	}
	return nil
}

func postDoList(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func listServers(cmd *cobra.Command, _ []string) {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "listServers")
	query, _, err := sqrl.
		Select("foreign_server_name, foreign_data_wrapper_name, authorization_identifier").
		From("information_schema.foreign_servers").
		ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return
	}
	log.Tracef("query: %s", query)
	rows, err := dbConnection.Query(cmd.Context(), query)
	if err != nil {
		log.Errorf("error querying for servers: %s", err)
		return
	}
	defer rows.Close()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Wrapper", "Owner", "Hostname", "Port", "DB Name"})
	servers := make([]ServerObject, 0)
	var serverName, wrapperName, owner string
	for rows.Next() {
		err = rows.Scan(&serverName, &wrapperName, &owner)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		servers = append(servers, ServerObject{
			Name:    serverName,
			Wrapper: wrapperName,
			Owner:   owner,
		})
	}
	rows.Close()
	for _, server := range servers {
		optionsQuery := fmt.Sprintf("SELECT option_name, option_value FROM information_schema.foreign_server_options WHERE foreign_server_name = '%s'", server.Name)
		log.Tracef("query: %s", optionsQuery)
		optionsRows, err := dbConnection.Query(cmd.Context(), optionsQuery)
		if err != nil {
			log.Errorf("error querying server options: %s", err)
			continue
		}
		var optionName, optionValue, hostName, port, dbName string
		for optionsRows.Next() {
			err = optionsRows.Scan(&optionName, &optionValue)
			if err != nil {
				log.Errorf("error scanning result row: %s", err)
				continue
			}
			switch optionName {
			case "host":
				hostName = optionValue
			case "port":
				port = optionValue
			case "dbname":
				dbName = optionValue
			default:
				log.Tracef("unknown option: %s", optionName)
			}
		}
		table.Append([]string{server.Name, server.Wrapper, server.Owner, hostName, port, dbName})
		optionsRows.Close()
	}
	table.Render()
}

func listExtension(cmd *cobra.Command, _ []string) {
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "listExtension")
	query, _, err := sqrl.
		Select("extname", "extversion").
		From("pg_extension").
		ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return
	}
	rows, err := dbConnection.Query(cmd.Context(), query)
	if err != nil {
		log.Errorf("error querying for extensions: %s", err)
		return
	}
	defer rows.Close()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Version"})
	var extname, extversion string
	for rows.Next() {
		err = rows.Scan(&extname, &extversion)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			return
		}
		table.Append([]string{
			extname,
			extversion,
		})
	}
	table.Render()
}
