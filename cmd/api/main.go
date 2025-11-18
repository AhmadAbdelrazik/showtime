package main

import (
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"log"
	"log/slog"
	"os"

	"github.com/AhmadAbdelrazik/showtime/internal/controllers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/AhmadAbdelrazik/showtime/cmd/api/docs"
)

// @title           Showtime API
// @version         1.0
// @description     API for managing theaters and shows.
// @termsOfService  http://swagger.io/terms/

// @contact.name   Ahmad Abdelrazik
// @contact.url    https://www.github.com/AhmadAbdelrazik
// @contact.email  ahmadabdelrazik159@gmail.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
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

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	app.Routes(r)

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}

}
