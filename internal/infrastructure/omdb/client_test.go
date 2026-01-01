package omdb_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/AhmadAbdelrazik/showtime/internal/infrastructure/omdb"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/internal/services"
)

func TestClient_GetMovie(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		apiKey string
		// Named input parameters for target function.
		movieId   string
		want      *models.Movie
		wantErr   bool
		wantedErr error
	}{
		// TODO: Add test cases.
		{
			name:    "fetch batman movie",
			apiKey:  "20d1919",
			movieId: "tt0372784",
			want: &models.Movie{
				ImdbID:     "tt0372784",
				Title:      "Batman Begins",
				Year:       "2005",
				Rated:      "PG-13",
				Runtime:    "140 min",
				Genre:      "Action, Crime, Drama",
				Director:   "Cristopher Nolan",
				Poster:     "https://m.media-amazon.com/images/M/MV5BMzA2NDQzZDEtNDU5Ni00YTlkLTg2OWEtYmQwM2Y1YTBjMjFjXkEyXkFqcGc@._V1_SX300.jpg",
				ImdbRating: "8.2",
			},
			wantErr:   false,
			wantedErr: nil,
		},
		{
			name:      "invalid movie id",
			apiKey:    "20d1919",
			movieId:   "tt03724",
			want:      nil,
			wantErr:   true,
			wantedErr: services.ErrInvalidMovieId,
		},
		{
			name:      "invalid api key",
			apiKey:    "20919",
			movieId:   "tt0372784",
			want:      nil,
			wantErr:   true,
			wantedErr: omdb.ErrInvalidApiKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := omdb.NewClient(tt.apiKey)
			got, gotErr := c.GetMovie(context.Background(), tt.movieId)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("GetMovie() failed: %v", gotErr)
				} else if !errors.Is(gotErr, tt.wantedErr) {
					t.Errorf("GetMovie() error mismatch\n"+
						"got: %v\n"+"want: %v\n", gotErr, tt.wantedErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("GetMovie() succeeded unexpectedly")
			}

			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMovie() = %v, want %v", got, tt.want)
			}
		})
	}

}
