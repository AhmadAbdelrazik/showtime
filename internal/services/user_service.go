package services

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

var (
	ErrInvalidUserRole   = errors.New("invalid user role")
	ErrIncorrectPassword = errors.New("incorrect password")
)

type UserService struct {
	models *models.Model
}

func (s *UserService) Signup(input SignupInput) (*models.User, error) {
	roles := []string{"customer", "admin", "manager"}
	if !slices.Contains(roles, input.Role) {
		return nil, fmt.Errorf(
			"%w: valid roles are (%v)",
			ErrInvalidUserRole,
			strings.Join(roles, " - "),
		)
	}

	password, err := models.NewPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username: input.Username,
		Email:    input.Email,
		Name:     input.Name,
		Role:     input.Role,
		Password: password,
	}

	// Add to database
	if err := s.models.Users.Create(user); err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicate):
			return nil, ErrDuplicate
		default:
			return nil, err
		}
	}

	slog.Info(
		"user has been created successfully",
		"username", user.Username,
		"id", user.ID,
		"email", user.Email,
	)

	return user, nil
}

func (s *UserService) Login(input LoginInput) (*models.User, error) {
	user, err := s.models.Users.FindByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return nil, ErrUserNotFound
		default:
			return nil, err
		}
	}

	if match := user.Password.ComparePassword(input.Password); !match {
		return nil, ErrIncorrectPassword
	}

	return user, nil
}

func (s *UserService) FindById(userId int) (*models.User, error) {
	return s.models.Users.Find(userId)
}

type SignupInput struct {
	Username string
	Email    string
	Name     string
	Role     string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}
