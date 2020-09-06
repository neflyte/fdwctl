package cmd

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/neflyte/fdwctl/internal/config"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"github.com/neflyte/fdwctl/internal/util"
	"github.com/spf13/cobra"
)

var (
	desiredStateCmd = &cobra.Command{
		Use:               "apply",
		Short:             "Apply a desired state",
		Long:              "Apply a desired state configuration to the FDW database",
		PersistentPreRunE: preDoDesiredState,
		PersistentPostRun: postDoDesiredState,
		Run:               doDesiredState,
	}
)

func preDoDesiredState(cmd *cobra.Command, _ []string) error {
	var err error

	log := logger.Log(cmd.Context()).
		WithField("function", "preDoDesiredState")
	dbConnection, err = database.GetConnection(cmd.Context(), config.Instance().GetDatabaseConnectionString())
	if err != nil {
		return logger.ErrorfAsError(log, "error getting database connection: %s", err)
	}
	return nil
}

func postDoDesiredState(cmd *cobra.Command, _ []string) {
	database.CloseConnection(cmd.Context(), dbConnection)
}

func doDesiredState(cmd *cobra.Command, _ []string) {
	log := logger.Log(cmd.Context()).
		WithField("function", "doDesiredState")
	// Apply Extensions
	err := applyExtensions(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error applying extensions: %s", err)
		return
	}
	// Convert DState servers to ForeignServers
	dStateServers := config.Instance().DesiredState.Servers
	// List servers in DB
	dbServers, err := util.GetServers(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error getting foreign servers: %s", err)
		return
	}
	// Diff servers
	serversInDBButNotInDState, serversInDStateButNotInDB, serversAlreadyInDB := util.DiffForeignServers(dStateServers, dbServers)
	// Remove servers in DB but not in DState
	for _, serverNotInDState := range serversInDBButNotInDState {
		err = util.DropServer(cmd.Context(), dbConnection, serverNotInDState.Name, true)
		if err != nil {
			log.Errorf("error dropping server %s that is not in desired state: %s", serverNotInDState.Name, err)
			return
		}
		log.Infof("server %s dropped", serverNotInDState.Name)
	}
	// Create servers that are in DState but not yet in DB
	for _, serverNotInDB := range serversInDStateButNotInDB {
		err = util.CreateServer(cmd.Context(), dbConnection, serverNotInDB)
		if err != nil {
			log.Errorf("error creating server: %s", err)
			return
		}
		log.Infof("server %s created", serverNotInDB.Name)
	}
	// Update servers that were already in the DB
	for _, serverAlreadyInDB := range serversAlreadyInDB {
		dsServer := util.FindForeignServer(config.Instance().DesiredState.Servers, serverAlreadyInDB.Name)
		if dsServer == nil {
			log.Errorf("cannot find desired state server %s", serverAlreadyInDB.Name)
			return
		}
		dbServer := *dsServer
		if !dbServer.Equals(serverAlreadyInDB) {
			err = util.UpdateServer(cmd.Context(), dbConnection, serverAlreadyInDB)
			if err != nil {
				log.Errorf("error updating server: %s", err)
				return
			}
			log.Infof("server %s updated", serverAlreadyInDB.Name)
		} else {
			log.Debugf("server %s is no different from the database; skipping it", serverAlreadyInDB.Name)
		}
	}
	// Collect a list of servers to process UserMap and Schemas for
	serversToProcess := make([]model.ForeignServer, 0)
	serversToProcess = append(serversToProcess, serversInDStateButNotInDB...)
	serversToProcess = append(serversToProcess, serversAlreadyInDB...)
	// Process UserMaps and Schemas
	for _, serverAlreadyInDB := range serversToProcess {
		err = applyUserMaps(cmd.Context(), dbConnection, serverAlreadyInDB)
		if err != nil {
			log.Errorf("error applying usermaps for server %s: %s", serverAlreadyInDB.Name, err)
			return
		}
		err = applySchemas(cmd.Context(), dbConnection, serverAlreadyInDB)
		if err != nil {
			log.Errorf("error applying schemas for server %s: %s", serverAlreadyInDB.Name, err)
			return
		}
	}
	log.Info("desired state applied.")
}

func applyExtensions(ctx context.Context, dbConnection *sqlx.DB) error {
	log := logger.Log(ctx).
		WithField("function", "applyExtensions")
	// List extensions in DB
	dbExts, err := util.GetExtensions(ctx, dbConnection)
	if err != nil {
		log.Errorf("error getting extensions: %s", err)
		return err
	}
	// Diff extensions
	_, extAdd := util.DiffExtensions(config.Instance().DesiredState.Extensions, dbExts)
	// NOTE: Don't remove extensions with abandon since we might remove something that's needed
	/*// Remove extensions
	for _, extToRemove := range extRemove {
		err = util.DropExtension(ctx, dbConnection, extToRemove)
		if err != nil {
			return logger.ErrorfAsError(log, "error dropping extension: %s", err)
		}
		log.Infof("extension %s dropped", extToRemove.Name)
	}*/
	// Add extensions
	for _, extToAdd := range extAdd {
		err = util.CreateExtension(ctx, dbConnection, extToAdd)
		if err != nil {
			return logger.ErrorfAsError(log, "error creating extension: %s", err)
		}
		log.Infof("extension %s created", extToAdd.Name)
	}
	return nil
}

