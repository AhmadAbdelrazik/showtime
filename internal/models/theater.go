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

type Theater struct {
	ID        int       `json:"id"`
	ManagerID int       `json:"manager_id"`
	Manager   *User     `json:"-"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Halls     []Hall    `json:"halls"`
}

func (t Theater) HasHall(code string) bool {
	for _, h := range t.Halls {
		if h.Code == code {
			return true
		}
	}
	return false
}

type TheaterModel struct {
	db *sql.DB
}

func (m *TheaterModel) Create(theater *Theater) error {
	query := `INSERT INTO theaters(manager_id, name, city,
	address) VALUES ($1, $2, $3, $4) RETURNING id, created_at,
	updated_at`

	args := []any{
		theater.ManagerID,
		theater.Name,
		theater.City,
		theater.Address,
	}

	err := m.db.QueryRow(query, args...).Scan(
		&theater.ID,
		&theater.CreatedAt,
		&theater.UpdatedAt,
	)
	if err != nil {
		slog.Error("SQL Database Failure", "error", err)
		return err
	}

	return nil
}

func (m *TheaterModel) Search(f TheaterFilter) ([]Theater, error) {
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

	var theaters []Theater
	for rows.Next() {
		var theater Theater
		err := rows.Scan(
			&theater.ID,
			&theater.ManagerID,
			&theater.Name,
			&theater.City,
			&theater.Address,
			&theater.CreatedAt,
			&theater.UpdatedAt,
		)

		if err != nil {
			slog.Error("Scan Failure", "error", err)
			return nil, err
		}

		theaters = append(theaters, theater)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Scan Failure", "error", err)
		return nil, err
	}

	return theaters, nil
}

func (m *TheaterModel) Find(id int) (*Theater, error) {
	query := `SELECT t.manager_id, t.name, t.city, t.address, t.created_at,
	t.updated_at, u.id, u.username, u.email, u.name, u.created_at, u.updated_at,
	h.id, h.theater_id, h.name, h.code, h.created_at, h.updated_at
	FROM theaters AS t
	JOIN users AS u ON u.id = t.manager_id
	LEFT JOIN halls AS h ON t.id = h.theater_id
	WHERE t.id = $1 and t.deleted_at IS NULL`

	rows, err := m.db.Query(query, id)
	if err != nil {
		slog.Error("SQL Database Failure", "error", err)
		return nil, err
	}

	defer rows.Close()

	theater := &Theater{
		ID:      id,
		Manager: &User{},
		Halls:   []Hall{},
	}

	type HallDB struct {
		ID        sql.NullInt32
		TheaterID sql.NullInt32
		Name      sql.NullString
		Code      sql.NullString
		CreatedAt sql.NullTime
		UpdatedAt sql.NullTime
	}

	first := true
	for rows.Next() {
		first = false
		h := HallDB{}
		var hall Hall

		err := rows.Scan(
			&theater.ManagerID,
			&theater.Name,
			&theater.City,
			&theater.Address,
			&theater.CreatedAt,
			&theater.UpdatedAt,
			&theater.Manager.ID,
			&theater.Manager.Username,
			&theater.Manager.Email,
			&theater.Manager.Name,
			&theater.Manager.CreatedAt,
			&theater.Manager.UpdatedAt,
			&h.ID,
			&h.TheaterID,
			&h.Name,
			&h.Code,
			&h.CreatedAt,
			&h.UpdatedAt,
		)

		if err != nil {
			slog.Error("Scan Failure", "error", err)
			return nil, err
		}

		if h.ID.Valid {
			hall.ID = int(h.ID.Int32)
			hall.TheaterID = int(h.TheaterID.Int32)
			hall.Name = h.Name.String
			hall.Code = h.Code.String
			hall.CreatedAt = h.CreatedAt.Time
			hall.UpdatedAt = h.UpdatedAt.Time

			theater.Halls = append(theater.Halls, hall)
		}
	}

	if first {
		return nil, ErrNotFound
	}

	return theater, nil
}

func (m *TheaterModel) Update(theater *Theater) error {
	query := `UPDATE theaters 
	SET name = $1, city = $2, address = $3, updated_at = NOW()
	WHERE id = $4 AND updated_at = $5 AND deleted_at IS NULL
	RETURNING updated_at`
	args := []any{theater.Name, theater.City, theater.Address, theater.ID, theater.UpdatedAt}

	err := m.db.QueryRow(query, args...).Scan(&theater.UpdatedAt)
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

func (m *TheaterModel) Delete(id int) error {
	tx, err := m.db.Begin()
	if err != nil {
		slog.Error("SQL Database Failure", "error", err)
		return err
	}

	query := `UPDATE theaters SET deleted_at = NOW() WHERE id = $1`

	result, err := tx.Exec(query, id)
	if err != nil {
		tx.Rollback()
		slog.Error("SQL Database Failure", "error", err)
		return err
	}

	if rows, err := result.RowsAffected(); err != nil {
		tx.Rollback()
		slog.Error("SQL Database Failure", "error", err)
		return err
	} else if rows == 0 {
		tx.Rollback()
		return ErrNotFound
	}

	query = `UPDATE halls SET deleted_at = NOW() WHERE theater_id = $1`
	result, err = tx.Exec(query, id)
	if err != nil {
		tx.Rollback()
		slog.Error("SQL Database Failure", "error", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		slog.Error("SQL Database Failure", "error", err)
		return err
	}

	return nil
}

type TheaterFilter struct {
	Name   *string `form:"name"`
	City   *string `form:"city"`
	SortBy *string `form:"sort_by"`
	Limit  *uint   `form:"limit"`
	Offset *uint   `form:"offset"`
}

func (f *TheaterFilter) Validate(v *validator.Validator) {
	if f.SortBy != nil {
		validSortValues := []string{"name", "-name", "city", "-city"}
		sort := *f.SortBy
		v.Check(slices.Contains(validSortValues, sort), "sort", "invalid sort value")
	}

	if f.Limit != nil {
		v.Check(*f.Limit <= 100, "limit", "must be at most 100")
	}

	if f.Name != nil {
		v.Check(len(*f.Name) <= 50, "name", "must be at most 50 characters")
	}

	if f.City != nil {
		v.Check(len(*f.City) <= 30, "city", "must be at most 30 characters")
	}
}

func (f *TheaterFilter) Build() (string, []any, error) {
	q := sq.Select(`id, manager_id, name, city, address, created_at,
		updated_at`).From("theaters").Where("deleted_at IS NULL")

	if f.Name != nil {
		q = q.Where(sq.Expr(
			"to_tsvector('english', name) @@ plainto_tsquery('english', ?)",
			*f.Name,
		))
	}

	if f.City != nil {
		q = q.Where(sq.Expr(
			"to_tsvector('english', city) @@ plainto_tsquery('english', ?)",
			*f.City,
		))
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
