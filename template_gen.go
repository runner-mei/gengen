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

var structText = `[[$var := camelizeDownFirst .class.Name]][[$class := .class]]
type [[.class.Name]] struct {
[[- range $field := .class.Fields ]]
  [[goify $field.Name  true]] [[gotype $field.Type]] ` + "`" + `json:"[[underscore $field.Name]]
  [[- if omitempty $field]],omitempty[[end -]]
  " xorm:"[[underscore $field.Name]]
  [[- if eq $field.Name "id"]] pk autoincr
  [[- else if eq $field.Name "created_at"]] created
  [[- else if eq $field.Name "updated_at"]] updated
  [[- end -]]
  [[- if $field.IsUniquely -]]
    [[- if ne $field.Name "id"]] unique[[end -]]
  [[- end]]
  [[- if $field.IsRequired -]]
    [[- if ne $field.Name "id"]] notnull[[end -]]
  [[- end -]]
  [[- if $field.DefaultValue]] default('[[$field.DefaultValue]]')[[end]]"` + "`" + `
[[- end]]

  [[- range $belongsTo := .class.BelongsTo ]]
    [[- if fieldExists $class $belongsTo.Name | not ]][[$belongsTo.AttributeName false]] int64 ` + "`" + `json:"[[$belongsTo.AttributeName true]]" xorm:"[[$belongsTo.AttributeName true]]"` + "`" + `
    [[- end]]
  [[- end]]
}

func ([[camelizeDownFirst .class.Name]] *[[.class.Name]]) TableName() string {
  return "[[if .class.Table]][[.class.Table]][[else]][[pluralize .class.Name | underscore]][[end]]"
}

func ([[camelizeDownFirst .class.Name]] *[[.class.Name]]) Validate(validation *revel.Validation) bool {
[[- range $column := .class.Fields]]
  [[- if ne $column.Name "id"]]
    [[- if $column.IsRequired]]
      validation.Required([[$var]].[[goify $column.Name true]]).Key("[[$var]].[[goify $column.Name true]]")
    [[- else if eq $column.Format "email"]]
      [[- if not $column.IsRequired]]
        if "" != [[$var]].[[goify $column.Name true]] {
          validation.Email([[$var]].[[goify $column.Name true]]).Key("[[$var]].[[goify $column.Name true]]")
        }
      [[- else]]
        validation.Email([[$var]].[[goify $column.Name true]]).Key("[[$var]].[[goify $column.Name true]]")
      [[- end]]
    [[- else if $column.Restrictions]]
      [[- if $column.Restrictions.MinLength]]
           validation.MinSize([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MinLength]]).Key("[[$var]].[[goify $column.Name true]]")
      [[- end]]
      [[- if $column.Restrictions.MaxLength]]
           validation.MaxSize([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MaxLength]]).Key("[[$var]].[[goify $column.Name true]]")
      [[- end]]
      [[- if $column.Restrictions.Length]]
           validation.Length([[$var]].[[goify $column.Name true]], [[$column.Restrictions.Length]]).Key("[[$var]].[[goify $column.Name true]]")
      [[- end]]
      [[- if $column.Restrictions.MaxValue]]
        [[- if $column.Restrictions.MinValue]]
           validation.Range([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MinValue]], [[$column.Restrictions.MaxValue]]).Key("[[$var]].[[goify $column.Name true]]")
        [[- else]]
           validation.Max([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MaxValue]]).Key("[[$var]].[[goify $column.Name true]]")
        [[- end]]
      [[- else if $column.Restrictions.MinValue]]
        validation.Min([[$var]].[[goify $column.Name true]], [[$column.Restrictions.MinValue]]).Key("[[$var]].[[goify $column.Name true]]")
      [[- end]]
    [[- end]]
  [[- end]]
[[- end]]
  return validation.HasErrors()
}

func KeyFor[[pluralize .class.Name]](key string) string {
  switch key {[[range $column := .class.Fields]]
  case "[[$column.Name]]":
     return "[[$var]].[[goify $column.Name true]]"[[end]]
  }
  return key
}`

