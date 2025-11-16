package controllers

import "github.com/gin-gonic/gin"

func (a *Application) Routes(r *gin.Engine) {
	api := r.Group("/api")

	api.GET("/logout", a.userLogoutHandler)
	api.POST("/login", a.userLoginHandler)
	api.POST("/signup", a.userSignupHandler)

	auth := api.Group("/")
	auth.Use(a.AuthMiddleware())

	auth.GET("/user-info", a.UserDetailsHandler)
}
