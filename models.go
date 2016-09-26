package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

// GenerateModelsCommand - 生成数据库模型代码
type GenerateModelsCommand struct {
	dbBase
	ns   string
	file string

	root           string
	templateHeader *template.Template
	templateModel  *template.Template
}

// Flags - 申明参数
func (cmd *GenerateModelsCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.initFlags(fs)
	flag.StringVar(&cmd.ns, "namespace", "models", "the namespace name")
	flag.StringVar(&cmd.file, "file", "models.go", "the output target")
	return fs
}

func (cmd *GenerateModelsCommand) init() error {
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
		"list_create": func(columns []Column) interface{} {
			filterdColumns := make([]Column, 0, len(columns))
			for _, column := range columns {
				found := false
				for _, s := range []string{"id"} {
					if s == column.DbName {
						found = true
						break
					}
				}
				if !found {
					filterdColumns = append(filterdColumns, column)
				}
			}
			return filterdColumns
		},
		"list_update": func(columns []Column) interface{} {
			filterdColumns := make([]Column, 0, len(columns))
			for _, column := range columns {
				found := false
				for _, s := range []string{"id", "created_at"} {
					if s == column.DbName {
						found = true
						break
					}
				}
				if !found {
					filterdColumns = append(filterdColumns, column)
				}
			}
			return filterdColumns
		},
		"list_join": func(columns []Column) interface{} {
			var buf bytes.Buffer
			for _, column := range columns {
				buf.WriteString("\"")
				buf.WriteString(column.DbName)
				buf.WriteString("\",")
			}
			if buf.Len() > 0 {
				buf.Truncate(buf.Len() - 1)
			}
			return buf.String()
		},

		"toNullName": func(s string) string {
			// switch s {
			// case "type":
			//  return "_type"
			// case "if":
			//  return "_if"
			// case "int":
			//  return "_int"
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
		"firstLower": func(s string) string {
			if "" == s {
				return s
			}
			return strings.ToLower(s[:1]) + s[1:]
		},
		//"ToNullValue": ToNullValueFromPostgres,
	}

	var e error
	cmd.templateHeader, e = template.New("default").Funcs(funcs).Parse(template_header_text)
	if nil != e {
		return e
	}

	cmd.templateModel, e = template.New("default").Funcs(funcs).Parse(template_model_text)
	if nil != e {
		return e
	}

	cmd.templateModel.New("toNullValue").Parse(template_sql_null_value)
	if nil != e {
		return e
	}
	return nil
}

// Run - 生成数据库模型代码
func (cmd *GenerateModelsCommand) Run(args []string) {
	if e := cmd.init(); nil != e {
		log.Println(e)
		return
	}

	tables, e := cmd.GetAllTables()
	if nil != e {
		log.Println(e)
		return
	}

	out := os.Stderr
	switch strings.ToLower(cmd.file) {
	case "stdout":
		out = os.Stdout
	case "stderr":
		out = os.Stderr
	case "":
		out = os.Stderr
	default:
		out, e = os.OpenFile(cmd.file, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0)
		if nil != e {
			log.Println(e)
			return
		}
		target := filepath.Join(filepath.Dir(cmd.file), "base.go")
		if e = copyFile(cmd.ns, embede_text, target); nil != e {
			log.Println(e)
			return
		}
	}

	if e := cmd.templateHeader.Execute(out, map[string]interface{}{
		"Namespace": cmd.ns,
	}); nil != e {
		log.Println(e)
		return
	}

	for _, table := range tables {
		log.Println("GEN ", table.TableName)
		if e := cmd.genrateFromTable(out, table); nil != e {
			log.Println(e)
			return
		}
	}
}

func (cmd *GenerateModelsCommand) genrateFromTable(out io.Writer, table Table) error {
	return cmd.templateModel.Execute(out, map[string]interface{}{
		"Namespace": cmd.ns,
		"table":     table,
		"columns":   table.Columns,
	})
}

func copyFile(ns, content, target string) error {
	out, e := os.OpenFile(target, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0)
	if nil != e {
		return e
	}
	defer out.Close()

	buf := bufio.NewReader(strings.NewReader(content))
	for {
		line, _, e := buf.ReadLine()
		if nil != e {
			if io.EOF == e {
				break
			}
			return e
		}

		ss := bytes.Fields(bytes.TrimSpace(line))
		if 2 == len(ss) && "package" == string(ss[0]) {
			out.WriteString("package " + ns + "\r\n")
		} else {
			out.Write(line)
			out.Write([]byte("\r\n"))
		}
	}
	return nil
}

