package main

import (
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/lann/builder"
)

var ErrNotUpdated = errors.New("no record is updated")
var ErrNotDeleted = errors.New("no record is deleted")

func isPostgersql(db interface{}) bool {
	return true
}

func isPlaceholderWithDollar(value interface{}) bool {
	return true
}

type ViewModel struct {
	TableName   string
	ColumnNames []string
}

func (self *ViewModel) Where(exprs ...Expr) squirrel.StatementBuilderType {
	if len(exprs) == 1 {
		return builder.Append(squirrel.StatementBuilder, "WhereParts", exprs[0]).(squirrel.StatementBuilderType)
	}
	sqlizers := make([]squirrel.Sqlizer, 0, len(exprs))
	for _, exp := range exprs {
		sqlizers = append(sqlizers, exp)
	}

	return builder.Append(squirrel.StatementBuilder, "WhereParts", squirrel.And(sqlizers)).(squirrel.StatementBuilderType)
}

func (self *ViewModel) UpdateBy(db squirrel.BaseRunner, values map[string]interface{}, pred interface{}, args ...interface{}) (int64, error) {
	sql := squirrel.Update(self.TableName)
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

func (self *ViewModel) DeleteBy(db squirrel.BaseRunner, pred interface{}, args ...interface{}) (int64, error) {
	result, e := squirrel.Delete(self.TableName).Where(pred, args).RunWith(db).Exec()
	if nil != e {
		return 0, e
	}
	return result.RowsAffected()
}

type DbModel struct {
	ViewModel
	KeyNames []string
}

func (self *DbModel) UpdateByPrimaryKey(db squirrel.BaseRunner, values map[string]interface{}, keys ...interface{}) error {
	sql := squirrel.Update(self.TableName)
	if isPlaceholderWithDollar(db) {
		sql = sql.PlaceholderFormat(squirrel.Dollar)
	}

	for key, value := range values {
		sql = sql.Set(key, value)
	}

	cond := squirrel.Eq{}
	for idx, key := range keys {
		cond[self.KeyNames[idx]] = key
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

func (self *DbModel) DeleteByPrimaryKey(db squirrel.BaseRunner, keys ...interface{}) error {
	sql := squirrel.Delete(self.TableName)
	if isPlaceholderWithDollar(db) {
		sql = sql.PlaceholderFormat(squirrel.Dollar)
	}
	cond := squirrel.Eq{}
	for idx, key := range keys {
		cond[self.KeyNames[idx]] = key
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

func (self *ColumnModel) EQU(value interface{}) Expr {
	return Expr{Column: self, Operator: "=", Value: value}
}

func (self *ColumnModel) NEQ(value interface{}) Expr {
	return Expr{Column: self, Operator: "<>", Value: value}
}

func (self *ColumnModel) EXISTS(value interface{}) Expr {
	return Expr{Column: self, Operator: "EXISTS", Value: value}
}

type Expr struct {
	Column   *ColumnModel
	Operator string
	Value    interface{}
}

func (self Expr) ToSql() (string, []interface{}, error) {
	if sqlizer, ok := self.Value.(squirrel.Sqlizer); ok {
		sub_sqlstr, sub_args, e := sqlizer.ToSql()
		if nil != e {
			return "", nil, e
		}
		return self.Column.Name + " " + self.Operator + " " + sub_sqlstr, sub_args, nil
	}
	return self.Column.Name + " " + self.Operator + " ? ", []interface{}{self.Value}, nil
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
