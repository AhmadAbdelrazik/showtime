package services

import (
	"errors"
	"fmt"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

type HallService struct {
	models *models.Model
}

func (s *HallService) Find(theaterId int, hallCode string) (*models.Hall, error) {
	hall, err := s.models.Halls.FindByCode(theaterId, hallCode)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return nil, ErrHallNotFound
		default:
			return nil, err
		}
	}
	return hall, nil
}

func (s *HallService) Create(user *models.User, hall *models.Hall, theaterId int) error {
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
		return fmt.Errorf("%w: creating halls is available for theater's manager only.", ErrUnauthorized)
	}

	if err := s.models.Halls.Create(hall); err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicate):
			return ErrDuplicate
		default:
			return err
		}
	}

	return nil
}

func (s *HallService) Update(user *models.User, theaterId int, hallCode string, input UpdateHallInput) (*models.Hall, error) {
	hall, err := s.models.Halls.FindByCode(theaterId, hallCode)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return nil, ErrHallNotFound
		default:
			return nil, err
		}
	}

	// check authorization
	if !isHallManagerOrAdmin(user, hall) {
		return nil, fmt.Errorf("%w: creating halls is available for theater's manager only.", ErrUnauthorized)
	}

	if input.Name != nil {
		hall.Name = *input.Name
	}

	if err := s.models.Halls.Update(hall); err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict),
			errors.Is(err, models.ErrDuplicate):
			return nil, ErrEditConflict
		default:
			return nil, err
		}
	}

	return hall, nil
}

func (s *HallService) Delete(user *models.User, theaterId int, hallCode string) error {
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
		return fmt.Errorf("%w: hall can be removed only by the theater manager", ErrUnauthorized)
	}

	if err := s.models.Halls.DeleteByCode(hallCode); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return ErrHallNotFound
		default:
			return err
		}
	}

	return nil
}

type UpdateHallInput struct {
	Name *string
}
