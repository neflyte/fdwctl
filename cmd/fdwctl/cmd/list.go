package cmd

import (
	"errors"
	"github.com/elgris/sqrl"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	listCmd = &cobra.Command{
		Use:     "list <objectType>",
		Short:   "List objects",
		PreRunE: preDoList,
		Run:     doList,
	}
)

type ForeignServersResult struct {
	ServerName  string `db:"foreign_server_name"`
	WrapperName string `db:"foreign_data_wrapper_name"`
	Owner       string `db:"authorization_identifier"`
}

func preDoList(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("object type is required")
	}
	return nil
}

func doList(cmd *cobra.Command, args []string) {
	log := logger.Root().WithField("function", "doList")
	objectType := strings.TrimSpace(args[0])
	switch objectType {
	case "server":
		conn, err := database.GetConnection(cmd.Context(), connectionString)
		if err != nil {
			log.Errorf("error getting database connection: %s", err)
			return
		}
		defer database.CloseConnection(cmd.Context(), conn)
		query, args, err := sqrl.
			Select("foreign_server_name, foreign_data_wrapper_name, authorization_identifier").
			From("information_schema.foreign_servers").
			PlaceholderFormat(sqrl.Dollar).
			ToSql()
		rows, err := conn.Query(cmd.Context(), query, args)
		if err != nil {
			log.Errorf("error querying for servers: %s", err)
			return
		}
		defer rows.Close()
		fservers := make([]ForeignServersResult, 0)
		fserver := new(ForeignServersResult)
		for rows.Next() {
			err = rows.Scan(fserver)
			if err != nil {
				log.Errorf("error scanning result row: %s", err)
				return
			}
			fservers = append(fservers, *fserver)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Wrapper", "Owner"})
		for _, svr := range fservers {
			table.Append([]string{
				svr.ServerName,
				svr.WrapperName,
				svr.Owner,
			})
		}
		table.Render()
	default:
		log.Errorf("unknown object type: %s", objectType)
	}
}