var controllerText = `[[$modelName := .modelName -]][[$class := .class -]]
import (
  "[[.projectPath]]/app"
  "[[.projectPath]]/app/libs"
  "[[.projectPath]]/app/models"
  "[[.projectPath]]/app/routes"

  "github.com/revel/revel"
  "github.com/three-plus-three/forms"
  "github.com/runner-mei/orm"
  "upper.io/db.v3"
)

// [[.controllerName]] - 控制器
type [[.controllerName]] struct {
  [[if .baseController]][[.baseController]][[else]]*revel.Controller[[end]]
}

// Index 列出所有记录
func (c [[.controllerName]]) Index() revel.Result {
  var cond orm.Cond
  if name := c.Params.Get("query"); name != "" {
    cond = orm.Cond{"name LIKE": "%" + name + "%"}
  }

  total, err := c.Lifecycle.DB.[[.modelName]]().Where().And(cond).Count()
  if err != nil {
    c.ViewArgs["errors"] = err.Error()
    return c.Render(err)
  }

  var pageIndex, pageSize int
  c.Params.Bind(&pageIndex, "pageIndex")
  c.Params.Bind(&pageSize, "pageSize")
  if pageSize <= 0 {
    pageSize = libs.DEFAULT_SIZE_PER_PAGE
  }

  var [[camelizeDownFirst .modelName]] []models.[[.class.Name]]
  err = c.Lifecycle.DB.[[.modelName]]().Where().
    And(cond).
    Offset(pageIndex * pageSize).
    Limit(pageSize).
    All(&[[camelizeDownFirst .modelName]])
  if err != nil {
    c.ViewArgs["errors"] = err.Error()
    return c.Render()
  }

  [[if .class.BelongsTo -]]
  if len([[camelizeDownFirst .modelName]]) > 0 {
    var errList []string
    [[range $belongsTo := .class.BelongsTo -]]
    [[- $targetName := pluralize $belongsTo.Target -]]
    [[- $varName := camelizeDownFirst $targetName]]
    var [[camelizeDownFirst $belongsTo.Target]]IDList = make([]int64, 0, len([[camelizeDownFirst $modelName]]))
    for idx := range [[camelizeDownFirst $modelName]] {
      [[camelizeDownFirst $belongsTo.Target]]IDList = append([[camelizeDownFirst $belongsTo.Target]]IDList, [[camelizeDownFirst $modelName]][idx].[[$belongsTo.AttributeName false]])
    }

    var [[$varName]] []models.[[$belongsTo.Target]]
    err = c.Lifecycle.DB.[[$targetName]]().Where().
      And(orm.Cond{"id IN": [[camelizeDownFirst $belongsTo.Target]]IDList}).
      All(&[[$varName]])
    if err != nil {
      errList = append(errList, "load [[$belongsTo.Target]] fail, " + err.Error())
    } else {
      var [[$varName]]ByID = make(map[int64]models.[[$belongsTo.Target]], len([[$varName]]))
      for idx := range [[$varName]] {
        [[$varName]]ByID[ [[$varName]][idx].ID ] = [[$varName]][idx]
      }
      c.ViewArgs["[[$varName]]"] = [[$varName]]ByID
    }
    [[- end]]
    if len(errList) > 0 {
      c.ViewArgs["errors"] = errList
    }
  }
  [[- end]]

  paginator := libs.NewPaginator(c.Request.Request, pageSize, total)
  return c.Render([[camelizeDownFirst .modelName]], paginator)
}

[[if newDisabled .class | not -]]
// New 编辑新建记录
func (c [[.controllerName]]) New() revel.Result {
  [[if .class.BelongsTo -]]
  var errList []string
  var err error
  [[range $belongsTo := .class.BelongsTo ]]  
  [[$targetName := pluralize $belongsTo.Target]][[$varName := camelizeDownFirst $targetName]]var [[$varName]] []models.[[$belongsTo.Target]]
  err = c.Lifecycle.DB.[[$targetName]]().Where().
    All(&[[$varName]])
  if err != nil {
    errList = append(errList, "load [[$belongsTo.Target]] fail, " + err.Error())
    c.ViewArgs["[[$varName]]"] = []forms.InputChoice{}
  } else {
    [[$field := field $class $belongsTo.Name -]]
    var opt[[$targetName]] = make([]forms.InputChoice, 0, len([[$varName]]))
    for _, o := range [[$varName]] {
      opt[[$targetName]] = append(opt[[$targetName]], forms.InputChoice{
        Value: strconv.FormatInt(int64(o.ID),10),
        Label: o.[[displayForBelongsTo $field]],
      })
    }
    c.ViewArgs["[[$varName]]"] = opt[[$targetName]]
  }
  [[- end]]

  if len(errList) > 0 {
    c.ViewArgs["errors"] = errList
  }
  [[- end]]
  return c.Render()
}

// Create 创建记录
func (c [[.controllerName]]) Create([[camelizeDownFirst .class.Name]] *models.[[.class.Name]]) revel.Result {
  if [[camelizeDownFirst .class.Name]].Validate(c.Validation) {
    c.Validation.Keep()
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].New())
  }

  _, err := c.Lifecycle.DB.[[.modelName]]().Insert([[camelizeDownFirst .class.Name]])
  if err != nil {
    if oerr, ok := err.(*orm.Error); ok {
      for _, validation := range oerr.Validations {
        c.Validation.Error(validation.Message).Key(models.KeyFor[[.modelName]](validation.Key))
      }
      c.Validation.Keep()
    }
    c.Flash.Error(err.Error())
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].New())
  }

  c.Flash.Success(revel.Message(c.Request.Locale, "insert.success"))
  return c.Redirect(routes.[[.controllerName]].Index())
}
[[- end]]

[[if editDisabled .class | not -]]
// Edit 编辑指定 id 的记录
func (c [[.controllerName]]) Edit(id int64) revel.Result {
  var [[camelizeDownFirst .class.Name]] models.[[.class.Name]]
  err := c.Lifecycle.DB.[[.modelName]]().Id(id).Get(&[[camelizeDownFirst .class.Name]])
  if err != nil {
    if err == orm.ErrNotFound {
      c.Flash.Error(revel.Message(c.Request.Locale, "update.record_not_found"))
    } else {
      c.Flash.Error(err.Error())
    }
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].Index())
  }

  [[if .class.BelongsTo -]]
  var errList []string
  [[range $belongsTo := .class.BelongsTo ]]  
  [[$targetName := pluralize $belongsTo.Target]][[$varName := camelizeDownFirst $targetName]]var [[$varName]] []models.[[$belongsTo.Target]]
  err = c.Lifecycle.DB.[[$targetName]]().Where().
    All(&[[$varName]])
  if err != nil {
    errList = append(errList, "load [[$belongsTo.Target]] fail, " + err.Error())
    c.ViewArgs["[[$varName]]"] = []forms.InputChoice{}
  } else {
    [[$field := field $class $belongsTo.Name -]]
    var opt[[$targetName]] = make([]forms.InputChoice, 0, len([[$varName]]))
    for _, o := range [[$varName]] {
      opt[[$targetName]] = append(opt[[$targetName]], forms.InputChoice{
        Value: strconv.FormatInt(int64(o.ID),10),
        Label: o.[[displayForBelongsTo $field]],
      })
    }
    c.ViewArgs["[[$varName]]"] = opt[[$targetName]]
  }

  if len(errList) > 0 {
    c.ViewArgs["errors"] = errList
  }
  [[- end]]
  [[- end]]

  return c.Render([[camelizeDownFirst .class.Name]])
}

// Update 按 id 更新记录
func (c [[.controllerName]]) Update([[camelizeDownFirst .class.Name]] *models.[[.class.Name]]) revel.Result {
  if [[camelizeDownFirst .class.Name]].Validate(c.Validation) {
    c.Validation.Keep()
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].Edit(int64([[camelizeDownFirst .class.Name]].ID)))
  }

  err := c.Lifecycle.DB.[[.modelName]]().Id([[camelizeDownFirst .class.Name]].ID).Update([[camelizeDownFirst .class.Name]])
  if err != nil {
    if err == orm.ErrNotFound {
      c.Flash.Error(revel.Message(c.Request.Locale, "update.record_not_found"))
    } else {
      if oerr, ok := err.(*orm.Error); ok {
        for _, validation := range oerr.Validations {
          c.Validation.Error(validation.Message).Key(models.KeyFor[[.modelName]](validation.Key))
        }
        c.Validation.Keep()
      }
      c.Flash.Error(err.Error())
    }
    c.FlashParams()
    return c.Redirect(routes.[[.controllerName]].Edit(int64([[camelizeDownFirst .class.Name]].ID)))
  }
  c.Flash.Success(revel.Message(c.Request.Locale, "update.success"))
  return c.Redirect(routes.[[.controllerName]].Index())
}
[[- end]]

[[if deleteDisabled .class | not -]]
[[if .class.PrimaryKey]]
// Delete 按 primaryKey 删除记录
func (c [[.controllerName]]) Delete([[- range $idx, $fieldName := .class.PrimaryKey]]
[[- $field := field $.class $fieldName]]
[[- if ne $idx 0]], [[end]][[$fieldName]] [[gotype $field.Type]]
[[- end]]) revel.Result {
  var cond = orm.Cond{}
[[- range $fieldName := .class.PrimaryKey]]
  [[- $field := field $.class $fieldName]]
  [[- if eq $field.Type "integer" "number" "biginteger" "int" "int8" "int16" "int32" "int64" "uint" "uint8" "uint16" "uint32" "uint64" "float" "float32" "float64"]]
    if [[$fieldName]] == 0 {
      c.Flash.Error("[[$fieldName]] is missing")
      return c.Redirect(routes.[[$.controllerName]].Index())
    }
  [[- else if eq $field.Type "datetime"]]
    if [[$fieldName]].IsZero() {
      c.Flash.Error("[[$fieldName]] is missing")
      return c.Redirect(routes.[[$.controllerName]].Index())
    } 
  [[- else if eq $field.Type "ipAddress" "ipaddress" "net.IP" "macAddress" "net.HardwareAddress"]]
    if [[$fieldName]] == nil {
      c.Flash.Error("[[$fieldName]] is missing")
      return c.Redirect(routes.[[$.controllerName]].Index())
    }
  [[- else]]
    if [[$fieldName]] == "" {
      c.Flash.Error("[[$fieldName]] is missing")
      return c.Redirect(routes.[[$.controllerName]].Index())
    }
  [[- end]]
  cond["[[$field.Name]] ="] = [[$fieldName]]
[[- end]]

  rowsEffected, err :=  c.Lifecycle.DB.[[.modelName]]().Where(cond).Delete()
  if nil != err {
    if err == orm.ErrNotFound {
      c.Flash.Error(revel.Message(c.Request.Locale, "delete.record_not_found"))
    } else {
      c.Flash.Error(err.Error())
    }
  } else if rowsEffected <= 0 {
    c.Flash.Error(revel.Message(c.Request.Locale, "delete.record_not_found"))
  } else {
    c.Flash.Success(revel.Message(c.Request.Locale, "delete.success"))
  }
  return c.Redirect(routes.[[.controllerName]].Index())
}
[[else]]
// Delete 按 id 删除记录
func (c [[.controllerName]]) Delete(id int64) revel.Result {
  err :=  c.Lifecycle.DB.[[.modelName]]().Id(id).Delete()
  if nil != err {
    if err == orm.ErrNotFound {
      c.Flash.Error(revel.Message(c.Request.Locale, "delete.record_not_found"))
    } else {
      c.Flash.Error(err.Error())
    }
  } else {
    c.Flash.Success(revel.Message(c.Request.Locale, "delete.success"))
  }
  return c.Redirect(routes.[[.controllerName]].Index())
}

// DeleteByIDs 按 id 列表删除记录
func (c [[.controllerName]]) DeleteByIDs(id_list []int64) revel.Result {
  if len(id_list) == 0 {
    c.Flash.Error("请至少选择一条记录！")
    return c.Redirect(routes.[[.controllerName]].Index())
  }
  _, err :=  c.Lifecycle.DB.[[.modelName]]().Where().And(orm.Cond{"id IN": id_list}).Delete()
  if nil != err {
    c.Flash.Error(err.Error())
  } else {
    c.Flash.Success(revel.Message(c.Request.Locale, "delete.success"))
  }
  return c.Redirect(routes.[[.controllerName]].Index())
}
[[- end]]
[[- end]]`

