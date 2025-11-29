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
	query := `SELECT theater_id, name, code, created_at, updated_at
	FROM halls
	WHERE theater_id = $1 AND code = $2`

	hall := &Hall{
		ID: id,
	}

	err := m.db.QueryRow(query, id).Scan(
		&hall.TheaterID,
		&hall.Name,
		&hall.Code,
		&hall.CreatedAt,
		&hall.UpdatedAt,
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

	return hall, nil
}

func (m *HallModel) FindByCode(theaterID int, code string) (*Hall, error) {
	query := `SELECT id, name, created_at, updated_at
	FROM halls
	WHERE theater_id = $1 AND code = $2`

	hall := &Hall{
		TheaterID: theaterID,
		Code:      code,
	}

	err := m.db.QueryRow(query, theaterID, code).Scan(
		&hall.ID,
		&hall.Name,
		&hall.CreatedAt,
		&hall.UpdatedAt,
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
