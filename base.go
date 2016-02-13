package main

import (
	"errors"

	"github.com/Masterminds/squirrel"
)

var ErrNotUpdated = errors.New("no record is updated")
var ErrNotDeleted = errors.New("no record is deleted")

func isPlaceholderWithDollar(value interface{}) bool {
	return true
}

type DbModel struct {
	TableName   string
	ColumnNames []string
	KeyNames    []string
}

func (self *DbModel) UpdateBy(db squirrel.BaseRunner, values map[string]interface{}, pred interface{}, args ...interface{}) (int64, error) {
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

func (self *DbModel) UpdateByKey(db squirrel.BaseRunner, values map[string]interface{}, keys ...interface{}) error {
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

func (self *DbModel) DeleteBy(db squirrel.BaseRunner, pred interface{}, args ...interface{}) (int64, error) {
	result, e := squirrel.Delete(self.TableName).Where(pred, args).RunWith(db).Exec()
	if nil != e {
		return 0, e
	}
	return result.RowsAffected()
}

func (self *DbModel) DeleteByKey(db squirrel.BaseRunner, keys ...interface{}) error {
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
