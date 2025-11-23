package controllers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AhmadAbdelrazik/showtime/internal/httputil"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/pkg/validator"
	"github.com/gin-gonic/gin"
)

// MoviesSearch godoc
//
//	@Summary		Movies Search
//	@Description	Search movies based on name or city
//	@Tags			movies
//	@Produce		json
//	@Param			title	query		string	flase	"movie title"
//	@Param			director	query		string	flase	"movie director"
//	@Param			release_year	query		int	flase	"release year"
//	@Param			sort_by	query		string	flase	"sort by title or release year"
//	@Param			limit	query		integer	flase	"limit"
//	@Param			offset	query		integer	flase	"offset"
//	@Success		200		{object}		SearchMoviesResponse
//	@Failure		400		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/movies [get]
func (h *Application) searchMoviesHandler(c *gin.Context) {
	var filters models.MovieFilter

	if err := c.ShouldBindQuery(&filters); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err)
		return
	}

	v := validator.New()
	if filters.Validate(v); !v.Valid() {
		httputil.NewValidationError(c, v.Errors)
		return
	}

	movies, err := h.models.Movies.Search(filters)
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
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid parameter: (id must be integer)"))
		return
	}

	movie, err := h.models.Movies.Find(int(id))
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

// CreateMovie godoc
//
//	@Summary		Create Movie
//	@Description	Creates a new movie
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			input	body		CreateMovieInput	true	"new movie data"
//	@Success		201		{object}	CreateMovieResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		401		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/movies [post]
func (h *Application) createMovieHandler(c *gin.Context) {
	var input CreateMovieInput

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

	movie := &models.Movie{
		Title:       input.Title,
		Director:    input.Director,
		ReleaseYear: *input.ReleaseYear,
		IMDBLink:    input.IMDBLink,
	}

	// error is checked in the validator
	duration, _ := time.ParseDuration(input.Duration)
	movie.Duration = duration

	if err := h.models.Movies.Create(movie); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusCreated, CreateMovieResponse{
		Message: "movie created sucessfully",
		Movie:   *movie,
	})
}

// UpdateMovie godoc
//
//	@Summary		Update Movie
//	@Description	Update an existing movie
//	@Tags			movies
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"movie id"
//	@Param			input	body		UpdateMovieInput	true	"updated movie data"
//	@Success		201		{object}	UpdateMovieResponse
//	@Failure		400		{object}	httputil.ValidationError
//	@Failure		401		{object}	httputil.HTTPError
//	@Failure		404		{object}	httputil.HTTPError
//	@Failure		409		{object}	httputil.HTTPError
//	@Failure		500		{object}	httputil.HTTPError
//	@Router			/api/movies/{id} [patch]
func (h *Application) updateMovieHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid parameter: (id must be integer)"))
		return
	}

	var input UpdateMovieInput

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

	movie, err := h.models.Movies.Find(int(id))
	if err != nil {
		switch {
		case errors.Is(err, models.ErrNotFound):
			httputil.NewError(c, http.StatusNotFound, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Director != nil {
		movie.Director = *input.Director
	}
	if input.ReleaseYear != nil {
		movie.ReleaseYear = *input.ReleaseYear
	}
	if input.Duration != nil {
		// error is checked in the validator
		duration, _ := time.ParseDuration(*input.Duration)
		movie.Duration = duration
	}
	if input.IMDBLink != nil {
		movie.IMDBLink = *input.IMDBLink
	}

	if err := h.models.Movies.Update(movie); err != nil {
		switch {
		case errors.Is(err, models.ErrEditConflict):
			httputil.NewError(c, http.StatusConflict, err)
		default:
			httputil.NewError(c, http.StatusInternalServerError, err)
		}
		return
	}

	c.JSON(http.StatusOK, UpdateMovieResponse{
		Message: "movie updated sucessfully",
		Movie:   *movie,
	})
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
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, errors.New("invalid parameter: (id must be integer)"))
		return
	}

	if err := h.models.Movies.Delete(int(id)); err != nil {
		httputil.NewError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, DeleteMovieResponse{Message: "Deleted Successfully"})
}

