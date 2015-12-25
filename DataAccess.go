package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	_ "github.com/lib/pq"
	"github.com/rakyll/command"
)

// Table entity in table `information_schema.tables`
type Table struct {
	Schema string `db:"table_schema" json:"-"`
	Name   string `db:"table_name" json:"name"`
}

// Column entity in table `information_schema.columns`
type Column struct {
	Name       string `db:"column_name" json:"name"`
	Nullable   string `db:"is_nullable" json:"-"`
	DataType   string `db:"udt_name"    json:"data_type"`
	PrimaryKey bool   `db:"primary_key" json:"primary"`
	IsSequence bool   `db:"is_sequence" json:"is_sequence"`
}

type dataAccess struct {
}

// GetByTable use to select columns from `information_schema.tables` of inputed tableName.
func (self *dataAccess) GetByTable(db *sql.DB, tableSchema, tableName string) ([]Column, error) {
	queryString := fmt.Sprintf(`SELECT
        t.column_name,
        t.is_nullable,
        t.udt_name,
        t.column_name = kcu.column_name as primary_key,
        t.column_default IS NOT NULL AND t.column_default LIKE 'nextval%%' as is_sequence
    FROM
        information_schema.columns t
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
    WHERE t.table_schema = '%s' and t.table_name = '%s'`, tableSchema, tableName)
	rows, e := db.Query(queryString)
	if nil != e {
		return nil, e
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var primaryKey sql.NullBool
		var isSequence sql.NullBool

		var column Column
		if e := rows.Scan(&column.Name,
			&column.Nullable,
			&column.DataType,
			&primaryKey,
			&isSequence); nil != e {
			return nil, e
		}
		if primaryKey.Valid {
			column.PrimaryKey = primaryKey.Bool
		}
		if isSequence.Valid {
			column.IsSequence = isSequence.Bool
		}
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

// GetAll use to select all tables from `information_schema.tables`.
func (self *dataAccess) GetAll(db *sql.DB, tableSchema string) ([]Table, error) {
	queryString := fmt.Sprintf(`SELECT
            t.table_schema, t.table_name
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
            t.table_schema = '%s'`, tableSchema)

	rows, e := db.Query(queryString)
	if nil != e {
		return nil, e
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		if e := rows.Scan(&table.Schema, &table.Name); nil != e {
			return nil, e
		}
		tables = append(tables, table)
	}
	return tables, rows.Err()
}

var DB = dataAccess{}

type GenerateModelsCommand struct {
	db_drv    string
	db_url    string
	db_schema string
	ns        string

	root string
}

func (cmd *GenerateModelsCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	flag.StringVar(&cmd.db_url, "db_url", "host=127.0.0.1 port=35432 dbname=tpt user=tpt password=extreme sslmode=disable", "the db url")
	flag.StringVar(&cmd.db_drv, "db_drv", "postgres", "the db driver")
	flag.StringVar(&cmd.db_schema, "db_schema", "public", "the db schema")
	flag.StringVar(&cmd.ns, "namespace", "models", "the namespace name")
	return fs
}

func (cmd *GenerateModelsCommand) Run(args []string) {
	db, e := sql.Open(cmd.db_drv, cmd.db_url)
	if nil != e {
		log.Println(e)
		return
	}
	defer db.Close()

	tables, e := DB.GetAll(db, cmd.db_schema)
	if nil != e {
		log.Println(e)
		return
	}

	for _, table := range tables {
		log.Println("GEN ", table.Name)

		columns, e := DB.GetByTable(db, cmd.db_schema, table.Name)
		if nil != e {
			log.Println("failed to read columns for", table.Name, "- ", e)
			return
		}

		if e := cmd.GenrateFromTable(table, columns); nil != e {
			log.Println(e)
			return
		}
	}
}

func (cmd *GenerateModelsCommand) GenrateFromTable(table Table, columns []Column) error {
	tmpl, e := template.New("default").Parse(models_template)
	if nil != e {
		return e
	}
	for idx, column := range columns {
		columns[idx].DataType = ToGoTypeFromPostgres(column.DataType)
	}
	return tmpl.Execute(os.Stdout, map[string]interface{}{
		"Namespace": cmd.ns,
		"table":     table,
		"columns":   columns,
	})
}

func init() {
	command.On("generate", "从数据库的表模型生成代码", &GenerateModelsCommand{}, nil)
}

// Name       string `db:"column_name" json:"name"`
// Nullable   string `db:"is_nullable" json:"-"`
// DataType   string `db:"udt_name"    json:"data_type"`
// PrimaryKey bool   `db:"primary_key" json:"primary"`
// IsSequence bool   `db:"is_sequence" json:"is_sequence"`

var models_template = `// file is generated by gengen
package {{.Namespace}}

type {{.table.Name}} struct { {{range $x := .columns }}
  {{$x.Name}} {{$x.DataType}}{{end}}
}

var _{{.table.Name}} = struct{
  } {
}
`

func ToGoTypeFromPostgres(nm string) string {
	switch nm {
	case "bool":
		return "bool"
	case "int4":
		return "int"
	case "int8":
		return "int64"
	case "float4":
		return "float"
	case "float8", "numeric":
		return "float64"
	case "varchar", "text":
		return "string"
	case "timestamp", "timestamptz":
		return "time.Time"
	case "cidr":
		return "net.IP"
	case "macaddr":
		return "string"
	default:
		panic("'" + nm + "' is unsupported")
	}
}
