package services

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/AhmadAbdelrazik/showtime/internal/httputil"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

var (
	ErrUnauthorized    = errors.New("unauthorized access")
	ErrTheaterNotFound = errors.New("theater not found")
	ErrEditConflict    = errors.New("edit conflict")
)

type TheaterService struct {
	models *models.Model
}

func (s *TheaterService) Search(filters models.TheaterFilter) ([]models.Theater, error) {
	return s.models.Theaters.Search(filters)
}

func (s *TheaterService) Find(theaterID int) (*models.Theater, error) {
	return s.models.Theaters.Find(theaterID)
}

func (s *TheaterService) Create(user *models.User, theater *models.Theater) error {
	if !isManagerOrAdmin(user) {
		return ErrUnauthorized
	}

	return s.models.Theaters.Create(theater)
}

func (s *TheaterService) Update(user *models.User, theaterId int, input UpdateTheaterInput) (*models.Theater, error) {
	theater, err := s.models.Theaters.Find(theaterId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return nil, ErrTheaterNotFound
		default:
			return nil, err
		}
	}

	if !isTheaterManagerOrAdmin(user, theater) {
		return nil, fmt.Errorf("%w: theater info can be updated by theater manager only", ErrUnauthorized)
	}

	if input.Name != nil {
		theater.Name = *input.Name
	}
	if input.City != nil {
		theater.City = *input.City
	}
	if input.Address != nil {
		theater.Address = *input.Address
	}

	if err := s.models.Theaters.Update(theater); err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			return nil, ErrEditConflict
		default:
			return nil, err
		}
	}

	return theater, nil
}

func (s *TheaterService) Delete(user *models.User, theaterId int) error {
	theater, err := s.models.Theaters.Find(theaterId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return ErrTheaterNotFound
		default:
			return err
		}
	}

	if !isTheaterManagerOrAdmin(user, theater) {
		return fmt.Errorf("%w: theater can be deleted by theater manager only", ErrUnauthorized)
	}

	return s.models.Theaters.Delete(theaterId)
}

type UpdateTheaterInput struct {
	Name    *string
	City    *string
	Address *string
}
