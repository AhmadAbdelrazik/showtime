package services

import (
	"errors"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

var (
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrTheaterNotFound = errors.New("theater not found")
	ErrHallNotFound    = errors.New("hall not found")
	ErrMovieNotFound   = errors.New("movie not found")
	ErrShowNotFound    = errors.New("show not found")
	ErrDuplicate       = errors.New("duplicate resource")
	ErrEditConflict    = errors.New("edit conflict")
)

type Service struct {
	Theaters *TheaterService
	Halls    *HallService
	Shows    *ShowService
}

func New(model *models.Model) *Service {

	return &Service{
		Theaters: &TheaterService{model},
		Shows:    &ShowService{model},
		Halls:    &HallService{model},
	}
}
