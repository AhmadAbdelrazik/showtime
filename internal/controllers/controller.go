package controllers

import (
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/pkg/cache"
)

type Controller struct {
	models *models.Model
	cache  *cache.Cache
}

func New(dsn string) (*Controller, error) {
	model, err := models.New(dsn)
	if err != nil {
		return nil, err
	}

	return &Controller{
		models: model,
		cache:  cache.New(),
	}, nil
}
