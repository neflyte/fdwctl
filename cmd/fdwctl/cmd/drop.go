package cmd

import (
	"strings"

	"github.com/neflyte/fdwctl/lib/config"
	"github.com/neflyte/fdwctl/lib/database"
	"github.com/neflyte/fdwctl/lib/logger"
	"github.com/neflyte/fdwctl/lib/model"
	"github.com/neflyte/fdwctl/lib/util"
	"github.com/spf13/cobra"
)

const (
	dropServerCmdMinArgCount = 2
)

var (
	dropCmd = &cobra.Command{
		Use:               "drop",
		Short:             "Drop (delete) objects",
		PersistentPreRunE: preDoDrop,
		PersistentPostRun: postDoDrop,
	}
	dropExtensionCmd = &cobra.Command{
		Use:   "extension <extension name>",
		Short: "Drop a PG extension (usually postgres_fdw)",
		Run:   dropExtension,
		Args:  cobra.MinimumNArgs(1),
	}
	dropServerCmd = &cobra.Command{
		Use:   "server <server name>",
		Short: "Drop a foreign server",
		Run:   dropServer,
		Args:  cobra.MinimumNArgs(1),
	}
	dropUsermapCmd = &cobra.Command{
		Use:   "usermap <server name> <local user>",
		Short: "Drop a user mapping for a foreign server",
		Run:   dropUsermap,
		Args:  cobra.MinimumNArgs(dropServerCmdMinArgCount),
	}
	dropSchemaCmd = &cobra.Command{
		Use:   "schema <schema name>",
		Short: "Drop a schema",
		Run:   dropSchema,
		Args:  cobra.MinimumNArgs(1),
	}
	cascadeDrop   bool
	dropLocalUser bool
)

func init() {
	dropUsermapCmd.Flags().BoolVar(&dropLocalUser, "droplocal", false, "also drop the local USER object")

	dropCmd.PersistentFlags().BoolVar(&cascadeDrop, "cascade", false, "drop objects with CASCADE option")
	dropCmd.AddCommand(dropExtensionCmd)
	dropCmd.AddCommand(dropServerCmd)
	dropCmd.AddCommand(dropUsermapCmd)
	dropCmd.AddCommand(dropSchemaCmd)
}

func preDoDrop(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.Log(cmd.Context()).
		WithField("function", "preDoDrop")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().GetDatabaseConnectionString())
	if err != nil {
		return logger.ErrorfAsError(log, "error getting database connection: %s", err)
	}
	return nil
}

func postDoDrop(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func dropExtension(cmd *cobra.Command, args []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "dropExtension")
	dropExtName := strings.TrimSpace(args[0])
	err := util.DropExtension(cmd.Context(), dbConnection, model.Extension{
		Name: dropExtName,
	})
	if err != nil {
		log.Errorf("error dropping extension %s: %s", dropExtName, err)
		return
	}
	log.Infof("extension %s dropped", dropExtName)
}

func dropServer(cmd *cobra.Command, args []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "dropServer")
	dsServerName := strings.TrimSpace(args[0])
	err := util.DropServer(cmd.Context(), dbConnection, dsServerName, cascadeDrop)
	if err != nil {
		log.Errorf("error dropping server: %s", err)
		return
	}
	log.Infof("server %s dropped", dsServerName)
}

func dropUsermap(cmd *cobra.Command, args []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "dropUsermap")
	duServerName := strings.TrimSpace(args[0])
	if duServerName == "" {
		log.Errorf("server name is required")
		return
	}
	duLocalUser := strings.TrimSpace(args[1])
	if duLocalUser == "" {
		log.Errorf("local user name is required")
		return
	}
	err := util.DropUserMap(cmd.Context(), dbConnection, model.UserMap{
		ServerName: duServerName,
		LocalUser:  duLocalUser,
	}, dropLocalUser)
	if err != nil {
		log.Errorf("error dropping user mapping: %s", err)
		return
	}
	log.Infof("user mapping for %s dropped", duLocalUser)
}

func dropSchema(cmd *cobra.Command, args []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "dropSchema")
	dsSchemaName := strings.TrimSpace(args[0])
	if dsSchemaName == "" {
		log.Errorf("schema name is required")
		return
	}
	err := util.DropSchema(cmd.Context(), dbConnection, model.Schema{
		LocalSchema: dsSchemaName,
	}, cascadeDrop)
	if err != nil {
		log.Errorf("error dropping schema: %s", err)
		return
	}
	log.Infof("schema %s dropped", dsSchemaName)
}