func getPrimaryKey(columns []Column) (bool, []Column) {
	primaryKeys := make([]Column, 0, 4)
	for _, column := range columns {
		if column.IsPrimaryKey {
			primaryKeys = append(primaryKeys, column)
		}
	}
	return 1 != len(primaryKeys), primaryKeys
}

var template_header_text = `// file is generated by gengen
package {{.Namespace}}

import (
  "github.com/Masterminds/squirrel"
  "time"
  "net"
  "database/sql"
  "github.com/lib/pq"
  "errors"
)

// type SelectBuilder interface{
//   Columns(columns ...string) squirrel.Sqlizer
// }
`

var template_model_text = `type {{.table.ClassName}} struct { {{range $x := .columns }}{{if eq "Id" $x.GoName}}
  {{$x.GoName}} int64` + "\t`json:\"id,omitempty\"`" + `
  {{else}}{{$x.GoName}} {{$x.GoType}}` + "\t`json:\"{{$x.DbName}},omitempty\"`" + `{{end}}
{{end}}
}

{{if not .table.IsView }}

{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
func (self *{{.table.ClassName}}) CreateIt(db squirrel.BaseRunner) ({{$pk.GoType}}, error){ {{else}}
func (self *{{.table.ClassName}}) CreateIt(db squirrel.BaseRunner) error { {{end}}
  return {{.table.ClassName}}Model.CreateIt(db, self)
}

func (self *{{.table.ClassName}}) UpdateIt(db squirrel.BaseRunner) error {
  return {{.table.ClassName}}Model.UpdateIt(db, self)
}

func (self *{{.table.ClassName}}) DeleteIt(db squirrel.BaseRunner) error { 
  return {{.table.ClassName}}Model.DeleteIt(db, self)
}
{{end}}

type {{firstLower .table.ClassName}}Columns struct{
  {{range $x := .columns }}{{ToUpper $x.GoName}} ColumnModel
  {{end}}
}

type {{firstLower .table.ClassName}}Model struct{
  {{if .table.IsView }}ViewModel{{else}}DbModel{{end}}
  C {{firstLower .table.ClassName}}Columns
}

func (self *{{firstLower .table.ClassName}}Model) scan(scanner squirrel.RowScanner) (*{{.table.ClassName}}, error){
  var value {{.table.ClassName}}
  {{$columns := .columns}}{{range $x := .columns }}{{if $x.IsNullable}}
  var {{toNullName $x.DbName}} {{ToNullType $x.DbType}}{{end}}{{end}}

  e := scanner.Scan({{range $idx, $x := .columns }}{{if not $x.IsNullable}}&value.{{$x.GoName}}{{else}}&{{toNullName $x.DbName}}{{end}}{{if last $columns $idx | not}},
    {{end}}{{end}})
  if nil != e {
    return nil, e
  }

  {{range $x := .columns }}{{if $x.IsNullable}}
  if {{toNullName $x.DbName}}.Valid { {{template "toNullValue" $x}}}
  {{end}}{{end}}

  return &value, nil
}

func (self *{{firstLower .table.ClassName}}Model) QueryRowWith(db squirrel.QueryRower, builder squirrel.SelectBuilder) (*{{.table.ClassName}}, error){
  if isPlaceholderWithDollar(db) {
    builder = builder.PlaceholderFormat(squirrel.Dollar)
  }
  return self.scan(squirrel.QueryRowWith(db, builder.Columns(self.ColumnNames...).From(self.TableName)))
}

func (self *{{firstLower .table.ClassName}}Model) QueryWith(db squirrel.Queryer, builder squirrel.SelectBuilder) ([]*{{.table.ClassName}}, error){
  if isPlaceholderWithDollar(db) {
    builder = builder.PlaceholderFormat(squirrel.Dollar)
  }

  rows, e := squirrel.QueryWith(db, builder.Columns(self.ColumnNames...).From(self.TableName))
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

func (self *{{firstLower .table.ClassName}}Model) FindById(db squirrel.QueryRower, id int64) (*{{.table.ClassName}}, error){
  builder := squirrel.Select().From(self.TableName).Where(squirrel.Eq{"id": id})
  return self.QueryRowWith(db, builder)
}

{{if not .table.IsView }}


{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
func (self *{{firstLower .table.ClassName}}Model) CreateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) ({{$pk.GoType}}, error){ {{else}}
func (self *{{firstLower .table.ClassName}}Model) CreateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) error { {{end}}
    {{if .table.HasCreatedAt}}value.CreatedAt = time.Now()
    {{end}}{{if .table.HasUpdatedAt}}value.UpdatedAt = time.Now()
    {{end}}{{$columns := .columns | list_create}}builder := squirrel.Insert(self.TableName).Columns({{list_join $columns}}).
    Values({{range $idx, $x := $columns }}value.{{$x.GoName}}{{if last $columns $idx | not}},
    {{end}}{{end}})

  if isPlaceholderWithDollar(db) {
    builder = builder.PlaceholderFormat(squirrel.Dollar)
  }

{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}{{if $pk.IsSequence}}
  if isPostgersql(db) {
    if e := builder.Suffix("RETURNING \"{{$pk.DbName}}\"").RunWith(db).
        QueryRow().Scan(&value.{{$pk.GoName}}); nil != e {
      return value.{{$pk.GoName}}, e
    }

    return value.{{$pk.GoName}}, nil
  }

  result, e := builder.RunWith(db).Exec();
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
  result, e := builder.RunWith(db).Exec();
  if nil != e {
    return value.{{$pk.GoName}}, e
  }
  _, e = result.RowsAffected()
  return value.{{$pk.GoName}}, e
}
{{end}}{{else}}
  result, e := builder.RunWith(db).Exec();
  if nil != e {
    return e
  }
  _, e = result.RowsAffected()
  return e
}
{{end}}

{{$columns := .columns}}
func (self *{{firstLower .table.ClassName}}Model) UpdateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) (error) {
  {{if .table.HasUpdatedAt}}value.UpdatedAt = time.Now()
  {{end}}{{$columns := .columns | list_update}}builder := squirrel.Update(self.TableName).{{range $idx, $x := $columns }}
    {{if not $x.IsPrimaryKey}}Set("{{$x.DbName}}", value.{{$x.GoName}}).{{end}}{{end}}
    Where(squirrel.Eq{ {{range $column := .columns}} {{if $column.IsPrimaryKey}}"{{$column.DbName}}": value.{{$column.GoName}}, 
      {{end}}{{end}} })

  if isPlaceholderWithDollar(db) {
    builder = builder.PlaceholderFormat(squirrel.Dollar)
  }

  result, e := builder.RunWith(db).Exec();
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


func (self *{{firstLower .table.ClassName}}Model) DeleteIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) error { {{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
  return self.DeleteById(db, value.{{$pk.GoName}}) {{else}}
  _, e := self.DeleteBy(db, squirrel.Eq{ {{range $column := .columns}} 
      "{{$column.DbName}}": value.{{$column.GoName}},
    {{end}} })
  return e {{end}}
}


{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
func (self *{{firstLower .table.ClassName}}Model) DeleteById(db squirrel.BaseRunner, key {{$pk.GoType}}) error {
  _, e := self.DeleteBy(db, squirrel.Eq{"{{$pk.DbName}}": key})
  return e
}
{{end}}


{{end}} {{/* isView end */}}

var {{.table.ClassName}}Model = {{firstLower .table.ClassName}}Model{
  {{if .table.IsView }}ViewModel: ViewModel{TableName: "{{.table.TableName}}",
  ColumnNames: []string{ {{range $x := .columns }} "{{$x.DbName}}", 
  {{end}} }},{{else}}DbModel: DbModel{ViewModel: ViewModel{TableName: "{{.table.TableName}}",
  ColumnNames: []string{ {{range $x := .columns }}  "{{$x.DbName}}", 
  {{end}}}},
  KeyNames: []string{ {{range $x := .columns }} {{if $x.IsPrimaryKey }}"{{$x.DbName}}", 
  {{end}}{{end}} },
  },{{end}}
  C: {{firstLower .table.ClassName}}Columns{ {{range $x := .columns }}{{ToUpper $x.GoName}}: ColumnModel{"{{$x.DbName}}"},
    {{end}}},
}
`

var template_sql_null_value = `{{if eq .DbType "bool"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Bool
    {{else if eq .GoType "int64"}}
      value.{{.GoName}} = {{toNullName .DbName}}.Int64
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
    {{else if eq .DbType "json"}}
      value.{{.GoName}} = {{toNullName .DbName}}.String
    {{else if eq .DbType "jsonb"}}
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
	case "json", "jsonb":
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
	case "json", "jsonb":
		return "sql.NullString"
	default:
		panic("'" + nm + "' is unsupported")
	}
}
