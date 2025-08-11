package main

import (
	routes "jwtauth/routes"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main(){
	// basically get env file is to load the certain parameters, for the project, its behaviour.
	err := godotenv.Load(".env")

	if err != nil{
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")

	if port == ""{
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	// this is basically the routes that we are using, to find the information that we need.
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func (c *gin.Context)  {
		c.JSON(200, gin.H{
			"success":"Access granted for api-1",
		})
	})

	router.GET("/api-2", func (c *gin.Context)  {
		c.JSON(200, gin.H{
			"success":"Access granted for api-2",
		})
	})

	router.Run(":" + port)
}