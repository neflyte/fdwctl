package cmd

import (
	"fmt"
	"github.com/jmoiron/sqlx"
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
		RunE:  listServers,
	}
	listExtensionCmd = &cobra.Command{
		Use:   "extension",
		Short: "List extensions",
		RunE:  listExtension,
	}
	listUsermapCmd = &cobra.Command{
		Use:   "usermap [server name]",
		Short: "List user mappings",
		RunE:  listUsermap,
	}
	listSchemaCmd = &cobra.Command{
		Use:   "schema",
		Short: "List schemas that contain foreign tables",
		RunE:  listSchema,
	}
	dbConnection *sqlx.DB
)

func init() {
	listCmd.AddCommand(listServerCmd)
	listCmd.AddCommand(listExtensionCmd)
	listCmd.AddCommand(listUsermapCmd)
	listCmd.AddCommand(listSchemaCmd)
}

func preDoList(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.Log(cmd.Context()).
		WithField("function", "preDoList")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().GetDatabaseConnectionString())
	if err != nil {
		return logger.ErrorfAsError(log, "error getting database connection: %s", err)
	}
	return nil
}

func postDoList(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func listServers(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.Log(cmd.Context()).
		WithField("function", "listServers")
	servers, err := util.GetServers(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error getting servers: %s", err)
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Wrapper", "Owner", "Hostname", "Port", "DB Name"})
	for _, server := range servers {
		table.Append([]string{server.Name, server.Wrapper, server.Owner, server.Host, fmt.Sprintf("%d", server.Port), server.DB})
	}
	table.Render()
	return nil
}

func listExtension(cmd *cobra.Command, _ []string) error {
	log := logger.Log(cmd.Context()).
		WithField("function", "listExtension")
	exts, err := util.GetExtensions(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error getting extensions: %s", err)
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Version"})
	for _, ext := range exts {
		table.Append([]string{ext.Name, ext.Version})
	}
	table.Render()
	return nil
}

func listUsermap(cmd *cobra.Command, args []string) error {
	var err error
	log := logger.Log(cmd.Context()).
		WithField("function", "listUsermap")
	foreignServer := ""
	if len(args) > 0 {
		foreignServer = strings.TrimSpace(args[0])
	}
	usermaps, err := util.GetUserMapsForServer(cmd.Context(), dbConnection, foreignServer)
	if err != nil {
		log.Errorf("error getting usermaps for server %s: %s", foreignServer, err)
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Local User", "Remote User", "Remote Password", "Remote Server"})
	for _, usermap := range usermaps {
		table.Append([]string{usermap.LocalUser, usermap.RemoteUser, usermap.RemoteSecret.Value, usermap.ServerName})
	}
	table.Render()
	return nil
}

func listSchema(cmd *cobra.Command, _ []string) error {
	log := logger.Log(cmd.Context()).
		WithField("function", "listSchema")
	schemas, err := util.GetSchemas(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error getting schemas: %s", err)
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Schema Name", "Foreign Server", "Remote Schema"})
	for _, schema := range schemas {
		table.Append([]string{schema.LocalSchema, schema.ServerName, schema.RemoteSchema})
	}
	table.Render()
	return nil
}
