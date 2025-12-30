package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/AhmadAbdelrazik/showtime/internal/httputil"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/pkg/validator"
	"github.com/gin-gonic/gin"
)

// getHall godoc
//
//	@Summary		Get Hall
//	@Description	Get Hall by ID
//	@Tags			halls
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Param			code	path		string	true	"hall code"
//	@Success		200	{object}	models.Hall
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/api/theaters/{id}/halls/{code} [get]
func (h *Application) getHallHandler(c *gin.Context) {
	hallCode := c.Param("code")
	theaterIdStr := c.Param("id")
	theaterID, err := strconv.ParseInt(theaterIdStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid theater id"))

	}

	hall, err := h.models.Halls.FindByCode(int(theaterID), hallCode)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, GetHallResponse{*hall})
}

// CreateHall godoc
//
//	@Summary		Create Hall
//	@Description	Creates a new theater's hall
//	@Tags			halls
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Param			input	body		CreateHallInput	true	"new hall data"
//	@Success		201		{object}	CreateHallResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		401		{object}	httputil.HTTPError
//	@Failure		403		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/theaters/{id}/halls [post]
func (h *Application) createHallHandler(c *gin.Context) {
	// parse input (parameter + body)
	var input CreateHallInput

	theaterIdStr := c.Param("id")
	theaterID, err := strconv.ParseInt(theaterIdStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid theater id"))

	}

	user := c.MustGet("user").(*models.User)

	// validate body input
	if err := c.ShouldBind(&input); err != nil {
		httputil.NewValidationError(c, input.Errors())
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		httputil.NewValidationError(c, v.Errors)
		return
	}

	hall := &models.Hall{
		TheaterID: int(theaterID),
		Name:      input.Name,
		Code:      input.Code,
	}

	// fetch theater from db
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
	if !isTheaterManagerOrAdmin(user, theater) {
		httputil.NewError(
			c,
			http.StatusForbidden,
			errors.New("creating halls is available for theater's manager only."),
		)
		return
	}

	// add hall
	if err := h.models.Halls.Create(hall); err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicate):
			httputil.NewError(c, http.StatusConflict, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	// return success
	c.JSON(http.StatusCreated, CreateHallResponse{"hall created successfully", *hall})
}

// UpdateHall godoc
//
//	@Summary		Update Hall
//	@Description	Updates a new theater's hall
//	@Tags			halls
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Param			input	body		UpdateHallInput	true	"new hall data"
//	@Success		201		{object}	UpdateHallResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		401		{object}	httputil.HTTPError
//	@Failure		403		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/theaters/{id}/halls/{code} [patch]
func (h *Application) updateHallHandler(c *gin.Context) {
	// parse input (parameter + body)
	var input UpdateHallInput

	hallCode := c.Param("code")
	theaterIdStr := c.Param("id")
	theaterID, err := strconv.ParseInt(theaterIdStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid theater id"))

	}

	user := c.MustGet("user").(*models.User)

	// validate body input
	if err := c.ShouldBind(&input); err != nil {
		httputil.NewValidationError(c, input.Errors())
		return
	}

	v := validator.New()
	if input.Validate(v); !v.Valid() {
		httputil.NewValidationError(c, v.Errors)
		return
	}

	// add hall
	hall, err := h.models.Halls.FindByCode(int(theaterID), hallCode)
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
	if !isHallManagerOrAdmin(user, hall) {
		httputil.NewError(
			c,
			http.StatusForbidden,
			errors.New("creating halls is available for theater's manager only."),
		)
		return
	}

	if input.Name != nil {
		hall.Name = *input.Name
	}

	if err := h.models.Halls.Update(hall); err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict),
			errors.Is(err, models.ErrDuplicate):
			httputil.NewError(c, http.StatusConflict, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	// return success
	c.JSON(http.StatusCreated, CreateHallResponse{"hall created successfully", *hall})
}

// deleteHall godoc
//
//	@Summary		Delete Hall
//	@Description	Delete theater's hall by Hall Code
//	@Tags			halls
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Param			code	path		string	true	"hall code"
//	@Success		200	{object}	DeleteHallResponse
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		401	{object}	httputil.HTTPError
//	@Failure		403	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/api/theaters/{id}/halls/{code} [delete]
func (h *Application) deleteHallResponse(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	hallCode := c.Param("code")
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid parameter: (id must be integer)"))
		return
	}

	theater, err := h.models.Theaters.Find(int(id))
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	if !isTheaterManagerOrAdmin(user, theater) {
		httputil.NewError(
			c,
			http.StatusForbidden,
			errors.New("hall can be removed only by the theater manager"),
		)
		return
	}

	if err := h.models.Halls.DeleteByCode(hallCode); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, DeleteHallResponse{Message: "Deleted Successfully"})
}

type GetHallResponse struct {
	Hall models.Hall `json:"hall"`
}

type CreateHallInput struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func (i *CreateHallInput) Validate(v *validator.Validator) {
	v.Check(len(strings.TrimSpace(i.Name)) > 0, "name", "required")
	v.Check(len(i.Name) <= 30, "name", "must be at most 50 characters")
	v.Check(len(i.Name) > 5, "name", "must be at least 5 characters")

	v.Check(len(strings.TrimSpace(i.Code)) > 0, "code", "required")
	v.Check(validator.AlphanumRX.MatchString(i.Code), "code", "must not contain any spaces or special characters")
	v.Check(len(i.Code) <= 10, "code", "must be at most 50 characters")
}

func (i *CreateHallInput) Errors() map[string]string {
	return map[string]string{
		"name": "required",
		"code": "required",
	}
}

type CreateHallResponse struct {
	Message string      `json:"message"`
	Hall    models.Hall `json:"hall"`
}

type UpdateHallInput struct {
	Name *string `json:"name"`
}

func (i *UpdateHallInput) Validate(v *validator.Validator) {
	if i.Name != nil {
		v.Check(len(strings.TrimSpace(*i.Name)) > 0, "name", "required")
		v.Check(len(*i.Name) <= 30, "name", "must be at most 50 characters")
		v.Check(len(*i.Name) > 5, "name", "must be at least 5 characters")
	}
}

func (i *UpdateHallInput) Errors() map[string]string {
	return map[string]string{
		"name": "required",
		"code": "required",
	}
}

type UpdateHallResponse struct {
	Message string      `json:"message"`
	Hall    models.Hall `json:"hall"`
}

type DeleteHallResponse struct {
	Message string `json:"message"`
}
