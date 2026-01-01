package services

import (
	"context"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

type MovieService struct {
	models   *models.Model
	provider MovieProvider
}

type MovieProvider interface {
	GetMovie(ctx context.Context, movieId string) (*models.Movie, error)
}