var viewEditText = `{{set . "title" "编辑[[.controllerName]]"}}
{{append . "moreScripts" "[[.customPath]]/public/js/[[underscore .controllerName]]/[[underscore .controllerName]].js"}}
{{template "[[if .layouts]][[.layouts]][[end]]header.html" .}}
<div class="ibox float-e-margins">
    <div class="ibox-title">
        [[edit_label .class]]
        <div class="ibox-tools"></div>
    </div>
    <div class="ibox-content">
        <form action="{{url "[[.controllerName]].Update" }}" method="POST" class="form-horizontal" id="[[underscore .controllerName]]-edit">
        <input type="hidden" name="_method" value="PUT">
        {{hidden_field . "[[camelizeDownFirst .class.Name]].ID" | render}}
        {{- $inEditMode := .inEditMode}}{{ set . "inEditMode" false}}
        {{template "[[.controllerName]]/edit_fields.html" .}}
        {{- set . "inEditMode" $inEditMode}}
        <div class="form-group">
            <div class="col-lg-offset-2 col-lg-10">
                <button type="submit" class="btn btn-info controls">修改</button>
                <a href="{{url "[[.controllerName]].Index" }}" class="btn btn-info controls">取消</a>
            </div>
        </div>
        </form>
    </div>
</div>
{{template "[[if .layouts]][[.layouts]][[end]]footer.html" .}}`

