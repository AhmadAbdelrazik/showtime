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
	Find(ctx context.Context, movieId string) (*models.Movie, error)
	Search(ctx context.Context, movieName string) ([]models.Movie, error)
}
