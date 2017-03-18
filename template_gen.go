package main

import "errors"

var embededText = `package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/Masterminds/squirrel"
	"github.com/lann/builder"
)

// ErrNotUpdated - 表示没有更新任何记录
var ErrNotUpdated = errors.New("no record is updated")

// ErrPrimaryKeyInvalid - 主 key 是无效的。
var ErrPrimaryKeyInvalid = errors.New("primary key is invalid")

// ErrNotDeleted - 表示没有删除任何记录
var ErrNotDeleted = errors.New("no record is deleted")

func isPostgersql(db interface{}) bool {
	return true
}

func isPlaceholderWithDollar(value interface{}) bool {
	return true
}

// Fields 代表多个字段和值
type Fields map[string]interface{}

// ThrowUpdateFailWithPrimaryKeyInvalid 返回一个 主键无效的错误
func ThrowUpdateFailWithPrimaryKeyInvalid(tableName string) error {
	return errors.New("update fail becase primary key of '" + tableName + "' is invalid")
}

// ThrowDeleteFailWithPrimaryKeyInvalid 返回一个 主键无效的错误
func ThrowDeleteFailWithPrimaryKeyInvalid(tableName string) error {
	return errors.New("delete fail becase primary key of '" + tableName + "' is invalid")
}

// JSON 代表一个数据库中一个 json
type JSON []byte

// String 将字节数组转成一个 JSON 对象
func (js JSON) String() string {
	if len(js) == 0 {
		return "{}"
	}
	return string(js)
}

// MarshalJSON returns *m as the JSON encoding of m.
func (js *JSON) MarshalJSON() ([]byte, error) {
	if len(*js) == 0 {
		return []byte("{}"), nil
	}
	return *js, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (js *JSON) UnmarshalJSON(data []byte) error {
	if js == nil {
		return errors.New("models.JSON: UnmarshalJSON on nil pointer")
	}
	*js = append((*js)[0:0], data...)
	return nil
}

// ToJSON 将字节数组转成一个 JSON 对象, 注意它将不拷贝字节数据
func ToJSON(bs []byte) JSON {
	return JSON(bs)
}

// ToJSONCopy 将字节数组转成一个 JSON 对象, 注意它将拷贝字节数据
func ToJSONCopy(bs []byte) JSON {
	data := make([]byte, len(bs))
	copy(data, bs)
	return JSON(data)
}

// ToJSONFromAny 将一个对象转成一个 JSON 对象
func ToJSONFromAny(v interface{}) JSON {
	bs, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return JSON(bs)
}

// ViewModel - 数据库视图模型
type ViewModel struct {
	TableName   string
	ColumnNames []string
}

// Count - 统计符合条件的记录数
func (viewModel *ViewModel) Count(db squirrel.QueryRower, exprs ...Sqlizer) (count int64, err error) {
	selectBuilder := viewModel.Where(exprs...).Select("count(*)").From(viewModel.TableName)
	if isPlaceholderWithDollar(db) {
		selectBuilder = selectBuilder.PlaceholderFormat(squirrel.Dollar)
	}
	err = squirrel.QueryRowWith(db, selectBuilder).Scan(&count)
	return
}

// Where - 生成查询语句， 如 Where(UserModel.C.NAME.EQU('小明')).Select()
func (viewModel *ViewModel) Where(exprs ...Sqlizer) squirrel.StatementBuilderType {
	if len(exprs) == 0 {
		return squirrel.StatementBuilder
	}
	if len(exprs) == 1 {
		return builder.Append(squirrel.StatementBuilder, "WhereParts", exprs[0]).(squirrel.StatementBuilderType)
	}
	sqlizers := make([]squirrel.Sqlizer, 0, len(exprs))
	for _, exp := range exprs {
		sqlizers = append(sqlizers, exp)
	}

	return builder.Append(squirrel.StatementBuilder, "WhereParts", squirrel.And(sqlizers)).(squirrel.StatementBuilderType)
}

func (viewModel *ViewModel) Update(db squirrel.BaseRunner, values map[string]interface{}, args ...squirrel.Sqlizer) (int64, error) {
	sql := squirrel.Update(viewModel.TableName)
	if isPlaceholderWithDollar(db) {
		sql = sql.PlaceholderFormat(squirrel.Dollar)
	}

	for key, value := range values {
		sql = sql.Set(key, value)
	}

	sql = sql.Where(squirrel.And(args))

	result, e := sql.RunWith(db).Exec()
	if nil != e {
		return 0, e
	}
	return result.RowsAffected()
}

func (viewModel *ViewModel) UpdateBy(db squirrel.BaseRunner, values map[string]interface{}, pred interface{}, args ...interface{}) (int64, error) {
	sql := squirrel.Update(viewModel.TableName)
	if isPlaceholderWithDollar(db) {
		sql = sql.PlaceholderFormat(squirrel.Dollar)
	}

	for key, value := range values {
		sql = sql.Set(key, value)
	}

	sql = sql.Where(pred, args)

	result, e := sql.RunWith(db).Exec()
	if nil != e {
		return 0, e
	}
	return result.RowsAffected()
}

func (viewModel *ViewModel) Delete(db squirrel.BaseRunner, exprs ...Sqlizer) (int64, error) {
	sq := viewModel.Where(exprs...).Delete(viewModel.TableName)
	if isPlaceholderWithDollar(db) {
		sq = sq.PlaceholderFormat(squirrel.Dollar)
	}
	result, e := sq.RunWith(db).Exec()
	if nil != e {
		return 0, e
	}
	return result.RowsAffected()
}

func (viewModel *ViewModel) DeleteBy(db squirrel.BaseRunner, pred interface{}, args ...interface{}) (int64, error) {
	sq := squirrel.Delete(viewModel.TableName).Where(pred, args)
	if isPlaceholderWithDollar(db) {
		sq = sq.PlaceholderFormat(squirrel.Dollar)
	}

	result, e := sq.RunWith(db).Exec()
	if nil != e {
		return 0, e
	}
	return result.RowsAffected()
}

type DbModel struct {
	ViewModel
	KeyNames []string
}

func (dbModel *DbModel) UpdateByPrimaryKey(db squirrel.BaseRunner, values map[string]interface{}, keys ...interface{}) error {
	sql := squirrel.Update(dbModel.TableName)
	if isPlaceholderWithDollar(db) {
		sql = sql.PlaceholderFormat(squirrel.Dollar)
	}

	for key, value := range values {
		sql = sql.Set(key, value)
	}

	cond := squirrel.Eq{}
	for idx, key := range keys {
		cond[dbModel.KeyNames[idx]] = key
	}
	sql = sql.Where(cond)

	result, e := sql.RunWith(db).Exec()
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

func (dbModel *DbModel) DeleteByPrimaryKey(db squirrel.BaseRunner, keys ...interface{}) error {
	sql := squirrel.Delete(dbModel.TableName)
	if isPlaceholderWithDollar(db) {
		sql = sql.PlaceholderFormat(squirrel.Dollar)
	}
	cond := squirrel.Eq{}
	for idx, key := range keys {
		cond[dbModel.KeyNames[idx]] = key
	}

	result, e := sql.Where(cond).RunWith(db).Exec()
	if nil != e {
		return e
	}
	rowsAffected, e := result.RowsAffected()
	if nil != e {
		return e
	}

	if 0 == rowsAffected {
		return ErrNotDeleted
	}
	return nil
}

type ColumnModel struct {
	Name string
}

func (model *ColumnModel) Field(name string) *columnModel {
	column := &columnModel{origin: model}
	return column.Field(name)
}

func (model *ColumnModel) TableAlias(alias string) *columnModel {
	column := &columnModel{origin: model}
	return column.TableAlias(alias)
}

func (model *ColumnModel) IsNULL() Expr {
	column := &columnModel{origin: model}
	return column.IsNULL()
}

func (model *ColumnModel) IsNotNULL() Expr {
	column := &columnModel{origin: model}
	return column.IsNotNULL()
}

func (model *ColumnModel) EQU(value interface{}) Expr {
	column := &columnModel{origin: model}
	return column.EQU(value)
}

func (model *ColumnModel) GT(value interface{}) Expr {
	column := &columnModel{origin: model}
	return column.GT(value)
}

func (model *ColumnModel) GTE(value interface{}) Expr {
	column := &columnModel{origin: model}
	return column.GTE(value)
}

func (model *ColumnModel) LT(value interface{}) Expr {
	column := &columnModel{origin: model}
	return column.LT(value)
}

func (model *ColumnModel) LTE(value interface{}) Expr {
	column := &columnModel{origin: model}
	return column.LTE(value)
}

func (model *ColumnModel) IN(values ...interface{}) Expr {
	column := &columnModel{origin: model}
	return column.IN(values...)
}

func (model *ColumnModel) NEQ(value interface{}) Expr {
	column := &columnModel{origin: model}
	return column.NEQ(value)
}

func (model *ColumnModel) EXISTS(value interface{}) Expr {
	column := &columnModel{origin: model}
	return column.EXISTS(value)
}

func (model *ColumnModel) LIKE(value string) Expr {
	column := &columnModel{origin: model}
	return column.LIKE(value)
}

func (model *ColumnModel) Search(lang, value string) Expr {
	column := &columnModel{origin: model}
	return column.Search(lang, value)
}

type columnModel struct {
	origin     *ColumnModel
	subField   string
	tableAlias string
}

func (column *columnModel) Field(name string) *columnModel {
	column.subField = name
	return column
}

func (column *columnModel) TableAlias(alias string) *columnModel {
	column.tableAlias = alias
	return column
}

func (column *columnModel) Name() string {
	if "" == column.tableAlias {
		if column.subField == "" {
			return column.origin.Name
		}
		return column.origin.Name + "->>'" + column.subField + "'"
	}

	if column.subField == "" {
		return column.tableAlias + "." + column.origin.Name
	}
	return column.tableAlias + "." + column.origin.Name + "->>'" + column.subField + "'"
}

func (model *columnModel) IsNULL() Expr {
	return Expr{Column: model, Operator: "IS", Value: "NULL"}
}

func (model *columnModel) IsNotNULL() Expr {
	return Expr{Column: model, Operator: "IS", Value: "NOT NULL"}
}

func (model *columnModel) EQU(value interface{}) Expr {
	return Expr{Column: model, Operator: "=", Value: value}
}

func (model *columnModel) GT(value interface{}) Expr {
	return Expr{Column: model, Operator: ">", Value: value}
}

func (model *columnModel) GTE(value interface{}) Expr {
	return Expr{Column: model, Operator: ">=", Value: value}
}

func (model *columnModel) LT(value interface{}) Expr {
	return Expr{Column: model, Operator: "<", Value: value}
}

func (model *columnModel) LTE(value interface{}) Expr {
	return Expr{Column: model, Operator: "<=", Value: value}
}

func (model *columnModel) IN(values ...interface{}) Expr {
	if len(values) == 0 {
		panic(errors.New("values is empty"))
	}

	if len(values) == 1 {
		val := reflect.ValueOf(values[0])
		if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
			if val.Len() == 0 {
				panic(errors.New("values is empty"))
			}
		}
	}
	return Expr{Column: model, Operator: "IN", Value: values}
}

func (model *columnModel) NEQ(value interface{}) Expr {
	return Expr{Column: model, Operator: "<>", Value: value}
}

func (model *columnModel) EXISTS(value interface{}) Expr {
	return Expr{Column: model, Operator: "EXISTS", Value: value}
}

func (model *columnModel) LIKE(value string) Expr {
	return Expr{Column: model, Operator: "LIKE", Value: value}
}

func (model *columnModel) Search(lang, value string) Expr {
	if "" == lang {
		lang = "english"
	}
	return Expr{Column: model, Operator: "@@", Value: "plainto_tsquery('" + lang + "','" + value + "')"}
}

type Expr struct {
	Column   *columnModel
	Operator string
	Value    interface{}
}

func (model Expr) ToSql() (string, []interface{}, error) {
	if sqlizer, ok := model.Value.(squirrel.Sqlizer); ok {
		subSqlstr, subArgs, e := sqlizer.ToSql()
		if nil != e {
			return "", nil, e
		}
		return model.Column.Name() + " " + model.Operator + " " + subSqlstr, subArgs, nil
	}

	if "IS" == model.Operator {
		return model.Column.Name() + " IS " + fmt.Sprint(model.Value), nil, nil
	}
	if "IN" == model.Operator {
		var buf bytes.Buffer
		buf.WriteString(model.Column.Name())
		buf.WriteString(" IN (")
		oldLength := buf.Len()
		JoinObjects(&buf, model.Value)
		if oldLength == buf.Len() {
			return "", nil, errors.New("ToSql: values is empty in the 'IN' case.")
		}
		buf.Truncate(buf.Len() - 1)
		buf.WriteString(") ")
		return buf.String(), nil, nil
	}
	return model.Column.Name() + " " + model.Operator + " ? ", []interface{}{model.Value}, nil
}

type existsExpr struct {
	sqlizer squirrel.Sqlizer
}

func (exists existsExpr) ToSql() (string, []interface{}, error) {
	subSqlstr, subArgs, e := exists.sqlizer.ToSql()
	if nil != e {
		return "", nil, e
	}
	return "EXISTS (" + subSqlstr + " )", subArgs, nil
}

func EXISTS(sqlizer squirrel.Sqlizer) *existsExpr {
	return &existsExpr{sqlizer: sqlizer}
}

func JoinObjects(buf *bytes.Buffer, value interface{}) {
	if inner, ok := value.([]interface{}); ok {
		for _, v := range inner {
			JoinObjects(buf, v)
		}
	} else if inner, ok := value.([]uint); ok {
		for _, v := range inner {
			buf.WriteString(fmt.Sprint(v))
			buf.WriteString(",")
		}
	} else if inner, ok := value.([]int); ok {
		for _, v := range inner {
			buf.WriteString(fmt.Sprint(v))
			buf.WriteString(",")
		}
	} else if inner, ok := value.([]uint64); ok {
		for _, v := range inner {
			buf.WriteString(fmt.Sprint(v))
			buf.WriteString(",")
		}
	} else if inner, ok := value.([]int64); ok {
		for _, v := range inner {
			buf.WriteString(fmt.Sprint(v))
			buf.WriteString(",")
		}
	} else {
		valVal := reflect.ValueOf(value)
		if valVal.Kind() == reflect.Array || valVal.Kind() == reflect.Slice {
			for i := 0; i < valVal.Len(); i++ {
				buf.WriteString(fmt.Sprint(valVal.Index(i).Interface()))
				buf.WriteString(",")
			}
		} else {
			buf.WriteString(fmt.Sprint(value))
			buf.WriteString(",")
		}
	}
}

// Sqlizer is the interface that wraps the ToSql method.
//
// ToSql returns a SQL representation of the Sqlizer, along with a slice of args
// as passed to e.g. database/sql.Exec. It can also return an error.
type Sqlizer interface {
	ToSql() (string, []interface{}, error)
}

// Execer is the interface that wraps the Exec method.
//
// Exec executes the given query as implemented by database/sql.Exec.
type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// Queryer is the interface that wraps the Query method.
//
// Query executes the given query as implemented by database/sql.Query.
type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// QueryRower is the interface that wraps the QueryRow method.
//
// QueryRow executes the given query as implemented by database/sql.QueryRow.
type QueryRower interface {
	QueryRow(query string, args ...interface{}) RowScanner
}

// BaseRunner groups the Execer and Queryer interfaces.
type BaseRunner interface {
	Execer
	Queryer
}

// Runner groups the Execer, Queryer, and QueryRower interfaces.
type Runner interface {
	Execer
	Queryer
	QueryRower
}

// RowScanner is the interface that wraps the Scan method.
//
// Scan behaves like database/sql.Row.Scan.
type RowScanner interface {
	Scan(...interface{}) error
}

// Row wraps database/sql.Row to let squirrel return new errors on Scan.
type Row struct {
	RowScanner
	err error
}

// Scan returns Row.err or calls RowScanner.Scan.
func (r *Row) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return r.RowScanner.Scan(dest...)
}
`

