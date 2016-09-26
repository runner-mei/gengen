package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

// GenerateControllerCommand - 生成控制器
type GenerateControllerCommand struct {
	dbBase

	root     string
	override bool
}

// Flags - 申明参数
func (cmd *GenerateControllerCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.initFlags(fs)
	flag.StringVar(&cmd.root, "root", "", "the root directory")
	flag.BoolVar(&cmd.override, "override", false, "")
	return fs
}

func (cmd *GenerateControllerCommand) init() error {
	if "" == cmd.root {
		for _, s := range []string{"conf/routes", "../conf/routes", "../../conf/routes", "../../conf/routes"} {
			if st, e := os.Stat(s); nil == e && nil != st && !st.IsDir() {
				cmd.root, _ = filepath.Abs(filepath.Join(s, "..", ".."))
				break
			}
		}

		if "" == cmd.root {
			return errors.New("root directory isn't found")
		}
	}
	return nil
}

// Run - 生成代码
func (cmd *GenerateControllerCommand) Run(args []string) {
	if len(args) == 0 {
		fmt.Println("table name is missing.")
		return
	}

	if e := cmd.init(); nil != e {
		fmt.Println(e)
		return
	}

	tables, e := cmd.GetAllTables()
	if nil != e {
		log.Println(e)
		return
	}

	for _, table := range tables {
		found := false
		for _, tableName := range args {
			if strings.Contains(table.TableName, tableName) {
				found = true
				break
			}
		}
		if !found {
			continue
		}

		if e := cmd.genrateControllerFromTable(&table); nil != e {
			log.Println(e)
			return
		}
	}
}

func (cmd *GenerateControllerCommand) genrateControllerFromTable(table *Table) error {
	controllerRoot := filepath.Join(cmd.root, "app", "controllers")
	if _, err := os.Stat(controllerRoot); os.IsNotExist(err) {
		os.MkdirAll(controllerRoot, 0)
	} else {
		fmt.Println(err)
	}

	controllerName := Pluralize(table.ClassName)
	uname := Underscore(controllerName)
	params := map[string]interface{}{
		"table":           table,
		"columns":         table.Columns,
		"controllerName":  controllerName,
		"controllerUName": uname,
		"dbPrefix":        cmd.dbPrefix,
	}

	return executeTempate(cmd.override, controllerTemplate, params, filepath.Join(controllerRoot, uname+".go"))
}

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
	"CamelCase":   CamelCase,
	"Underscore":  Underscore,
	"Pluralize":   Pluralize,
	"Singularize": Singularize,
	"Tableize":    Tableize,
	"Capitalize":  Capitalize,
	"Typeify":     Typeify,
	"ToUpper":     strings.ToUpper,
	"TrimPrefix":  strings.TrimPrefix,
	"ToNullType":  ToNullTypeFromPostgres,
}

var controllerTemplate = template.Must(template.New("default").Funcs(funcs).Parse(controllerText))

var controllerText = `{{$tableNoPrefix := TrimPrefix .table.TableName .dbPrefix}}{{$varName := Singularize $tableNoPrefix}}package controllers

import (
  "mc/app"
  "mc/app/libs"
  "mc/app/models"
  "mc/app/routes"

  "github.com/revel/revel"
)

// {{.controllerName}} - 控制器
type {{.controllerName}} struct {
  *revel.Controller
}

// 列出所有记录
func (c {{.controllerName}}) Index(pageIndex uint64) revel.Result {
  //var exprs []models.Expr
  //if "" != name {
  //  exprs = append(exprs, models.{{.table.ClassName}}Model.C.NAME.LIKE("%"+name+"%"))
  //}

  total, err := models.{{.table.ClassName}}Model.Count(app.DbRunner) // exprs...)

  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Render(err)
  }

  {{$tableNoPrefix}}, err := models.{{.table.ClassName}}Model.QueryWith(app.DbRunner, models.{{.table.ClassName}}Model.Where(). //exprs...).
    Select().Limit(libs.DEFAULT_SIZE_PER_PAGE).Offset(pageIndex*libs.DEFAULT_SIZE_PER_PAGE))
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Render()
  }
  paginator := libs.NewPaginator(c.Request.Request, libs.DEFAULT_SIZE_PER_PAGE, total)
  return c.Render({{$tableNoPrefix}}, paginator)
}

// 编辑新建记录
func (c {{.controllerName}}) New() revel.Result {
  return c.Render()
}

// 创建记录
func (c {{.controllerName}}) Create({{$varName}} *models.{{.table.ClassName}}) revel.Result {
  if {{$varName}}.Validate(c.Validation) {
    c.Validation.Keep()
    c.FlashParams()
    return c.Redirect(routes.{{.controllerName}}.New())
  }

  _, err := {{$varName}}.CreateIt(app.DbRunner)
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Redirect(routes.{{.controllerName}}.New())
  }

  c.Flash.Success(revel.Message(c.Request.Locale, "insert.success"))
  return c.Redirect(routes.{{.controllerName}}.Index(0, ""))
}

// 编辑指定 id 的记录
func (c {{.controllerName}}) Edit(id int64) revel.Result {
  {{$varName}}, err := models.{{.table.ClassName}}s.FindByID(app.DbRunner, id)
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Redirect(routes.{{.controllerName}}.Index(0, ""))
  }
  return c.Render({{$varName}})
}


// 按 id 更新记录
func (c {{.controllerName}}) Update({{$varName}} *models.{{.table.ClassName}}) revel.Result {
  if {{$varName}}.Validate(c.Validation) {
    c.Validation.Keep()
    c.FlashParams()
    return c.Redirect(routes.{{.controllerName}}.Edit(int64({{$varName}}.Id)))
  }

  err := {{$varName}}.UpdateIt(app.DbRunner)
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Redirect(routes.{{.controllerName}}.Edit(int64({{$varName}}.Id)))
  }
  c.Flash.Success(revel.Message(c.Request.Locale, "update.success"))
  return c.Redirect(routes.{{.controllerName}}.Index(0, ""))
}

// 按 id 删除记录
func (c {{.controllerName}}) Delete(id int64) revel.Result {
  {{$varName}} := &models.{{.table.ClassName}}{Id: id}
  err := {{$varName}}.DeleteIt(app.DbRunner)
  if nil != err {
    c.Flash.Error(err.Error())
  } else {
    c.Flash.Success(revel.Message(c.Request.Locale, "delete.success"))
  }
  return c.Redirect({{.controllerName}}.Index)
}

// 按 id 列表删除记录
func (c {{.controllerName}}) DeleteByIDs(id_list []int64) revel.Result {
  _, err := models.{{.table.ClassName}}Model.Delete(app.DbRunner, models.{{.table.ClassName}}Model.C.ID.IN(id_list))
  if nil != err {
    return c.RenderError(err)
  } else {
    return c.RenderJson(revel.Message(c.Request.Locale, "delete.success"))
  }
}
`