var viewFieldsText = `[[$class := .class]]
[[- $instaneName := camelizeDownFirst .class.Name]]
[[- define "lengthLimit"]][[end]]
[[- range $column := .class.Fields]]

  [[- if isID $column]]
  [[- else if editDisabled $column]]
  [[- else if isBelongsTo $class  $column ]]
    {{select_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" .[[belongsToClassName $class  $column | pluralize | camelizeDownFirst]] [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else if valueInAnnotations $column "enumerationSource" ]]
    {{select_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" .global.[[valueInAnnotations $column "enumerationSource"]] [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else if hasEnumerations $column ]]
    {{select_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" "[[jsEnumeration $column.Restrictions.Enumerations | js]]" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else if $column.Format ]]
    [[- if eq $column.Format "ip" ]]
      {{ipaddress_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" | render}}
    [[- else if eq $column.Format "email" ]]
      {{email_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
    [[- end]]
  [[- else if eq $column.Type "string" ]]
    [[- if isClob $column ]]
    {{textarea_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" 3  0 | [[template "lengthLimit" $column]] render}}
    [[- else]]
    {{text_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
    [[- end]]
  [[- else if eq $column.Type "integer" "number" "biginteger" "int" "int64" "uint" "uint64" "float" "float64" ]]
    [[- if $column.Restrictions]]
      [[- if $column.Restrictions.MinValue]]
        [[- if $column.Restrictions.MaxValue]]
          {{number_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
        [[- else]]
          {{number_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
        [[- end]]
      [[- else if $column.Restrictions.MaxValue]]
        {{number_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
      [[- end]]
    [[- else]]
      {{number_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
    [[- end]]
  [[- else if eq $column.Type "boolean" "bool" ]]
    {{checkbox_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else if eq $column.Type "password" ]]
    {{password_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else if eq $column.Type "time" ]]
    {{time_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else if eq $column.Type "datetime" ]]
    {{datetime_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else if eq $column.Type "date" ]]
    {{date_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else if eq $column.Type "map" ]]
    {{map_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- else]]
    {{text_field . "[[$instaneName]].[[goify $column.Name true]]" "[[localizeName $column]]:" [[if and $column.IsReadOnly]]| f_setEditMode .inEditMode [[end]] | render}}
  [[- end]]
[[- end]]`