var ns = `package [[.namespace]]
`

var handler = ``

var structText = `
type [[.class.Name]] struct {[[range $field := .class.Fields ]]
  [[goify $field.Name  true]] [[gotype $field.Type]] ` + "`" + `json:"[[underscore $field.Name]][[if omitempty $field]],omitempty[[end]]" xorm:"[[underscore $field.Name]][[if eq $field.Name "id"]] pk autoincr[[else if eq $field.Name "created_at"]] created[[else if eq $field.Name "updated_at"]] updated[[end]][[if $field.IsUniquely]][[if ne $field.Name "id"]] unique[[end]][[end]][[if $field.DefaultValue]] default('[[$field.DefaultValue]]')[[end]]"` + "`" + `[[end]]
}


func ([[camelizeDownFirst .class.Name]] *[[.class.Name]]) TableName() string {
  return "[[if .class.Table]][[.class.Table]][[else]][[pluralize .class.Name | underscore]][[end]]"
}

func ([[camelizeDownFirst .class.Name]] *[[.class.Name]]) Validate(validation *revel.Validation) bool {[[$var := camelizeDownFirst .class.Name]]
[[range $column := .class.Fields]]
[[if ne $column.Name "id"]][[if $column.IsRequired]]
  validation.Required([[$var]].[[goify $column.Name true]]).Key("[[$var]].[[goify $column.Name true]]")
  [[else if $column.Restrictions]][[if $column.Restrictions.MinLength]]
             validation.MinSize([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MinLength]]).Key("[[$var]].[[goify $column.Name true]]")
        [[end]][[if $column.Restrictions.MaxLength]]
             validation.MaxSize([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MaxLength]]).Key("[[$var]].[[goify $column.Name true]]")
        [[end]][[if $column.Restrictions.Length]]
             validation.Length([[$var]].[[goify $column.Name true]], [[$column.Restrictions.Length]]).Key("[[$var]].[[goify $column.Name true]]")
        [[end]][[if $column.Restrictions.MaxValue]][[if $column.Restrictions.MinValue]]
             validation.Range([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MinValue]], [[$column.Restrictions.MaxValue]]).Key("[[$var]].[[goify $column.Name true]]")
          [[else]]
             validation.Max([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MaxValue]]).Key("[[$var]].[[goify $column.Name true]]")
        [[end]][[else if $column.Restrictions.MinValue]]
             validation.Min([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MinValue]]).Key("[[$var]].[[goify $column.Name true]]")
        [[end]][[end]]
[[end]][[end]]
  return validation.HasErrors()
}`

