package omdb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/internal/services"
)

var ErrInvalidApiKey = errors.New("invalid api key")

type Client struct {
	apiKey string
}

func NewClient(apiKey string) *Client {
	return &Client{apiKey}
}

func (c *Client) GetMovie(_ context.Context, movieId string) (*models.Movie, error) {
	resp, err := http.Get(fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&i=%v", c.apiKey, movieId))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var success findSuccessResponse
	if err := json.Unmarshal(raw, &success); err == nil && success.Response == "True" {
		return &models.Movie{
			ImdbID:     success.ImdbID,
			Title:      success.Title,
			Year:       success.Year,
			Rated:      success.Rated,
			Runtime:    success.Runtime,
			Genre:      success.Genre,
			Director:   success.Director,
			Poster:     success.Poster,
			ImdbRating: success.ImdbRating,
		}, nil
	}

	var error errorResponse
	if err := json.Unmarshal(raw, &error); err == nil && success.Response == "False" {
		switch error.Error {
		case "Incorrect IMDb ID.":
			return nil, services.ErrInvalidMovieId
		case "Invalid API key!":
			return nil, ErrInvalidApiKey
		default:
			return nil, errors.New(error.Error)
		}
	}

	return nil, errors.New("unknown response shape")
}

func (c *Client) Search(_ context.Context, title, year string) ([]models.Movie, error) {
	url := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&s=%v", c.apiKey, title)

	if year != "" {
		url += fmt.Sprintf("&y=%v", year)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var success searchSuccessResponse
	if err := json.Unmarshal(raw, &success); err == nil && success.Response == "True" {
		movies := make([]models.Movie, len(success.Search))

		for i, m := range success.Search {
			movies[i] = models.Movie{
				ImdbID: m.ImdbID,
				Title:  m.Title,
				Year:   m.Year,
				Poster: m.Poster,
			}
		}

		return movies, nil
	}

	var error errorResponse
	if err := json.Unmarshal(raw, &error); err == nil && success.Response == "False" {
		switch error.Error {
		case "Incorrect IMDb ID.":
			return nil, services.ErrInvalidMovieId
		case "Movie not found":
			return nil, services.ErrMovieNotFound
		default:
			return nil, errors.New(error.Error)
		}
	}

	return nil, errors.New("unknown response shape")
}
