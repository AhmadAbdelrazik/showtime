package services

import (
	"context"
	"errors"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

var ErrInvalidMovieId = errors.New("invalid movie id")

type MovieProvider interface {
	GetMovie(ctx context.Context, movieId string) (*models.Movie, error)
	Search(ctx context.Context, title, year string) ([]models.Movie, error)
}