var controllerText = `import (
  "[[.projectPath]]/app"
  "[[.projectPath]]/app/libs"
  "[[.projectPath]]/app/models"
  "[[.projectPath]]/app/routes"

  "github.com/revel/revel"
  "github.com/runner-mei/orm"
  "upper.io/db.v3"
)

// [[.controllerName]] - 控制器
type [[.controllerName]] struct {
  [[if .baseController]][[.baseController]][[else]]*revel.Controller[[end]]
}

// 列出所有记录
func (c [[.controllerName]]) Index(pageIndex int) revel.Result {
  //var exprs []db.Expr
  //if "" != name {
  //  exprs = append(exprs, models.[[.class.Name]]s.C.NAME.LIKE("%"+name+"%"))
  //}


  total, err := c.Lifecycle.DB.[[.modelName]]().Where().Count()
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Render(err)
  }

  var [[camelizeDownFirst .modelName]] []models.[[.class.Name]]
  err = c.Lifecycle.DB.[[.modelName]]().Where().
    Limit(libs.DEFAULT_SIZE_PER_PAGE).
    Offset(pageIndex * libs.DEFAULT_SIZE_PER_PAGE).
    All(&[[camelizeDownFirst .modelName]])
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Render()
  }
  paginator := libs.NewPaginator(c.Request.Request, libs.DEFAULT_SIZE_PER_PAGE, total)
  return c.Render([[camelizeDownFirst .modelName]], paginator)
}

// 编辑新建记录
func (c [[.controllerName]]) New() revel.Result {
  return c.Render()
}

// 创建记录
func (c [[.controllerName]]) Create([[camelizeDownFirst .class.Name]] *models.[[.class.Name]]) revel.Result {
  if [[camelizeDownFirst .class.Name]].Validate(c.Validation) {
    c.Validation.Keep()
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].New())
  }

  _, err := c.Lifecycle.DB.[[.modelName]]().Insert([[camelizeDownFirst .class.Name]])
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].New())
  }

  c.Flash.Success(revel.Message(c.Request.Locale, "insert.success"))
  return c.Redirect(routes.[[.controllerName]].Index(0))
}

// 编辑指定 id 的记录
func (c [[.controllerName]]) Edit(id int64) revel.Result {
  var [[camelizeDownFirst .class.Name]] models.[[.class.Name]]
  err := c.Lifecycle.DB.[[.modelName]]().Id(id).Get(&[[camelizeDownFirst .class.Name]])
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].Index(0))
  }
  return c.Render([[camelizeDownFirst .class.Name]])
}


// 按 id 更新记录
func (c [[.controllerName]]) Update([[camelizeDownFirst .class.Name]] *models.[[.class.Name]]) revel.Result {
  if [[camelizeDownFirst .class.Name]].Validate(c.Validation) {
    c.Validation.Keep()
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].Edit(int64([[camelizeDownFirst .class.Name]].ID)))
  }

  err := c.Lifecycle.DB.[[.modelName]]().Id([[camelizeDownFirst .class.Name]].ID).Update([[camelizeDownFirst .class.Name]])
  if err != nil {
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].Edit(int64([[camelizeDownFirst .class.Name]].ID)))
  }
  c.Flash.Success(revel.Message(c.Request.Locale, "update.success"))
  return c.Redirect(routes.[[.controllerName]].Index(0))
}

// 按 id 删除记录
func (c [[.controllerName]]) Delete(id int64) revel.Result {
  err :=  c.Lifecycle.DB.[[.modelName]]().Id(id).Delete()
  if nil != err {
    c.Flash.Error(err.Error())
  } else {
    c.Flash.Success(revel.Message(c.Request.Locale, "delete.success"))
  }
  return c.Redirect([[.controllerName]].Index)
}

// 按 id 列表删除记录
func (c [[.controllerName]]) DeleteByIDs(id_list []int64) revel.Result {
  if len(id_list) == 0 {
    c.Flash.Error("请至少选择一条记录！")
    return c.Redirect([[.controllerName]].Index)
  }
  _, err :=  c.Lifecycle.DB.[[.modelName]]().Where().And(orm.Cond{"id IN": id_list}).Delete()
  if nil != err {
    c.Flash.Error(err.Error())
  } else {
    c.Flash.Success(revel.Message(c.Request.Locale, "delete.success"))
  }
  return c.Redirect([[.controllerName]].Index)
}`

