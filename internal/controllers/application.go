package controllers

import (
	"time"

	"github.com/AhmadAbdelrazik/showtime/internal/services"
	"github.com/AhmadAbdelrazik/showtime/pkg/cache"
)

type RateLimit struct {
	Rate            float64
	Burst           int
	CleanupDuration time.Duration
	Enabled         bool
}

type Config struct {
	RateLimit RateLimit
}

type Application struct {
	services *services.Service
	cache    *cache.Cache
	cfg      *Config
}

func New(service *services.Service, cache *cache.Cache, config *Config) *Application {
	return &Application{
		services: service,
		cache:    cache,
		cfg:      config,
	}
}
