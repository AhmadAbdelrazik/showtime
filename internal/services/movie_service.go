package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

type MovieService struct {
	models   *models.Model
	provider MovieProvider
}

}
