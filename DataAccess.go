package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
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
	db_prefix string

	root            string
	template_header *template.Template
	template_model  *template.Template
}

func (cmd *GenerateModelsCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	flag.StringVar(&cmd.db_url, "db_url", "host=127.0.0.1 port=35432 dbname=tpt user=tpt password=extreme sslmode=disable", "the db url")
	flag.StringVar(&cmd.db_drv, "db_drv", "postgres", "the db driver")
	flag.StringVar(&cmd.db_schema, "db_schema", "public", "the db schema")
	flag.StringVar(&cmd.ns, "namespace", "models", "the namespace name")
	flag.StringVar(&cmd.db_prefix, "db_prefix", "tpt_", "the db prefix name")
	return fs
}

func (cmd *GenerateModelsCommand) Init() error {
	var funcs = template.FuncMap{
		"last": func(v interface{}, i int) (bool, error) {
			rv := reflect.ValueOf(v)
			if rv.Kind() != reflect.Slice {
				return false, errors.New("not a slice")
			}
			return rv.Len()-1 == i, nil
		},
		"CamelCase":   CamelCase,
		"Underscore":  Underscore,
		"Pluralize":   Pluralize,
		"Singularize": Singularize,
		"Tableize":    Tableize,
		"Capitalize":  Capitalize,
		"Typeify":     Typeify,
		"ToUpper":     strings.ToUpper,
		"ToGoType":    ToGoTypeFromPostgres,
		"ToNullType":  ToNullTypeFromPostgres,
		"ToNullValue": ToNullValueFromPostgres,
	}

	var e error
	cmd.template_header, e = template.New("default").Funcs(funcs).Parse(template_header_text)
	if nil != e {
		return e
	}

	cmd.template_model, e = template.New("default").Funcs(funcs).Parse(template_model_text)
	if nil != e {
		return e
	}
	return nil
}

func (cmd *GenerateModelsCommand) Run(args []string) {
	if e := cmd.Init(); nil != e {
		log.Println(e)
		return
	}

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

	out := os.Stdout

	if e := cmd.template_header.Execute(out, map[string]interface{}{
		"Namespace": cmd.ns,
	}); nil != e {
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

		if e := cmd.GenrateFromTable(out, table, columns); nil != e {
			log.Println(e)
			return
		}
	}
}

func (cmd *GenerateModelsCommand) GenrateFromTable(out io.Writer, table Table, columns []Column) error {
	table.Name = strings.TrimPrefix(table.Name, cmd.db_prefix)

	return cmd.template_model.Execute(out, map[string]interface{}{
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

var template_header_text = `// file is generated by gengen
package {{.Namespace}}

import "github.com/Masterminds/squirrel"


type SelectBuilder interface{
  Columns(columns ...string) Sqlizer
}
`

var template_model_text = `type {{Typeify .table.Name}} struct { {{range $x := .columns }}
  {{$x.Name}} {{ToGoType $x.DataType}}{{end}}
}

type _{{Typeify .table.Name}}Model struct{
  column_names []string 
} 

func (self *_{{Typeify .table.Name}}Model) scan(scanner squirrel.RowScanner) (*{{Typeify .table.Name}}, error){
  var value = {{Typeify .table.Name}}
  {{$columns := .columns}}{{range $x := .columns }}{{if ne $x.Nullable "YES"}}
  var {{$x.Name}} {{ToNullType $x.DataType}}{{end}}{{end}}

  e := scanner.Scan({{range $idx, $x := .columns }}{{if ne $x.Nullable "yes"}}{{$x.Name}}{{else}}value.{{$x.Name}}{{end}}{{if last $columns $idx | not}},
    {{end}}{{end}})
  if nil != e {
    return nil, e
  }

  {{range $x := .columns }}{{if ne $x.Nullable "YES"}}
  if {{$x.Name}}.Valid {
    value.{{$x.Name}} = {{ToNullValue $x.Name $x.DataType}}
  }
  {{end}}{{end}}

  return nil, errors.New("NOT IMPLEMENTED")
}

func (self *_{{Typeify .table.Name}}Model) queryRowWith(builder SelectBuilder, db squirrel.QueryRower) (*{{Typeify .table.Name}}, error){
  return self.scan(squirrel.QueryRowWith(builder.Columns(self.columns), db))
}

func (self *_{{Typeify .table.Name}}Model) queryWith(builder SelectBuilder, db squirrel.Queryer) ([]*{{Typeify .table.Name}}, error){
  rows, e := squirrel.QueryWith(builder.Columns(self.columns), db)
  if nil != e {
    return nil, e
  }
  results := make([]*{{Typeify .table.Name}}, 0, 4)
  for rows.Next() {
    v, e := self.scan(rows)
    if nil != e {
      return nil, e
    }
  }
  return results, rows.Err()
}

var {{Typeify .table.Name}}Model = _{{Typeify .table.Name}}Model{
  column_names: []string{ {{range $x := .columns }} "{{$x.Name}}", 
  {{end}} }
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

func ToNullTypeFromPostgres(nm string) string {
	switch nm {
	case "bool":
		return "sql.NullBool"
	case "int4":
		return "sql.NullInt64"
	case "int8":
		return "sql.NullInt64"
	case "float4":
		return "sql.NullFloat64"
	case "float8", "numeric":
		return "sql.NullFloat64"
	case "varchar", "text":
		return "sql.NullString"
	case "timestamp", "timestamptz":
		return "pq.NullTime"
	case "cidr":
		return "sql.NullString"
	case "macaddr":
		return "sql.NullString"
	default:
		panic("'" + nm + "' is unsupported")
	}
}

func ToNullValueFromPostgres(valueName, typeName string) string {
	switch typeName {
	case "bool":
		return valueName + ".Bool"
	case "int4":
		return valueName + ".Int64"
	case "int8":
		return valueName + ".Int64"
	case "float4":
		return valueName + ".Float64"
	case "float8", "numeric":
		return valueName + ".Float64"
	case "varchar", "text":
		return valueName + ".String"
	case "timestamp", "timestamptz":
		return valueName + ".Time"
	case "cidr":
		return valueName + ".String"
	case "macaddr":
		return valueName + ".String"
	default:
		panic("'" + typeName + "' is unsupported")
	}
}
