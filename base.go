package main

import (
	"bytes"
	"database/sql"
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

// JSON 代表一个数据库中一个 json
type JSON []byte

// ToJSON 将字节数组转成一个 JSON 对象
func ToJSON(bs []byte) JSON {
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
		return column.origin.Name + "->'" + column.subField + "'"
	}

	if column.subField == "" {
		return column.tableAlias + "." + column.origin.Name
	}
	return column.tableAlias + "." + column.origin.Name + "->'" + column.subField + "'"
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
