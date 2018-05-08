package models

import (
	"database/sql"
	"errors"
)

var ErrIncompleteEntity = errors.New("incomplete entity")

func notFoundOrErr(err error) (bool, error) {
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func errOr(a, b error) error {
	if a == nil {
		return b
	}
	return a
}
