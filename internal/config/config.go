package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	ErrConfigError = errors.New("config error")
)

type Config struct {
	DSN         string
	Port        string
	Environment string
	OmdbApiKey  string
	RateLimit   struct {
		Enabled         bool
		Rate            float64
		Burst           int
		CleanupDuration time.Duration
	}
}

func Load() (*Config, error) {
	godotenv.Load()

	rateLimitEnabled, err := strconv.ParseBool(os.Getenv("RATELIMIT_ENABLED"))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse RATELIMIT_ENABLED)", ErrConfigError)
	}

	rateLimitrate, err := strconv.ParseFloat(os.Getenv("RATELIMIT_RATE"), 64)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse RATELIMIT_RATE)", ErrConfigError)
	}
	rateLimitBurst, err := strconv.Atoi(os.Getenv("RATELIMIT_BURST"))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse RATELIMIT_BURST)", ErrConfigError)
	}
	rateLimitCleanupDuration, err := time.ParseDuration(os.Getenv("RATELIMIT_CLEANUP_DURATION"))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse RATELIMIT_CLEANUP_DURATION)", ErrConfigError)
	}

	return &Config{
		DSN:         os.Getenv("DB_DSN"),
		Port:        os.Getenv("PORT"),
		Environment: os.Getenv("ENVIRONMENT"),
		OmdbApiKey:  os.Getenv("OMDB_APIKEY"),
		RateLimit: struct {
			Enabled         bool
			Rate            float64
			Burst           int
			CleanupDuration time.Duration
		}{
			Enabled:         rateLimitEnabled,
			Rate:            rateLimitrate,
			Burst:           rateLimitBurst,
			CleanupDuration: rateLimitCleanupDuration,
		},
	}, nil
}
