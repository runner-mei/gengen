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
		"columns_remove_foreign_keys": func(columns []Column) interface{} {
			filterdColumns := make([]Column, 0, len(columns))
			for _, column := range columns {
				if !column.IsForeignKey {
					filterdColumns = append(filterdColumns, column)
				}
			}
			return filterdColumns
		},
		"columns_count_foreign_keys": func(columns []Column) interface{} {
			count := 0
			for _, column := range columns {
				if column.IsForeignKey {
					count++
				}
			}
			return count
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
			return "null" + CamelCase(s)
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
		"notSQLSupport": func(typ string) bool {
			return "net.IP" == typ || "JSON" == typ
		},
		"firstLower": func(s string) string {
			if "" == s {
				return s
			}
			return strings.ToLower(s[:1]) + s[1:]
		},
		"isIntegerType": func(s string) bool {
			s = strings.ToLower(s)
			return s == "int" || s == "int32" || s == "int64" ||
				s == "uint" || s == "uint32" || s == "uint64"
		},
		//"ToNullValue": ToNullValueFromPostgres,
	}

	var e error
	cmd.templateHeader, e = template.New("template_header_text").Funcs(funcs).Parse(template_header_text)
	if nil != e {
		return e
	}

	cmd.templateModel, e = template.New("template_model_text").Funcs(funcs).Parse(template_model_text)
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
  {{$columns := .columns}}{{range $x := .columns }}{{if $x.IsNullable}}var {{toNullName $x.DbName}} {{ToNullType $x.DbType}}
  {{else if eq "net.IP" $x.GoType}}var {{toNullName $x.DbName}} sql.NullString
  {{end}}{{end}}

  e := scanner.Scan({{range $idx, $x := .columns }}{{if eq "net.IP" $x.GoType}}&{{toNullName $x.DbName}}{{else if $x.IsNullable}}&{{toNullName $x.DbName}}{{else}}&value.{{$x.GoName}}{{end}}{{if last $columns $idx | not}},
    {{end}}{{end}})
  if nil != e {
    return nil, e
  }

  {{range $x := .columns }}{{if $x.IsNullable}}
  {{template "toNullValue" $x}}
  {{else if  eq "net.IP" $x.GoType}}
  {{template "toNullValue" $x}}
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
  defer rows.Close()
  
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

{{if not .table.IsCombinedKey}}
func (self *{{firstLower .table.ClassName}}Model) FindByID(db squirrel.QueryRower, id int64) (*{{.table.ClassName}}, error){
  builder := squirrel.Select().From(self.TableName).Where(squirrel.Eq{"id": id})
  return self.QueryRowWith(db, builder)
}
{{else if not .table.IsView}}
func (self *{{firstLower .table.ClassName}}Model) FindByKey(db squirrel.QueryRower, {{range $idx, $column := .table.PrimaryKey}}{{firstLower $column.GoName}} {{$column.GoType}}{{if last $columns $idx | not}},
      {{end}}{{end}}) (*{{.table.ClassName}}, error){
  builder := squirrel.Select().From(self.TableName).Where({{range $idx, $column := .table.PrimaryKey}}squirrel.Eq{"{{$column.DbName}}": {{firstLower $column.GoName}} }{{if last $columns $idx | not}},
      {{end}}{{end}})
  return self.QueryRowWith(db, builder)
}
{{end}}

