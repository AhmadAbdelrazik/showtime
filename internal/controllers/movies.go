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

// MoviesSearch godoc
//
//	@Summary		Movies Search
//	@Description	Search movies based on title
//	@Tags			movies
//	@Produce		json
//	@Param			title	query		string	flase	"movie title"
//	@Param			year	query		int	flase	"release year"
//	@Success		200		{object}		SearchMoviesResponse
//	@Failure		400		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/movies [get]
func (h *Application) searchMoviesHandler(c *gin.Context) {
	var filters movieFilter

	if err := c.ShouldBindQuery(&filters); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	v := validator.New()
	if filters.Validate(v); !v.Valid() {
		httputil.NewValidationError(c, v.Errors)
		return
	}

	movies, err := h.services.Movies.Search(*filters.Title, *filters.Year)
	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, SearchMoviesResponse{movies})
}

// getMovie godoc
//
//	@Summary		Get Movie
//	@Description	Get Movie by ID
//	@Tags			movies
//	@Produce		json
//	@Param			id	path		int	true	"movie id"
//	@Success		200	{object}	models.Movie
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/api/movies/{id} [get]
func (h *Application) getMovieHandler(c *gin.Context) {
	id := c.Param("id")

	movie, err := h.services.Movies.Find(id)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, movie)
}

// deleteMovie godoc
//
//	@Summary		Delete Movie
//	@Description	Delete Movie by ID
//	@Tags			movies
//	@Produce		json
//	@Param			id	path		int	true	"movie id"
//	@Success		200	{object}	DeleteMovieResponse
//	@Failure		400	{object}	httputil.HTTPError
//	@Failure		401	{object}	httputil.HTTPError
//	@Failure		404	{object}	httputil.HTTPError
//	@Failure		500	{object}	httputil.HTTPError
//	@Router			/api/movies/{id} [delete]
func (h *Application) deleteMovieHandler(c *gin.Context) {
	movieId := c.Param("id")
	user := c.MustGet("user").(*models.User)

	if err := h.services.Movies.Delete(user, movieId); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, DeleteMovieResponse{Message: "Deleted Successfully"})
}

type SearchMoviesResponse struct {
	Movies []models.Movie `json:"movies"`
}

type DeleteMovieResponse struct {
	Message string `json:"message"`
}

type movieFilter struct {
	Title *string `form:"title"`
	Year  *string `form:"year"`
}

func (f *movieFilter) Validate(v *validator.Validator) {
	if f.Title != nil {
		v.Check(len(*f.Title) <= 100, "title", "must be at most 100 characters")
	}

	if f.Year != nil {
		year, err := strconv.Atoi(*f.Year)
		if err != nil {
			v.AddError("year", "must be a valid year")
		}
		v.Check(
			year >= 1900 && year <= time.Now().Year(),
			"year",
			"must be between 1900 and this year",
		)
	}
}
