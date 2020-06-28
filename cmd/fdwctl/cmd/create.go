package cmd

import (
	"errors"
	"fmt"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/spf13/cobra"
	"strings"
)

var (
	createCmd = &cobra.Command{
		Use:       "create <objectType>",
		Short:     "Create objects",
		PreRunE:   preDoCreate,
		PostRunE:  postDoCreate,
		Run:       doCreate,
		Args:      cobra.MinimumNArgs(1),
		ValidArgs: []string{"server", "extension"},
	}
	serverHost   string
	serverPort   string
	serverDBName string
)

func init() {
	createCmd.Flags().StringVar(&serverHost, "serverhost", "", "hostname of the remote PG server")
	createCmd.Flags().StringVar(&serverPort, "serverport", "5432", "port of the remote PG server")
	createCmd.Flags().StringVar(&serverDBName, "serverdbname", "", "database name on remote PG server")
}

func preDoCreate(cmd *cobra.Command, args []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoCreate")
	if connectionString == "" {
		return errors.New("fdw database connection string is required")
	}
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

func postDoCreate(cmd *cobra.Command, _ []string) error {
	database.CloseConnection(cmd.Context(), dbConnection)
	return nil
}

func doCreate(cmd *cobra.Command, args []string) {
	log := logger.Root().WithField("function", "doCreate")
	objectType := strings.TrimSpace(args[0])
	switch objectType {
	case "server":
		createServer(cmd)
	case "extension":
		createExtension(cmd)
	default:
		log.Errorf("unknown object type: %s", objectType)
	}
}

func createExtension(cmd *cobra.Command) {
	log := logger.Root().WithField("function", "createExtension")
	query := "CREATE EXTENSION IF NOT EXISTS postgres_fdw"
	_, err := dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error creating fdw extension: %s", err)
		return
	}
	log.Info("extension postgres_fdw created")
}

func createServer(cmd *cobra.Command) {
	log := logger.Root().WithField("function", "createExtension")
	if serverHost == "" {
		log.Errorf("server hostname is required")
		return
	}
	if serverPort == "" {
		log.Errorf("server port is required")
		return
	}
	if serverDBName == "" {
		log.Errorf("database name is required")
		return
	}
	hostSlug := strings.Replace(serverHost, ".", "-", -1)
	serverName := fmt.Sprintf("%s_%s_%s", hostSlug, serverPort, serverDBName)
	query := fmt.Sprintf(
		"CREATE SERVER %s FOREIGN DATA WRAPPER postgres_fdw OPTIONS (host '%s', port '%s', dbname '%s')",
		serverName,
		serverHost,
		serverPort,
		serverDBName,
	)
	_, err := dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error creating server: %s", err)
		return
	}
	log.Infof("server %s created", serverName)
}
