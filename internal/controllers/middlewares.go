package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware Authenticate the user and add User struct to the request's context
func (h *Application) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		value, err := c.Cookie("SESSION_ID")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"unauthorized": "login credentials needed"})
			return
		}

		userID := h.cache.Get(value)
		if userID == "" {
			c.SetCookie("SESSION_ID", "", -1, "/", "", false, false)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"unauthorized": "login credentials needed"})
			return
		}

		// Should never fail since we store user ids in the cache.
		id, err := strconv.ParseInt(userID, 10, 32)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"unauthorized": "login credentials needed"})
			return
		}

		// Should never fail except if the account has been deleted
		user, err := h.models.Users.Find(int(id))
		if err != nil {
			switch {
			case errors.Is(err, models.ErrNotFound):
				c.SetCookie("SESSION_ID", "", -1, "/", "", false, false)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"unauthorized": "login credentials needed"})
			default:
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{})
			}
			return
		}

		c.Set("user", user)

		c.Next()
	}
}
