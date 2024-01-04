package cmd

import (
	"context"
	"database/sql"

	"github.com/neflyte/fdwctl/lib/config"
	"github.com/neflyte/fdwctl/lib/database"
	"github.com/neflyte/fdwctl/lib/logger"
	"github.com/neflyte/fdwctl/lib/model"
	"github.com/neflyte/fdwctl/lib/util"
	"github.com/spf13/cobra"
)

var (
	desiredStateCmd = &cobra.Command{
		Use:               "apply",
		Short:             "Apply a desired state",
		Long:              "Apply a desired state configuration to the FDW database",
		PersistentPreRunE: preDoDesiredState,
		PersistentPostRun: postDoDesiredState,
		RunE:              doDesiredState,
	}
	desiredStateRecreateSchemas = false
)

func init() {
	desiredStateCmd.Flags().BoolVar(&desiredStateRecreateSchemas, "recreateschemas", false, "flag indicating that foreign schemas should be re-created")
}

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

func doDesiredState(cmd *cobra.Command, _ []string) error {
	log := logger.Log(cmd.Context()).
		WithField("function", "doDesiredState")
	// Apply Extensions
	err := applyExtensions(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error applying extensions: %s", err)
		return err
	}
	// Convert DState servers to ForeignServers
	dStateServers := config.Instance().DesiredState.Servers
	// List servers in DB
	dbServers, err := util.GetServers(cmd.Context(), dbConnection)
	if err != nil {
		log.Errorf("error getting foreign servers: %s", err)
		return err
	}
	// Diff servers
	serversInDBButNotInDState, serversInDStateButNotInDB, serversAlreadyInDB := util.DiffForeignServers(dStateServers, dbServers)
	log.Tracef(
		"serversInDBButNotInDState: %#v, serversInDStateButNotInDB: %#v, serversAlreadyInDB: %#v",
		serversInDBButNotInDState,
		serversInDStateButNotInDB,
		serversAlreadyInDB,
	)
	// Remove servers in DB but not in DState
	for _, serverNotInDState := range serversInDBButNotInDState {
		log.Debugf("removing server %s", serverNotInDState.Name)
		err = util.DropServer(cmd.Context(), dbConnection, serverNotInDState.Name, true)
		if err != nil {
			log.Errorf("error dropping server %s that is not in desired state: %s", serverNotInDState.Name, err)
			return err
		}
		log.Infof("server %s dropped", serverNotInDState.Name)
	}
	// Create servers that are in DState but not yet in DB
	for _, serverNotInDB := range serversInDStateButNotInDB {
		log.Debugf("creating server %s", serverNotInDB.Name)
		err = util.CreateServer(cmd.Context(), dbConnection, serverNotInDB)
		if err != nil {
			log.Errorf("error creating server: %s", err)
			return err
		}
		log.Infof("server %s created", serverNotInDB.Name)
	}
	// Update servers that were already in the DB
	for _, serverAlreadyInDB := range serversAlreadyInDB {
		log.Debugf("updating server %s", serverAlreadyInDB.Name)
		dsServer := util.FindForeignServer(config.Instance().DesiredState.Servers, serverAlreadyInDB.Name)
		if dsServer == nil {
			log.Errorf("cannot find desired state server %s", serverAlreadyInDB.Name)
			return err
		}
		dbServer := *dsServer
		if !dbServer.Equals(serverAlreadyInDB) {
			err = util.UpdateServer(cmd.Context(), dbConnection, serverAlreadyInDB)
			if err != nil {
				log.Errorf("error updating server: %s", err)
				return err
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
	for _, serverToProcess := range serversToProcess {
		log.Debugf("applying user maps to %s", serverToProcess.Name)
		err = applyUserMaps(cmd.Context(), dbConnection, serverToProcess)
		if err != nil {
			log.Errorf("error applying usermaps for server %s: %s", serverToProcess.Name, err)
			return err
		}
		log.Debugf("applying schemas to %s", serverToProcess.Name)
		err = applySchemas(cmd.Context(), dbConnection, serverToProcess)
		if err != nil {
			log.Errorf("error applying schemas for server %s: %s", serverToProcess.Name, err)
			return err
		}
	}
	log.Info("desired state applied.")
	return nil
}

func applyExtensions(ctx context.Context, dbConnection *sql.DB) error {
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
		log.Debugf("creating extension %s", extToAdd.Name)
		err = util.CreateExtension(ctx, dbConnection, extToAdd)
		if err != nil {
			return logger.ErrorfAsError(log, "error creating extension: %s", err)
		}
		log.Infof("extension %s created", extToAdd.Name)
	}
	return nil
}

func applyUserMaps(ctx context.Context, dbConnection *sql.DB, server model.ForeignServer) error {
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
		log.Debugf("removing usermap for local user %s", usermapToRemove.LocalUser)
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
		log.Debugf("adding usermap for local user %s", usermapToAdd.LocalUser)
		err = util.CreateUserMap(ctx, dbConnection, usermapToAdd)
		if err != nil {
			log.Errorf("error creating user map for local user %s: %s", usermapToAdd.LocalUser, err)
			return err
		}
		log.Infof("user mapping %s -> %s created", usermapToAdd.LocalUser, usermapToAdd.RemoteUser)
	}
	// Update usermaps that are already there
	for _, usermapToUpdate := range usModify {
		usermapToUpdate.ServerName = dsServer.Name
		log.Debugf("updating usermap for local user %s", usermapToUpdate.LocalUser)
		dbUserMap := util.FindUserMap(dbServerUsermaps, usermapToUpdate.LocalUser)
		if dbUserMap == nil {
			return logger.ErrorfAsError(log, "cannot find user mapping for local user %s", usermapToUpdate.LocalUser)
		}
		// if util.SecretIsDefined(usermapToUpdate.RemoteSecret) {
		if usermapToUpdate.RemoteSecret.IsDefined() {
			remoteSecret := ""
			remoteSecret, err = util.GetSecret(ctx, usermapToUpdate.RemoteSecret)
			if err != nil {
				return logger.ErrorfAsError(log, "error getting secret value: %s", err)
			}
			usermapToUpdate.RemoteSecret.Value = remoteSecret
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

func applySchemas(ctx context.Context, dbConnection *sql.DB, server model.ForeignServer) error {
	log := logger.Log(ctx).
		WithField("function", "applySchemas")
	// Get DB remote schemas
	dbSchemas, err := util.GetSchemasForServer(ctx, dbConnection, server.Name)
	if err != nil {
		log.Errorf("error getting remote schemas: %s", err)
		return err
	}
	// Get DState schemas
	dStateSchemas := server.Schemas
	// Diff schemas
	schRemove, schAdd, schModify := util.DiffSchemas(dStateSchemas, dbSchemas)
	log.Tracef("schRemove: %#v, schAdd: %#v, schModify: %#v", schRemove, schAdd, schModify)
	// Drop schemas not in DState
	for _, schemaToRemove := range schRemove {
		log.Debugf("removing schema %s", schemaToRemove.LocalSchema)
		err = util.DropSchema(ctx, dbConnection, schemaToRemove, true)
		if err != nil {
			log.Errorf("error dropping local schema %s: %s", schemaToRemove.LocalSchema, err)
			return err
		}
		log.Infof("local schema %s dropped", schemaToRemove.LocalSchema)
	}
	// Import schemas in DState but not imported
	for _, schemaToAdd := range schAdd {
		log.Debugf("adding schema %s", schemaToAdd.LocalSchema)
		err = util.ImportSchema(ctx, dbConnection, server.Name, schemaToAdd)
		if err != nil {
			log.Errorf("error importing into local schema %s: %s", schemaToAdd.LocalSchema, err)
			return err
		}
		log.Infof("foreign schema %s imported", schemaToAdd.RemoteSchema)
	}
	// Drop + Re-Import all other schemas
	for _, schemaToModify := range schModify {
		log.Debugf("modifying schema %s", schemaToModify.LocalSchema)
		if desiredStateRecreateSchemas {
			// Drop
			log.Debugf("recreating schema %s (drop)", schemaToModify.LocalSchema)
			err = util.DropSchema(ctx, dbConnection, schemaToModify, true)
			if err != nil {
				log.Errorf("error dropping local schema %s: %s", schemaToModify.LocalSchema, err)
				return err
			}
			// Import
			log.Debugf("recreating schema %s (import)", schemaToModify.LocalSchema)
			err = util.ImportSchema(ctx, dbConnection, server.Name, schemaToModify)
			if err != nil {
				log.Errorf("error importing into local schema %s: %s", schemaToModify.LocalSchema, err)
				return err
			}
			// Done
			log.Infof("foreign schema %s re-imported", schemaToModify.RemoteSchema)
		} else {
			log.Infof("foreign schema %s exists; will not re-create it", schemaToModify.RemoteSchema)
		}
	}
	return nil
}