{{if not .table.IsView }}
{{$oldValues := .}}
{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
func (self *{{firstLower .table.ClassName}}Model) CreateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) ({{$pk.GoType}}, error){ {{else}}
func (self *{{firstLower .table.ClassName}}Model) CreateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) error { {{end}}
    {{if .table.HasCreatedAt}}value.CreatedAt = time.Now()
    {{end}}{{if .table.HasUpdatedAt}}value.UpdatedAt = time.Now()
    {{end}}{{$columns := .columns | list_create}}{{$fkeyCount := columns_count_foreign_keys $columns}}{{if eq $fkeyCount 0 }}builder := squirrel.Insert(self.TableName).Columns({{list_join $columns}}).
    Values({{range $idx, $x := $columns }}value.{{$x.GoName}}{{if notSQLSupport $x.GoType}}.String(){{end}}{{if last $columns $idx | not}},
    {{end}}{{end}})
  {{else if eq $fkeyCount 1}}{{range $idx, $x := $columns }}{{if $x.IsForeignKey}}{{set $oldValues "foreignKey" $x}}{{end}}{{end}}var builder squirrel.InsertBuilder

  if value.{{.foreignKey.GoName}} <= 0 {
    builder = squirrel.Insert(self.TableName).Columns({{columns_remove_foreign_keys $columns | list_join}}).
      Values({{range $idx, $x := columns_remove_foreign_keys $columns }}value.{{$x.GoName}}{{if notSQLSupport $x.GoType}}.String(){{end}}{{if last $columns $idx | not}},
      {{end}}{{end}})
  } else {
    builder = squirrel.Insert(self.TableName).Columns({{list_join $columns}}).
      Values({{range $idx, $x := $columns }}value.{{$x.GoName}}{{if notSQLSupport $x.GoType}}.String(){{end}}{{if last $columns $idx | not}},
      {{end}}{{end}})
  }
  {{else}} {{/*  if eq $fkeyCount 1 */}}
  columnNames := []string{ {{columns_remove_foreign_keys $columns | list_join}} }
  columnValues := []interface{}{ {{range $idx, $x := columns_remove_foreign_keys $columns }}value.{{$x.GoName}}{{if notSQLSupport $x.GoType}}.String(){{end}}{{if last $columns $idx | not}},
      {{end}}{{end}}}

  {{range $idx, $x := $columns }}{{if $x.IsForeignKey}}if value.{{$x.GoName}} > 0 {
    columnNames = append(columnNames, "{{$x.DbName}}")
    columnValues = append(columnValues, value.{{$x.GoName}})
  }
  {{end}}{{end}}

  var  builder = squirrel.Insert(self.TableName).Columns(columnNames...).
      Values(columnValues...)
  {{end}} {{/*  if eq $fkeyCount 1 */}}
  if isPlaceholderWithDollar(db) {
    builder = builder.PlaceholderFormat(squirrel.Dollar)
  }

{{if .table.IsCombinedKey}}
  result, e := builder.RunWith(db).Exec();
  if nil != e {
    return e
  }
  _, e = result.RowsAffected()
  return e
}
{{else}}
{{$pk := index .table.PrimaryKey 0}}
{{if $pk.IsSequence}}
  if isPostgersql(db) {
    if e := builder.Suffix("RETURNING \"{{$pk.DbName}}\"").RunWith(db).
        QueryRow().Scan(&value.{{$pk.GoName}}); nil != e {
      return 0, e
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
    return 0, e
  }
  value.{{$pk.GoName}} = {{$pk.GoType}}(pk){{end}}
  return value.{{$pk.GoName}}, e
}
{{else}} {{/* IsSequence */}}
  result, e := builder.RunWith(db).Exec();
  if nil != e {
    return value.{{$pk.GoName}}, e
  }

  _, e = result.RowsAffected()
  return value.{{$pk.GoName}}, e
}
{{end}} {{/* IsSequence end */}}
{{end}}

{{$columns := .columns}}
func (self *{{firstLower .table.ClassName}}Model) UpdateIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) (error) {
  {{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}{{if isIntegerType $pk.GoType}}if 0 == value.{{$pk.GoName}} {
    return ThrowPrimaryKeyInvalid(self)
  }

  {{end}}{{if eq "string" $pk.GoType}}if "" == value.{{$pk.GoName}} {
    return ThrowPrimaryKeyInvalid(self)
  }

  {{end}}{{end}}{{if .table.HasUpdatedAt}}value.UpdatedAt = time.Now()
  {{end}}{{$columns := .columns | list_update}}builder := squirrel.Update(self.TableName).
    {{range $idx, $x := $columns }}{{if not $x.IsForeignKey}}{{if not $x.IsPrimaryKey}}Set("{{$x.DbName}}", value.{{$x.GoName}}{{if notSQLSupport $x.GoType}}.String(){{end}}).
    {{end}}{{end}}{{end}}Where({{range $idx, $column := .table.PrimaryKey}}squirrel.Eq{"{{$column.DbName}}": value.{{$column.GoName}} }{{if last $columns $idx | not}},
      {{end}}{{end}})

  {{range $idx, $x := $columns }}{{if $x.IsForeignKey}}if value.{{$x.GoName}} > 0 {
    builder = builder.Set("{{$x.DbName}}", value.{{$x.GoName}})
  }
  {{end}}{{end}}

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
    return ErrNotUpdated
  }
  return nil
}


