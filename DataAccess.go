package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

// select constraint_column_usage.table_catalog as ftable_catalog,
//        constraint_column_usage.table_schema as ftable_schema,
//        constraint_column_usage.table_name  as ftable_name,
//        constraint_column_usage.column_name  as fcolumn_name,
//        constraint_column_usage.constraint_name
//   from information_schema.constraint_column_usage, information_schema.referential_constraints
//   where constraint_column_usage.constraint_catalog = referential_constraints.constraint_catalog and
//         constraint_column_usage.constraint_schema = referential_constraints.constraint_schema and
//         constraint_column_usage.constraint_name = referential_constraints.constraint_name;

// select *
//   from information_schema.key_column_usage, information_schema.referential_constraints
//   where key_column_usage.constraint_catalog = referential_constraints.constraint_catalog and
//         key_column_usage.constraint_schema = referential_constraints.constraint_schema and
//         key_column_usage.constraint_name = referential_constraints.constraint_name and
//         key_column_usage.table_catalog = 'tpt_models_test' and
//         key_column_usage.table_schema = 'public' and
//         key_column_usage.table_name = 'tpt_tree_nodes';

// Table entity in table `information_schema.tables`
type Table struct {
	Schema        string
	TableName     string
	ClassName     string
	IsView        bool
	Columns       []Column
	IsCombinedKey bool
	PrimaryKey    []Column
	HasCreatedAt  bool
	HasUpdatedAt  bool
}

// Column entity in table `information_schema.columns`
type Column struct {
	DbName string
	GoName string
	DbType string
	GoType string

	IsNullable   bool
	IsPrimaryKey bool
	IsForeignKey bool
	IsSequence   bool
}

type dbBase struct {
	dbDrv     string
	dbURL     string
	dbCatalog string
	dbSchema  string
	dbPrefix  string
}

func (cmd *dbBase) initFlags(fs *flag.FlagSet) *flag.FlagSet {
	flag.StringVar(&cmd.dbURL, "db_url", "host=127.0.0.1 port=35432 dbname=tpt user=tpt password=extreme sslmode=disable", "the db url")
	flag.StringVar(&cmd.dbDrv, "db_drv", "postgres", "the db driver")
	flag.StringVar(&cmd.dbCatalog, "db_catalog", "tpt", "the db schema")
	flag.StringVar(&cmd.dbSchema, "db_schema", "public", "the db schema")
	flag.StringVar(&cmd.dbPrefix, "db_prefix", "tpt_", "the db prefix name")
	return fs
}

// GetAll use to select all tables from `information_schema.tables`.
func (cmd *dbBase) GetAllTables() ([]Table, error) {
	db, e := sql.Open(cmd.dbDrv, cmd.dbURL)
	if nil != e {
		return nil, e
	}
	defer db.Close()

	queryString := fmt.Sprintf(`SELECT
            distinct t.table_name, t.table_schema, t.table_type
        FROM
            information_schema.tables t
        LEFT JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
             ON tc.table_catalog = t.table_catalog
             AND tc.table_schema = t.table_schema
             AND tc.table_name = t.table_name
             AND tc.constraint_type = 'PRIMARY KEY'
        LEFT JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
             ON kcu.table_catalog = tc.table_catalog
             AND kcu.table_schema = tc.table_schema
             AND kcu.table_name = tc.table_name
             AND kcu.constraint_name = tc.constraint_name
        WHERE
            t.table_catalog = '%s' AND
            t.table_schema = '%s'`, cmd.dbCatalog, cmd.dbSchema)

	rows, e := db.Query(queryString)
	if nil != e {
		return nil, e
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		var tableType string
		if e := rows.Scan(&table.TableName, &table.Schema, &tableType); nil != e {
			return nil, e
		}

		if "view" == strings.ToLower(tableType) {
			table.IsView = true
		}

		columns, e := cmd.getByTable(db, cmd.dbCatalog, cmd.dbSchema, table.TableName)
		if nil != e {
			return nil, errors.New("failed to read columns for " + table.TableName + " - " + e.Error())
		}

		table.Columns = columns
		table.IsCombinedKey, table.PrimaryKey = getPrimaryKey(table.Columns)
		table.ClassName = Typeify(strings.TrimPrefix(table.TableName, cmd.dbPrefix))

		//if "tpt_network_devices" == table.TableName {
		//	fmt.Println(table.TableName, table.IsCombinedKey, table.PrimaryKey)
		//}
		for _, column := range columns {
			if "created_at" == column.DbName {
				table.HasCreatedAt = true
			}
			if "updated_at" == column.DbName {
				table.HasUpdatedAt = true
			}
		}

		tables = append(tables, table)
	}
	return tables, rows.Err()
}

