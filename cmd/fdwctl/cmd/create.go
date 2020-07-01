package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/util"
	"github.com/spf13/cobra"
	"strings"
)

var (
	createCmd = &cobra.Command{
		Use:               "create <object type>",
		Short:             "Create objects",
		PersistentPreRunE: preDoCreate,
		PersistentPostRun: postDoCreate,
	}
	createServerCmd = &cobra.Command{
		Use:   "server",
		Short: "Create a foreign server",
		Run:   createServer,
	}
	createExtensionCmd = &cobra.Command{
		Use:   "extension",
		Short: "Create the postgres_fdw extension",
		Run:   createExtension,
	}
	createUsermapCmd = &cobra.Command{
		Use:   "usermap",
		Short: "Create a user mapping for a foreign server",
		Run:   createUsermap,
	}
	createSchemaCmd = &cobra.Command{
		Use:   "schema <local schema name>",
		Short: "Create (import) a schema from a foreign server",
		Run:   createSchema,
	}
	serverHost       string
	serverPort       string
	serverDBName     string
	localUser        string
	remoteUser       string
	remotePassword   string
	serverName       string
	localSchemaName  string
	remoteSchemaName string
	csServerName     string
)

func init() {
	createServerCmd.Flags().StringVar(&serverHost, "serverhost", "", "hostname of the remote PG server")
	createServerCmd.Flags().StringVar(&serverPort, "serverport", "5432", "port of the remote PG server")
	createServerCmd.Flags().StringVar(&serverDBName, "serverdbname", "", "database name on remote PG server")
	_ = createServerCmd.MarkFlagRequired("serverhost")
	_ = createServerCmd.MarkFlagRequired("serverport")
	_ = createServerCmd.MarkFlagRequired("serverdbname")

	createUsermapCmd.Flags().StringVar(&serverName, "servername", "", "foreign server name")
	createUsermapCmd.Flags().StringVar(&localUser, "localuser", "", "local user name")
	createUsermapCmd.Flags().StringVar(&remoteUser, "remoteuser", "", "remote user name")
	createUsermapCmd.Flags().StringVar(&remotePassword, "remotepassword", "", "remote user password")
	_ = createUsermapCmd.MarkFlagRequired("servername")
	_ = createUsermapCmd.MarkFlagRequired("localuser")
	_ = createUsermapCmd.MarkFlagRequired("remoteuser")
	_ = createUsermapCmd.MarkFlagRequired("remotepassword")

	createSchemaCmd.Flags().StringVar(&localSchemaName, "localschema", "", "local schema name")
	createSchemaCmd.Flags().StringVar(&csServerName, "servername", "", "foreign server name")
	createSchemaCmd.Flags().StringVar(&remoteSchemaName, "remoteschema", "", "the remote schema to import")
	_ = createSchemaCmd.MarkFlagRequired("localschema")
	_ = createSchemaCmd.MarkFlagRequired("servername")
	_ = createSchemaCmd.MarkFlagRequired("remoteschema")

	createCmd.AddCommand(createServerCmd)
	createCmd.AddCommand(createExtensionCmd)
	createCmd.AddCommand(createUsermapCmd)
	createCmd.AddCommand(createSchemaCmd)
}

func preDoCreate(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoCreate")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().FDWConnection)
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
	log := logger.Root().
		WithContext(cmd.Context()).
		WithField("function", "createExtension")
	query := "CREATE EXTENSION IF NOT EXISTS postgres_fdw"
	_, err := dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error creating fdw extension: %s", err)
		return
	}
	log.Info("extension postgres_fdw created")
}

func createServer(cmd *cobra.Command, _ []string) {
	log := logger.Root().
		WithContext(cmd.Context()).
		WithField("function", "createServer")
	hostSlug := strings.Replace(serverHost, ".", "-", -1)
	serverSlug := fmt.Sprintf("%s_%s_%s", hostSlug, serverPort, serverDBName)
	query := fmt.Sprintf(
		"CREATE SERVER %s FOREIGN DATA WRAPPER postgres_fdw OPTIONS (host '%s', port '%s', dbname '%s')",
		serverSlug,
		serverHost,
		serverPort,
		serverDBName,
	)
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error creating server: %s", err)
		return
	}
	log.Infof("server %s created", serverSlug)
}

func createUsermap(cmd *cobra.Command, _ []string) {
	log := logger.Root().
		WithContext(cmd.Context()).
		WithField("function", "createUsermap")
	err := util.EnsureUser(cmd.Context(), dbConnection, localUser, remotePassword)
	if err != nil {
		log.Errorf("error ensuring local user exists: %s", err)
		return
	}
	query := fmt.Sprintf("CREATE USER MAPPING FOR %s SERVER %s OPTIONS (user '%s', password '%s')", localUser, serverName, remoteUser, remotePassword)
	log.Tracef("query: %s", query)
	_, err = dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error creating user mapping: %s", err)
		return
	}
	log.Infof("user mapping %s -> %s created", localUser, remoteUser)
}

func createSchema(cmd *cobra.Command, _ []string) {
	log := logger.Root().
		WithContext(cmd.Context()).
		WithField("function", "createSchema")
	err := util.EnsureSchema(cmd.Context(), dbConnection, localSchemaName)
	if err != nil {
		log.Errorf("error ensuring local schema exists: %s", err)
		return
	}
	// TODO: support LIMIT TO and EXCEPT
	query := fmt.Sprintf("IMPORT FOREIGN SCHEMA %s FROM SERVER %s INTO %s", remoteSchemaName, csServerName, localSchemaName)
	log.Tracef("query: %s", query)
	_, err = dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error importing foreign schema: %s", err)
		return
	}
	log.Infof("foreign schema %s imported", remoteSchemaName)
}