var viewEditText = `{{set . "title" "编辑[[.controllerName]]"}}
{{append . "moreStyles" "/public/css/form.css"}}
{{append . "moreScripts" "/public/js/[[underscore .controllerName]]/[[underscore .controllerName]].js"}}
{{template "[[if .layouts]][[.layouts]][[end]]header.html" .}}
<div class="widget stacked">
    <div class="widget-header">
        <h3>编辑[[.controllerName]]</h3>
    </div>
    <div class="widget-content">
        <form action="{{url "[[.controllerName]].Update" }}" method="POST" class="form-horizontal" id="insert">
        <input type="hidden" name="_method" value="PUT">
        {{with $field := field "[[camelizeDownFirst .class.Name]].ID" .}}<input type="hidden" name="{{$field.Name}}" value="{{if $field.Flash}}{{$field.Flash}}{{else}}{{$field.Value}}{{end}}">{{end}}
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
{{template "[[if .layouts]][[.layouts]][[end]]footer.html" .}}`

var viewFieldsText = `[[define "lengthLimit"]][[end]][[$instaneName := camelizeDownFirst .class.Name]] [[range $column := .class.Fields]][[if isID $column]][[else if eq $column.Type "string" ]][[if isClob $column ]]{{textarea_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" | [[template "lengthLimit" $column]] render}}
[[else]]{{text_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" | render}}
[[end]][[else if eq $column.Type "integer" "number" "biginteger" ]]{{number_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" | render}}
[[else if eq $column.Type "boolean" "bool" ]]{{checkbox_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" | render}}
[[else if eq $column.Type "password" ]]{{password_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" | render}}
[[else if editDisabled $column]][[end]][[end]]`

