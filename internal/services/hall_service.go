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

func (s *HallService) Create(input CreateHallInput) (*models.Hall, error) {
	theater, err := s.models.Theaters.Find(input.TheaterID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return nil, ErrTheaterNotFound
		default:
			return nil, err
		}
	}

	if !isTheaterManagerOrAdmin(input.User, theater) {
		return nil, fmt.Errorf("%w: creating halls is available for theater's manager only.", ErrUnauthorized)
	}

	hall := &models.Hall{
		TheaterID: input.TheaterID,
		ManagerID: input.User.ID,
		Name:      input.Hall.Name,
		Code:      input.Hall.Code,
		Seats:     newSeating(input.Hall.Rows, input.Hall.SeatsPerRow),
	}

	if err := s.models.Halls.Create(hall); err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicate):
			return nil, ErrDuplicate
		default:
			return nil, err
		}
	}

	return hall, nil
}

func (s *HallService) Update(input UpdateHallInput) (*models.Hall, error) {
	hall, err := s.models.Halls.FindByCode(input.TheaterId, input.HallCode)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return nil, ErrHallNotFound
		default:
			return nil, err
		}
	}

	// check authorization
	if !isHallManagerOrAdmin(input.User, hall) {
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

func newSeating(rows, seatsPerRow int) *models.Seating {
	seats := make([]models.Seat, rows*seatsPerRow)

	for i := range seats {
		seats[i].Row = string(rune('A' + i/seatsPerRow))
		seats[i].SeatNumber = i % seatsPerRow
	}

	return &models.Seating{
		Seats: seats,
	}
}

type CreateHallInput struct {
	User *models.User
	Hall struct {
		Name        string
		Code        string
		Rows        int
		SeatsPerRow int
	}
	TheaterID int
}

type UpdateHallInput struct {
	User      *models.User
	TheaterId int
	HallCode  string
	Name      *string
}
