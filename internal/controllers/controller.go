package controllers

import (
	"log/slog"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/pkg/cache"
)

type Application struct {
	models *models.Model
	cache  *cache.Cache
}

func New(dsn string) (*Application, error) {
	model, err := models.New(dsn)
	if err != nil {
		slog.Error("failed to create model", slog.String("error", err.Error()))
		return nil, err
	}

	return &Application{
		models: model,
		cache:  cache.New(),
	}, nil
}
