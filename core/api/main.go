package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	router.Use(func(ctx *gin.Context) {
		if ctx.Request.Header["Content-Length"][0] == "0" {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Payload should not be empty"})
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
	})

	router.Group("/api/v1")
}
