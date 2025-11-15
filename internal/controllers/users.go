package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/pkg/validator"
	"github.com/gin-gonic/gin"
)

func (h *Controller) PostUserSignupHandler(c *gin.Context) {
	// Input handling
	var input SignupInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		c.JSON(http.StatusBadRequest, v.Errors)
		return
	}

	user := &models.User{
		Username: input.Username,
		Email:    input.Email,
		Name:     input.Name,
	}

	password, err := models.NewPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	user.Password = password

	// Add to database

	if err := h.models.Users.Create(user); err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicate):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, nil)
		}
		return
	}

	sessionID := h.generateRandomString()
	h.cache.Set(sessionID, fmt.Sprint(user.ID))

	// session cookie
	cookie := &http.Cookie{
		Name:     "SESSION_ID",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		HttpOnly: true,
	}

	http.SetCookie(c.Writer, cookie)

	// Server Response

	c.JSON(http.StatusCreated, gin.H{})
}

type SignupInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (i SignupInput) Validate(v *validator.Validator) {
	v.Check(len(strings.TrimSpace(i.Username)) > 0, "username", "required")
	v.Check(len(i.Username) <= 50, "username", "must be at most 50 characters")

	v.Check(len(strings.TrimSpace(i.Name)) > 0, "name", "required")
	v.Check(len(i.Name) <= 50, "name", "must be at most 50 characters")

	v.Check(len(strings.TrimSpace(i.Email)) > 0, "email", "required")
	v.Check(validator.EmailRX.MatchString(i.Email), "email", "invalid email form")

	v.Check(len(strings.TrimSpace(i.Password)) > 0, "password", "required")
	v.Check(len(i.Password) >= 8, "password", "must be at least 8 characters")
	v.Check(len(i.Password) <= 50, "password", "must be at most 50 characters")
	v.Check(
		validator.PasswordRX.MatchString(i.Password),
		"password",
		"must contain at least 1 lowercase, 1 uppercase, 1 digit, and 1 special character",
	)
}

func (h *Controller) PostUserLoginHandler(c *gin.Context) {
	// Input handling
	var input LoginInput

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		c.JSON(http.StatusBadRequest, v.Errors)
		return
	}

	user, err := h.models.Users.FindByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid email or password"})
		default:
			c.JSON(http.StatusInternalServerError, nil)
		}
		return
	}

	if match := user.Password.ComparePassword(input.Password); !match {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid email or password"})
		return
	}

	sessionID := h.generateRandomString()
	h.cache.Set(sessionID, fmt.Sprint(user.ID))

	// session cookie
	cookie := &http.Cookie{
		Name:     "SESSION_ID",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		HttpOnly: true,
	}

	http.SetCookie(c.Writer, cookie)

	// Server Response
	c.JSON(http.StatusOK, gin.H{})
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (i LoginInput) Validate(v *validator.Validator) {
	v.Check(len(strings.TrimSpace(i.Email)) > 0, "email", "required")
	v.Check(validator.EmailRX.MatchString(i.Email), "email", "invalid email form")

	v.Check(len(strings.TrimSpace(i.Password)) > 0, "password", "required")
	v.Check(len(i.Password) >= 8, "password", "must be at least 8 characters")
	v.Check(len(i.Password) <= 50, "password", "must be at most 50 characters")
	v.Check(
		validator.PasswordRX.MatchString(i.Password),
		"password",
		"must contain at least 1 lowercase, 1 uppercase, 1 digit, and 1 special character",
	)
}
