package models

import (
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"
)

type Movie struct {
	ImdbID     string    `json:"imdb_id"`
	Title      string    `json:"title"`
	Year       string    `json:"year"`
	Rated      string    `json:"rated"`
	Runtime    string    `json:"runtime"`
	Genre      string    `json:"genre"`
	Director   string    `json:"director"`
	Poster     string    `json:"poster"`
	ImdbRating string    `json:"imdb_rating"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MovieModel struct {
	db *sql.DB
}

func (m *MovieModel) Create(movie *Movie) error {
	query := `INSERT INTO movies(imdb_id, title, year, rated, runtime, director, poster, imdb_rating)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING created_at, updated_at`
	args := []any{
		movie.ImdbID,
		movie.Title,
		movie.Year,
		movie.Rated,
		movie.Runtime,
		movie.Director,
		movie.Poster,
		movie.ImdbRating,
	}

	err := m.db.QueryRow(query, args...).Scan(
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "movies_pkey"):
			return ErrDuplicate
		default:
			slog.Error("SQL Database Failure", "error", err)
			return err
		}
	}

	return nil
}

func (m *MovieModel) Find(imdbId string) (*Movie, error) {
	query := `SELECT title, year, rated, runtime, genre, director,
	poster created_at, updated_at
	FROM movies
	WHERE imdb_id = $1 AND deleted_at IS NULL`

	movie := &Movie{ImdbID: imdbId}

	err := m.db.QueryRow(query, imdbId).Scan(
		&movie.Title,
		&movie.Year,
		&movie.Rated,
		&movie.Runtime,
		&movie.Genre,
		&movie.Director,
		&movie.Poster,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			slog.Error("SQL Database Failure", "error", err)
			return nil, err
		}
	}

	return movie, nil
}

func (m *MovieModel) Delete(imdbId string) error {
	query := `DELETE FROM movies WHERE imdb_id = $1`

	result, err := m.db.Exec(query, imdbId)
	if err != nil {
		slog.Error("SQL Database Failure", "error", err)
		return err
	}

	if rows, err := result.RowsAffected(); err != nil {
		slog.Error("SQL Database Failure", "error", err)
		return err
	} else if rows == 0 {
		return ErrNotFound
	}

	return nil
}
