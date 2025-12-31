package controllers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/AhmadAbdelrazik/showtime/internal/httputil"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/internal/services"
	"github.com/AhmadAbdelrazik/showtime/pkg/validator"
	"github.com/gin-gonic/gin"
)

// UserSignup godoc
//
//	@Summary		User Signup
//	@Description	Registers a new user account
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		SignupInput	true	"User signup data"
//	@Success		201		{object}	SignupResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		404		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/signup [post]
func (h *Application) userSignupHandler(c *gin.Context) {
	// Input handling
	var input SignupInput

	if err := c.ShouldBind(&input); err != nil {
		v := validator.New()
		input.Validate(v)
		httputil.NewValidationError(c, v.Errors)
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		httputil.NewValidationError(c, v.Errors)
		return
	}

	user, err := h.services.Users.Signup(services.SignupInput(input))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidUserRole):
			v := validator.New()
			v.AddError("role", err.Error())
			httputil.NewValidationError(c, v.Errors)
		case errors.Is(err, services.ErrDuplicate):
			httputil.NewError(c, http.StatusConflict, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	h.addAuthSessionId(user, c)
	// Server Response

	c.JSON(http.StatusCreated, SignupResponse{
		Message: "Created successfully",
		User:    *user,
	})
}

// UserLogin godoc
//
//	@Summary		User Login
//	@Description	Login existing users to their accounts
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			input	body		LoginInput	true	"User login data"
//	@Success		200		{object}	LoginResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		403		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/login [post]
func (h *Application) userLoginHandler(c *gin.Context) {
	// Input handling
	var input LoginInput

	if err := c.ShouldBind(&input); err != nil {
		v := validator.New()
		input.Validate(v)
		httputil.NewValidationError(c, v.Errors)
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		httputil.NewValidationError(c, v.Errors)
		return
	}

	user, err := h.services.Users.Login(services.LoginInput(input))
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUserNotFound),
			errors.Is(err, services.ErrIncorrectPassword):
			httputil.NewError(c, http.StatusForbidden, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	h.addAuthSessionId(user, c)

	// Server Response
	c.JSON(http.StatusOK, LoginResponse{"logged in successfully"})
}

// UserLogout godoc
//
//	@Summary		User Logout
//	@Description	Logout users from the system.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	LogoutResponse
//	@Router			/api/logout [get]
func (h *Application) userLogoutHandler(c *gin.Context) {
	// session cookie
	cookie := &http.Cookie{
		Name:     "SESSION_ID",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	}

	http.SetCookie(c.Writer, cookie)

	// Server Response
	c.JSON(http.StatusOK, LogoutResponse{"Logged out successfully"})
}

// UserDetailsHandler godoc
//
//	@Summary		User Details
//	@Description	Get details of the current user
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	models.User
//	@Router			/api/user-info [get]
func (a *Application) UserDetailsHandler(c *gin.Context) {
	slog.Debug("retreiving user model from the request key-value")
	user := c.MustGet("user").(*models.User)
	slog.Debug("retreived successfully")

	c.JSON(http.StatusOK, user)
}

func (h *Application) addAuthSessionId(user *models.User, c *gin.Context) {
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
}

type SignupInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

type SignupResponse struct {
	Message string      `json:"message"`
	User    models.User `json:"user"`
}

func (i SignupInput) Validate(v *validator.Validator) {
	v.Check(len(strings.TrimSpace(i.Username)) > 0, "username", "required")
	v.Check(len(i.Username) <= 50, "username", "must be at most 50 characters")

	v.Check(len(strings.TrimSpace(i.Name)) > 0, "name", "required")
	v.Check(len(i.Name) <= 50, "name", "must be at most 50 characters")

	v.Check(len(strings.TrimSpace(i.Email)) > 0, "email", "required")
	v.Check(validator.EmailRX.MatchString(i.Email), "email", "invalid email form")

	v.Check(len(strings.TrimSpace(i.Role)) > 0, "role", "required")
	v.Check(len(i.Role) <= 10, "role", "must be at most 10 characters")

	v.Check(len(strings.TrimSpace(i.Password)) > 0, "password", "required")
	v.Check(len(i.Password) >= 8, "password", "must be at least 8 characters")
	v.Check(len(i.Password) <= 50, "password", "must be at most 50 characters")
	v.Check(
		validator.LowerRX.MatchString(i.Password),
		"password",
		"must contain at least 1 lowercase character",
	)
	v.Check(
		validator.UpperRX.MatchString(i.Password),
		"password",
		"must contain at least 1 uppercase character",
	)
	v.Check(
		validator.NumberRX.MatchString(i.Password),
		"password",
		"must contain at least a number",
	)
	v.Check(
		validator.SpecialRX.MatchString(i.Password),
		"password",
		"must contain at least 1 special character ( !@#$%&* )",
	)
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
}

func (i LoginInput) Validate(v *validator.Validator) {
	v.Check(len(strings.TrimSpace(i.Email)) > 0, "email", "required")
	v.Check(validator.EmailRX.MatchString(i.Email), "email", "invalid email form")

	v.Check(len(strings.TrimSpace(i.Password)) > 0, "password", "required")
	v.Check(len(i.Password) >= 8, "password", "must be at least 8 characters")
	v.Check(len(i.Password) <= 50, "password", "must be at most 50 characters")
	v.Check(
		validator.LowerRX.MatchString(i.Password),
		"password",
		"must contain at least 1 lowercase character",
	)
	v.Check(
		validator.UpperRX.MatchString(i.Password),
		"password",
		"must contain at least 1 uppercase character",
	)
	v.Check(
		validator.NumberRX.MatchString(i.Password),
		"password",
		"must contain at least a number",
	)
	v.Check(
		validator.SpecialRX.MatchString(i.Password),
		"password",
		"must contain at least 1 special character ( !@#$%&* )",
	)
}

type LogoutResponse struct {
	Message string `json:"message"`
}
