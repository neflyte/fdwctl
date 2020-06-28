package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/spf13/cobra"
	"strings"
)

var (
	editCmd = &cobra.Command{
		Use:               "edit",
		Short:             "Edit objects",
		PersistentPreRunE: preDoEdit,
		PersistentPostRun: postDoEdit,
	}
	editServerCmd = &cobra.Command{
		Use:   "server <server name>",
		Short: "Edit a server object",
		Run:   editServer,
		Args:  cobra.MinimumNArgs(1),
	}
	editServerName   string
	editServerHost   string
	editServerPort   string
	editServerDBName string
)

func init() {
	editServerCmd.Flags().StringVar(&editServerName, "servername", "", "the new name for the server object")
	editServerCmd.Flags().StringVar(&editServerHost, "serverhost", "", "the new hostname of the server object")
	editServerCmd.Flags().StringVar(&editServerPort, "serverport", "", "the new port of the server object")
	editServerCmd.Flags().StringVar(&editServerDBName, "serverdbname", "", "the new database name of the server object")
	editCmd.AddCommand(editServerCmd)
}

func preDoEdit(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoEdit")
	dbConnection, err = database.GetConnection(cmd.Context(), connectionString)
	if err != nil {
		log.Errorf("error getting database connection: %s", err)
		return err
	}
	return nil
}

func postDoEdit(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func editServer(cmd *cobra.Command, args []string) {
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "editServer")
	serverName := strings.TrimSpace(args[0])
	if serverName == "" {
		log.Errorf("server name is required")
		return
	}
	// Edit server hostname, port, and dbname first
	if editServerHost != "" || editServerPort != "" || editServerDBName != "" {
		query := fmt.Sprintf("ALTER SERVER %s OPTIONS (", serverName)
		opts := make([]string, 0)
		if editServerHost != "" {
			opts = append(opts, fmt.Sprintf("SET host '%s'", editServerHost))
		}
		if editServerPort != "" {
			opts = append(opts, fmt.Sprintf("SET port '%s'", editServerPort))
		}
		if editServerDBName != "" {
			opts = append(opts, fmt.Sprintf("SET dbname '%s'", editServerDBName))
		}
		query = fmt.Sprintf("%s %s )", query, strings.Join(opts, ","))
		log.Tracef("query: %s", query)
		_, err := dbConnection.Exec(cmd.Context(), query)
		if err != nil {
			log.Errorf("error editing server: %s", err)
			return
		}
		log.Infof("server %s edited", serverName)
	}
	// Rename server entry
	if editServerName != "" {
		query := fmt.Sprintf("ALTER SERVER %s RENAME TO %s", serverName, editServerName)
		log.Tracef("query: %s", query)
		_, err := dbConnection.Exec(cmd.Context(), query)
		if err != nil {
			log.Errorf("error renaming server object: %s", err)
			return
		}
		log.Infof("server %s renamed to %s", serverName, editServerName)
	}
}
