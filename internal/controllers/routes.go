package controllers

import "github.com/gin-gonic/gin"

func (a *Application) Routes(r *gin.Engine) {
	api := r.Group("/api")

	auth := api.Group("/")
	auth.Use(a.AuthMiddleware())

	// auth
	api.GET("/logout", a.userLogoutHandler)
	api.POST("/login", a.userLoginHandler)
	api.POST("/signup", a.userSignupHandler)

	auth.GET("/user-info", a.UserDetailsHandler)

	// theaters
	api.GET("/theaters", a.theatersSearchHandler)
	api.GET("/theaters/:id", a.getTheaterHandler)

	auth.POST("/theaters", a.createTheaterHandler)
	auth.PATCH("/theaters/:id", a.updateTheaterHandler)
	auth.DELETE("/theaters/:id", a.deleteTheaterHandler)

	// halls
	api.GET("/theaters/:id/halls/:code", a.getHallHandler)

	auth.POST("/theaters/:id/halls", a.createHallHandler)
	auth.PATCH("/theaters/:id/halls/:code", a.updateHallHandler)
	auth.DELETE("/theaters/:id/halls/:code", a.deleteHallResponse)
}
