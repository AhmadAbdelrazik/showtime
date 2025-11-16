package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Password  *Password `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserModel struct {
	db *sql.DB
}

func (m *UserModel) Create(u *User) error {
	query := `INSERT INTO users(username, name, email, hash) VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, updated_at`
	args := []any{u.Username, u.Name, u.Email, u.Password.hash}

	row := m.db.QueryRow(query, args...)
	err := row.Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_username_key"`):
			return fmt.Errorf("%w: user with this username already exists", ErrDuplicate)
		case strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_email_key"`):
			return fmt.Errorf("%w: user with this email already exists", ErrDuplicate)
		default:
			return err
		}
	}

	return nil
}

func (m *UserModel) Find(id int) (*User, error) {
	query := `SELECT username, name, email, hash, created_at,
	updated_at FROM users WHERE id = $1`

	user := &User{
		ID:       id,
		Password: &Password{},
	}

	row := m.db.QueryRow(query, id)
	err := row.Scan(
		&user.Username,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (m *UserModel) FindByUsername(username string) (*User, error) {
	query := `SELECT id, name, email, hash, created_at,
	updated_at FROM users WHERE username = $1`

	user := &User{
		Username: username,
		Password: &Password{},
	}

	row := m.db.QueryRow(query, username)
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}

func (m *UserModel) FindByEmail(email string) (*User, error) {
	query := `SELECT id, name, username, hash, created_at,
	updated_at FROM users WHERE email = $1`

	user := &User{
		Email:    email,
		Password: &Password{},
	}

	row := m.db.QueryRow(query, email)
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Username,
		&user.Password.hash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return user, nil
}
