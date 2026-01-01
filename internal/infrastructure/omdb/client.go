package omdb

import (
	"context"
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

func (c *Client) Find(ctx context.Context, movieId string) (*models.Movie, error) {
	resp, err := http.Get(fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&i=%v", c.apiKey, movieId))
	if err != nil {
		return nil, err
	}

}

func (c *Client) Search(ctx context.Context, movieName string) ([]models.Movie, error) {
	panic("not implemented yet")
}
