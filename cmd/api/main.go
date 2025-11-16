package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/AhmadAbdelrazik/showtime/internal/controllers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env file")
		os.Exit(1)
	}

	// setup structured logging (slog)
	loggerOpts := &slog.HandlerOptions{}
	if os.Getenv("ENVIRONMENT") == "DEVELOPMENT" || os.Getenv("ENVIRONMENT") == "TESTING" {
		loggerOpts.Level = slog.LevelDebug
	} else {
		loggerOpts.Level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, loggerOpts))
	slog.SetDefault(logger)

	// Initialize controllers with dsn for Models
	app, err := controllers.New(os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal("Error Loading Controller" + err.Error())
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	app.Routes(r)

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}

}
