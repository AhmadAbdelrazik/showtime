package services

import (
	"errors"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

var (
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrTheaterNotFound = errors.New("theater not found")
	ErrHallNotFound    = errors.New("hall not found")
	ErrDuplicate       = errors.New("duplicate resource")
	ErrEditConflict    = errors.New("edit conflict")
)

type Service struct {
	Theaters *TheaterService
	Halls    *HallService
}

func New(model *models.Model) *Service {

	return &Service{
		Theaters: &TheaterService{model},
		Halls:    &HallService{model},
	}
}
