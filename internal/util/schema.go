package util

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/elgris/sqrl"

	"github.com/neflyte/fdwctl/internal/database"
	"github.com/neflyte/fdwctl/internal/logger"
	"github.com/neflyte/fdwctl/internal/model"
)

// ensureSchema verifies that a schema with the supplied name exists and if it does not then it will be created
func ensureSchema(ctx context.Context, dbConnection *sql.DB, schemaName string) error {
	log := logger.Log(ctx).
		WithField("function", "ensureSchema")
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
	schemaRows, err := dbConnection.Query(query, args...)
	if err != nil {
		log.Errorf("error checking for schema: %s", err)
		return err
	}
	defer database.CloseRows(ctx, schemaRows)
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
	if schemaRows.Err() != nil {
		log.Errorf("error iterating result rows: %s", schemaRows.Err())
		return schemaRows.Err()
	}
	if !localSchemaExists {
		log.Debug("schema does not exist; creating")
		query = fmt.Sprintf(`CREATE SCHEMA "%s"`, schemaName)
		log.Tracef("query: %s", query)
		_, err = dbConnection.Exec(query)
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

// getEnums returns a list of ENUM types
func getEnums(ctx context.Context, dbConnection *sql.DB) ([]string, error) {
	log := logger.Log(ctx).
		WithField("function", "getEnums")
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
	enumRows, err := dbConnection.Query(query, args...)
	if err != nil {
		log.Errorf("error querying enums: %s", err)
		return nil, err
	}
	defer database.CloseRows(ctx, enumRows)
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
	if enumRows.Err() != nil {
		log.Errorf("error iterating result rows: %s", enumRows.Err())
		return nil, enumRows.Err()
	}
	return enums, nil
}

// getSchemaEnumsUsedInTables returns a list of ENUM types that are used in tables of the specified schema
func getSchemaEnumsUsedInTables(ctx context.Context, dbConnection *sql.DB, schemaName string) ([]string, error) {
	log := logger.Log(ctx).
		WithField("function", "getSchemaEnumsUsedInTables")
	query, args, err := sqrl.
		Select("DISTINCT cuu.udt_name").
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
	enumRows, err := dbConnection.Query(query, args...)
	if err != nil {
		log.Errorf("error querying enums: %s", err)
		return nil, err
	}
	defer database.CloseRows(ctx, enumRows)
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
	if enumRows.Err() != nil {
		log.Errorf("error iterating result rows: %s", enumRows.Err())
		return nil, enumRows.Err()
	}
	return enums, nil
}

// getEnumStrings returns a list of string entries from the specified ENUM type
func getEnumStrings(ctx context.Context, dbConnection *sql.DB, enumType string) ([]string, error) {
	log := logger.Log(ctx).
		WithField("function", "getEnumStrings")
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
	enumRows, err := dbConnection.Query(query, args...)
	if err != nil {
		log.Errorf("error querying enum data: %s", err)
		return nil, err
	}
	defer database.CloseRows(ctx, enumRows)
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
	if enumRows.Err() != nil {
		log.Errorf("error iterating result rows: %s", enumRows.Err())
		return nil, enumRows.Err()
	}
	return enumStrings, nil
}

// GetSchemasForServer returns a list of foreign schemas
func GetSchemasForServer(ctx context.Context, dbConnection *sql.DB, serverName string) ([]model.Schema, error) {
	log := logger.Log(ctx).
		WithField("function", "GetSchemasForServer")
	qbuilder := sqrl.
		Select("DISTINCT ft.foreign_table_schema", "ft.foreign_server_name", "ftos.option_value AS remote_schema").
		From("information_schema.foreign_tables ft").
		Join("information_schema.foreign_table_options ftos ON " +
			"ftos.foreign_table_schema = ft.foreign_table_schema " +
			"AND ftos.foreign_table_catalog = ft.foreign_table_catalog " +
			"AND ftos.foreign_table_name = ft.foreign_table_name " +
			"AND ftos.option_name = 'schema_name'")
	if serverName != "" {
		qbuilder = qbuilder.Where(sqrl.Eq{"ft.foreign_server_name": serverName})
	}
	query, args, _ := qbuilder.PlaceholderFormat(sqrl.Dollar).ToSql()
	log.Tracef("query: %s; args: %#v", query, args)
	schemaRows, err := dbConnection.Query(query, args...)
	if err != nil {
		log.Errorf("error listing schemas: %s", err)
		return nil, err
	}
	defer database.CloseRows(ctx, schemaRows)
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
	if schemaRows.Err() != nil {
		log.Errorf("error iterating result rows: %s", schemaRows.Err())
		return nil, schemaRows.Err()
	}
	return schemas, nil
}

// DiffSchemas takes two lists of schemas and produces a list of schemas that migrate the second list (dbSchemas)
// to equal the first (dStateSchemas). The first list (dStateSchemas) is the desired state; the second list (dbSchemas) is the
// current state. A list of schemas to remove, schemas to add, and schemas to modify are returned.
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

// DropSchema drops a database schema with optional CASCADE
func DropSchema(ctx context.Context, dbConnection *sql.DB, schema model.Schema, cascadeDrop bool) error {
	// TODO: Figure out if it's feasible to also drop the foreign ENUMs as well to make the drop as clean as possible
	log := logger.Log(ctx).
		WithField("function", "DropSchema")
	if schema.LocalSchema == "" {
		return logger.ErrorfAsError(log, "local schema name is required")
	}
	query := fmt.Sprintf(`DROP SCHEMA "%s"`, schema.LocalSchema)
	if cascadeDrop {
		query = fmt.Sprintf("%s CASCADE", query)
	}
	log.Tracef("query: %s", query)
	_, err := dbConnection.Exec(query)
	if err != nil {
		log.Errorf("error dropping schema: %s", err)
		return err
	}
	return nil
}

// importSchemaEnums attempts to create ENUM types locally that represent ENUM types used in the remote schema
func importSchemaEnums(ctx context.Context, dbConnection *sql.DB, schema model.Schema) error {
	log := logger.Log(ctx).
		WithField("function", "importSchemaEnums")
	fdbConnStr := ResolveConnectionString(schema.ENUMConnection, &schema.ENUMSecret)
	fdbConn, err := database.GetConnection(ctx, fdbConnStr)
	if err != nil {
		log.Errorf("error connecting to foreign database: %s", err)
		return err
	}
	defer database.CloseConnection(ctx, fdbConn)
	remoteEnums, err := getSchemaEnumsUsedInTables(ctx, fdbConn, schema.RemoteSchema)
	if err != nil {
		log.Errorf("error getting remote ENUMs: %s", err)
		return err
	}
	sort.Strings(remoteEnums)
	// Get a list of local enums, too
	localEnums, err := getEnums(ctx, dbConnection)
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
		enumStrings, err = getEnumStrings(ctx, fdbConn, remoteEnum)
		if err != nil {
			log.Errorf("error getting enum values: %s", err)
			return err
		}
		query := fmt.Sprintf(`CREATE TYPE "%s" AS ENUM (`, remoteEnum)
		quotedEnumStrings := make([]string, 0)
		for _, enumString := range enumStrings {
			quotedEnumStrings = append(quotedEnumStrings, fmt.Sprintf("'%s'", enumString))
		}
		query = fmt.Sprintf("%s %s )", query, strings.Join(quotedEnumStrings, ","))
		log.Tracef("query: %s", query)
		_, err = dbConnection.Exec(query)
		if err != nil {
			log.Errorf("error creating local enum type: %s", err)
			return err
		}
		log.Infof("enum type %s created", remoteEnum)
	}
	return nil
}

// ImportSchema attempts to import a remote schema from a foreign server into a local schema, optionally importing
// ENUM types used in the remote schema as well.
func ImportSchema(ctx context.Context, dbConnection *sql.DB, serverName string, schema model.Schema) error {
	log := logger.Log(ctx).
		WithField("function", "ImportSchema")
	// Sanity Check
	if serverName == "" {
		return logger.ErrorfAsError(log, "server name is required")
	}
	if schema.ImportENUMs && schema.ENUMConnection == "" {
		return logger.ErrorfAsError(log, "enum database connection string is required when importing enums")
	}
	// Ensure the local schema exists
	err := ensureSchema(ctx, dbConnection, schema.LocalSchema)
	if err != nil {
		log.Errorf("error ensuring local schema exists: %s", err)
		return err
	}
	if schema.ImportENUMs {
		err = importSchemaEnums(ctx, dbConnection, schema)
		if err != nil {
			log.Errorf("error importing foreign enums: %s", err)
			return err
		}
	}
	// TODO: support LIMIT TO and EXCEPT
	sb := new(strings.Builder)
	sb.WriteString(`IMPORT FOREIGN SCHEMA "`)
	sb.WriteString(schema.RemoteSchema)
	sb.WriteString(`" FROM SERVER "`)
	sb.WriteString(serverName)
	sb.WriteString(`" INTO "`)
	sb.WriteString(schema.LocalSchema)
	sb.WriteString(`"`)
	log.Tracef("query: %s", sb.String())
	_, err = dbConnection.Exec(sb.String())
	if err != nil {
		log.Errorf("error importing foreign schema: %s", err)
		return err
	}
	// If there are permissions to configure then configure them
	for _, user := range schema.SchemaGrants.Users {
		log.Debugf("applying grants to schema %s for user %s", schema.LocalSchema, user)
		// GRANT USAGE ON SCHEMA xxxx TO yyyy
		sb := new(strings.Builder)
		sb.WriteString(`GRANT USAGE ON SCHEMA "`)
		sb.WriteString(schema.LocalSchema)
		sb.WriteString(`" TO "`)
		sb.WriteString(user)
		sb.WriteString(`"`)
		query := sb.String()
		log.Tracef("query: %s", query)
		_, err = dbConnection.Exec(query)
		if err != nil {
			log.Errorf("error granting usage to local user: %s", err)
			return err
		}
		// GRANT SELECT ON ALL TABLES IN SCHEMA xxxx TO yyyy
		sb = new(strings.Builder)
		sb.WriteString(`GRANT SELECT ON ALL TABLES IN SCHEMA "`)
		sb.WriteString(schema.LocalSchema)
		sb.WriteString(`" TO "`)
		sb.WriteString(user)
		sb.WriteString(`"`)
		query = sb.String()
		log.Tracef("query: %s", query)
		_, err = dbConnection.Exec(query)
		if err != nil {
			log.Errorf("error granting select to local user: %s", err)
			return err
		}
	}
	return nil
}
