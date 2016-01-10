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
	Schema        string
	TableName     string
	ClassName     string
	IsView        bool
	Columns       []Column
	IsCombinedKey bool
	PrimaryKey    []Column
}

// Column entity in table `information_schema.columns`
type Column struct {
	DbName string
	GoName string
	DbType string
	GoType string

	IsNullable   bool
	IsPrimaryKey bool
	IsSequence   bool
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
		if isSequence.Valid {
			column.IsSequence = isSequence.Bool
		}

		column.GoName = CamelCase(column.DbName)
		column.GoType = ToGoTypeFromDbType(column.DbType)
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

// GetAll use to select all tables from `information_schema.tables`.
func (self *dataAccess) GetAll(db *sql.DB, tableSchema string) ([]Table, error) {
	queryString := fmt.Sprintf(`SELECT
            t.table_schema, t.table_name, t.table_type
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
		var tableType string
		if e := rows.Scan(&table.Schema, &table.TableName, &tableType); nil != e {
			return nil, e
		}
		if "view" == strings.ToLower(tableType) {
			table.IsView = true
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
		"set": func(values map[string]interface{}, name string, value interface{}) interface{} {
			values[name] = value
			return ""
		},

		"toNullName": func(s string) string {
			// switch s {
			// case "type":
			// 	return "_type"
			// case "if":
			// 	return "_if"
			// case "int":
			// 	return "_int"
			// }
			return "null_" + s
		},
		"CamelCase":   CamelCase,
		"Underscore":  Underscore,
		"Pluralize":   Pluralize,
		"Singularize": Singularize,
		"Tableize":    Tableize,
		"Capitalize":  Capitalize,
		"Typeify":     Typeify,
		"ToUpper":     strings.ToUpper,
		"ToNullType":  ToNullTypeFromPostgres,
		//"ToNullValue": ToNullValueFromPostgres,
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

	cmd.template_model.New("toNullValue").Parse(template_sql_null_value)
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
		log.Println("GEN ", table.TableName)
		columns, e := DB.GetByTable(db, cmd.db_schema, table.TableName)
		if nil != e {
			log.Println("failed to read columns for", table.TableName, "- ", e)
			return
		}
		table.Columns = columns
		table.IsCombinedKey, table.PrimaryKey = GetPrimaryKey(table.Columns)
		table.ClassName = Typeify(strings.TrimPrefix(table.TableName, cmd.db_prefix))

		if e := cmd.GenrateFromTable(out, table, columns); nil != e {
			log.Println(e)
			return
		}
	}
}

func (cmd *GenerateModelsCommand) GenrateFromTable(out io.Writer, table Table, columns []Column) error {
	return cmd.template_model.Execute(out, map[string]interface{}{
		"Namespace": cmd.ns,
		"table":     table,
		"columns":   columns,
	})
}

func GetPrimaryKey(columns []Column) (bool, []Column) {
	primaryKeys := make([]Column, 0, 4)
	for _, column := range columns {
		if column.IsPrimaryKey {
			primaryKeys = append(primaryKeys, column)
		}
	}
	return 1 != len(primaryKeys), primaryKeys
}

func init() {
	command.On("generate", "从数据库的表模型生成代码", &GenerateModelsCommand{}, nil)
}

var template_header_text = `// file is generated by gengen
package {{.Namespace}}

import "github.com/Masterminds/squirrel"

func isPostgersql(db interface{}) bool {
  return true
}

func isPlaceholderWithDollar(db interface{}) bool {
  return true
}

// type SelectBuilder interface{
//   Columns(columns ...string) squirrel.Sqlizer
// }
`

var template_model_text = `type {{.table.ClassName}} struct { {{range $x := .columns }}
  {{$x.GoName}} {{$x.GoType}}{{end}}
}

type _{{.table.ClassName}}Model struct{
  table_name   string
  column_names []string 
}

func (self *_{{.table.ClassName}}Model) scan(scanner squirrel.RowScanner) (*{{.table.ClassName}}, error){
  var value {{.table.ClassName}}
  {{$columns := .columns}}{{range $x := .columns }}{{if $x.IsNullable}}
  var {{toNullName $x.DbName}} {{ToNullType $x.DbType}}{{end}}{{end}}

  e := scanner.Scan({{range $idx, $x := .columns }}{{if not $x.IsNullable}}value.{{$x.GoName}}{{else}}{{toNullName $x.DbName}}{{end}}{{if last $columns $idx | not}},
    {{end}}{{end}})
  if nil != e {
    return nil, e
  }

  {{range $x := .columns }}{{if $x.IsNullable}}
  if {{toNullName $x.DbName}}.Valid { {{template "toNullValue" $x}}}
  {{end}}{{end}}

  return &value, nil
}

func (self *_{{.table.ClassName}}Model) queryRowWith(db squirrel.QueryRower, builder squirrel.SelectBuilder) (*{{.table.ClassName}}, error){
  return self.scan(squirrel.QueryRowWith(db, builder.Columns(self.column_names...).From(self.table_name)))
}

func (self *_{{.table.ClassName}}Model) queryWith(db squirrel.Queryer, builder squirrel.SelectBuilder) ([]*{{.table.ClassName}}, error){
  rows, e := squirrel.QueryWith(db, builder.Columns(self.column_names...).From(self.table_name))
  if nil != e {
    return nil, e
  }
  results := make([]*{{.table.ClassName}}, 0, 4)
  for rows.Next() {
    v, e := self.scan(rows)
    if nil != e {
      return nil, e
    }
    results = append(results, v)
  }
  return results, rows.Err()
}

func (self *_{{.table.ClassName}}Model) FindById(db squirrel.QueryRower, id int64) (*{{.table.ClassName}}, error){
  builder := squirrel.Select(self.column_names...).From(self.table_name).Where(squirrel.Eq{"id": id})
  return self.queryRowWith(db, builder)
}

{{if not .table.IsView }}


{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
func (self *_{{.table.ClassName}}Model) CreateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) ({{$pk.GoType}}, error){ {{else}}
func (self *_{{.table.ClassName}}Model) CreateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) error { {{end}}
  sql := squirrel.Insert(self.table_name).Columns(self.column_names[1:]...).
    Values({{$columns := .columns}}{{range $idx, $x := .columns }}value.{{$x.GoName}}{{if last $columns $idx | not}},
    {{end}}{{end}})

  if isPlaceholderWithDollar(db) {
    sql = sql.PlaceholderFormat(squirrel.Dollar)
  }

{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}{{if $pk.IsSequence}}
  if isPostgersql(db) {
    if e := sql.Suffix("RETURNING \"{{$pk.GoName}}\"").RunWith(db).
        QueryRow().Scan(&value.{{$pk.GoName}}); nil != e {
      return value.{{$pk.GoName}}, e
    }

    return value.{{$pk.GoName}}, nil
  }

  result, e := sql.RunWith(db).Exec();
  if nil != e {
    return value.{{$pk.GoName}}, e
  }
  {{if eq $pk.GoType "int64"}}value.{{$pk.GoName}}, e = result.LastInsertId()
  {{else}}pk, e := result.LastInsertId()
  if nil != e {
    return value.{{$pk.GoName}}, e
  }
  value.{{$pk.GoName}} = {{$pk.GoType}}(pk)
  {{end}}
  return value.{{$pk.GoName}}, e
}
{{else}}
  result, e := sql.RunWith(db).Exec();
  if nil != e {
    return value.{{$pk.GoName}}, e
  }
  _, e = result.RowsAffected()
  return value.{{$pk.GoName}}, e
}
{{end}}{{else}}
  result, e := sql.RunWith(db).Exec();
  if nil != e {
    return e
  }
  _, e = result.RowsAffected()
  return e
}
{{end}}

{{$columns := .columns}}
func (self *_{{.table.ClassName}}Model) UpdateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) (error) {
  sql := squirrel.Update(self.table_name).{{range $idx, $x := .columns }}
    {{if not $x.IsPrimaryKey}}Set("{{$x.DbName}}", value.{{$x.GoName}}).{{end}}{{end}}
    Where(squirrel.Eq{ {{range $column := .columns}} {{if $column.IsPrimaryKey}}"{{$column.DbName}}": value.{{$column.GoName}}, 
      {{end}}{{end}} })

  if isPlaceholderWithDollar(db) {
    sql = sql.PlaceholderFormat(squirrel.Dollar)
  }

  result, e := sql.RunWith(db).Exec();
  if nil != e {
    return e
  }
  rowsAffected, e := result.RowsAffected()
  if nil != e {
    return e
  }
  if 0 == rowsAffected {
    return errors.New("update failed")
  }
  return nil
}

