package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// GenerateViewCommand - 生成视图
type GenerateViewCommand struct {
	generateBase
}

// Run - 生成代码
func (cmd *GenerateViewCommand) Run(args []string) {
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

		if e := cmd.genrateViewsFromTable(&table); nil != e {
			log.Println(e)
			return
		}
	}
}

func (cmd *GenerateViewCommand) genrateViewsFromTable(table *Table) error {
	controllerName := Pluralize(table.ClassName)
	viewRoot := filepath.Join(cmd.root, "app", "views", controllerName)
	if _, err := os.Stat(viewRoot); os.IsNotExist(err) {
		os.MkdirAll(viewRoot, 0)
	} else {
		fmt.Println(err)
	}
	uname := Underscore(controllerName)
	params := map[string]interface{}{
		"table":           table,
		"columns":         table.Columns,
		"controllerName":  controllerName,
		"controllerUName": uname,
		"dbPrefix":        cmd.dbPrefix,
	}
	err := executeTempate(cmd.override, indexTemplate, params, filepath.Join(viewRoot, "index.html"))
	if err != nil {
		return err
	}

	err = executeTempate(cmd.override, editTemplate, params, filepath.Join(viewRoot, "edit.html"))
	if err != nil {
		os.Remove(filepath.Join(viewRoot, "index.html"))
		return err
	}

	err = executeTempate(cmd.override, fieldsTemplate, params, filepath.Join(viewRoot, "edit_fields.html"))
	if err != nil {
		os.Remove(filepath.Join(viewRoot, "index.html"))
		os.Remove(filepath.Join(viewRoot, "edit.html"))
		return err
	}

	err = executeTempate(cmd.override, newTemplate, params, filepath.Join(viewRoot, "new.html"))
	if err != nil {
		os.Remove(filepath.Join(viewRoot, "index.html"))
		os.Remove(filepath.Join(viewRoot, "edit.html"))
		os.Remove(filepath.Join(viewRoot, "edit_fields.html"))
		return err
	}

	err = executeTempate(cmd.override, quickTemplate, params, filepath.Join(viewRoot, "quick-bar.html"))
	if err != nil {
		os.Remove(filepath.Join(viewRoot, "index.html"))
		os.Remove(filepath.Join(viewRoot, "edit.html"))
		os.Remove(filepath.Join(viewRoot, "edit_fields.html"))
		os.Remove(filepath.Join(viewRoot, "quick-bar.html"))
		return err
	}
	return nil
}

func executeTempate(override bool, tpl *template.Template, params interface{}, fname string) error {
	var out *os.File
	var err error

	if !override {
		out, err = os.OpenFile(fname, os.O_CREATE|os.O_EXCL, 0)
	} else {
		out, err = os.OpenFile(fname, os.O_CREATE|os.O_TRUNC, 0)
	}
	if nil != err {
		return err
	}
	defer out.Close()

	if err := tpl.Execute(out, params); err != nil {
		out.Close()
		os.Remove(fname)
		return err
	}
	return nil
}

var indexTemplate = template.Must(template.New("default").Delims("[[", "]]").Funcs(funcs).Parse(indexText))

var indexText = `[[$tableNoPrefix := TrimPrefix .table.TableName .dbPrefix]][[$varName := Singularize $tableNoPrefix]]
{{set . "title" "调度[[$varName]]"}}
{{append . "moreScripts" "/public/js/[[$varName]].js"}}
{{template "layouts/header.html" .}}

<div class="widget stacked">
  <div class="gui-list">
    {{template "[[.controllerName]]/quick-bar.html" .}}
    <table class="table table-bordered table-striped table-highlight ">
      <thead>
      <tr>
        <th><input type="checkbox" class="all-checker"></th>[[range $x := .columns]]
        <th><nobr> [[$x.DbName]] </nobr></th>[[end]]
      </tr>
      </thead>
      {{range $k,$v:=.[[$tableNoPrefix]]}}
      <tr>
        <td><input type="checkbox" class="row-checker" key="{{$v.Id}}" url="{{url "[[.controllerName]].Edit" $v.Id}}" id="row-checker"></td>
        [[range $column := .columns]]
        <td>{{$v.[[$column.GoName]]}}</td>[[end]]
      </tr>
      {{end}}
    </table>
    </div>
</div>

{{template "layouts/paginator.html" .}}
{{template "layouts/footer.html" .}}
`