func applyUserMaps(ctx context.Context, dbConnection *sqlx.DB, server model.ForeignServer) error {
	log := logger.Log(ctx).
		WithField("function", "applyUserMaps")
	// List Usermaps for this server in the DB
	dbServerUsermaps, err := util.GetUserMapsForServer(ctx, dbConnection, server.Name)
	if err != nil {
		log.Errorf("error getting usermaps for server %s: %s", server.Name, err)
		return err
	}
	// List Usermaps for this server in DState
	dsServer := util.FindForeignServer(config.Instance().DesiredState.Servers, server.Name)
	if dsServer == nil {
		return logger.ErrorfAsError(log, "cannot find desired state server %s; THIS IS UNEXPECTED", server.Name)
	}
	dStateServerUsermaps := dsServer.UserMaps
	// Diff usermaps
	usRemove, usAdd, usModify := util.DiffUserMaps(dStateServerUsermaps, dbServerUsermaps)
	// Delete Usermaps not in DState
	for _, usermapToRemove := range usRemove {
		usermapToRemove.ServerName = dsServer.Name
		err = util.DropUserMap(ctx, dbConnection, usermapToRemove, true)
		if err != nil {
			log.Errorf("error dropping user map for local user %s: %s", usermapToRemove.LocalUser, err)
			return err
		}
		log.Infof("user mapping for %s dropped", usermapToRemove.LocalUser)
	}
	// Add Usermaps in DState but not in DB
	for _, usermapToAdd := range usAdd {
		usermapToAdd.ServerName = dsServer.Name
		err = util.CreateUserMap(ctx, dbConnection, usermapToAdd)
		if err != nil {
			log.Errorf("error creating user map for local user %s: %s", usermapToAdd.LocalUser, err)
			return err
		}
		log.Infof("user mapping %s -> %s created", usermapToAdd.LocalUser, usermapToAdd.RemoteUser)
	}
	// Update usermaps that are already there
	for _, usermapToUpdate := range usModify {
		dbUserMap := util.FindUserMap(dbServerUsermaps, usermapToUpdate.LocalUser)
		if dbUserMap == nil {
			return logger.ErrorfAsError(log, "cannot find user mapping for local user %s", usermapToUpdate.LocalUser)
		}
		if !usermapToUpdate.Equals(*dbUserMap) {
			err = util.UpdateUserMap(ctx, dbConnection, usermapToUpdate)
			if err != nil {
				log.Errorf("error updating user map for local user %s: %s", usermapToUpdate.LocalUser, err)
				return err
			}
			log.Infof("user mapping %s -> %s updated", usermapToUpdate.LocalUser, usermapToUpdate.RemoteUser)
		} else {
			log.Debugf("user mapping %s -> %s is no different from the database; skipping it", usermapToUpdate.LocalUser, usermapToUpdate.RemoteUser)
		}
	}
	return nil
}

func applySchemas(ctx context.Context, dbConnection *sqlx.DB, server model.ForeignServer) error {
	log := logger.Log(ctx).
		WithField("function", "applySchemas")
	// Get DB remote schemas
	dbSchemas, err := util.GetSchemas(ctx, dbConnection)
	if err != nil {
		log.Errorf("error getting remote schemas: %s", err)
		return err
	}
	// Get DState schemas
	dStateSchemas := server.Schemas
	// Diff schemas
	schRemove, schAdd, schModify := util.DiffSchemas(dStateSchemas, dbSchemas)
	// Drop schemas not in DState
	for _, schemaToRemove := range schRemove {
		err = util.DropSchema(ctx, dbConnection, schemaToRemove, true)
		if err != nil {
			log.Errorf("error dropping local schema %s: %s", schemaToRemove.LocalSchema, err)
			return err
		}
		log.Infof("local schema %s dropped", schemaToRemove.LocalSchema)
	}
	// Import schemas in DState but not imported
	for _, schemaToAdd := range schAdd {
		err = util.ImportSchema(ctx, dbConnection, server.Name, schemaToAdd)
		if err != nil {
			log.Errorf("error importing into local schema %s: %s", schemaToAdd.LocalSchema, err)
			return err
		}
		log.Infof("foreign schema %s imported", schemaToAdd.RemoteSchema)
	}
	// Drop + Re-Import all other schemas
	for _, schemaToModify := range schModify {
		// Drop
		err = util.DropSchema(ctx, dbConnection, schemaToModify, true)
		if err != nil {
			log.Errorf("error dropping local schema %s: %s", schemaToModify.LocalSchema, err)
			return err
		}
		// Import
		err = util.ImportSchema(ctx, dbConnection, server.Name, schemaToModify)
		if err != nil {
			log.Errorf("error importing into local schema %s: %s", schemaToModify.LocalSchema, err)
			return err
		}
		// Done
		log.Infof("foreign schema %s re-imported", schemaToModify.RemoteSchema)
	}
	return nil
}
