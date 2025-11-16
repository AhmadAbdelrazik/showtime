package controllers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware Authenticate the user and add User struct to the request's context
func (h *Application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		slog.Debug("Parsing Session Cookie")
		value, err := c.Cookie("SESSION_ID")
		if err != nil {
			slog.Debug("Cookie was not found with the request")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"unauthorized": "login credentials needed"})
			return
		}

		slog.Debug("Check if cookie value exist in the cache")
		userID := h.cache.Get(value)
		if userID == "" {
			slog.Debug("Cookie was not found in the cache")
			c.SetCookie("SESSION_ID", "", -1, "/", "", false, false)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"unauthorized": "login credentials needed"})
			return
		}

		// Should never fail since we store user ids in the cache.
		slog.Debug("Parsing user ID returning from the cache")
		id, err := strconv.ParseInt(userID, 10, 32)
		if err != nil {
			slog.Error("Cache contained a non-integer value for user id", "value", userID)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"unauthorized": "login credentials needed"})
			return
		}

		// Should never fail except if the account has been deleted or internal server error
		slog.Debug("Fetching user from the database")
		user, err := h.models.Users.Find(int(id))
		if err != nil {
			switch {
			case errors.Is(err, models.ErrNotFound):
				slog.Debug("Attempting to access a deleted user", "userID", id)
				c.SetCookie("SESSION_ID", "", -1, "/", "", false, false)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"unauthorized": "login credentials needed"})
			default:
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			}
			return
		}

		slog.Debug("adding user model in the request key-value store")
		c.Set("user", user)

		c.Next()
	}
}