var viewIndexText = `{{set . "title" "[[.controllerName]]"}}
{{append . "moreScripts" "/public/js/[[underscore .controllerName]]/[[underscore .controllerName]].js"}}
{{template "[[if .layouts]][[.layouts]][[end]]header.html" .}}

<div class="widget stacked">
  <div class="gui-list">
    {{template "[[.controllerName]]/quick-bar.html" .}}
    <table class="table table-bordered table-striped table-highlight ">
      <thead>
      <tr>
        <th><input type="checkbox" class="all-checker"></th>[[range $field := .class.Fields]][[if needDisplay $field]]
        <th><nobr>[[localizeName $field]]</nobr></th>[[end]][[end]]
        <th>操作</th>
      </tr>
      </thead>
      {{range $v := .[[camelizeDownFirst .modelName]]}}
      <tr>
        <td><input type="checkbox" class="row-checker" key="{{$v.ID}}" url="{{url "[[.controllerName]].Edit" $v.ID}}" id="row-checker"></td>
        [[range $column := .class.Fields]][[if needDisplay $column]]
        <td>{{$v.[[goify $column.Name true]]}}</td>[[end]][[end]]
        <td>
          <a href='{{url "[[.controllerName]].Edit" $v.ID}}'>编辑</a>
          <form id='[[underscore .controllerName]]-delete-{{$v.ID}}' action="{{url "[[.controllerName]].Delete"  $v.ID}}" method="POST" class="form-horizontal">
            <input type="hidden" name="_method" value="DELETE">
            <input type="hidden" name="id" value="{{$v.ID}}">
              <a href="javascript:document.getElementById('[[underscore .controllerName]]-delete-{{$v.ID}}').submit()">
                <i class="icon-search"></i> 删除
              </a>
            </form>
        </td>
      </tr>
      {{end}}
    </table>
    {{template "[[if .layouts]][[.layouts]][[end]]paginator.html" .}}
    </div>
</div>

{{template "[[if .layouts]][[.layouts]][[end]]footer.html" .}}`

