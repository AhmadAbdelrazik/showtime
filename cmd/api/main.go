package main

import (
	"log"
	"os"

	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	_, err := models.New(os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := gin.Default()
	if err := r.Run(); err != nil {
		log.Fatal(err)
	}
}
