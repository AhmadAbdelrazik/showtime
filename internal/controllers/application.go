package controllers

import (
	"github.com/AhmadAbdelrazik/showtime/internal/config"
	"github.com/AhmadAbdelrazik/showtime/internal/services"
	"github.com/AhmadAbdelrazik/showtime/pkg/cache"
)

type Application struct {
	services *services.Service
	cache    *cache.Cache
	cfg      *config.Config
}

func New(service *services.Service, cache *cache.Cache, config *config.Config) *Application {
	return &Application{
		services: service,
		cache:    cache,
		cfg:      config,
	}
}
