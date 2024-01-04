package util

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/neflyte/fdwctl/lib/database"
	"github.com/neflyte/fdwctl/lib/logger"
	"github.com/neflyte/fdwctl/lib/model"
)

const (
	sqlGetExtensions   = `SELECT extname, extversion FROM pg_extension`
	sqlCreateExtension = `CREATE EXTENSION IF NOT EXISTS "%s"`
	sqlDropExtension   = `DROP EXTENSION IF EXISTS "%s"`
)

// GetExtensions returns a list of installed extensions
func GetExtensions(ctx context.Context, dbConnection *sql.DB) ([]model.Extension, error) {
	log := logger.Log(ctx).
		WithField("function", "GetExtensions")
	exts := make([]model.Extension, 0)
	rows, err := dbConnection.Query(sqlGetExtensions)
	if err != nil {
		log.Errorf("error querying for extensions: %s", err)
		return nil, err
	}
	defer database.CloseRows(ctx, rows)
	for rows.Next() {
		extension := new(model.Extension)
		err = rows.Scan(&extension.Name, &extension.Version)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			return nil, err
		}
		exts = append(exts, *extension)
	}
	if rows.Err() != nil {
		log.Errorf("error iterating result rows: %s", rows.Err())
		return nil, rows.Err()
	}
	return exts, nil
}

// DiffExtensions takes two lists of extensions and produces a list of extensions that migrate the second list (dbExts)
// to equal the first (dStateExts). The first list (dStateExts) is the desired state; the second list (dbExts) is the
// current state. A list of extensions to remove and extensions to add are returned.
func DiffExtensions(dStateExts []model.Extension, dbExts []model.Extension) (extRemove []model.Extension, extAdd []model.Extension) {
	extRemove = make([]model.Extension, 0)
	extAdd = make([]model.Extension, 0)
	// extRemove
	for _, dbExt := range dbExts {
		foundDStateExt := false
		for _, dStateExt := range dStateExts {
			if dStateExt.Equals(dbExt) {
				foundDStateExt = true
				break
			}
		}
		if !foundDStateExt {
			extRemove = append(extRemove, dbExt)
		}
	}
	// extAdd
	for _, dStateExt := range dStateExts {
		foundDBExt := false
		for _, dbExt := range dbExts {
			if dbExt.Equals(dStateExt) {
				foundDBExt = true
				break
			}
		}
		if !foundDBExt {
			extAdd = append(extAdd, dStateExt)
		}
	}
	return
}

// CreateExtension creates a postgres extension in the database
func CreateExtension(ctx context.Context, dbConnection *sql.DB, ext model.Extension) error {
	log := logger.Log(ctx).
		WithField("function", "CreateExtension")
	_, err := dbConnection.Exec(fmt.Sprintf(sqlCreateExtension, ext.Name))
	if err != nil {
		return logger.ErrorfAsError(log, "error creating extension %s: %s", ext.Name, err)
	}
	return nil
}

// DropExtension drops a postgres extension from the database
func DropExtension(ctx context.Context, dbConnection *sql.DB, ext model.Extension) error {
	log := logger.Log(ctx).
		WithField("function", "DropExtension")
	_, err := dbConnection.Exec(fmt.Sprintf(sqlDropExtension, ext.Name))
	if err != nil {
		return logger.ErrorfAsError(log, "error dropping extension %s: %s", ext.Name, err)
	}
	return nil
}
