package models

import (
	"database/sql"
)

type scanner interface {
	Scan(...interface{}) error
}

type scannerFrom interface {
	ScanFrom(scanner) error
}

type querier interface {
	QueryRow(string, ...interface{}) *sql.Row
	Query(string, ...interface{}) (*sql.Rows, error)
}

type executor interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type database interface {
	querier
	executor
	Begin() (*sql.Tx, error)
}

//Creator is a struct that has a method that creates a table
type Creator interface {
	CreateTable(ex executor) error
	CreatePriority() int
}

//Savable to database
type Savable interface {
	Save(database) error
}
