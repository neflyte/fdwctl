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
