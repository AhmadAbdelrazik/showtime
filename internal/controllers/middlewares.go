package controllers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/AhmadAbdelrazik/showtime/internal/httputil"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// AuthMiddleware Authenticate the user and add User struct to the request's context
func (h *Application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Debug("Parsing Session Cookie")
		value, err := c.Cookie("SESSION_ID")
		if err != nil {
			slog.Debug("Cookie was not found with the request")
			httputil.NewError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		slog.Debug("Check if cookie value exist in the cache")
		userID := h.cache.Get(value)
		if userID == "" {
			slog.Debug("Cookie was not found in the cache")
			c.SetCookie("SESSION_ID", "", -1, "/", "", false, false)
			httputil.NewError(c, http.StatusUnauthorized, errors.New("unauthorized"))
			c.Abort()
			return
		}

		// Should never fail since we store user ids in the cache.
		slog.Debug("Parsing user ID returning from the cache")
		id, err := strconv.ParseInt(userID, 10, 32)
		if err != nil {
			slog.Error("Cache contained a non-integer value for user id", "value", userID)
			httputil.NewError(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}

		// Should never fail except if the account has been deleted or internal server error
		slog.Debug("Fetching user from the database")
		user, err := h.services.Users.FindById(int(id))
		if err != nil {
			switch {
			case errors.Is(err, models.ErrNotFound):
				slog.Debug("Attempting to access a deleted user", "userID", id)
				c.SetCookie("SESSION_ID", "", -1, "/", "", false, false)
				httputil.NewError(c, http.StatusUnauthorized, err)
			default:
				httputil.NewError(c, http.StatusInternalServerError, err)
			}
			c.Abort()
			return
		}

		slog.Debug("adding user model in the request key-value store")
		c.Set("user", user)

		c.Next()
	}
}

// RateLimit limits the number of requests for each user
func RateLimit(limitRate float64, burst int, cleanupDuration time.Duration) gin.HandlerFunc {
	type Limiter struct {
		limit       *rate.Limiter
		lastRequest time.Time
	}

	freq := struct {
		m map[string]*Limiter
		sync.RWMutex
	}{
		m: make(map[string]*Limiter),
	}

	// cleanup
	go func() {
		freq.RWMutex.Lock()
		for ip, limiter := range freq.m {
			if time.Since(limiter.lastRequest) > cleanupDuration {
				delete(freq.m, ip)
			}
		}
		freq.RWMutex.Unlock()
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		freq.RWMutex.RLock()
		_, ok := freq.m[ip]
		freq.RWMutex.RUnlock()

		freq.RWMutex.Lock()
		if !ok {
			freq.m[ip] = &Limiter{
				limit:       rate.NewLimiter(rate.Limit(limitRate), burst),
				lastRequest: time.Now(),
			}
		} else {
			if !freq.m[ip].limit.Allow() {
				httputil.NewError(c, http.StatusTooManyRequests, errors.New("Too many requests"))
				c.Abort()
			}
			freq.m[ip].lastRequest = time.Now()
		}
		freq.RWMutex.Unlock()

		c.Next()
	}
}
