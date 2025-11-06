package models

import (
	"database/sql"
	"errors"
)

var (
	ErrDuplicateKey = errors.New("duplicate key")
	ErrNotFound     = errors.New("not found")
)

type Model struct {
	Users *UserModel
}

func New(dsn string) (*Model, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Model{
		Users: &UserModel{db},
	}, nil
}
