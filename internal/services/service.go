package services

import (
	"errors"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

var (
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrUserNotFound    = errors.New("user not found")
	ErrTheaterNotFound = errors.New("theater not found")
	ErrHallNotFound    = errors.New("hall not found")
	ErrMovieNotFound   = errors.New("movie not found")
	ErrShowNotFound    = errors.New("show not found")
	ErrDuplicate       = errors.New("duplicate resource")
	ErrEditConflict    = errors.New("edit conflict")
)

type Service struct {
	Theaters *TheaterService
	Shows    *ShowService
	Halls    *HallService
	Users    *UserService
}

func New(model *models.Model) *Service {

	return &Service{
		Theaters: &TheaterService{model},
		Shows:    &ShowService{model},
		Halls:    &HallService{model},
		Users:    &UserService{model},
	}
}
