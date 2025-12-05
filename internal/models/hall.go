package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type Hall struct {
	ID        int       `json:"id"`
	TheaterID int       `json:"theater_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Schedule  *Schedule `json:"schedule"`
}

type HallModel struct {
	db *sql.DB
}

var (
	ErrInvalidSchedule = errors.New("invalid schedule")
)

type Schedule struct {
	From  time.Time
	To    time.Time
	Shows []Show
}

func (s Schedule) IsFree(show Show) error {
	if show.StartTime.Before(s.From) || show.EndTime.After(s.To) {
		return fmt.Errorf("%w: schedule is out of the selected range (%v to %v)", ErrInvalidSchedule, s.From, s.To)
	}

	for _, sh := range s.Shows {
		if sh.StartTime.Before(show.EndTime) && show.StartTime.Before(sh.EndTime) {
			return fmt.Errorf(
				"%w: contradiction with screening of %v from %v to %v",
				ErrInvalidSchedule,
				sh.MovieTitle,
				sh.StartTime,
				sh.EndTime,
			)
		}
	}

	return nil
}

func (m *HallModel) Create(hall *Hall) error {
	query := `INSERT INTO halls(theater_id, name, code)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at`
	args := []any{hall.TheaterID, hall.Name, hall.Code}

	err := m.db.QueryRow(query, args...).Scan(
		&hall.ID,
		&hall.CreatedAt,
		&hall.UpdatedAt,
	)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "halls_theater_id_code_key"):
			return fmt.Errorf("%w: hall with code %v already exists", ErrDuplicate, hall.Code)
		case strings.Contains(err.Error(), "halls_theater_id_fkey"):
			return fmt.Errorf("%w: theater with id %v doesn't exist", ErrDuplicate, hall.TheaterID)
		default:
			slog.Error("SQL Database Failure", "error", err)
			return err
		}
	}

	return nil
}

func (m *HallModel) Find(id int) (*Hall, error) {
	return m.FindWithSchedule(id, time.Now(), time.Now().Add(time.Hour*24*7))
}

func (m *HallModel) FindByCode(theaterID int, code string) (*Hall, error) {
	return m.FindByCodeWithSchedule(theaterID, code, time.Now(), time.Now().Add(time.Hour*24*7))
}

func (m *HallModel) FindByCodeWithSchedule(theaterID int, code string, from, to time.Time) (*Hall, error) {
	query := `SELECT h.theater_id, h.name, h.id,
	h.created_at, h.updated_at, s.id, h.theater_id, h.id,
	h.code, m.id, m.title, m.imdb_link, s.start_time,
	s.end_time, s.created_at, s.updated_at
	FROM halls AS h
	JOIN shows AS s on s.hall_id = h.id
	JOIN movies AS m on s.movie_id = m.id
	WHERE h.theater_id = $1 AND h.code = $2 AND s.end_time >= $3 AND s.start_time <= $4`

	args := []any{theaterID, code, from, to}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		slog.Error("SQL Database Failure", "error", err)
		return nil, err
	}
	defer rows.Close()

	hall := &Hall{
		Code: code,
		Schedule: &Schedule{
			From:  from,
			To:    to,
			Shows: []Show{},
		},
	}

	type ShowDB struct {
		ID            sql.NullInt32
		TheaterID     sql.NullInt32
		HallID        sql.NullInt32
		HallCode      sql.NullString
		MovieID       sql.NullInt32
		MovieTitle    sql.NullString
		MovieIMDBLink sql.NullString
		StartTime     sql.NullTime
		EndTime       sql.NullTime
		CreatedAt     sql.NullTime
		UpdatedAt     sql.NullTime
	}

	first := true
	for rows.Next() {
		first = false
		var s ShowDB
		var show Show

		err := rows.Scan(
			&hall.TheaterID,
			&hall.Name,
			&hall.ID,
			&hall.CreatedAt,
			&hall.UpdatedAt,
			&s.ID,
			&s.TheaterID,
			&s.HallID,
			&s.HallCode,
			&s.MovieID,
			&s.MovieTitle,
			&s.MovieIMDBLink,
			&s.StartTime,
			&s.EndTime,
			&s.CreatedAt,
			&s.UpdatedAt,
		)

		if err != nil {
			slog.Error("Scan Failure", "error", err)
			return nil, err
		}

		if s.ID.Valid {
			show.ID = int(s.ID.Int32)
			show.TheaterID = int(s.TheaterID.Int32)
			show.HallID = int(s.HallID.Int32)
			show.HallCode = s.HallCode.String
			show.MovieID = int(s.MovieID.Int32)
			show.MovieTitle = s.MovieTitle.String
			show.MovieIMDBLink = s.MovieIMDBLink.String
			show.StartTime = s.StartTime.Time
			show.EndTime = s.EndTime.Time
			show.CreatedAt = s.CreatedAt.Time
			show.UpdatedAt = s.UpdatedAt.Time

			hall.Schedule.Shows = append(hall.Schedule.Shows, show)
		}

	}

	if first {
		return nil, ErrNotFound
	}

	return hall, nil
}

func (m *HallModel) FindWithSchedule(id int, from, to time.Time) (*Hall, error) {
	query := `SELECT h.theater_id, h.name, h.code,
	h.created_at, h.updated_at, s.id, h.theater_id, h.id,
	h.code, m.id, m.title, m.imdb_link, s.start_time,
	s.end_time, s.created_at, s.updated_at
	FROM halls AS h
	JOIN shows AS s on s.hall_id = h.id
	JOIN movies AS m on s.movie_id = m.id
	WHERE h.id = $1 AND s.end_time >= $2 AND s.start_time <= $3`

	rows, err := m.db.Query(query, id, from, to)
	if err != nil {
		slog.Error("SQL Database Failure", "error", err)
		return nil, err
	}
	defer rows.Close()

	hall := &Hall{
		ID: id,
		Schedule: &Schedule{
			From:  from,
			To:    to,
			Shows: []Show{},
		},
	}

	type ShowDB struct {
		ID            sql.NullInt32
		TheaterID     sql.NullInt32
		HallID        sql.NullInt32
		HallCode      sql.NullString
		MovieID       sql.NullInt32
		MovieTitle    sql.NullString
		MovieIMDBLink sql.NullString
		StartTime     sql.NullTime
		EndTime       sql.NullTime
		CreatedAt     sql.NullTime
		UpdatedAt     sql.NullTime
	}

	first := true
	for rows.Next() {
		first = false
		var s ShowDB
		var show Show

		err := rows.Scan(
			&hall.TheaterID,
			&hall.Name,
			&hall.Code,
			&hall.CreatedAt,
			&hall.UpdatedAt,
			&s.ID,
			&s.TheaterID,
			&s.HallID,
			&s.HallCode,
			&s.MovieID,
			&s.MovieTitle,
			&s.MovieIMDBLink,
			&s.StartTime,
			&s.EndTime,
			&s.CreatedAt,
			&s.UpdatedAt,
		)

		if err != nil {
			slog.Error("Scan Failure", "error", err)
			return nil, err
		}

		if s.ID.Valid {
			show.ID = int(s.ID.Int32)
			show.TheaterID = int(s.TheaterID.Int32)
			show.HallID = int(s.HallID.Int32)
			show.HallCode = s.HallCode.String
			show.MovieID = int(s.MovieID.Int32)
			show.MovieTitle = s.MovieTitle.String
			show.MovieIMDBLink = s.MovieIMDBLink.String
			show.StartTime = s.StartTime.Time
			show.EndTime = s.EndTime.Time
			show.CreatedAt = s.CreatedAt.Time
			show.UpdatedAt = s.UpdatedAt.Time

			hall.Schedule.Shows = append(hall.Schedule.Shows, show)
		}

	}

	if first {
		return nil, ErrNotFound
	}

	return hall, nil
}

func (m *HallModel) Update(hall *Hall) error {
	query := `UPDATE halls
	SET name = $1, code = $2, updated_at = NOW()
	WHERE id = $3 AND updated_at = $4
	RETURNING updated_at`
	args := []any{hall.Name, hall.Code, hall.ID, hall.UpdatedAt}

	err := m.db.QueryRow(query, args...).Scan(&hall.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		case strings.Contains(err.Error(), "halls_theater_id_code_key"):
			return fmt.Errorf("%w: hall with code %v already exists", ErrDuplicate, hall.Code)
		default:
			slog.Error("SQL Database Failure", "error", err)
			return err
		}
	}

	return nil
}

func (m *HallModel) Delete(id int) error {
	query := `DELETE FROM halls WHERE id = $1`

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

func (m *HallModel) DeleteByCode(code string) error {
	query := `DELETE FROM halls WHERE code = $1`

	result, err := m.db.Exec(query, code)
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
