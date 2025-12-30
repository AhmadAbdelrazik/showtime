package controllers

import (
	"crypto/rand"
	"encoding/base32"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
)

func (Application) generateRandomString() string {
	b := make([]byte, 16)
	rand.Read(b)

	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)

	return s
}

func isManagerOrAdmin(user *models.User) bool {
	return user.Role == "manager" || user.Role == "admin"
}

func isTheaterManagerOrAdmin(user *models.User, theater *models.Theater) bool {
	return theater.ManagerID == user.ID || user.Role == "admin"
}

func isHallManagerOrAdmin(user *models.User, hall *models.Hall) bool {
	return hall.ManagerID == user.ID || user.Role == "admin"
}
