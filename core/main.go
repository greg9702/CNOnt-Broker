package main

import (
	"CNOnt-Broker/core/api/controllers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	helloController := controllers.HelloController{}

	v1 := router.Group("/api/v1")

	v1.GET("/hello", helloController.GetHello)
	v1.PUT("/hello/:id", helloController.EchoNumber)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router.Run(":" + port)
}