var viewIndexText = `[[$raw := .]]{{$raw := .}}{{set . "title" "[[.controllerName]]"}}
{{if eq .RunMode "dev"}}
{{append . "moreScripts" "/public/js/plugins/bootbox/bootbox.js"}}
{{else}}
{{append . "moreScripts" "/public/js/plugins/bootbox/bootbox.min.js"}}
{{end}}
{{append . "moreScripts" "[[.customPath]]/public/js/[[underscore .controllerName]]/[[underscore .controllerName]].js"}}
{{template "[[if .layouts]][[.layouts]][[end]]header.html" .}}

<div class="ibox float-e-margins">
    <div class="ibox-title">
        [[index_label .class]]
        <div class="ibox-tools"></div>
    </div>
    <div class="ibox-content">
    {{template "[[.controllerName]]/quick-bar.html" .}}
    <table class="table table-bordered table-striped table-highlight ">
      <thead>
      <tr>
        [[- if hasAllFeatures $raw.class "editDisabled" "deleteDisabled" | not -]]
        <th><input type="checkbox" id="[[underscore .controllerName]]-all-checker" /></th>
        [[- end]]
        [[- range $field := .class.Fields]]
          [[- if needDisplay $field]]
            [[- $bt := belongsTo $raw.class $field]]
            [[- if $bt ]]
              [[- $refClass := class $bt.Target]]
              [[- $referenceFields := referenceFields $field]]
              [[- range $rField := $referenceFields ]]
                [[- $referenceField := field $refClass $rField.Name]]
        <th>{{table_column_title . "[[$field.Name]]" "[[if $rField.Label]][[$rField.Label]][[else]][[localizeName $referenceField]][[end]]"}}</th>
              [[- end]]
            [[- else]]
        <th>{{table_column_title . "[[$field.Name]]" "[[localizeName $field ]]"}}</th>
            [[- end -]]
          [[- end]]
        [[- end]]
        [[- if hasAllFeatures $raw.class "editDisabled" "deleteDisabled" | not]]
        {{- if current_user_has_write_permission $raw "[[underscore .controllerName]]"}}
        <th>操作</th>
        {{- end}}
        [[- end]]
      </tr>
      </thead>
      <tbody>
      {{- range $idx $v := .[[camelizeDownFirst .modelName]]}}
        <tr>
        [[- if hasAllFeatures $raw.class "editDisabled" "deleteDisabled" | not -]]
          <td><input type="checkbox" class="[[underscore .controllerName]]-row-checker"
          [[- if $raw.class.PrimaryKey]]
            [[- if deleteDisabled $raw.class | not]] del.url="{{url "[[.controllerName]].Delete" 
              [[- range $fieldName := $raw.class.PrimaryKey]] $v.[[goify $fieldName true]]
              [[- end]]}}"
            [[- end]]
            [[- if editDisabled $raw.class | not]] edit.url="{{url "[[.controllerName]].Edit" 
              [[- range $fieldName := $raw.class.PrimaryKey]] $v.[[goify $fieldName true]]
              [[- end]]}}"
            [[- end]]
          [[- else]]
            key="{{$v.ID}}"[[- if editDisabled $raw.class | not]] url="{{url "[[.controllerName]].Edit" $v.ID}}"
          [[- end -]]/></td>
        [[- end]]
        [[- end]]
          [[- range $column := .class.Fields]]
            [[- if needDisplay $column]]
              [[- $bt := belongsTo $raw.class $column]]
              [[- if $bt ]]
                  [[- $refClass := class $bt.Target]]
                  [[- $referenceFields := referenceFields $column]]
              {{- $rValue := index $raw.[[pluralize $refClass.Name | camelizeDownFirst]] $v.[[goify $column.Name true]]}}
                  [[- range $rField := $referenceFields ]]
                    [[- $referenceField := field $refClass $rField.Name]]
              <td>{{$rValue.[[goify $referenceField.Name true]]}}</td>
                  [[- end]]
              [[- else]]
              <td>{{[[if eq $column.Type "date"]]date [[else if eq $column.Type "datetime"]]datetime [[else if eq $column.Type "time"]]time [[end]]$v.[[goify $column.Name true]]}}</td>
              [[- end]]
            [[- end]]
          [[- end]]



          [[- if hasAllFeatures $raw.class "editDisabled" "deleteDisabled" | not -]]
            [[- if $raw.class.PrimaryKey]]
              [[- if editDisabled $raw.class | not]]
                {{if current_user_has_edit_permission $raw "[[underscore .controllerName]]" -}}
              <a href='{{url "[[.controllerName]].Edit"
                [[- range $fieldName := $raw.class.PrimaryKey]] $v.[[goify $fieldName true]]
                [[- end]]}}'>编辑</a>
                {{- end}}
              [[- end]]
              [[- if deleteDisabled $raw.class | not]]
                {{if current_user_has_del_permission $raw "[[underscore .controllerName]]" -}}
               <form id='[[underscore .controllerName]]-delete-{{$idx}}' action="{{url "[[.controllerName]].Delete"}}" method="POST" class="form-inline" role="form" style="display: inline;">
                  <input type="hidden" name="_method" value="DELETE">
                [[- range $fieldName := $raw.class.PrimaryKey]]
                  <input type="hidden" name="[[$fieldName]]" value="{{$v.[[goify $fieldName true]]}}">
                [[- end -]]
                  <a href="javascript:document.getElementById('[[underscore .controllerName]]-delete-{{$idx}}').submit()">
                      <i class="icon-search"></i> 删除
                    </a>
                </form>
                {{- end}}
              [[- end]]
            [[- else]]

              {{if current_user_has_write_permission $raw "[[underscore .controllerName]]"}}<td>
              [[if editDisabled $raw.class | not -]]
                {{if current_user_has_edit_permission $raw "[[underscore .controllerName]]" -}}
                <a href='{{url "[[.controllerName]].Edit" $v.ID}}'>编辑</a>
                {{- end}}
              [[- end]]
              [[- if deleteDisabled $raw.class | not -]]
                {{- if current_user_has_del_permission $raw "[[underscore .controllerName]]"}}
                <form id='[[underscore .controllerName]]-delete-{{$v.ID}}' action="{{url "[[.controllerName]].Delete" $v.ID}}" method="POST" class="form-inline" role="form" style="display: inline;">
                  <input type="hidden" name="_method" value="DELETE">
                  <input type="hidden" name="id" value="{{$v.ID}}">
                    <a href="javascript:document.getElementById('[[underscore .controllerName]]-delete-{{$v.ID}}').submit()">
                      <i class="icon-search"></i> 删除
                    </a>
                </form>
                {{- end}}
              [[- end]]
            [[- end]]
            </td>
            {{- end}}
          [[- end]]
        </tr>
      <tbody>
      {{- end}}
    </table>
    {{template "[[if .layouts]][[.layouts]][[end]]paginator.html" .}}
    </div>
</div>

{{template "[[if .layouts]][[.layouts]][[end]]footer.html" .}}`

var viewNewText = `{{set . "title" "新建[[.controllerName]]"}}
{{append . "moreStyles" "/public/css/form.css"}}
{{append . "moreScripts" "[[.customPath]]/public/js/[[underscore .controllerName]]/[[underscore .controllerName]].js"}}
{{template "[[if .layouts]][[.layouts]][[end]]header.html" .}}
<div class="ibox float-e-margins">
    <div class="ibox-title">
        [[new_label .class]]
        <div class="ibox-tools"></div>
    </div>
    <div class="ibox-content">
        <form action="{{url "[[.controllerName]].Create" }}" method="POST" class="form-horizontal" id="[[underscore .controllerName]]-insert">
        {{- $inEditMode := .inEditMode}}{{ set . "inEditMode" true}}
        {{template "[[.controllerName]]/edit_fields.html" .}}
        {{- set . "inEditMode" $inEditMode}}
        <div class="form-group">
            <div class="col-lg-offset-2 col-lg-10">
                <button type="submit" class="btn btn-info controls">新建</button>
                <a href="{{url "[[.controllerName]].Index" }}" class="btn btn-info controls">取消</a>
            </div>
        </div>
        </form>
    </div>
</div>
{{template "[[if .layouts]][[.layouts]][[end]]footer.html" .}}`

