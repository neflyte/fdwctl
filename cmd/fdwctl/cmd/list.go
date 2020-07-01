package cmd

import (
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
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
		Short: "List foreign server objects",
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
	dbConnection *pgx.Conn
)

func init() {
	listCmd.AddCommand(listServerCmd)
	listCmd.AddCommand(listExtensionCmd)
	listCmd.AddCommand(listUsermapCmd)
}

func preDoList(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoList")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().FDWConnection)
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
		Select(
			"fs.foreign_server_name",
			"fs.foreign_data_wrapper_name",
			"fs.authorization_identifier",
			"fsoh.option_value AS hostname",
			"fsop.option_value AS port",
			"fsod.option_value AS dbname",
		).From("information_schema.foreign_servers fs").
		Join("information_schema.foreign_server_options fsoh ON fsoh.foreign_server_name = fs.foreign_server_name AND fsoh.option_name = 'host'").
		Join("information_schema.foreign_server_options fsop ON fsop.foreign_server_name = fs.foreign_server_name AND fsop.option_name = 'port'").
		Join("information_schema.foreign_server_options fsod ON fsod.foreign_server_name = fs.foreign_server_name AND fsod.option_name = 'dbname'").
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
	var serverName, wrapperName, owner, hostName, port, dbName string
	for rows.Next() {
		err = rows.Scan(&serverName, &wrapperName, &owner, &hostName, &port, &dbName)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		table.Append([]string{serverName, wrapperName, owner, hostName, port, dbName})
	}
	rows.Close()
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
	sb := sqrl.
		Select("u.authorization_identifier", "ou.option_value", "op.option_value", "s.srvname").
		From("information_schema.user_mappings u").
		Join("information_schema.user_mapping_options ou ON ou.authorization_identifier = u.authorization_identifier AND ou.option_name = 'user'").
		Join("information_schema.user_mapping_options op ON op.authorization_identifier = u.authorization_identifier AND op.option_name = 'password'").
		Join("pg_user_mappings s ON s.usename = u.authorization_identifier")
	if foreignServer != "" {
		sb = sb.Where(sqrl.Eq{"s.srvname": foreignServer})
	}
	query, qArgs, err := sb.PlaceholderFormat(sqrl.Dollar).ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return
	}
	log.Tracef("query: %s, args: %#v", query, qArgs)
	mappingRows, err := dbConnection.Query(cmd.Context(), query, qArgs...)
	if err != nil {
		log.Errorf("error selecting user mappings: %s", err)
		return
	}
	defer mappingRows.Close()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Local User", "Remote User", "Remote Password", "Remote Server"})
	var user, optUser, optPass, srvName string
	for mappingRows.Next() {
		err = mappingRows.Scan(&user, &optUser, &optPass, &srvName)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		table.Append([]string{user, optUser, optPass, srvName})
	}
	table.Render()
}
