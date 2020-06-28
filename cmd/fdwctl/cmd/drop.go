package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/spf13/cobra"
	"strings"
)

var (
	dropCmd = &cobra.Command{
		Use:               "drop",
		Short:             "Drop (delete) objects",
		PersistentPreRunE: preDoDrop,
		PersistentPostRun: postDoDrop,
	}
	dropExtensionCmd = &cobra.Command{
		Use:   "extension",
		Short: "drop the postgres_fdw extension",
		Run:   dropExtension,
	}
	dropServerCmd = &cobra.Command{
		Use:   "server <server name>",
		Short: "drop a foreign server",
		Run:   dropServer,
		Args:  cobra.MinimumNArgs(1),
	}
	cascadeDrop bool
)

func init() {
	dropCmd.PersistentFlags().BoolVar(&cascadeDrop, "cascade", false, "drop objects with CASCADE option")
	dropCmd.AddCommand(dropExtensionCmd)
	dropCmd.AddCommand(dropServerCmd)
}

func preDoDrop(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoDrop")
	dbConnection, err = database.GetConnection(cmd.Context(), connectionString)
	if err != nil {
		log.Errorf("error getting database connection: %s", err)
		return err
	}
	return nil
}

func postDoDrop(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func dropExtension(cmd *cobra.Command, _ []string) {
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "dropExtension")
	query := "DROP EXTENSION IF EXISTS postgres_fdw"
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error dropping fdw extension: %s", err)
		return
	}
	log.Info("extension postgres_fdw dropped")
}

func dropServer(cmd *cobra.Command, args []string) {
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "dropServer")
	serverName := strings.TrimSpace(args[0])
	if serverName == "" {
		log.Errorf("server name is required")
		return
	}
	query := fmt.Sprintf("DROP SERVER %s", serverName)
	if cascadeDrop {
		query = fmt.Sprintf("%s CASCADE", query)
	}
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error dropping server: %s", err)
		return
	}
	log.Infof("server %s dropped", serverName)
}
