package controllers

import (
	"log/slog"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/internal/services"
	"github.com/AhmadAbdelrazik/showtime/pkg/cache"
)

type Application struct {
	services *services.Service
	cache    *cache.Cache
}

func New(dsn string) (*Application, error) {
	model, err := models.New(dsn)
	if err != nil {
		slog.Error("failed to create model", slog.String("error", err.Error()))
		return nil, err
	}

	return &Application{
		services: services.New(model),
		cache:    cache.New(),
	}, nil
}