var viewNewText = `{{set . "title" "新建[[.controllerName]]"}}
{{append . "moreStyles" "/public/css/form.css"}}
{{append . "moreScripts" "/public/js/[[underscore .controllerName]]/[[underscore .controllerName]].js"}}
{{template "[[if .layouts]][[.layouts]][[end]]header.html" .}}
<div class="widget stacked">
    <div class="widget-header">
        <h3>新建[[.controllerName]]</h3>
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
{{template "[[if .layouts]][[.layouts]][[end]]footer.html" .}}`

var viewQuickText = `<div class="quick-bar">
    <ul class="quick-actions ">
        <li>
            <a id='[[underscore .controllerName]]-new' href='{{url "[[.controllerName]].New"}}'  class="grid-action" method="" mode="*" confirm="" client="false" target="_self">
                <i class="icon-add"></i>添加
            </a>
        </li>
        <li>
            <a id='[[underscore .controllerName]]-edit' href='' url='{{url "[[.controllerName]].Edit"}}'  class="grid-action update" method="" mode="1" confirm="" client="false" target="_self">
                <i class="icon-edit"></i>编辑
            </a>
        </li>
        <li>
            <a id='[[underscore .controllerName]]-delete' href='' url='{{url "[[.controllerName]].DeleteByIDs"}}'  class="grid-action delete" mode="+" target="_self">
                <i class="icon-delete"></i> 删除
            </a>
        </li>
    </ul>
    <!--
    <ul class="quick-actions ">
        <form action="" method="get" class="form-action" id="[[underscore .controllerName]]-quick-form" >
            <li>
                <label>
                    <span>名称</span><input type="text" name="name">
                </label>
            </li>
            <li>
                <a href="javascript:document.getElementById('[[underscore .controllerName]]-quick-form').submit()" class="grid-action" method="" mode="*">
                    <i class="icon-search"></i> 查询
                </a>
            </li>
        </form>
    </ul>-->
</div>`