var fieldsTemplate = template.Must(template.New("default").Delims("[[", "]]").Funcs(funcs).Parse(fieldsText))

var fieldsText = `[[$tableNoPrefix := TrimPrefix .table.TableName .dbPrefix]][[$varName := Singularize $tableNoPrefix]]
                  [[range $column := .columns]]{{text_field . "[[$varName]].[[$column.GoName]]" "[[$column.GoName]]:" | f_addClass "span5" | render}}
                  [[end]]`

var editTemplate = template.Must(template.New("default").Delims("[[", "]]").Funcs(funcs).Parse(editText))

var editText = `[[$tableNoPrefix := TrimPrefix .table.TableName .dbPrefix]][[$varName := Singularize $tableNoPrefix]]
{{set . "title" "编辑[[$varName]]"}}
{{append . "moreStyles" "/public/css/form.css"}}
{{append . "moreScripts" "/public/js/[[$varName]].js"}}
{{template "layouts/header.html" .}}
<div class="widget stacked">
    <div class="widget-header">
        <h3>编辑[[$varName]]</h3>
    </div>
    <div class="widget-content">
        <form action="{{url "[[.controllerName]].Update" }}" method="POST" class="form-horizontal" id="insert">
        <input type="hidden" name="_method" value="PUT">
        {{with $field := field "{{$varName}}.Id" .}}
        <input type="hidden" name="{{$field.Name}}" value="{{or $field.Flash $field.Value}}">
        {{end}}
        {{template "[[.controllerName]]/edit_fields.html" .}}
        <div class="controls-group">
            <div class="controls controls-row">
                <button type="submit" class="btn btn-info controls">修改</button>
                <button type="submit" class="btn btn-info controls">取消</button>
            </div>
        </div>
        </form>
    </div>
</div>
{{template "layouts/footer.html" .}}`

var newTemplate = template.Must(template.New("default").Delims("[[", "]]").Funcs(funcs).Parse(newText))

var newText = `[[$tableNoPrefix := TrimPrefix .table.TableName .dbPrefix]][[$varName := Singularize $tableNoPrefix]]
{{set . "title" "新建[[$varName]]"}}
{{append . "moreStyles" "/public/css/form.css"}}
{{append . "moreScripts" "/public/js/[[$varName]].js"}}
{{template "layouts/header.html" .}}
<div class="widget stacked">
    <div class="widget-header">
        <h3>新建[[$varName]]</h3>
    </div>
    <div class="widget-content">
            <form action="{{url "[[.controllerName]].Create"}}" method="post" class="form-horizontal" id="insert">
            {{template "[[.controllerName]]/edit_fields.html" .}}
            <div class="control-group">
                <div class="controls">
                    <button type="submit" class="btn btn-info">插入</button>
                </div>
            </div>
            </form>
    </div>
</div>
{{template "layouts/footer.html" .}}`

var quickTemplate = template.Must(template.New("default").Delims("[[", "]]").Funcs(funcs).Parse(quickText))

var quickText = `
<div class="quick-bar">
    <ul class="quick-actions ">
        <li>
            <a href='{{url "SchdJobs.New"}}'  class="grid-action" method="" mode="*" confirm="" client="false" target="_self">
                <i class="icon-add"></i>添加
            </a>
        </li>
        <li>
            <a  href="#"  class="grid-action update" method="" mode="1" confirm="" client="false" target="_self">
                <i class="icon-edit"></i>编辑
            </a>
        </li>
        <li>
            <a href="#"  class="grid-action delete" mode="+" target="_self">
                <i class="icon-delete"></i> 删除
            </a>
        </li>
    </ul>
    <ul class="quick-actions ">
        <form action="" method="get" class="form-action" id="form" >
            <li>
                <label>
                    <span>名称</span><input type="text" name="name">
                </label>
            </li>
            <li>
                <a href="javascript:document.getElementById('form').submit()" class="grid-action" method="" mode="*">
                    <i class="icon-search"></i> 查询
                </a>
            </li>
        </form>
    </ul>
</div>
`
