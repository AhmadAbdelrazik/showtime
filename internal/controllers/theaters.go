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

// TheatersSearch godoc
//
//	@Summary		Theaters Search
//	@Description	Search theaters based on name or city
//	@Tags			theaters
//	@Produce		json
//	@Param			name	query		string	flase	"theater name"
//	@Param			city	query		string	flase	"theater city"
//	@Param			sort_by	query		string	flase	"sort by city or name"
//	@Param			limit	query		integer	flase	"limit"
//	@Param			offset	query		integer	flase	"offset"
//	@Success		200		{object}		TheaterSearchResponse
//	@Failure		400		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/theaters [get]
func (h *Application) theatersSearchHandler(c *gin.Context) {
	var filters models.TheaterFilter

	if err := c.ShouldBindQuery(&filters); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	theaters, err := h.models.Theaters.Search(filters)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, TheaterSearchResponse{theaters})
}

// getTheater godoc
//
//	@Summary		Get Theater
//	@Description	Get Theater by ID
//	@Tags			theaters
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Success		200	{object}	models.Theater
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/api/theaters/{id} [get]
func (h *Application) getTheaterHandler(c *gin.Context) {
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

	c.JSON(http.StatusOK, theater)
}

// CreateTheater godoc
//
//	@Summary		Create Theater
//	@Description	Creates a new theater
//	@Tags			theaters
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateTheaterInput	true	"new theater data"
//	@Success		201		{object}	CreateTheaterResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		401		{object}	httputil.HTTPError
//	@Failure		403		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/theaters [post]
func (h *Application) createTheaterHandler(c *gin.Context) {
	var input CreateTheaterInput

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

	theater := &models.Theater{
		Name:      input.Name,
		City:      input.City,
		Address:   input.Address,
		ManagerID: user.ID,
	}

	if err := h.models.Theaters.Create(theater); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, CreateTheaterResponse{
		Message: "theater created sucessfully",
		Theater: *theater,
	})
}

// UpdateTheater godoc
//
//	@Summary		Update Theater
//	@Description	Update an existing theater
//	@Tags			theaters
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"theater id"
//	@Param			input	body		UpdateTheaterInput	true	"updated theater data"
//	@Success		201		{object}	UpdateTheaterResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		401		{object}	httputil.HTTPError
//	@Failure		403		{object}	httputil.HTTPError
//	@Failure		404		{object}	httputil.HTTPError
//	@Failure		409		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/theaters/{id} [put]
func (h *Application) updateTheaterHandler(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid parameter: (id must be integer)"))
		return
	}

	var input UpdateTheaterInput

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

	if theater.ManagerID != user.ID {
		httputil.NewError(
			c,
			http.StatusForbidden,
			errors.New("theater info can be updated by theater manager only"),
		)
		return
	}

	if input.Name != nil {
		theater.Name = *input.Name
	}
	if input.City != nil {
		theater.City = *input.City
	}
	if input.Address != nil {
		theater.Address = *input.Address
	}

	if err := h.models.Theaters.Update(theater); err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			httputil.NewError(c, http.StatusConflict, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, UpdateTheaterResponse{
		Message: "theater updated sucessfully",
		Theater: *theater,
	})
}

// deleteTheater godoc
//
//	@Summary		Delete Theater
//	@Description	Delete Theater by ID
//	@Tags			theaters
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Success		200	{object}	DeleteTheaterResponse
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		401	{object}	httputil.HTTPError
//	@Failure		403	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/api/theaters/{id} [delete]
func (h *Application) deleteTheaterHandler(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

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

	if theater.ManagerID != user.ID {
		httputil.NewError(
			c,
			http.StatusForbidden,
			errors.New("theater info can be updated by theater manager only"),
		)
		return
	}

	if err := h.models.Theaters.Delete(int(id)); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, DeleteTheaterResponse{Message: "Deleted Successfully"})
}

type TheaterSearchResponse struct {
	Theaters []models.Theater `json:"theaters"`
}

type CreateTheaterInput struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	Address string `json:"address"`
}

func (i *CreateTheaterInput) Validate(v *validator.Validator) {
	v.Check(len(strings.TrimSpace(i.Name)) > 0, "name", "required")
	v.Check(len(i.Name) <= 50, "name", "must be at most 50 characters")
	v.Check(len(i.Name) > 5, "name", "must be at least 5 characters")

	v.Check(len(strings.TrimSpace(i.City)) > 0, "city", "required")
	v.Check(len(i.City) <= 30, "city", "must be at most 30 characters")
	v.Check(len(i.City) > 5, "city", "must be at least 5 characters")

	v.Check(len(strings.TrimSpace(i.Address)) > 0, "address", "required")
	v.Check(len(i.Address) <= 100, "address", "must be at most 100 characters")
	v.Check(len(i.Address) > 5, "address", "must be at least 5 characters")
}

type CreateTheaterResponse struct {
	Message string         `json:"message"`
	Theater models.Theater `json:"theater"`
}

type UpdateTheaterInput struct {
	Name    *string `json:"name"`
	City    *string `json:"city"`
	Address *string `json:"address"`
}

func (i *UpdateTheaterInput) Validate(v *validator.Validator) {
	if i.Name != nil {
		v.Check(len(strings.TrimSpace(*i.Name)) > 0, "name", "required")
		v.Check(len(*i.Name) <= 50, "name", "must be at most 50 characters")
		v.Check(len(*i.Name) > 5, "name", "must be at least 5 characters")
	}

	if i.City != nil {
		v.Check(len(strings.TrimSpace(*i.City)) > 0, "city", "required")
		v.Check(len(*i.City) <= 30, "city", "must be at most 30 characters")
		v.Check(len(*i.City) > 5, "city", "must be at least 5 characters")
	}

	if i.Address != nil {
		v.Check(len(strings.TrimSpace(*i.Address)) > 0, "address", "required")
		v.Check(len(*i.Address) <= 100, "address", "must be at most 100 characters")
		v.Check(len(*i.Address) > 5, "address", "must be at least 5 characters")
	}
}

type UpdateTheaterResponse struct {
	Message string         `json:"message"`
	Theater models.Theater `json:"theater"`
}

type DeleteTheaterResponse struct {
	Message string `json:"message"`
}
