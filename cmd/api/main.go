package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/AhmadAbdelrazik/showtime/internal/config"
	"github.com/AhmadAbdelrazik/showtime/internal/controllers"
	"github.com/AhmadAbdelrazik/showtime/internal/infrastructure/omdb"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/internal/services"
	"github.com/AhmadAbdelrazik/showtime/pkg/cache"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/AhmadAbdelrazik/showtime/cmd/api/docs"
)

//	@title			Showtime API
//	@version		1.0
//	@description	API for managing theaters and shows.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Ahmad Abdelrazik
//	@contact.url	https://www.github.com/AhmadAbdelrazik
//	@contact.email	ahmadabdelrazik159@gmail.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @host		localhost:8080
// @BasePath	/api/v1
func main() {
	// 1. Load configurations
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// 2. Initialize Loggers.
	setupLogger(cfg.Environment)

	// 3. Initialize Services Dependencies
	models, err := models.New(cfg.DSN)
	if err != nil {
		slog.Error("failed to create model", slog.String("error", err.Error()))
		os.Exit(1)
	}

	omdbClient := omdb.NewClient(cfg.OmdbApiKey)

	cache := cache.New()
	service := services.New(models, omdbClient)

	// 4. Initialize HTTP Server Dependencies
	app := controllers.New(service, cache, cfg)

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	app.Routes(r)

	if err := r.Run(); err != nil {
		log.Fatal(err)
	}

}

func setupLogger(env string) {
	// setup structured logging (slog)
	loggerOpts := &slog.HandlerOptions{}
	if env == "DEVELOPMENT" || env == "TESTING" {
		loggerOpts.Level = slog.LevelDebug
	} else {
		loggerOpts.Level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, loggerOpts))
	slog.SetDefault(logger)
}
