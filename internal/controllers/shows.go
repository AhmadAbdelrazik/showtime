package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AhmadAbdelrazik/showtime/internal/httputil"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/pkg/validator"
	"github.com/gin-gonic/gin"
)

// ShowsSearch godoc
//
//	@Summary		Shows Search
//	@Description	Search shows based on movie or theater
//	@Tags			shows
//	@Produce		json
//	@Param			movie_title	query		string	flase	"show title"
//	@Param			theater_name	query		string	flase	"show title"
//	@Param			theater_city	query		string	flase	"show title"
//	@Param			start_date	query		string	flase	"show title"
//	@Param			end_date	query		string	flase	"show title"
//	@Param			sort_by	query		string	flase	"sort by title or release year"
//	@Param			limit	query		integer	flase	"limit"
//	@Param			offset	query		integer	flase	"offset"
//	@Success		200		{object}		SearchShowsResponse
//	@Failure		400		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/shows [get]
func (h *Application) searchShowsHandler(c *gin.Context) {
	var filters models.ShowFilter

	if err := c.ShouldBindQuery(&filters); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	v := validator.New()
	if filters.Validate(v); !v.Valid() {
		httputil.NewValidationError(c, v.Errors)
		return
	}

	shows, err := h.models.Shows.Search(filters)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, SearchShowsResponse{shows})
}

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

	hall, err := h.models.Halls.FindByCode(int(theaterID), input.HallCode)
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
			errors.New("creating shows is available for theater's manager only."),
		)
		return
	}

	movie, err := h.models.Movies.Find(input.MovieID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	movieDuration, err := time.ParseDuration(movie.Duration)
	if err != nil {
		panic(err)
	}

	if input.EndTime.Sub(input.StartTime) < movieDuration {
		v := validator.New()
		v.AddError("duration", fmt.Sprintf("movie duration (%v) is longer than reserved time", movie.Duration))
		httputil.NewValidationError(c, v.Errors)
		return
	}

	show := &models.Show{
		MovieID:   movie.ID,
		TheaterID: hall.TheaterID,
		HallCode:  input.HallCode,
		StartTime: input.StartTime,
		EndTime:   input.EndTime,
	}

	if err := h.models.Shows.Create(show); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusCreated, CreateShowResponse{
		Message: "show created successfully",
	})
}

// getShow godoc
//
//	@Summary		Get Show
//	@Description	Get Show by ID
//	@Tags			shows
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Param			show_id	path		string	true	"show code"
//	@Success		200	{object}	models.Show
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/api/theaters/{id}/shows/{show_id} [get]
func (h *Application) getShowHandler(c *gin.Context) {
	theaterIdStr := c.Param("id")
	theaterID, err := strconv.Atoi(theaterIdStr)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid theater id"))
	}
	showIdStr := c.Param("showId")
	showId, err := strconv.Atoi(showIdStr)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid show id"))
	}

	show, err := h.models.Shows.Find(showId)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	if show.TheaterID != theaterID {
		httputil.NewError(c, http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, show)
}

// deleteShow godoc
//
//	@Summary		Delete Show
//	@Description	Delete theater's show by Show ID
//	@Tags			shows
//	@Produce		json
//	@Param			id	path		int	true	"theater id"
//	@Param			show_id	path		string	true	"show id"
//	@Success		200	{object}	DeleteShowResponse
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		401	{object}	httputil.HTTPError
//	@Failure		403	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/api/theaters/{id}/shows/{show_id} [delete]
func (h *Application) deleteShowHandler(c *gin.Context) {
	theaterIdStr := c.Param("id")
	theaterID, err := strconv.Atoi(theaterIdStr)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid theater id"))
	}
	showIdStr := c.Param("showId")
	showId, err := strconv.Atoi(showIdStr)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid show id"))
	}

	user := c.MustGet("user").(*models.User)

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

	if !isTheaterManagerOrAdmin(user, theater) {
		httputil.NewError(
			c,
			http.StatusForbidden,
			errors.New("deleting shows is available for theater's manager only."),
		)
		return
	}

	if err := h.models.Shows.Delete(showId); err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, DeleteShowResponse{Message: "Deleted Successfully"})
}

type SearchShowsResponse struct {
	Shows []models.Show `json:"shows"`
}

type CreateShowInput struct {
	MovieID   int
	HallCode  string
	StartTime time.Time
	EndTime   time.Time
}

func (i *CreateShowInput) Validate(v *validator.Validator) {
	v.Check(len(strings.TrimSpace(i.HallCode)) > 0, "hall_code", "required")
	v.Check(validator.AlphanumRX.MatchString(i.HallCode), "hall_code", "must not contain any spaces or special characters")
	v.Check(len(i.HallCode) <= 10, "hall_code", "must be at most 50 characters")

	v.Check(i.StartTime.Before(i.EndTime), "start_time", "can't be after end_time")
	v.Check(i.EndTime.Sub(i.StartTime).Minutes() == 0, "duration", "duration difference must be in hours e.g. 1h, 3h, etc...")
}

type CreateShowResponse struct {
	Message string      `json:"message"`
	Show    models.Show `json:"show"`
}

type DeleteShowResponse struct {
	Message string `json:"message"`
}
