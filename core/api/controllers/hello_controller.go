package controllers

import (
	"CNOnt-Broker/core/kubernetes/client"

	"github.com/gin-gonic/gin"
)

// HelloController is example controller
type HelloController struct {
	kubernetesClient *client.KubernetesClient
}

// NewHelloController creates new HelloController instance
func NewHelloController(client *client.KubernetesClient) *HelloController {
	h := HelloController{client}
	return &h
}

// GetHello returns hello message and prints all pods to the console
func (h *HelloController) GetHello(ctx *gin.Context) {
	h.kubernetesClient.GetAllPods()
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