var viewQuickText = `    <div class="quick-actions btn-group m-b">
        [[if newDisabled .class | not -]]
        {{- if current_user_has_new_permission . "[[underscore .controllerName]]"}}
        <a id='[[underscore .controllerName]]-new' href='{{url "[[.controllerName]].New"}}'  class="btn btn-outline btn-default" method="" mode="*" confirm="" client="false" target="_self">
            <i class="fa fa-add"></i>添加
        </a>
        {{- end}}
        [[- end]]
        [[if editDisabled .class | not -]]
        {{- if current_user_has_edit_permission . "[[underscore .controllerName]]"}}
        <a id='[[underscore .controllerName]]-edit' href='' url='{{url "[[.controllerName]].Edit"}}'  class="btn btn-outline btn-default" method="" mode="1" confirm="" client="false" target="_self">
            <i class="fa fa-edit"></i>编辑
        </a>
        {{- end}}
        [[- end]]
        [[if deleteDisabled .class | not -]]
        {{- if current_user_has_del_permission . "[[underscore .controllerName]]"}}
        <a id='[[underscore .controllerName]]-delete' href='' url='{{url "[[.controllerName]].DeleteByIDs"}}'  class="btn btn-outline btn-default" mode="+" target="_self">
            <i class="fa fa-trash"></i> 删除
        </a>
        {{- end}}
        [[- end]]
        [[- if fieldExists .class "name" -]]
        <form action="{{url "[[.controllerName]].Index" 0 0}}" method="POST" id='[[underscore .controllerName]]-quick-form' class="form-inline"  style="display: inline;">
            <input type="text" name="query">
            <a href="javascript:document.getElementById('[[underscore .controllerName]]-quick-form').submit()" >
                <i class="fa fa-search"></i> 查询
            </a>
        </form>
        [[- end]]
    </div>`

var viewJsText = `$(function () {
    var urlPrefix = $("#urlPrefix").val();

    $("#[[underscore .controllerName]]-all-checker").on("click", function () {
        var all_checked =  this.checked
        $(".[[underscore .controllerName]]-row-checker").each(function(){
            this.checked = all_checked
            return true;
        });
        return true;
    });


    $("#[[underscore .controllerName]]-delete").on("click", function () {
        bootbox.confirm("确认删除选定信息？", function(result){
            if (!result) {
                return;
            }

            var f = document.createElement("form");
            f.action = $("#[[underscore .controllerName]]-delete").attr("url");
            f.method="POST";
            var inputField = document.createElement("input");
            inputField.type = "hidden";
            inputField.name = "_method";
            inputField.value = "DELETE";

            $(".[[underscore .controllerName]]-row-checker:checked").each(function (i) {
                var inputField = document.createElement("input");
                inputField.type = "hidden";
                inputField.name = "id_list[]";
                inputField.value = $(this).attr("key");
                f.appendChild(inputField);
            });

            document.body.appendChild(f);
            f.submit();
        })
        return false
    });

    $("#[[underscore .controllerName]]-edit").on("click", function () {
        var elements = $(".[[underscore .controllerName]]-row-checker:checked");
        if (elements.length == 1) {
            window.location.href= elements.first().attr("url");
        } else if (elements.length == 0) {
            bootbox.alert('请选择一条记录！')
        } else {
            bootbox.alert('你选择了多条记录，请选择一条记录！')
        }

        return false
    });
});
`

