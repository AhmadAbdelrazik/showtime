package controllers

import (
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/go-playground/validator/v10"
)

type Controller struct {
	models *models.Model
	v      *validator.Validate
}

func New(dsn string) (*Controller, error) {
	model, err := models.New(dsn)
	if err != nil {
		return nil, err
	}

	return &Controller{
		models: model,
		v:      validator.New(validator.WithRequiredStructEnabled()),
	}, nil
}
