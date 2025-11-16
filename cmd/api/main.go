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
	loggerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, loggerOpts))
	slog.SetDefault(logger)

	slog.Debug("loading .env file")
	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load .env file")
		os.Exit(1)
	}
	slog.Debug("loaded successfully")

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