type SearchMoviesResponse struct {
	Movies []models.Movie `json:"movies"`
}

type CreateMovieInput struct {
	Title       string `json:"title"`
	Director    string `json:"director"`
	ReleaseYear *int   `json:"release_year"`
	Duration    string `json:"duration"`
	IMDBLink    string `json:"imdb_link"`
}

func (i *CreateMovieInput) Validate(v *validator.Validator) {
	v.Check(len(strings.TrimSpace(i.Title)) > 0, "title", "required")
	v.Check(len(i.Title) <= 100, "title", "must be at most 50 characters")
	v.Check(len(i.Title) > 5, "title", "must be at least 5 characters")

	v.Check(len(strings.TrimSpace(i.Director)) > 0, "director", "required")
	v.Check(len(i.Director) <= 50, "director", "must be at most 30 characters")
	v.Check(len(i.Director) > 5, "director", "must be at least 5 characters")

	v.Check(i.ReleaseYear != nil, "release_year", "required")
	v.Check(
		*i.ReleaseYear >= 1900 && *i.ReleaseYear <= time.Now().Year(),
		"release_year",
		"must be between 1900 and this year",
	)

	if _, err := time.ParseDuration(i.Duration); err != nil {
		v.AddError("duration", "invalid duration format. must be similar to (2h10m)")
	}

	u, err := url.Parse(i.IMDBLink)
	if err != nil {
		v.AddError("imdb_link", "invalid url")
		return
	}
	v.Check(u.Scheme == "https", "imdb_link", "invalid url: scheme must be (https)")

	allowed := "imdb.com"
	host := u.Hostname()

	if host != allowed && !strings.HasSuffix(host, "."+allowed) {
		v.AddError("imdb_link", "invalid url: host must be imdb.com")
	}
}

type CreateMovieResponse struct {
	Message string       `json:"message"`
	Movie   models.Movie `json:"movie"`
}

type UpdateMovieInput struct {
	Title       *string `json:"title"`
	Director    *string `json:"director"`
	ReleaseYear *int    `json:"release_year"`
	Duration    *string `json:"duration"`
	IMDBLink    *string `json:"imdb_link"`
}

func (i *UpdateMovieInput) Validate(v *validator.Validator) {

	if i.Title != nil {
		v.Check(len(strings.TrimSpace(*i.Title)) > 0, "title", "required")
		v.Check(len(*i.Title) <= 100, "title", "must be at most 50 characters")
		v.Check(len(*i.Title) > 5, "title", "must be at least 5 characters")
	}

	if i.Director != nil {
		v.Check(len(strings.TrimSpace(*i.Director)) > 0, "director", "required")
		v.Check(len(*i.Director) > 5, "director", "must be at least 5 characters")
		v.Check(len(*i.Director) <= 50, "director", "must be at most 30 characters")
	}

	if i.ReleaseYear != nil {
		v.Check(
			*i.ReleaseYear >= 1900 && *i.ReleaseYear <= time.Now().Year(),
			"release_year",
			"must be between 1900 and this year",
		)
	}

	if i.Duration != nil {
		if _, err := time.ParseDuration(*i.Duration); err != nil {
			v.AddError("duration", "invalid duration format. must be similar to (2h10m)")
		}
	}

	if i.IMDBLink != nil {
		u, err := url.Parse(*i.IMDBLink)
		if err != nil {
			v.AddError("imdb_link", "invalid url")
			return
		}
		v.Check(u.Scheme == "https", "imdb_link", "invalid url: scheme must be (https)")

		allowed := "imdb.com"
		host := u.Hostname()

		if host != allowed && !strings.HasSuffix(host, "."+allowed) {
			v.AddError("imdb_link", "invalid url: host must be imdb.com")
		}
	}
}

type UpdateMovieResponse struct {
	Message string       `json:"message"`
	Movie   models.Movie `json:"movie"`
}

type DeleteMovieResponse struct {
	Message string `json:"message"`
}
