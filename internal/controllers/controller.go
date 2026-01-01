package controllers

import (
	"github.com/AhmadAbdelrazik/showtime/internal/services"
	"github.com/AhmadAbdelrazik/showtime/pkg/cache"
)

type Application struct {
	services *services.Service
	cache    *cache.Cache
}

func New(service *services.Service) (*Application, error) {

	return &Application{
		services: service,
		cache:    cache.New(),
	}, nil
}
