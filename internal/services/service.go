package services

import "github.com/AhmadAbdelrazik/showtime/internal/models"

type Service struct {
	Theaters *TheaterService
}

func New(model *models.Model) *Service {

	return &Service{
		Theaters: &TheaterService{model},
	}
}
