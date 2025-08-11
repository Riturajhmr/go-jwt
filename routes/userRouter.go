package routes

import(
	controller "jwtauth/controllers"
	"jwtauth/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine){

	// we are using middleware, because after login the token is generated, and the token determines who have 
	//how much authority in the database to access, which is held on middleware folder.
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}
