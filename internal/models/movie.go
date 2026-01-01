package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/AhmadAbdelrazik/showtime/pkg/validator"
	sq "github.com/Masterminds/squirrel"
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

func (m *MovieModel) Search(f MovieFilter) ([]Movie, error) {
	query, args, err := f.Build()
	if err != nil {
		slog.Error("filter build error", "filter", query)
		return nil, err
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		slog.Error("SQL Database Failure", "error", err)
		return nil, err
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var m Movie
		err := rows.Scan(
			&m.ImdbID,
			&m.Title,
			&m.Year,
			&m.Rated,
			&m.Runtime,
			&m.Genre,
			&m.Director,
			&m.Poster,
			&m.CreatedAt,
			&m.UpdatedAt,
		)

		if err != nil {
			slog.Error("Scan Failure", "error", err)
			return nil, err
		}

		movies = append(movies, m)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Scan Failure", "error", err)
		return nil, err
	}

	return movies, nil
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
	query := `UPDATE movies SET deleted_at = NOW() WHERE imdb_id = $1`

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

type MovieFilter struct {
	Title  *string `form:"title"`
	Year   *int    `form:"year"`
	SortBy *string `form:"sort_by"`
	Limit  *uint   `form:"limit"`
	Offset *uint   `form:"offset"`
}

func (f *MovieFilter) Validate(v *validator.Validator) {
	if f.SortBy != nil {
		validSortValues := []string{
			"title",
			"-title",
			"year",
			"-year",
		}
		sort := *f.SortBy
		v.Check(slices.Contains(validSortValues, sort), "sort", "invalid sort value")
	}

	if f.Limit != nil {
		v.Check(*f.Limit <= 100, "limit", "must be at most 100")
	}

	if f.Title != nil {
		v.Check(len(*f.Title) <= 100, "title", "must be at most 100 characters")
	}

	if f.Year != nil {
		year := *f.Year
		v.Check(
			year >= 1900 && year <= time.Now().Year(),
			"release_year",
			"must be between 1900 and this year",
		)
	}
}

func (f *MovieFilter) Build() (string, []any, error) {
	q := sq.Select(`imdb_id, title, year, rated, runtime, genre, director, poster
		created_at, updated_at`).From("movies").Where("deleted_at IS NULL")

	if f.Title != nil {
		q = q.Where(sq.Expr(
			"to_tsvector('english', title) @@ plainto_tsquery('english', ?)",
			*f.Title,
		))
	}

	if f.Year != nil {
		q = q.Where(sq.Eq{"release_year": *f.Year})
	}

	if f.SortBy != nil {
		sort := *f.SortBy
		if strings.HasPrefix(sort, "-") {
			sort, _ = strings.CutPrefix(sort, "-")
			q = q.OrderBy(fmt.Sprint(sort, " DESC"))
		} else {
			q = q.OrderBy(sort)
		}
	}

	if f.Limit != nil {
		q = q.Limit(uint64(*f.Limit))
	} else {
		q = q.Limit(20)
	}

	if f.Offset != nil {
		q = q.Offset(uint64(*f.Offset))
	}

	return q.PlaceholderFormat(sq.Dollar).ToSql()
}
