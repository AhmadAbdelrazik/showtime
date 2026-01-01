package omdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

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

	var input movieResponse

	decoder := json.NewDecoder(resp.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&input); err != nil {
		return nil, err
	}

	return &models.Movie{
		ImdbID:     input.ImdbID,
		Title:      input.Title,
		Year:       input.Year,
		Rated:      input.Rated,
		Runtime:    input.Runtime,
		Genre:      input.Genre,
		Director:   input.Director,
		Poster:     input.Poster,
		ImdbRating: input.ImdbRating,
	}, nil
}