func (self *_{{.table.ClassName}}Model) UpdateBy(db squirrel.BaseRunner, values map[string]interface{}, pred interface{}, args ...interface{}) (int64, error) {
  sql := squirrel.Update(self.table_name)
    for key, value := range values {
      sql = sql.Set(key, value)
    }

  sql = sql.Where(pred, args)
  if isPlaceholderWithDollar(db) {
    sql = sql.PlaceholderFormat(squirrel.Dollar)
  }

  result, e := sql.RunWith(db).Exec();
  if nil != e {
    return 0, e
  }
  return result.RowsAffected()
}

func (self *_{{.table.ClassName}}Model) DeleteIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) error { {{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
  return self.DeleteById(db, value.{{$pk.GoName}}) {{else}}
  _, e := self.DeleteBy(db, squirrel.Eq{ {{range $column := .columns}} 
      "{{$column.DbName}}": value.{{$column.GoName}},
    {{end}} })
  return e {{end}}
}


{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
func (self *_{{.table.ClassName}}Model) DeleteById(db squirrel.BaseRunner, key {{$pk.GoType}}) error {
  _, e := self.DeleteBy(db, squirrel.Eq{"{{$pk.DbName}}": key})
  return e
}
{{end}}

func (self *_{{.table.ClassName}}Model) DeleteBy(db squirrel.BaseRunner, pred interface{}, args ...interface{}) (int64, error) {
  result, e :=  squirrel.Delete("").From(self.table_name).
    Where(pred, args).RunWith(db).Exec();
  if nil != e {
    return 0, e
  }
  return result.RowsAffected()
}


{{end}} {{/* isView end */}}

var {{.table.ClassName}}Model = _{{.table.ClassName}}Model{
  table_name: "{{.table.TableName}}",
  column_names: []string{ {{range $x := .columns }} "{{$x.DbName}}", 
  {{end}} },
}
`

var template_sql_null_value = `{{if eq .DbType "bool"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Bool
    {{else if eq .DbType "int4"}}
      value.{{.GoName}} = int({{toNullName .DbName}}.Int64)
    {{else if eq .DbType "int8"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Int64
    {{else if eq .DbType "float4"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Float64
    {{else if eq .DbType "float8"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Float64
    {{else if eq .DbType "numeric"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Float64
    {{else if eq .DbType "varchar"}}
      value.{{.GoName}} = {{toNullName .DbName}}.String
    {{else if eq .DbType "text"}}
      value.{{.GoName}} = {{toNullName .DbName}}.String
    {{else if eq .DbType "timestamp"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Time
    {{else if eq .DbType "timestamptz"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Time
    {{else if eq .DbType "cidr"}}
      if "" != {{toNullName .DbName}}.String {
        ipValue := net.ParseIP({{toNullName .DbName}}.String)
        if nil != ipValue {
          value.{{.GoName}} = ipValue
        } else if cidr, _, e := net.ParseCIDR({{toNullName .DbName}}.String); nil == e {
          value.{{.GoName}} = cidr
        }
      }
    {{else if eq .DbType "macaddr"}}
      value.{{.GoName}} = {{toNullName .DbName}}.String
    {{else}}
      type({{.DbType}}) of value.{{.DbName}} is unsupported...........................................
    {{end}}`

func ToGoTypeFromDbType(nm string) string {
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
