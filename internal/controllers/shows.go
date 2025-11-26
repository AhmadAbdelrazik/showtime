package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AhmadAbdelrazik/showtime/internal/httputil"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/pkg/validator"
	"github.com/gin-gonic/gin"
)

// CreateShow godoc
//
//	@Summary		Create Show
//	@Description	Creates a new theater's show
//	@Tags			shows
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Param			input	body		CreateShowInput	true	"new show data"
//	@Success		201		{object}	CreateShowResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		401		{object}	httputil.HTTPError
//	@Failure		403		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/theaters/{id}/shows [post]
func (h *Application) createShowHandler(c *gin.Context) {

	// 1. Check if the user is authorized
	// 2. Check if the theater and hall exist
	// 3. Check if the hall is available at the specified time
	// 3.1

	var input CreateShowInput

	theaterIdStr := c.Param("id")
	theaterID, err := strconv.ParseInt(theaterIdStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid theater id"))

	}
	user := c.MustGet("user").(*models.User)

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

	theater, err := h.models.Theaters.Find(int(theaterID))
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	// check authorization
	if theater.ManagerID != user.ID {
		httputil.NewError(
			c,
			http.StatusForbidden,
			errors.New("creating halls is available for theater's manager only."),
		)
		return
	}

	hall, err := h.models.Halls.FindByCode(theater.ID, input.HallCode)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

}

type CreateShowInput struct {
	MovieID  int
	HallCode string
	From     time.Time
	To       time.Time
}

func (i *CreateShowInput) Validate(v *validator.Validator) {

}
