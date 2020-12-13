package util

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/elgris/sqrl"

	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
)

// GetExtensions returns a list of installed extensions
func GetExtensions(ctx context.Context, dbConnection *sql.DB) ([]model.Extension, error) {
	log := logger.Log(ctx).
		WithField("function", "GetExtensions")
	exts := make([]model.Extension, 0)
	query, _, err := sqrl.
		Select("extname", "extversion").
		From("pg_extension").
		ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return nil, err
	}
	rows, err := dbConnection.Query(query)
	if err != nil {
		log.Errorf("error querying for extensions: %s", err)
		return nil, err
	}
	defer database.CloseRows(ctx, rows)
	var extname, extversion string
	for rows.Next() {
		err = rows.Scan(&extname, &extversion)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			return nil, err
		}
		exts = append(exts, model.Extension{
			Name:    extname,
			Version: extversion,
		})
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
	_, err := dbConnection.Exec(fmt.Sprintf(`CREATE EXTENSION IF NOT EXISTS %s`, ext.Name))
	if err != nil {
		return logger.ErrorfAsError(log, "error creating extension %s: %s", ext.Name, err)
	}
	return nil
}

// DropExtension drops a postgres extension from the database
func DropExtension(ctx context.Context, dbConnection *sql.DB, ext model.Extension) error {
	log := logger.Log(ctx).
		WithField("function", "DropExtension")
	_, err := dbConnection.Exec(fmt.Sprintf(`DROP EXTENSION IF EXISTS %s`, ext.Name))
	if err != nil {
		return logger.ErrorfAsError(log, "error dropping extension %s: %s", ext.Name, err)
	}
	return nil
}
