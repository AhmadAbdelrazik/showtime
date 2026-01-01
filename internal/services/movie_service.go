package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

type MovieService struct {
	models   *models.Model
	provider MovieProvider
}

func (s *MovieService) Find(movieId string) (*models.Movie, error) {
	movie, err := s.models.Movies.Find(movieId)
	if err == nil {
		return movie, nil
	} else if err != models.ErrNotFound {
		return movie, err
	}

	movie, err = s.provider.GetMovie(context.Background(), movieId)
	if err != nil {
		return nil, err
	}

	if err := s.models.Movies.Create(movie); err != nil {
		slog.Error(fmt.Sprintf("failed to store movie with id %v", movieId))
		return movie, err
	}

	return movie, nil
}