var testCtlText = `
package tests

import (
	"net/url"
	"strconv"
	"strings"
	"[[.projectPath]]/app"
	"[[.projectPath]]/app/models"
)

// [[$varName := camelizeDownFirst .class.Name]] [[.controllerName]]Test 测试
type [[.controllerName]]Test struct {
	BaseTest
}

func (t [[.controllerName]]Test) TestIndex() {
	t.ClearTable("[[tableName .class]]")
	t.LoadFiles("tests/fixtures/[[underscore .controllerName]].yaml")	
	//conds := EQU{"name": "这是一个规则名,请替换成正确的值"}
	conds := EQU{}
	ruleId := t.GetIDFromTable("[[tableName .class]]", conds)

	t.Get(t.ReverseUrl("[[.controllerName]].Index"))
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
	//t.AssertContains("这是一个规则名,请替换成正确的值")

	var [[$varName]] models.[[.class.Name]]
	err :=  app.Lifecycle.DB.[[.controllerName]]().Id(ruleId).Get(&[[$varName]])
	if err != nil {
		t.Assertf(false, err.Error())
	}
	[[range $column := .class.Fields]][[if isID $column]][[else if eq $column.Name "created_at" "updated_at"]][[else if eq $column.Type "password"]][[else]]
	t.AssertContains(fmt.Sprint([[$varName]].[[goify $column.Name true]]))[[end]][[end]]
}

func (t [[.controllerName]]Test) TestNew() {
	t.ClearTable("[[tableName .class]]")
	t.Get(t.ReverseUrl("[[.controllerName]].New"))
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")
}

func (t [[.controllerName]]Test) TestCreate() {
	t.ClearTable("[[tableName .class]]")
	v := url.Values{}
	[[range $column := .class.Fields]][[if isID $column]][[else]]
  v.Set("[[$varName]].[[goify $column.Name true]]", "[[randomValue $column]]")
  [[end]][[end]]
  
  t.Post(t.ReverseUrl("[[.controllerName]].Create"), "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	t.AssertOk()

	//conds := EQU{"name": "这是一个规则名,请替换成正确的值"}
	conds := EQU{}
	ruleId := t.GetIDFromTable("[[tableName .class]]", conds)

	var [[$varName]] models.[[.class.Name]]
	err :=  app.Lifecycle.DB.[[.controllerName]]().Id(ruleId).Get(&[[$varName]])
	if err != nil {
		t.Assertf(false, err.Error())
	}
	[[range $column := .class.Fields]][[if isID $column]][[else if eq $column.Name "created_at" "updated_at"]][[else]]
	t.AssertEqual(fmt.Sprint([[$varName]].[[goify $column.Name true]]), v.Get("[[$varName]].[[goify $column.Name true]]"))[[end]][[end]]
}

func (t [[.controllerName]]Test) TestEdit() {
	t.ClearTable("[[tableName .class]]")
	t.LoadFiles("tests/fixtures/[[underscore .controllerName]].yaml")
	//conds := EQU{"name": "这是一个规则名,请替换成正确的值"}
	conds := EQU{}
	ruleId := t.GetIDFromTable("[[tableName .class]]", conds)
	t.Get(t.ReverseUrl("[[.controllerName]].Edit", ruleId))
	t.AssertOk()
	t.AssertContentType("text/html; charset=utf-8")

	var [[$varName]] models.[[.class.Name]]
	err :=  app.Lifecycle.DB.[[.controllerName]]().Id(ruleId).Get(&[[$varName]])
	if err != nil {
		t.Assertf(false, err.Error())
	}
	fmt.Println(string(t.ResponseBody))
	[[range $column := .class.Fields]][[if isID $column]][[else if eq $column.Name "created_at" "updated_at"]][[else if eq $column.Type "password"]][[else]]
	t.AssertContains(fmt.Sprint([[$varName]].[[goify $column.Name true]]))[[end]][[end]]
}

func (t [[.controllerName]]Test) TestUpdate() {
	t.ClearTable("[[tableName .class]]")
	t.LoadFiles("tests/fixtures/[[underscore .controllerName]].yaml")
	//conds := EQU{"name": "这是一个规则名,请替换成正确的值"}
	conds := EQU{}
	ruleId := t.GetIDFromTable("[[tableName .class]]", conds)
	v := url.Values{}
	v.Set("_method", "PUT")
	v.Set("[[$varName]].ID", strconv.FormatInt(ruleId, 10))

	[[range $column := .class.Fields]][[if isID $column]][[else]]
  v.Set("[[$varName]].[[goify $column.Name true]]", "[[randomValue $column]]")
  [[end]][[end]]


  t.Post(t.ReverseUrl("[[.controllerName]].Update"), "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	t.AssertOk()

	var [[$varName]] models.[[.class.Name]]
	err :=  app.Lifecycle.DB.[[.controllerName]]().Id(ruleId).Get(&[[$varName]])
	if err != nil {
		t.Assertf(false, err.Error())
	}
	[[range $column := .class.Fields]][[if isID $column]][[else if eq $column.Name "created_at" "updated_at"]][[else]]
	t.AssertEqual(fmt.Sprint([[$varName]].[[goify $column.Name true]]), v.Get("[[$varName]].[[goify $column.Name true]]"))
  [[end]][[end]]
}

func (t [[.controllerName]]Test) TestDelete() {
	t.ClearTable("[[tableName .class]]")
	t.LoadFiles("tests/fixtures/[[underscore .controllerName]].yaml")
	//conds := EQU{"name": "这是一个规则名,请替换成正确的值"}
	conds := EQU{}
	ruleId := t.GetIDFromTable("[[tableName .class]]", conds)
	t.Delete(t.ReverseUrl("[[.controllerName]].Delete", ruleId))
	t.AssertStatus(http.StatusOK)
	//t.AssertContentType("application/json; charset=utf-8")
	count := t.GetCountFromTable("[[tableName .class]]", nil)
	t.Assertf(count == 0, "count != 0, actual is %v", count)
}

func (t [[.controllerName]]Test) TestDeleteByIDs() {
	t.ClearTable("[[tableName .class]]")
	t.LoadFiles("tests/fixtures/[[underscore .controllerName]].yaml")
	//conds := EQU{"name": "这是一个规则名,请替换成正确的值"}
	conds := EQU{}
	ruleId := t.GetIDFromTable("[[tableName .class]]", conds)
	t.Delete(t.ReverseUrl("[[.controllerName]].DeleteByIDs", []interface{}{ruleId}))
	t.AssertStatus(http.StatusOK)
	//t.AssertContentType("application/json; charset=utf-8")
	count := t.GetCountFromTable("[[tableName .class]]", nil)
	t.Assertf(count == 0, "count != 0, actual is %v", count)
}
`

var testYamlText = `- table: '[[tableName .class]]'
  pk:
    id: 'PK_GENERATE([[underscore .class.Name]]_key)'
  fields:[[range $column := .class.Fields]][[if isID $column]][[else]]
    [[$column.Name]]: [[randomValue $column]][[end]][[end]]`

