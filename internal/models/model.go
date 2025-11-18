package models

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/Masterminds/squirrel"

	_ "github.com/lib/pq"
)

var (
	ErrDuplicate    = errors.New("duplicate")
	ErrNotFound     = errors.New("not found")
	ErrEditConflict = errors.New("edit conflict")
)

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

type Model struct {
	Users    *UserModel
	Theaters *TheaterModel
}

func New(dsn string) (*Model, error) {
	slog.Debug("Connecting to database")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		return nil, err
	}
	slog.Debug(
		"Connection to database established",
		slog.Group(
			"database",
			slog.String("database", "postgres"),
			slog.String("port", "5432"),
		),
	)

	if err := db.Ping(); err != nil {
		slog.Error("failed to ping database", slog.String("error", err.Error()))
		return nil, err
	}

	return &Model{
		Users:    &UserModel{db},
		Theaters: &TheaterModel{db},
	}, nil
}
