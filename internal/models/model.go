package models

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
)

var (
	ErrDuplicateKey = errors.New("duplicate")
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
