package cmd

import (
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/util"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strings"
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
		Short: "List foreign servers",
		Run:   listServers,
	}
	listExtensionCmd = &cobra.Command{
		Use:   "extension",
		Short: "List extensions",
		Run:   listExtension,
	}
	listUsermapCmd = &cobra.Command{
		Use:   "usermap [server name]",
		Short: "List user mappings",
		Run:   listUsermap,
	}
	listSchemaCmd = &cobra.Command{
		Use:   "schema",
		Short: "List schemas that contain foreign tables",
		Run:   listSchema,
	}
	dbConnection *pgx.Conn
)

func init() {
	listCmd.AddCommand(listServerCmd)
	listCmd.AddCommand(listExtensionCmd)
	listCmd.AddCommand(listUsermapCmd)
	listCmd.AddCommand(listSchemaCmd)
}

func preDoList(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoList")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().FDWConnection)
	if err != nil {
		return logger.ErrorfAsError(log, "error getting database connection: %s", err)
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
	servers, err := util.GetServers(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error getting servers: %s", err)
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Wrapper", "Owner", "Hostname", "Port", "DB Name"})
	for _, server := range servers {
		table.Append([]string{server.Name, server.Wrapper, server.Owner, server.Host, fmt.Sprintf("%d", server.Port), server.DB})
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
		table.Append([]string{extname, extversion})
	}
	table.Render()
}

func listUsermap(cmd *cobra.Command, args []string) {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "listUsermap")
	foreignServer := ""
	if len(args) > 0 {
		foreignServer = strings.TrimSpace(args[0])
	}
	usermaps, err := util.GetUserMapsForServer(cmd.Context(), dbConnection, foreignServer)
	if err != nil {
		log.Errorf("error getting usermaps for server %s: %s", foreignServer, err)
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Local User", "Remote User", "Remote Password", "Remote Server"})
	for _, usermap := range usermaps {
		table.Append([]string{usermap.LocalUser, usermap.RemoteUser, usermap.RemotePassword, usermap.ServerName})
	}
	table.Render()
}

func listSchema(cmd *cobra.Command, _ []string) {
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "listSchema")
	schemas, err := util.GetSchemas(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error getting schemas: %s", err)
		return
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Schema Name", "Foreign Server", "Remote Schema"})
	for _, schema := range schemas {
		table.Append([]string{schema.LocalSchema, schema.ServerName, schema.RemoteSchema})
	}
	table.Render()
}
