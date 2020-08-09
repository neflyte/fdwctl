package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"github.com/neflyte/fdwctl/internal/util"
	"github.com/spf13/cobra"
	"strconv"
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
		Use:   "server <server name>",
		Short: "Create a foreign server",
		Run:   createServer,
		Args:  cobra.MinimumNArgs(1),
	}
	createExtensionCmd = &cobra.Command{
		Use:   "extension <extension name>",
		Short: "Create a PG extension (usually postgres_fdw)",
		Run:   createExtension,
		Args:  cobra.MinimumNArgs(1),
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
	log := logger.Log(cmd.Context()).
		WithField("function", "preDoCreate")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().GetDatabaseConnectionString())
	if err != nil {
		return logger.ErrorfAsError(log, "error getting database connection: %s", err)
	}
	return nil
}

func postDoCreate(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func createExtension(cmd *cobra.Command, args []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "createExtension")
	extName := strings.TrimSpace(args[0])
	err := util.CreateExtension(cmd.Context(), dbConnection, model.Extension{
		Name: extName,
	})
	if err != nil {
		log.Errorf("error creating extension %s: %s", extName, err)
		return
	}
	log.Infof("extension %s created", extName)
}

func createServer(cmd *cobra.Command, args []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "createServer")
	serverSlug := strings.TrimSpace(args[0])
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
	portInt, err := strconv.Atoi(serverPort)
	if err != nil {
		log.Errorf("error converting port to integer: %s", err)
		return
	}
	err = util.CreateServer(cmd.Context(), dbConnection, model.ForeignServer{
		Name: serverSlug,
		Host: serverHost,
		Port: portInt,
		DB:   serverDBName,
	})
	if err != nil {
		log.Errorf("error creating server: %s", err)
		return
	}
	log.Infof("server %s created", serverSlug)
}

func createUsermap(cmd *cobra.Command, _ []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "createUsermap")
	err := util.EnsureUser(cmd.Context(), dbConnection, localUser, remotePassword)
	if err != nil {
		log.Errorf("error ensuring local user exists: %s", err)
		return
	}
	err = util.CreateUserMap(cmd.Context(), dbConnection, model.UserMap{
		ServerName: serverName,
		LocalUser:  localUser,
		RemoteUser: remoteUser,
		RemoteSecret: model.Secret{
			Value: remotePassword,
		},
	})
	if err != nil {
		log.Errorf("error creating user mapping: %s", err)
		return
	}
	log.Infof("user mapping %s -> %s created", localUser, remoteUser)
}

func createSchema(cmd *cobra.Command, _ []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "createSchema")
	err := util.ImportSchema(cmd.Context(), dbConnection, csServerName, model.Schema{
		ServerName:     csServerName,
		LocalSchema:    localSchemaName,
		RemoteSchema:   remoteSchemaName,
		ImportENUMs:    importEnums,
		ENUMConnection: importEnumConnection,
	})
	if err != nil {
		log.Errorf("error importing foreign schema: %s", err)
		return
	}
	log.Infof("foreign schema %s imported", remoteSchemaName)
}
