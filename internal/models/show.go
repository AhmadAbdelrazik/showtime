package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/AhmadAbdelrazik/showtime/pkg/validator"
	sq "github.com/Masterminds/squirrel"
)

type Show struct {
	ID            int       `json:"id"`
	TheaterID     int       `json:"theater_id,omitempty"`
	HallID        int       `json:"hall_id"`
	HallCode      string    `json:"hall_code,omitempty"`
	MovieID       int       `json:"movie_id"`
	MovieTitle    string    `json:"movie_title,omitempty"`
	MovieIMDBLink string    `json:"movie_imdb_link,omitempty"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ShowModel struct {
	db *sql.DB
}

func (m *ShowModel) Search(f ShowFilter) ([]Show, error) {
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

	var shows []Show
	for rows.Next() {
		var show Show
		err := rows.Scan(
			&show.TheaterID,
			&show.HallID,
			&show.HallCode,
			&show.MovieID,
			&show.MovieTitle,
			&show.MovieIMDBLink,
			&show.StartTime,
			&show.EndTime,
			&show.CreatedAt,
			&show.UpdatedAt,
		)

		if err != nil {
			slog.Error("Scan Failure", "error", err)
			return nil, err
		}

		shows = append(shows, show)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Scan Failure", "error", err)
		return nil, err
	}

	return shows, nil
}

func (m *ShowModel) Create(show *Show) error {
	query := `INSERT INTO shows(movie_id, hall_id, start_time, end_time)
	SELECT m.id, h.id, $4, $5
	FROM movies AS m
	JOIN halls AS h ON h.theater_id = $2 AND h.code = $3
	WHERE m.id = $1
	RETURNING id, created_at, updated_at
	`

	args := []any{
		show.MovieID,
		show.TheaterID,
		show.HallCode,
		show.StartTime,
		show.EndTime,
	}

	err := m.db.QueryRow(query, args...).Scan(
		&show.ID,
		&show.CreatedAt,
		&show.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return fmt.Errorf("%w: check if movie and hall exists", ErrNotFound)
		default:
			slog.Error("SQL Database Failure", "error", err)
			return err
		}
	}

	return nil
}

func (m *ShowModel) Find(id int) (*Show, error) {
	query := `SELECT h.theater_id, s.hall_id, h.code,
	s.movie_id, m.title, m.imdb_link, s.start_time, s.end_time,
	s.created_at, s.updated_at
	FROM shows AS s
	JOIN movies AS m on m.id = s.movie_id
	JOIN halls AS h on h.id = s.hall_id
	JOIN theaters AS t on t.id = h.theater_id
	WHERE s.id = $1`

	show := &Show{
		ID: id,
	}

	err := m.db.QueryRow(query, id).Scan(
		&show.TheaterID,
		&show.HallID,
		&show.HallCode,
		&show.MovieID,
		&show.MovieTitle,
		&show.MovieIMDBLink,
		&show.StartTime,
		&show.EndTime,
		&show.CreatedAt,
		&show.UpdatedAt,
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

	return show, nil
}

func (m *ShowModel) Delete(id int) error {
	query := `DELETE FROM shows WHERE id = $1`

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

type ShowFilter struct {
	MovieTitle  *string    `form:"movie_title"`
	TheaterName *string    `form:"theater_name"`
	TheaterCity *string    `form:"theater_city"`
	StartDate   *time.Time `form:"start_date" time_format:"2006-01-02"`
	EndDate     *time.Time `form:"end_date" time_format:"2006-01-02"`
	SortBy      *string    `form:"sort_by"`
	Limit       *uint      `form:"limit"`
	Offset      *uint      `form:"offset"`
}

func (f *ShowFilter) Validate(v *validator.Validator) {
	if f.SortBy != nil {
		validSortValues := map[string]string{
			"movie_title":   "m.title",
			"-movie_title":  "-m.title",
			"theater_name":  "t.name",
			"-theater_name": "-t.name",
			"theater_city":  "t.city",
			"-theater_city": "-t.city",
			"date":          "s.start_time",
			"-date":         "-s.start_time",
		}
		sort := *f.SortBy
		if _, ok := validSortValues[sort]; !ok {
			v.AddError("sort", "invalid sort value")
		}
	}

	if f.Limit != nil {
		v.Check(*f.Limit <= 100, "limit", "must be at most 100")
	}

	if f.MovieTitle != nil {
		v.Check(len(*f.MovieTitle) <= 100, "movie_title", "must be at most 100 characters")
	}

	if f.TheaterName != nil {
		v.Check(len(*f.TheaterName) <= 50, "theater_name", "must be at most 50 characters")
	}

	if f.TheaterCity != nil {
		v.Check(len(*f.TheaterCity) <= 30, "theater_city", "must be at most 30 characters")
	}
}

func (f *ShowFilter) Build() (string, []any, error) {
	q := sq.Select(`h.theater_id, s.hall_id, h.code,
		s.movie_id, m.title, m.imdb_link, s.start_time, s.end_time,
		s.created_at, s.updated_at`).From(`shows AS s`).Join(`movies AS m on m.id =
		s.movie_id`).Join(`halls AS h on h.id = s.hall_id`).Join(`theaters AS t on
		t.id = h.theater_id`)

	if f.MovieTitle != nil {
		q = q.Where(sq.Expr(
			"to_tsvector('english', m.title) @@ plainto_tsquery('english', ?)",
			*f.MovieTitle,
		))
	}

	if f.TheaterName != nil {
		q = q.Where(sq.Expr(
			"to_tsvector('english', t.name) @@ plainto_tsquery('english', ?)",
			*f.TheaterName,
		))
	}

	if f.TheaterCity != nil {
		q = q.Where(sq.Expr(
			"to_tsvector('english', t.city) @@ plainto_tsquery('english', ?)",
			*f.TheaterCity,
		))
	}

	if f.StartDate != nil {
		q = q.Where("s.start_time > ?", *f.StartDate)
	} else {
		q = q.Where("s.start_time > NOW()")
	}

	if f.EndDate != nil {
		q = q.Where("s.start_time < ?", *f.EndDate)
	} else {
		q = q.Where("s.start_time < ?", time.Now().Add(time.Hour*24*7))
	}

	if f.SortBy != nil {
		validSortValues := map[string]string{
			"movie_title":   "m.title",
			"-movie_title":  "-m.title",
			"theater_name":  "t.name",
			"-theater_name": "-t.name",
			"theater_city":  "t.city",
			"-theater_city": "-t.city",
			"date":          "s.start_time",
			"-date":         "-s.start_time",
		}
		sort, ok := validSortValues[*f.SortBy]
		if !ok {
			panic("invalid sort value")
		}

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