var testBaseText = `package tests

import (
	"cn/com/hengwei/commons/httputils"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"[[.projectPath]]/app"
	"strings"

	fixtures "github.com/AreaHQ/go-fixtures"
	"github.com/Masterminds/squirrel"
	"github.com/revel/revel"
	"github.com/revel/revel/testing"
)

// DbRunner wraps sql.DB to implement Runner.
type DbRunner struct {
	*sql.DB
}

// QueryRow wraps QueryRow to implement RowScanner.
func (r DbRunner) QueryRow(query string, args ...interface{}) squirrel.RowScanner {
	return r.DB.QueryRow(query, args...)
}

type EQU map[string]interface{}

type BaseTest struct {
	testing.TestSuite
}

func (t *BaseTest) Before() {
	println("================ Set up  =================")
	fmt.Println(app.Lifecycle.Env.Db.Models.Url())
	if !strings.Contains(app.Lifecycle.Env.Db.Models.Schema, "_test") {
		panic("runMode must is test.")
	}
}

func (t *BaseTest) After() {
	println("=============== Tear down ================")
}

func (t *BaseTest) DB() *sql.DB {
	return app.Lifecycle.DB.Engine.DB().DB
}

func (t *BaseTest) DataDB() *sql.DB {
	return app.Lifecycle.DataDB.Engine.DB().DB
}

func (t *BaseTest) DBRunable() squirrel.Runner{
	return &DbRunner{t.DB()}
}

func (t *BaseTest) DataDBRunable() squirrel.Runner{
	return &DbRunner{t.DataDB()}
}

func (t *BaseTest) ReverseUrl(args ...interface{}) string {
	s, e := revel.ReverseURL(args...)
	if e != nil {
		t.Assertf(false, e.Error())
		return ""
	}
	return string(s)
}

func (t *BaseTest) LoadFiles(filenames ...string) {
	if err := fixtures.LoadFiles(filenames, t.DB(), "postgres"); err != nil {
		t.Assertf(false, err.Error())
		return
	}
}

func (t *BaseTest) LoadFilesToData(filenames ...string) {
	if err := fixtures.LoadFiles(filenames, t.DataDB(), "postgres"); err != nil {
		t.Assertf(false, err.Error())
		return
	}
}

func (t *BaseTest) GetCountFromTable(table string, params EQU) (count int64) {
	return t.getCountFromTable(t.DBRunable(), table, params)
}

func (t *BaseTest) GetCountFromDataTable(table string, params EQU) (count int64) {
	return t.getCountFromTable(t.DataDBRunable(), table, params)
}

func (t *BaseTest) getCountFromTable(dbRunner squirrel.Runner, table string, params EQU) (count int64) {
	builder := squirrel.Select("count(*)").From(table)
	if len(params) > 0 {
		builder = builder.Where(squirrel.Eq(params))
	}
	builder = builder.PlaceholderFormat(squirrel.Dollar)

	fmt.Println(builder.ToSql())
	rs := squirrel.QueryRowWith(dbRunner, builder)
	if err := rs.Scan(&count); nil != err {
		t.Assertf(false, err.Error())
	}
	return count
}

func (t *BaseTest) GetIDFromTable(table string, params EQU) (id int64) {
	return t.getIDFromTable(t.DBRunable(), table, params)
}

func (t *BaseTest) GetIDFromDataTable(table string, params EQU) (id int64) {
	return t.getIDFromTable(t.DataDBRunable(), table, params)
}

func (t *BaseTest) getIDFromTable(dbRunner squirrel.Runner, table string, params EQU) (id int64) {
	builder := squirrel.Select("id").From(table)
	if len(params) > 0 {
		builder = builder.Where(squirrel.Eq(params))
	}
	builder = builder.PlaceholderFormat(squirrel.Dollar)
	rs := squirrel.QueryRowWith(dbRunner, builder)
	if err := rs.Scan(&id); nil != err {
		t.Assertf(false, err.Error())
	}
	return id
}

func (t *BaseTest) ClearTable(tableName string) {
	t.clearTable(t.DBRunable(), tableName)
}

func (t *BaseTest) ClearDataTable(tableName string) {
	t.clearTable(t.DataDBRunable(), tableName)
}

func (t *BaseTest) clearTable(dbRunner squirrel.Runner, tableName string) {
	if _, err := dbRunner.Exec("truncate table " + tableName + " cascade"); err != nil {
		t.Assertf(false, err.Error())
	}
}

func (t *BaseTest) ClearDB() {
	if _, err := t.DB().Exec("select clear_data_of_all_table()"); err != nil {
		t.Assertf(false, err.Error())
	}
}

func (t *BaseTest) ResponseAsJSONObject() map[string]interface{} {
	var res map[string]interface{}
	if err := json.Unmarshal(t.ResponseBody, &res); err != nil {
		t.Assertf(false, err.Error())
	}
	return res
}

func (t *BaseTest) ResponseAsJSONArray() []map[string]interface{} {
	var res []map[string]interface{}
	if err := json.Unmarshal(t.ResponseBody, &res); err != nil {
		t.Assertf(false, err.Error())
	}
	return res
}

func (t *BaseTest) ResponseAsArray() []interface{} {
	var res []interface{}
	if err := json.Unmarshal(t.ResponseBody, &res); err != nil {
		t.Assertf(false, err.Error())
	}
	return res
}`

var dbText = `
import (
  "github.com/go-xorm/xorm"
  "github.com/runner-mei/orm"
)

type DB struct {
  Engine *xorm.Engine
}

[[- range $class := .classes]]
func (db *DB) [[pluralize $class.Name]]() *orm.Collection {
  return orm.New(func() interface{}{
    return &[[$class.Name]]{}
  })(db.Engine)
}
[[- end]]

func InitTables(engine *xorm.Engine) error {
  beans := []interface{}{[[range $class := .classes]]
    &[[$class.Name]]{},[[end]]
  }

  if err := engine.CreateTables(beans...); err != nil {
    return err
  }

  for _, bean := range beans {
    if err := engine.CreateIndexes(bean); err != nil {
      if !strings.Contains(err.Error(), "already exists") {
        return err
      }
      revel.WARN.Println(err)
    }

    if err := engine.CreateUniques(bean); err != nil {
      if !strings.Contains(err.Error(), "already exists") &&
        !strings.Contains(err.Error(), "create unique index") {
        return err
      }
      revel.WARN.Println(err)
    }
  }
  return nil
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
	case "tests/test_ctl":
		return []byte(testCtlText)
	case "tests/test_yaml":
		return []byte(testYamlText)
	case "tests/test_base":
		return []byte(testBaseText)
	case "db":
		return []byte(dbText)
	default:
		panic(errors.New("template '" + nm + "' isn't default template."))
	}
}
