package controllers

import (
	"github.com/gin-gonic/gin"
)

// HelloController is example controller
type HelloController struct {
}

// GetHello returns hello message
func (h *HelloController) GetHello(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Hello",
	})
}

// EchoNumber responde with a number it received
func (h *HelloController) EchoNumber(ctx *gin.Context) {
	number := ctx.Param("id")
	ctx.JSON(200, gin.H{
		"number": number,
	})
}
