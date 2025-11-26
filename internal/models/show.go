package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

var (
	ErrInvalidSchedule = errors.New("invalid schedule")
)

type Schedule struct {
	From  time.Time
	To    time.Time
	Shows []Show
}

func (s Schedule) IsFree(show Show) error {
	return nil
}

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

func (m *ShowModel) Create(show *Show) error {
	query := `INSERT INTO shows(movie_id, hall_id,
	movie_imdb_link, start_time, end_time)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at, updated_at`
	args := []any{
		show.MovieID,
		show.HallID,
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
		case strings.Contains(err.Error(), "shows_hall_id_fkey"):
			return fmt.Errorf("%w: hall with id %v was not found", ErrNotFound, show.HallID)
		case strings.Contains(err.Error(), "shows_movie_id_fkey"):
			return fmt.Errorf("%w: movie with id %v was not found", ErrNotFound, show.MovieID)
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
