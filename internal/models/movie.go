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
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Director    string    `json:"director"`
	ReleaseYear int       `json:"release_year"`
	Duration    string    `json:"duration"`
	IMDBLink    string    `json:"imdb_link"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
			&m.ID,
			&m.Title,
			&m.Director,
			&m.ReleaseYear,
			&m.Duration,
			&m.IMDBLink,
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
	query := `INSERT INTO movies(title, director, release_year, duration, imdb_link)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at`
	args := []any{movie.Title, movie.Director, movie.ReleaseYear, movie.Duration, movie.IMDBLink}

	err := m.db.QueryRow(query, args...).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.UpdatedAt,
	)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "movies_title_release_year_key"):
			return fmt.Errorf("%w: movie with the same title and release year already exists", ErrDuplicate)
		default:
			slog.Error("SQL Database Failure", "error", err)
			return err
		}
	}

	return nil
}

func (m *MovieModel) Find(id int) (*Movie, error) {
	query := `SELECT title, director, release_year, duration, imdb_link, created_at, updated_at
	FROM movies
	WHERE id = $1`

	movie := &Movie{ID: id}

	err := m.db.QueryRow(query, id).Scan(
		&movie.Title,
		&movie.Director,
		&movie.ReleaseYear,
		&movie.Duration,
		&movie.IMDBLink,
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

func (m *MovieModel) Update(movie *Movie) error {
	query := `UPDATE movies
	SET title = $1, director = $2, release_year = $3,
	duration = $4, imdb_link = $5, updated_at = NOW()
	WHERE id = $6 AND updated_at = $7
	RETURNING updated_at`
	args := []any{
		movie.Title,
		movie.Director,
		movie.ReleaseYear,
		movie.Duration,
		movie.IMDBLink,
		movie.ID,
		movie.UpdatedAt,
	}

	err := m.db.QueryRow(query, args...).Scan(&movie.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			slog.Error("SQL Database Failure", "error", err)
			return err
		}
	}

	return nil
}

func (m *MovieModel) Delete(id int) error {
	query := `DELETE FROM movies WHERE id = $1`

	result, err := m.db.Exec(query, id)
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
	Title       *string `form:"title"`
	Director    *string `form:"director"`
	ReleaseYear *int    `form:"release_year"`
	SortBy      *string `form:"sort_by"`
	Limit       *uint   `form:"limit"`
	Offset      *uint   `form:"offset"`
}

func (f *MovieFilter) Validate(v *validator.Validator) {
	if f.SortBy != nil {
		validSortValues := []string{
			"title",
			"-title",
			"release_year",
			"-release_year",
			"director",
			"-director",
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

	if f.Director != nil {
		v.Check(len(*f.Director) <= 50, "director", "must be at most 50 characters")
	}

	if f.ReleaseYear != nil {
		year := *f.ReleaseYear
		v.Check(
			year >= 1900 && year <= time.Now().Year(),
			"release_year",
			"must be between 1900 and this year",
		)
	}
}

func (f *MovieFilter) Build() (string, []any, error) {
	q := sq.Select("id, title, director, release_year, duration, imdb_link, created_at, updated_at").From("movies")

	if f.Title != nil {
		q = q.Where(sq.Expr(
			"to_tsvector('english', title) @@ plainto_tsquery('english', ?)",
			*f.Title,
		))
	}

	if f.Director != nil {
		q = q.Where(sq.Expr(
			"to_tsvector('english', director) @@ plainto_tsquery('english', ?)",
			*f.Director,
		))
	}

	if f.ReleaseYear != nil {
		q = q.Where(sq.Eq{"release_year": *f.ReleaseYear})
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