var viewJsText = `$(function () {
    var urlPrefix = $("#urlPrefix").val();
    $("#[[underscore .controllerName]]-delete").on("click", function () {
        if (!confirm("确认删除选定信息？")){
            return false
        }

        var f = document.createElement("form");
        f.action = $("#[[underscore .controllerName]]-delete").attr("url");
        f.method="POST";
        var inputField = document.createElement("input");
        inputField.type = "hidden";
        inputField.name = "_method";
        inputField.value = "DELETE";

        $("#row-checker:checked").each(function (i) {
            var inputField = document.createElement("input");
            inputField.type = "hidden";
            inputField.name = "id_list[]";
            inputField.value = $(this).attr("key");
            f.appendChild(inputField);
        });

        document.body.appendChild(f);
        f.submit();
        return false
    });

    $("#[[underscore .controllerName]]-edit").on("click", function () {
        var elements = $("#row-checker:checked");
        if (elements.length == 1) {
            window.location.href= elements.first().attr("url");
        } else if (elements.length == 0) {
            alert('请选择一条记录！')
        } else {
            alert('你选择了多条记录，请选择一条记录！')
        }

        return false
    });
});
`

var dbText = `
import (
  "github.com/go-xorm/xorm"
  "github.com/runner-mei/orm"
)

type DB struct {
  Engine *xorm.Engine
}

[[range $class := .classes]]
func (db *DB) [[pluralize $class.Name]]() *orm.Collection {
  return orm.New(func() interface{}{
    return &[[$class.Name]]{}
  })(db.Engine)
}
[[end]]



func InitTables(engine *xorm.Engine) error {
  beans := []interface{}{[[range $class := .classes]]
  &[[$class.Name]]{},
[[end]]}
  return engine.CreateTables(beans...)
}
`

func textDefault(nm string) []byte {
	switch nm {
	case "base":
		return []byte(embededText)
	case "ns":
		return []byte(ns)
	case "handler":
		return []byte(handler)
	case "struct":
		return []byte(structText)
	case "controller":
		return []byte(controllerText)
	case "views/edit":
		return []byte(viewEditText)
	case "views/fields":
		return []byte(viewFieldsText)
	case "views/index":
		return []byte(viewIndexText)
	case "views/new":
		return []byte(viewNewText)
	case "views/quick":
		return []byte(viewQuickText)
	case "views/js":
		return []byte(viewJsText)
	case "db":
		return []byte(dbText)
	default:
		panic(errors.New("template '" + nm + "' isn't default template."))
	}
}