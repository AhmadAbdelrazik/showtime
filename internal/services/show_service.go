package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

var (
	ErrInvalidShowDuration = errors.New("movie duration is longer than reserved time")
)

type ShowService struct {
	models       *models.Model
	movieService *MovieService
}

func (s *ShowService) Search(filters models.ShowFilter) ([]models.Show, error) {
	return s.models.Shows.Search(filters)
}

func (s *ShowService) Create(user *models.User, theaterId int, input CreateShowInput) error {
	hall, err := s.models.Halls.FindByCode(theaterId, input.HallCode)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return ErrHallNotFound
		default:
			return nil
		}
	}

	// check authorization
	if !isHallManagerOrAdmin(user, hall) {
		return fmt.Errorf("%w: creating shows is available for theater's manager only.", ErrUnauthorized)
	}

	movie, err := s.movieService.Find(input.MovieID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return ErrMovieNotFound
		default:
			return err
		}
	}

	movieDuration, err := time.ParseDuration(movie.Runtime)
	if err != nil {
		panic(err)
	}

	if input.EndTime.Sub(input.StartTime) < movieDuration {
		return fmt.Errorf("%w (duration = %v)", ErrInvalidShowDuration, movie.Runtime)
	}

	show := &models.Show{
		MovieID:   movie.ImdbID,
		TheaterID: hall.TheaterID,
		HallCode:  input.HallCode,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
	}

	return s.models.Shows.Create(show)
}

func (s *ShowService) Find(theaterId, showId int) (*models.Show, error) {
	show, err := s.models.Shows.Find(showId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return nil, ErrShowNotFound
		default:
			return nil, err
		}
	}

	if show.TheaterID != theaterId {
		return nil, ErrShowNotFound
	}

	return show, nil
}

func (s *ShowService) Delete(user *models.User, theaterId, showId int) error {
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
		return fmt.Errorf("%w: deleting shows is available for theater's manager only.", ErrUnauthorized)
	}

	if err := s.models.Shows.Delete(showId); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			return ErrShowNotFound
		default:
			return err
		}
	}

	return nil
}

type CreateShowInput struct {
	MovieID   string
	HallCode  string
	StartTime time.Time
	EndTime   time.Time
}
