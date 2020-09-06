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

const (
	editUserCmdMinArgCount = 2
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
		Short: "Edit a foreign server",
		RunE:  editServer,
		Args:  cobra.MinimumNArgs(1),
	}
	editUsermapCmd = &cobra.Command{
		Use:   "usermap <server name> <local user>",
		Short: "Edit a user mapping for a foreign server",
		RunE:  editUsermap,
		Args:  cobra.MinimumNArgs(editUserCmdMinArgCount),
	}
	editServerName     string
	editServerHost     string
	editServerPort     string
	editServerDBName   string
	editRemoteUser     string
	editRemotePassword string
)

func init() {
	editServerCmd.Flags().StringVar(&editServerName, "servername", "", "the new name for the server object")
	editServerCmd.Flags().StringVar(&editServerHost, "serverhost", "", "the new hostname of the server object")
	editServerCmd.Flags().StringVar(&editServerPort, "serverport", "", "the new port of the server object")
	editServerCmd.Flags().StringVar(&editServerDBName, "serverdbname", "", "the new database name of the server object")
	editUsermapCmd.Flags().StringVar(&editRemoteUser, "remoteuser", "", "the new remote user name")
	editUsermapCmd.Flags().StringVar(&editRemotePassword, "remotepassword", "", "the new password for the remote user")
	editCmd.AddCommand(editServerCmd)
	editCmd.AddCommand(editUsermapCmd)
}

func preDoEdit(cmd *cobra.Command, _ []string) error {
	var err error
	log := logger.Log(cmd.Context()).
		WithField("function", "preDoEdit")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().GetDatabaseConnectionString())
	if err != nil {
		return logger.ErrorfAsError(log, "error getting database connection: %s", err)
	}
	return nil
}

func postDoEdit(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func editServer(cmd *cobra.Command, args []string) error {
	log := logger.Log(cmd.Context()).
		WithField("function", "editServer")
	esServerName := strings.TrimSpace(args[0])
	if esServerName == "" {
		log.Errorf("server name is required")
		return fmt.Errorf("server name is required")
	}
	portInt, err := strconv.Atoi(editServerPort)
	if err != nil {
		log.Errorf("error converting port to integer: %s", err)
		return err
	}
	fServer := model.ForeignServer{
		Name: esServerName,
		Host: editServerHost,
		Port: portInt,
		DB:   editServerDBName,
	}
	err = util.UpdateServer(cmd.Context(), dbConnection, fServer)
	if err != nil {
		log.Errorf("error editing server: %s", err)
		return err
	}
	log.Infof("server %s edited", esServerName)
	// Rename server entry
	if editServerName != "" {
		err = util.UpdateServerName(cmd.Context(), dbConnection, fServer, editServerName)
		if err != nil {
			log.Errorf("error renaming foreign server: %s", err)
			return err
		}
		log.Infof("server %s renamed to %s", esServerName, editServerName)
	}
	return nil
}

func editUsermap(cmd *cobra.Command, args []string) error {
	log := logger.Log(cmd.Context()).
		WithField("function", "editServer")
	euServerName := strings.TrimSpace(args[0])
	if euServerName == "" {
		log.Errorf("server name is required")
		return fmt.Errorf("server name is required")
	}
	euLocalUser := strings.TrimSpace(args[1])
	if euLocalUser == "" {
		log.Errorf("local user name is required")
		return fmt.Errorf("local user name is required")
	}
	err := util.UpdateUserMap(cmd.Context(), dbConnection, model.UserMap{
		ServerName: euServerName,
		LocalUser:  euLocalUser,
		RemoteUser: editRemoteUser,
		RemoteSecret: model.Secret{
			Value: editRemotePassword,
		},
	})
	if err != nil {
		log.Errorf("error editing user mapping: %s", err)
		return err
	}
	log.Infof("user mapping %s edited", euLocalUser)
	return nil
}
