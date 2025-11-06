package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Hash      []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) ValidatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword(u.Hash, []byte(password)) == nil
}

type UserModel struct {
	db *sql.DB
}

func (m *UserModel) Create(u *User) error {
	query := `INSERT INTO users(username, name, email, hash) VALUES ($1, $2, $3, $4)
	RETURNING id, created_at`
	args := []any{u.Username, u.Name, u.Email, u.Hash}

	row := m.db.QueryRow(query, args...)
	err := row.Scan(&u.ID, &u.CreatedAt)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "duplicate"):
			return fmt.Errorf("%w: user with this username already exists", ErrDuplicateKey)
		default:
			return err
		}
	}

	return nil
}

func (m *UserModel) FindByID(id int) (*User, error) {
	query := `SELECT username, name, email, hash, created_at,
	updated_at FROM users WHERE id = $1`

	user := &User{
		ID: id,
	}

	row := m.db.QueryRow(query, id)
	err := row.Scan(
		&user.Username,
		&user.Name,
		&user.Email,
		&user.Hash,
		&user.CreatedAt,
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
	}

	row := m.db.QueryRow(query, username)
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Hash,
		&user.CreatedAt,
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
		Email: email,
	}

	row := m.db.QueryRow(query, email)
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Username,
		&user.Hash,
		&user.CreatedAt,
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
