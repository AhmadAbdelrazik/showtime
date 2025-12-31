package services

import "github.com/AhmadAbdelrazik/showtime/internal/models"

func isManagerOrAdmin(user *models.User) bool {
	return user.Role == "manager" || user.Role == "admin"
}

func isTheaterManagerOrAdmin(user *models.User, theater *models.Theater) bool {
	return theater.ManagerID == user.ID || user.Role == "admin"
}

func isHallManagerOrAdmin(user *models.User, hall *models.Hall) bool {
	return hall.ManagerID == user.ID || user.Role == "admin"
}
