package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
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
	WHERE t.id = $1`

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
	WHERE id = $4 AND updated_at = $5
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
	query := `DELETE FROM theaters WHERE id = $1`

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

type TheaterFilter struct {
	Name   *string `form:"name"`
	City   *string `form:"city"`
	SortBy *string `form:"sort_by"`
	Limit  *uint   `form:"limit"`
	Offset *uint   `form:"offset"`
}

func (f *TheaterFilter) Build() (string, []any, error) {
	q := psql.Select("id, manager_id, name, city, address, created_at, updated_at").From("theaters")
	counter := 1

	if f.Name != nil {
		q = q.Where(fmt.Sprintf("to_tsvector('english', name) @@ to_tsquery('english', $%v)", counter))
		counter++
	}

	if f.City != nil {
		q = q.Where(fmt.Sprintf("to_tsvector('english', city) @@ to_tsquery('english', $%v)", counter))
		counter++
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
	}

	return q.ToSql()
}
