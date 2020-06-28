package cmd

import (
	"errors"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	listCmd = &cobra.Command{
		Use:      "list <objectType>",
		Short:    "List objects [server, extension]",
		PreRunE:  preDoList,
		PostRunE: postDoList,
		Run:      doList,
	}
	dbConnection *pgx.Conn
)

type ForeignServersResult struct {
	ServerName  string `db:"foreign_server_name"`
	WrapperName string `db:"foreign_data_wrapper_name"`
	Owner       string `db:"authorization_identifier"`
}

type ExtensionsResult struct {
	Name    string `db:"extname"`
	Version string `db:"extversion"`
}

func preDoList(cmd *cobra.Command, args []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoList")
	if len(args) == 0 {
		return errors.New("object type is required")
	}
	dbConnection, err = database.GetConnection(cmd.Context(), connectionString)
	if err != nil {
		log.Errorf("error getting database connection: %s", err)
		return err
	}
	return nil
}

func postDoList(cmd *cobra.Command, _ []string) error {
	database.CloseConnection(cmd.Context(), dbConnection)
	return nil
}

func doList(cmd *cobra.Command, args []string) {
	log := logger.Root().WithField("function", "doList")
	objectType := strings.TrimSpace(args[0])
	switch objectType {
	case "server":
		listServers(cmd)
	case "extension":
		listExtension(cmd)
	default:
		log.Errorf("unknown object type: %s", objectType)
	}
}

func listServers(cmd *cobra.Command) {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "listServers")
	query, _, _ := sqrl.
		Select("foreign_server_name, foreign_data_wrapper_name, authorization_identifier").
		From("information_schema.foreign_servers").
		ToSql()
	rows, err := dbConnection.Query(cmd.Context(), query)
	if err != nil {
		log.Errorf("error querying for servers: %s", err)
		return
	}
	defer rows.Close()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Wrapper", "Owner"})
	for rows.Next() {
		fserver := new(ForeignServersResult)
		err = rows.Scan(&fserver.ServerName, &fserver.WrapperName, &fserver.Owner)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		table.Append([]string{
			fserver.ServerName,
			fserver.WrapperName,
			fserver.Owner,
		})
	}
	table.Render()
}

func listExtension(cmd *cobra.Command) {
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "listExtension")
	query, _, _ := sqrl.
		Select("extname", "extversion").
		From("pg_extension").
		ToSql()
	rows, err := dbConnection.Query(cmd.Context(), query)
	if err != nil {
		log.Errorf("error querying for extensions: %s", err)
		return
	}
	defer rows.Close()
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Version"})
	for rows.Next() {
		ext := new(ExtensionsResult)
		err = rows.Scan(&ext.Name, &ext.Version)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			return
		}
		table.Append([]string{
			ext.Name,
			ext.Version,
		})
	}
	table.Render()
}
