package controllers

import (
	"CNOnt-Broker/core/kubernetes/client"
	"CNOnt-Broker/core/ontology"

	"github.com/gin-gonic/gin"
)

// HelloController is example controller
type HelloController struct {
	kubernetesClient *client.KubernetesClient
	ontology         *ontology.OntologyWrapper
}

// NewHelloController creates new HelloController instance
func NewHelloController(client *client.KubernetesClient, ontology *ontology.OntologyWrapper) *HelloController {
	h := HelloController{client, ontology}
	return &h
}

// GetHello returns hello message and prints all pods to the console
func (h *HelloController) GetHello(ctx *gin.Context) {
	h.kubernetesClient.GetAllPods()
	h.ontology.PrintClasses()
	ctx.JSON(200, gin.H{
		"message": "Hello",
	})
}

func (h *HelloController) PreviewDeployment(ctx *gin.Context) {
	res, err := h.kubernetesClient.PreviewDeployment()

	if err != nil {
		ctx.JSON(500, gin.H{})
		return
	}
	ctx.JSON(200, gin.H{
		"deployment": res,
	})
}

func (h *HelloController) CreateDeployment(ctx *gin.Context) {
	res := h.kubernetesClient.CreateDeployment()

	if res != nil {
		ctx.JSON(500, gin.H{})
		return
	}
	ctx.JSON(201, gin.H{})
}

func (h *HelloController) DeleteDeployment(ctx *gin.Context) {
	res := h.kubernetesClient.DeleteDeployment()

	if res != nil {
		ctx.JSON(404, gin.H{})
		return
	}
	ctx.JSON(204, gin.H{})
}

// EchoNumber responde with a number it received
func (h *HelloController) EchoNumber(ctx *gin.Context) {
	number := ctx.Param("id")
	ctx.JSON(200, gin.H{
		"number": number,
	})
}
