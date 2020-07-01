package util

import (
	"context"
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
)

func EnsureSchema(ctx context.Context, dbConnection *pgx.Conn, schemaName string) error {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "EnsureSchema")
	query, args, err := sqrl.
		Select("1").
		From("information_schema.schemata").
		Where(sqrl.Eq{"schema_name": schemaName}).
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return err
	}
	schemaRows, err := dbConnection.Query(ctx, query, args...)
	if err != nil {
		log.Errorf("error checking for schema: %s", err)
		return err
	}
	defer schemaRows.Close()
	localSchemaExists := false
	if schemaRows.Next() {
		var foo int
		err = schemaRows.Scan(&foo)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			return err
		}
		if foo == 1 {
			localSchemaExists = true
		}
	}
	if !localSchemaExists {
		log.Debug("schema does not exist; creating")
		query := fmt.Sprintf("CREATE SCHEMA %s", schemaName)
		log.Tracef("query: %s", query)
		_, err = dbConnection.Exec(ctx, query)
		if err != nil {
			log.Errorf("error creating schema: %s")
			return err
		}
		log.Infof("local schema %s created", schemaName)
		return nil
	}
	log.Infof("local schema %s exists", schemaName)
	return nil
}
