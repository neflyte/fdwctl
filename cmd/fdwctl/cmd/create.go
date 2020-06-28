package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/spf13/cobra"
	"strings"
)

var (
	createCmd = &cobra.Command{
		Use:               "create <objectType>",
		Short:             "Create objects",
		PersistentPreRunE: preDoCreate,
		PersistentPostRun: postDoCreate,
	}
	createServerCmd = &cobra.Command{
		Use:   "server",
		Short: "Create a server object",
		Run:   createServer,
	}
	createExtensionCmd = &cobra.Command{
		Use:   "extension",
		Short: "Create the postgres_fdw extension",
		Run:   createExtension,
	}
	serverHost   string
	serverPort   string
	serverDBName string
)

func init() {
	createServerCmd.Flags().StringVar(&serverHost, "serverhost", "", "hostname of the remote PG server")
	createServerCmd.Flags().StringVar(&serverPort, "serverport", "5432", "port of the remote PG server")
	createServerCmd.Flags().StringVar(&serverDBName, "serverdbname", "", "database name on remote PG server")
	_ = createServerCmd.MarkFlagRequired("serverhost")
	_ = createServerCmd.MarkFlagRequired("serverport")
	_ = createServerCmd.MarkFlagRequired("serverdbname")
	createCmd.AddCommand(createServerCmd)
	createCmd.AddCommand(createExtensionCmd)
}

func preDoCreate(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoCreate")
	dbConnection, err = database.GetConnection(cmd.Context(), connectionString)
	if err != nil {
		log.Errorf("error getting database connection: %s", err)
		return err
	}
	return nil
}

func postDoCreate(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func createExtension(cmd *cobra.Command, _ []string) {
	log := logger.Root().WithField("function", "createExtension")
	query := "CREATE EXTENSION IF NOT EXISTS postgres_fdw"
	_, err := dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error creating fdw extension: %s", err)
		return
	}
	log.Info("extension postgres_fdw created")
}

func createServer(cmd *cobra.Command, _ []string) {
	log := logger.Root().WithField("function", "createServer")
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
