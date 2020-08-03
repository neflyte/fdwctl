package cmd

import (
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/util"
	"github.com/spf13/cobra"
)

var (
	desiredStateCmd = &cobra.Command{
		Use:               "desiredstate",
		Short:             "Apply a desired state",
		Long:              "Apply a desired state configuration to the FDW database",
		PersistentPreRunE: preDoDesiredState,
		PersistentPostRun: postDoDesiredState,
		Run:               doDesiredState,
	}
)

func preDoDesiredState(cmd *cobra.Command, _ []string) error {
	var err error

	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "preDoDesiredState")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().FDWConnection)
	if err != nil {
		log.Errorf("error getting database connection: %s", err)
		return err
	}
	return nil
}

func postDoDesiredState(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func doDesiredState(cmd *cobra.Command, _ []string) {
	log := logger.
		Root().
		WithContext(cmd.Context()).
		WithField("function", "doDesiredState")
	// Convert DState servers to ForeignServers
	dStateServers := util.ForeignServersFromDesiredStateServers(config.Instance().DesiredState.Servers)
	// List servers in DB
	dbServers, err := util.GetServers(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error getting foreign servers: %s", err)
		return
	}
	// Diff servers
	serversInDBButNotInDState, serversInDStateButNotInDB, serversAlreadyInDB, _ := util.DiffForeignServers(dStateServers, dbServers)
	// Remove servers in DB but not in DState
	for _, serverNotInDState := range serversInDBButNotInDState {
		err := util.DropServer(cmd.Context(), dbConnection, serverNotInDState.Name, true)
		if err != nil {
			log.Errorf("error dropping server %s that is not in desired state: %s", serverNotInDState, err)
			return
		}
		log.Infof("server %s dropped", serverNotInDState)
	}
	// Create servers that are in DState but not yet in DB
	for _, serverNotInDB := range serversInDStateButNotInDB {
		err := util.CreateServer(cmd.Context(), dbConnection, serverNotInDB)
		if err != nil {
			log.Errorf("error creating server: %s", err)
			return
		}
		log.Infof("server %s created", serverNotInDB.Name)
	}
	// Update servers that were already in the DB
	for _, serverAlreadyInDB := range serversAlreadyInDB {
		err := util.UpdateServer(cmd.Context(), dbConnection, serverAlreadyInDB)
		if err != nil {
			log.Errorf("error updating server: %s", err)
			return
		}
		log.Infof("server %s updated", serverAlreadyInDB.Name)
		// List Usermaps for this server in the DB
		dbServerUsermaps, err := util.GetUserMapsForServer(cmd.Context(), dbConnection, serverAlreadyInDB.Name)
		if err != nil {
			log.Errorf("error getting usermaps for server %s: %s", serverAlreadyInDB.Name, err)
			return
		}
		// List Usermaps for this server in DState
		dsServer, err := util.FindDesiredStateServer(config.Instance().DesiredState.Servers, serverAlreadyInDB.Name)
		if err != nil {
			log.Errorf("cannot find desired state server %s; THIS IS UNEXPECTED", serverAlreadyInDB.Name)
			return
		}
		dStateServerUsermaps := util.UserMapsFromDesiredStateUserMaps(serverAlreadyInDB.Name, dsServer.UserMap)
		// Diff usermaps
		usRemove, usAdd, usModify, _ := util.DiffUserMaps(dStateServerUsermaps, dbServerUsermaps)
		// Delete Usermaps not in DState
		for _, usermapToRemove := range usRemove {
			err = util.DropUserMap(cmd.Context(), dbConnection, usermapToRemove, true)
			if err != nil {
				log.Errorf("error dropping user map for local user %s: %s", usermapToRemove.LocalUser, err)
				return
			}
			log.Infof("user mapping for %s dropped", usermapToRemove.LocalUser)
		}
		// Add Usermaps in DState but not in DB
		for _, usermapToAdd := range usAdd {
			err = util.CreateUserMap(cmd.Context(), dbConnection, usermapToAdd)
			if err != nil {
				log.Errorf("error creating user map for local user %s: %s", usermapToAdd.LocalUser, err)
				return
			}
		}
		// Update usermaps that are already there
		for _, usermapToUpdate := range usModify {
			err = util.UpdateUserMap(cmd.Context(), dbConnection, usermapToUpdate)
			if err != nil {
				log.Errorf("error updating user map for local user %s: %s", usermapToUpdate.LocalUser, err)
				return
			}
		}
		// Get DB remote schemas
		dbSchemas, err := util.GetSchemas(cmd.Context(), dbConnection)
		if err != nil {
			log.Errorf("error getting remote schemas: %s", err)
			return
		}
		// Get DState schemas
		dStateSchemas := util.SchemasFromDesiredStateSchemas(dsServer.Schemas)
		// Diff schemas
		schRemove, schAdd, schModify := util.DiffSchemas(dStateSchemas, dbSchemas)
		// Drop schemas not in DState
		for _, schemaToRemove := range schRemove {

		}
		// Import schemas in DState but not imported
		// Drop + Re-Import all other schemas
	}
}
