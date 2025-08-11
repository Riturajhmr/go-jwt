package routes

import (
	controller "jwtauth/controllers"

	"github.com/gin-gonic/gin"
)

// this function is basically made for the authentication purpose, to signin or register the users.
//and the signup and login functionality, will control by the controller package.

func AuthRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())
	incomingRoutes.GET("users/verify-email", controller.VerifyEmail())
}