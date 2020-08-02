package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/util"
	"github.com/spf13/cobra"
	"sort"
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
		Use:   "schema",
		Short: "Create (import) a schema from a foreign server",
		Run:   createSchema,
	}
	serverHost           string
	serverPort           string
	serverDBName         string
	localUser            string
	remoteUser           string
	remotePassword       string
	serverName           string
	localSchemaName      string
	remoteSchemaName     string
	csServerName         string
	importEnums          bool
	importEnumConnection string
)

func init() {
	createServerCmd.Flags().StringVar(&serverHost, "serverhost", "", "hostname of the remote PG server")
	createServerCmd.Flags().StringVar(&serverPort, "serverport", "5432", "port of the remote PG server")
	createServerCmd.Flags().StringVar(&serverDBName, "serverdbname", "", "database name on remote PG server")
	createServerCmd.Flags().StringVar(&csServerName, "servername", "", "foerign server name (optional)")
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
	createSchemaCmd.Flags().BoolVar(&importEnums, "importenums", false, "attempt to auto-create ENUMs locally before import")
	createSchemaCmd.Flags().StringVar(&importEnumConnection, "enumconnection", "", "connection string of database to import enums from")
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
	serverSlug := csServerName
	if serverSlug == "" {
		hostSlug := strings.Replace(serverHost, ".", "_", -1)
		serverSlug = fmt.Sprintf("%s_%s_%s", hostSlug, serverPort, serverDBName)
		/* If the server slug starts with a number, prepend "server_" to it
		   since PG doesn't like a number at the beginning of a server name. */
		log.Debugf("serverSlug: %s", serverSlug)
		if util.StartsWithNumber(serverSlug) {
			serverSlug = fmt.Sprintf("server_%s", serverSlug)
		}
	}
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
	// Sanity Check
	if importEnums && importEnumConnection == "" {
		log.Errorf("enum database connection string is required when importing enums")
		return
	}
	// Ensure the local schema exists
	err := util.EnsureSchema(cmd.Context(), dbConnection, localSchemaName)
	if err != nil {
		log.Errorf("error ensuring local schema exists: %s", err)
		return
	}
	if importEnums {
		fdbConn, err := database.GetConnection(cmd.Context(), importEnumConnection)
		if err != nil {
			log.Errorf("error connecting to foreign database: %s", err)
			return
		}
		defer database.CloseConnection(cmd.Context(), fdbConn)
		remoteEnums, err := util.GetSchemaEnumsUsedInTables(cmd.Context(), fdbConn, remoteSchemaName)
		if err != nil {
			log.Errorf("error getting remote ENUMs: %s", err)
			return
		}
		sort.Strings(remoteEnums)
		// Get a list of local enums, too
		localEnums, err := util.GetEnums(cmd.Context(), dbConnection)
		if err != nil {
			log.Errorf("error getting local ENUMs: %s", err)
			return
		}
		sort.Strings(localEnums)
		// Get enough data from remote database to re-create enums and then create them
		for _, remoteEnum := range remoteEnums {
			if sort.SearchStrings(localEnums, remoteEnum) != len(localEnums) {
				log.Debugf("local enum %s exists", remoteEnum)
				continue
			}
			enumStrings, err := util.GetEnumStrings(cmd.Context(), fdbConn, remoteEnum)
			if err != nil {
				log.Errorf("error getting enum values: %s", err)
				return
			}
			query := fmt.Sprintf("CREATE TYPE %s AS ENUM (", remoteEnum)
			quotedEnumStrings := make([]string, 0)
			for _, enumString := range enumStrings {
				quotedEnumStrings = append(quotedEnumStrings, fmt.Sprintf("'%s'", enumString))
			}
			query = fmt.Sprintf("%s %s )", query, strings.Join(quotedEnumStrings, ","))
			log.Tracef("query: %s", query)
			_, err = dbConnection.Exec(cmd.Context(), query)
			if err != nil {
				log.Errorf("error creating local enum type: %s", err)
				return
			}
			log.Infof("enum type %s created", remoteEnum)
		}
		// Close the foreign DB connection since we no longer need it
		database.CloseConnection(cmd.Context(), fdbConn)
	}
	// TODO: support LIMIT TO and EXCEPT
	query := fmt.Sprintf("IMPORT FOREIGN SCHEMA %s FROM SERVER %s INTO %s", remoteSchemaName, csServerName, localSchemaName) //nolint:gosec
	log.Tracef("query: %s", query)
	_, err = dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error importing foreign schema: %s", err)
		return
	}
	log.Infof("foreign schema %s imported", remoteSchemaName)
}
