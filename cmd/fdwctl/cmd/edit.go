package cmd

import (
	"fmt"
	"github.com/neflyte/fdwctl/internal/config"
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
		Short: "Edit a foreign server",
		Run:   editServer,
		Args:  cobra.MinimumNArgs(1),
	}
	editUsermapCmd = &cobra.Command{
		Use:   "usermap <server name> <local user>",
		Short: "Edit a user mapping for a foreign server",
		Run:   editUsermap,
		Args:  cobra.MinimumNArgs(2),
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
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoEdit")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().FDWConnection)
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
	esServerName := strings.TrimSpace(args[0])
	if esServerName == "" {
		log.Errorf("server name is required")
		return
	}
	// Edit server hostname, port, and dbname first
	if editServerHost != "" || editServerPort != "" || editServerDBName != "" {
		query := fmt.Sprintf("ALTER SERVER %s OPTIONS (", esServerName)
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
		log.Infof("server %s edited", esServerName)
	}
	// Rename server entry
	if editServerName != "" {
		query := fmt.Sprintf("ALTER SERVER %s RENAME TO %s", esServerName, editServerName)
		log.Tracef("query: %s", query)
		_, err := dbConnection.Exec(cmd.Context(), query)
		if err != nil {
			log.Errorf("error renaming server object: %s", err)
			return
		}
		log.Infof("server %s renamed to %s", esServerName, editServerName)
	}
}

func editUsermap(cmd *cobra.Command, args []string) {
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "editServer")
	euServerName := strings.TrimSpace(args[0])
	if euServerName == "" {
		log.Errorf("server name is required")
		return
	}
	euLocalUser := strings.TrimSpace(args[1])
	if euLocalUser == "" {
		log.Errorf("local user name is required")
		return
	}
	if editRemoteUser == "" && editRemotePassword == "" {
		log.Warn("no remote user name or password specified; nothing to do")
		return
	}
	optArgs := make([]string, 0)
	query := fmt.Sprintf("ALTER USER MAPPING FOR %s SERVER %s OPTIONS (", euLocalUser, euServerName)
	if editRemoteUser != "" {
		optArgs = append(optArgs, fmt.Sprintf("SET user '%s'", editRemoteUser))
	}
	if editRemotePassword != "" {
		optArgs = append(optArgs, fmt.Sprintf("SET password '%s'", editRemotePassword))
	}
	query = fmt.Sprintf("%s %s )", query, strings.Join(optArgs, ", "))
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(cmd.Context(), query)
	if err != nil {
		log.Errorf("error editing user mapping: %s", err)
		return
	}
	log.Infof("user mapping %s edited", euLocalUser)
}
