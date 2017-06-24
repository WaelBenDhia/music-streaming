package models

import "database/sql"

type scannable interface {
	Scan(...interface{}) error
}

type scanner interface {
	ScanFrom(scannable) error
}

type queriable interface {
	QueryRow(string, ...interface{}) *sql.Row
	Exec(string, ...interface{}) (sql.Result, error)
	Query(string, ...interface{}) (*sql.Rows, error)
}

type database interface {
	queriable
	Begin() (*sql.Tx, error)
}

func notFoundOrErr(err error) (bool, error) {
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
