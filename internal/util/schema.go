package util

import (
	"context"
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
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
	log.Tracef("query: %s, args: %#v", query, args)
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
			log.Errorf("error creating schema: %s", err)
			return err
		}
		log.Infof("local schema %s created", schemaName)
		return nil
	}
	log.Infof("local schema %s exists", schemaName)
	return nil
}

func GetEnums(ctx context.Context, dbConnection *pgx.Conn) ([]string, error) {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "GetEnums")
	query, args, err := sqrl.
		Select("typname").
		From("pg_type").
		Where(sqrl.Eq{"typtype": "e"}).
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return nil, err
	}
	log.Tracef("query: %s, args: %#v", query, args)
	enumRows, err := dbConnection.Query(ctx, query, args...)
	if err != nil {
		log.Errorf("error querying enums: %s", err)
		return nil, err
	}
	defer enumRows.Close()
	enums := make([]string, 0)
	var enumName string
	for enumRows.Next() {
		err = enumRows.Scan(&enumName)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		enums = append(enums, enumName)
	}
	return enums, nil
}

func GetSchemaEnumsUsedInTables(ctx context.Context, dbConnection *pgx.Conn, schemaName string) ([]string, error) {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "GetSchemaEnumsUsedInTables")
	query, args, err := sqrl.
		Select("cuu.udt_name").
		From("information_schema.column_udt_usage cuu").
		Join("pg_type t ON t.typname = cuu.udt_name").
		Where(sqrl.And{
			sqrl.Eq{"t.typtype": "e"},
			sqrl.Eq{"cuu.table_schema": schemaName},
		}).
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return nil, err
	}
	log.Tracef("query: %s, args: %#v", query, args)
	enumRows, err := dbConnection.Query(ctx, query, args...)
	if err != nil {
		log.Errorf("error querying enums: %s", err)
		return nil, err
	}
	defer enumRows.Close()
	enums := make([]string, 0)
	var enumName string
	for enumRows.Next() {
		err = enumRows.Scan(&enumName)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		enums = append(enums, enumName)
	}
	return enums, nil
}

func GetEnumStrings(ctx context.Context, dbConnection *pgx.Conn, enumType string) ([]string, error) {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "GetEnumStrings")
	query, args, err := sqrl.
		Select("e.enumlabel").
		From("pg_type t").
		Join("pg_enum e ON t.oid = e.enumtypid").
		Where(sqrl.Eq{"t.typname": enumType}).
		OrderBy("e.enumsortorder ASC").
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	if err != nil {
		log.Errorf("error creating query: %s", err)
		return nil, err
	}
	log.Tracef("query: %s, args: %#v", query, args)
	enumRows, err := dbConnection.Query(ctx, query, args...)
	if err != nil {
		log.Errorf("error querying enum data: %s", err)
		return nil, err
	}
	defer enumRows.Close()
	enumStrings := make([]string, 0)
	var enumString string
	for enumRows.Next() {
		err = enumRows.Scan(&enumString)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			continue
		}
		enumStrings = append(enumStrings, enumString)
	}
	return enumStrings, nil
}

func GetSchemas(ctx context.Context, dbConnection *pgx.Conn) ([]model.Schema, error) {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "GetSchemas")
	query, _, _ := sqrl.
		Select("DISTINCT ft.foreign_table_schema", "ft.foreign_server_name", "ftos.option_value AS remote_schema").
		From("information_schema.foreign_tables ft").
		Join("information_schema.foreign_table_options ftos ON " +
			"ftos.foreign_table_schema = ft.foreign_table_schema " +
			"AND ftos.foreign_table_catalog = ft.foreign_table_catalog " +
			"AND ftos.foreign_table_name = ft.foreign_table_name " +
			"AND ftos.option_name = 'schema_name'").
		PlaceholderFormat(sqrl.Dollar).
		ToSql()
	log.Tracef("query: %s", query)
	schemaRows, err := dbConnection.Query(ctx, query)
	if err != nil {
		log.Errorf("error listing schemas: %s", err)
		return nil, err
	}
	defer schemaRows.Close()
	schemas := make([]model.Schema, 0)
	var schemaName, foreignServer, remoteSchema string
	for schemaRows.Next() {
		err = schemaRows.Scan(&schemaName, &foreignServer, &remoteSchema)
		if err != nil {
			log.Errorf("error scanning result row: %s", err)
			return nil, err
		}
		schemas = append(schemas, model.Schema{
			LocalSchema:  schemaName,
			RemoteSchema: remoteSchema,
		})
	}
	return schemas, nil
}

func DiffSchemas(dStateSchemas []model.Schema, dbSchemas []model.Schema) (schRemove []model.Schema, schAdd []model.Schema, schModify []model.Schema) {
	schRemove = make([]model.Schema, 0)
	schAdd = make([]model.Schema, 0)
	schModify = make([]model.Schema, 0)
	// schRemove
	for _, dbSchema := range dbSchemas {
		foundDBSchema := false
		for _, dStateSchema := range dStateSchemas {
			if dbSchema.LocalSchema == dStateSchema.LocalSchema {
				foundDBSchema = true
				break
			}
		}
		if !foundDBSchema {
			schRemove = append(schRemove, dbSchema)
		}
	}
	// schAdd + achModify
	for _, dStateSchema := range dStateSchemas {
		foundDStateSchema := false
		for _, dbSchema := range dbSchemas {
			if dStateSchema.LocalSchema == dbSchema.LocalSchema {
				foundDStateSchema = true
				break
			}
		}
		if !foundDStateSchema {
			schAdd = append(schAdd, dStateSchema)
		} else {
			schModify = append(schModify, dStateSchema)
		}
	}
	return
}