func (self *{{firstLower .table.ClassName}}Model) DeleteIt(db squirrel.BaseRunner, value *{{.table.ClassName}}) error {
  {{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}return self.DeleteByID(db, value.{{$pk.GoName}})
  {{else}}count, err := self.DeleteBy(db, {{range $idx, $column := .table.PrimaryKey}}squirrel.Eq{"{{$column.DbName}}": value.{{$column.GoName}} }{{if last $columns $idx | not}},
      {{end}}{{end}})
  if err != nil {
    return err
  }
  if count == 0 {
    return ErrNotDeleted
  }
  return nil {{end}}
}

{{if not .table.IsCombinedKey}}{{$pk := index .table.PrimaryKey 0}}
func (self *{{firstLower .table.ClassName}}Model) DeleteByID(db squirrel.BaseRunner, key {{$pk.GoType}}) error {
  {{if isIntegerType $pk.GoType}}if 0 == key {
    return ThrowPrimaryKeyInvalid(self)
  }
  {{end}}{{if eq "string" $pk.GoType}}if "" == key {
    return ThrowPrimaryKeyInvalid(self)
  }
  {{end}}

  count, err := self.DeleteBy(db, squirrel.Eq{"{{$pk.DbName}}": key})
  if err != nil {
    return err
  }
  if count == 0 {
    return ErrNotDeleted
  }
  return nil
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

var template_sql_null_value = `{{if eq .DbType "bool"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.Bool
    }{{else if eq .GoType "int64"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.Int64
    }{{else if eq .DbType "int4"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = int({{toNullName .DbName}}.Int64)
    }{{else if eq .DbType "int8"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.Int64
    }{{else if eq .DbType "float4"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.Float64
    }{{else if eq .DbType "float8"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.Float64
    }{{else if eq .DbType "numeric"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.Float64
    }{{else if eq .DbType "varchar"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.String
    }{{else if eq .DbType "text"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.String
    }{{else if eq .DbType "json"}}
      value.{{.GoName}} = ToJSON({{toNullName .DbName}})
    {{else if eq .DbType "jsonb"}}
      value.{{.GoName}} = ToJSON({{toNullName .DbName}})
    {{else if eq .DbType "timestamp"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.Time
    }{{else if eq .DbType "timestamptz"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.Time
    }{{else if eq .GoType "net.IP"}}if {{toNullName .DbName}}.Valid { 
      if "" != {{toNullName .DbName}}.String {
        ipValue := net.ParseIP({{toNullName .DbName}}.String)
        if nil != ipValue {
          value.{{.GoName}} = ipValue
        } else if cidr, _, e := net.ParseCIDR({{toNullName .DbName}}.String); nil == e {
          value.{{.GoName}} = cidr
        }
      }
    }{{else if eq .DbType "cidr"}}if {{toNullName .DbName}}.Valid { 
      if "" != {{toNullName .DbName}}.String {
        ipValue := net.ParseIP({{toNullName .DbName}}.String)
        if nil != ipValue {
          value.{{.GoName}} = ipValue
        } else if cidr, _, e := net.ParseCIDR({{toNullName .DbName}}.String); nil == e {
          value.{{.GoName}} = cidr
        }
      }
    }{{else if eq .DbType "macaddr"}}if {{toNullName .DbName}}.Valid { 
      value.{{.GoName}} = {{toNullName .DbName}}.String
    }{{else}}if {{toNullName .DbName}}.Valid { 
      type({{.DbType}}) of value.{{.DbName}} is unsupported...........................................
    }{{end}}`

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
		return "JSON"
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
		return "[]byte" // "sql.RawBytes" -- sql: RawBytes isn't allowed on Row.Scan
	default:
		panic("'" + nm + "' is unsupported")
	}
}