func (cmd *dbBase) isForeignKey(db *sql.DB, tableCatalog, tableSchema, tableName, columnName string) (bool, error) {
	queryString := fmt.Sprintf(`SELECT count(*)
    FROM
        INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
      LEFT JOIN
        INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
      ON
        kcu.table_schema = tc.table_schema
        AND kcu.table_name = tc.table_name
        AND kcu.constraint_name = tc.constraint_name
    WHERE 
        tc.constraint_type = 'FOREIGN KEY'
        AND kcu.table_catalog = '%s'
        AND kcu.table_schema = '%s'
        AND kcu.table_name = '%s' 
        AND kcu.column_name = '%s'`, tableCatalog, tableSchema, tableName, columnName)

	var count int
	e := db.QueryRow(queryString).Scan(&count)
	if nil != e {
		return false, e
	}
	return count > 0, nil
}

// getByTable use to select columns from `information_schema.tables` of inputed tableName.
func (cmd *dbBase) getByTable(db *sql.DB, tableCatalog, tableSchema, tableName string) ([]Column, error) {
	queryString := fmt.Sprintf(`SELECT
        distinct t.column_name,
        t.is_nullable,
        t.udt_name,
        t.column_name = kcu.column_name as primary_key,
        t.column_default IS NOT NULL AND t.column_default LIKE 'nextval%%' as is_sequence
    FROM
        INFORMATION_SCHEMA.COLUMNS t
    LEFT JOIN
        INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
    ON
        tc.table_schema = t.table_schema
        AND tc.table_name = t.table_name
        AND tc.constraint_type = 'PRIMARY KEY'
    LEFT JOIN
        INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
    ON
        kcu.table_schema = tc.table_schema
        AND kcu.table_name = tc.table_name
        AND kcu.constraint_name = tc.constraint_name
    WHERE t.table_catalog = '%s' and t.table_schema = '%s' and t.table_name = '%s'`, tableCatalog, tableSchema, tableName)
	rows, e := db.Query(queryString)
	if nil != e {
		return nil, e
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var isNullable sql.NullString
		var primaryKey sql.NullBool
		var isSequence sql.NullBool

		var column Column
		if e := rows.Scan(&column.DbName,
			&isNullable,
			&column.DbType,
			&primaryKey,
			&isSequence); nil != e {
			return nil, e
		}

		if isNullable.Valid {
			column.IsNullable = strings.ToLower(isNullable.String) == "yes"
		}
		if primaryKey.Valid {
			column.IsPrimaryKey = primaryKey.Bool
		}
		isSequenceByForeignKey := false
		if isForeignKey, e := cmd.isForeignKey(db, tableCatalog, tableSchema, tableName, column.DbName); e == nil {
			if "id" == column.DbName { // for tpt_managed_objects
				column.IsPrimaryKey = true
				isSequenceByForeignKey = true
			} else {
				column.IsForeignKey = isForeignKey
			}
			//if "id" == column.DbName {
			//	fmt.Println(tableName, column.DbName, isForeignKey)
			//}
		}
		if isSequence.Valid {
			column.IsSequence = isSequence.Bool
		}
		if isSequenceByForeignKey {
			column.IsSequence = true
		}

		found := false
		for idx, col := range columns {
			if col.DbName == column.DbName {
				found = true

				if column.IsPrimaryKey {
					columns[idx].IsPrimaryKey = true
				}
				break
			}
		}
		if found {
			continue
		}

		column.GoName = CamelCase(column.DbName)
		if "id" == column.DbName && "int4" == column.DbType {
			column.GoType = "int64"
		} else if column.IsForeignKey {
			column.GoType = "int64"
		} else {
			column.GoType = ToGoTypeFromDbType(column.DbType)
		}
		columns = append(columns, column)
	}
	return columns, rows.Err()
}
