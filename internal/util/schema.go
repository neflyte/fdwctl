package util

import (
	"context"
	"fmt"
	"github.com/elgris/sqrl"
	"github.com/jackc/pgx/v4"
	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
	"sort"
	"strings"
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
		query = fmt.Sprintf("CREATE SCHEMA %s", schemaName)
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
			ServerName:   foreignServer,
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

func DropSchema(ctx context.Context, dbConnection *pgx.Conn, schema model.Schema, cascadeDrop bool) error {
	// TODO: Figure out if it's feasible to also drop the foreign ENUMs as well to make the drop as clean as possible
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "DropSchema")
	if schema.LocalSchema == "" {
		return logger.ErrorfAsError(log, "local schema name is required")
	}
	query := fmt.Sprintf("DROP SCHEMA %s", schema.LocalSchema)
	if cascadeDrop {
		query = fmt.Sprintf("%s CASCADE", query)
	}
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(ctx, query)
	if err != nil {
		log.Errorf("error dropping schema: %s", err)
		return err
	}
	return nil
}

func ImportSchemaEnums(ctx context.Context, dbConnection *pgx.Conn, schema model.Schema) error {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "ImportSchemaEnums")
	fdbConnStr := ResolveConnectionString(schema.ENUMConnection, &schema.ENUMSecret)
	fdbConn, err := database.GetConnection(ctx, fdbConnStr)
	if err != nil {
		log.Errorf("error connecting to foreign database: %s", err)
		return err
	}
	defer database.CloseConnection(ctx, fdbConn)
	remoteEnums, err := GetSchemaEnumsUsedInTables(ctx, fdbConn, schema.RemoteSchema)
	if err != nil {
		log.Errorf("error getting remote ENUMs: %s", err)
		return err
	}
	sort.Strings(remoteEnums)
	// Get a list of local enums, too
	localEnums, err := GetEnums(ctx, dbConnection)
	if err != nil {
		log.Errorf("error getting local ENUMs: %s", err)
		return err
	}
	sort.Strings(localEnums)
	// Get enough data from remote database to re-create enums and then create them
	for _, remoteEnum := range remoteEnums {
		if sort.SearchStrings(localEnums, remoteEnum) != len(localEnums) {
			log.Debugf("local enum %s exists", remoteEnum)
			continue
		}
		var enumStrings []string
		enumStrings, err = GetEnumStrings(ctx, fdbConn, remoteEnum)
		if err != nil {
			log.Errorf("error getting enum values: %s", err)
			return err
		}
		query := fmt.Sprintf("CREATE TYPE %s AS ENUM (", remoteEnum)
		quotedEnumStrings := make([]string, 0)
		for _, enumString := range enumStrings {
			quotedEnumStrings = append(quotedEnumStrings, fmt.Sprintf("'%s'", enumString))
		}
		query = fmt.Sprintf("%s %s )", query, strings.Join(quotedEnumStrings, ","))
		log.Tracef("query: %s", query)
		_, err = dbConnection.Exec(ctx, query)
		if err != nil {
			log.Errorf("error creating local enum type: %s", err)
			return err
		}
		log.Infof("enum type %s created", remoteEnum)
	}
	return nil
}

func ImportSchema(ctx context.Context, dbConnection *pgx.Conn, serverName string, schema model.Schema) error {
	log := logger.Root().
		WithContext(ctx).
		WithField("function", "ImportSchema")
	// Sanity Check
	if serverName == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	if schema.ImportENUMs && schema.ENUMConnection == "" {
		return logger.ErrorfAsError(log, "enum database connection string is required when importing enums")
	}
	// Ensure the local schema exists
	err := EnsureSchema(ctx, dbConnection, schema.LocalSchema)
	if err != nil {
		log.Errorf("error ensuring local schema exists: %s", err)
		return err
	}
	if schema.ImportENUMs {
		err = ImportSchemaEnums(ctx, dbConnection, schema)
		if err != nil {
			log.Errorf("error importing foreign enums: %s", err)
			return err
		}
	}
	// TODO: support LIMIT TO and EXCEPT
	query := fmt.Sprintf("IMPORT FOREIGN SCHEMA %s FROM SERVER %s INTO %s", schema.RemoteSchema, serverName, schema.LocalSchema)
	log.Tracef("query: %s", query)
	_, err = dbConnection.Exec(ctx, query)
	if err != nil {
		log.Errorf("error importing foreign schema: %s", err)
		return err
	}
	return nil
}
