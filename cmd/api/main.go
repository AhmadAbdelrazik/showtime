package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/AhmadAbdelrazik/showtime/internal/controllers"
	"github.com/AhmadAbdelrazik/showtime/internal/infrastructure/omdb"
	"github.com/AhmadAbdelrazik/showtime/internal/models"
	"github.com/AhmadAbdelrazik/showtime/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

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
	_, err := os.Stat(".env")
	if err == nil {
		// load .env
		if err := godotenv.Load(); err != nil {
			log.Fatal("failed to load .env file")
			os.Exit(1)
		}
	}

	setupLogger()
	dsn := getDSN()
	app, err := setupApplication(dsn, os.Getenv("OMDB_APIKEY"))
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

func getDSN() string {
	var dsn string

	if os.Getenv("DB_DSN") != "" {
		dsn = os.Getenv("DB_DSN")
	} else {
		dsn = fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_DATABASE"),
		)
	}
	return dsn
}

func setupLogger() {
	// setup structured logging (slog)
	loggerOpts := &slog.HandlerOptions{}
	if os.Getenv("ENVIRONMENT") == "DEVELOPMENT" || os.Getenv("ENVIRONMENT") == "TESTING" {
		loggerOpts.Level = slog.LevelDebug
	} else {
		loggerOpts.Level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, loggerOpts))
	slog.SetDefault(logger)
}

func setupApplication(dsn, omdbApiKey string) (*controllers.Application, error) {
	models, err := models.New(dsn)
	if err != nil {
		slog.Error("failed to create model", slog.String("error", err.Error()))
		return nil, err
	}

	omdbClient := omdb.NewClient(omdbApiKey)
	service := services.New(models, omdbClient)

	return controllers.New(service)
}
