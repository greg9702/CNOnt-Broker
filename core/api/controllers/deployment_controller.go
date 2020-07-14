package controllers

import (
	"CNOnt-Broker/core/kubernetes/client"
	"CNOnt-Broker/core/ontology"

	"github.com/gin-gonic/gin"
)

// DeploymentController is controller used for deployment requests
type DeploymentController struct {
	kubernetesClient *client.KubernetesClient
	ontologyWrapper  *ontology.OntologyWrapper
}

// NewDeploymentController creates new DeploymentController instance
func NewDeploymentController(client *client.KubernetesClient, ontology *ontology.OntologyWrapper) *DeploymentController {
	d := DeploymentController{client, ontology}
	return &d
}

// PreviewDeployment used for preview current ontology
func (d *DeploymentController) PreviewDeployment(ctx *gin.Context) {
	deployment, err := d.ontologyWrapper.BuildDeploymentConfiguration()

	if err != nil {
		ctx.JSON(422, gin.H{
			"error":   "Creating deployment configuration error",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"deployment": deployment,
	})
}

// CreateDeployment used for create deployment
func (d *DeploymentController) CreateDeployment(ctx *gin.Context) {

	deployment, err := d.ontologyWrapper.BuildDeploymentConfiguration()

	if err != nil {
		ctx.JSON(422, gin.H{
			"error":   "Creating deployment configuration error",
			"details": err.Error(),
		})
		return
	}

	var deploymentName string
	getName := deployment.Object["metadata"].(map[string]interface{})["name"]

	if getName == nil {
		ctx.JSON(422, gin.H{
			"error": "Could not get deployment name",
		})
		return
	}

	deploymentName = getName.(string)

	res := d.kubernetesClient.CreateDeployment(deployment)

	if res != nil {
		ctx.JSON(422, gin.H{
			"error":   "Creating deployment error",
			"name":    deploymentName,
			"details": res.Error(),
		})
		return
	}
	ctx.JSON(201, gin.H{
		"name": deploymentName,
	})
}

// DeleteDeployment used for deleting current deployment
func (d *DeploymentController) DeleteDeployment(ctx *gin.Context) {

	deployment, err := d.ontologyWrapper.BuildDeploymentConfiguration()

	if err != nil {
		ctx.JSON(422, gin.H{
			"error":   "Creating deployment configuration error",
			"details": err.Error(),
		})
		return
	}

	var deploymentName string
	getName := deployment.Object["metadata"].(map[string]interface{})["name"]

	if getName == nil {
		ctx.JSON(422, gin.H{
			"error": "Could not get deployment name",
		})
		return
	}

	deploymentName = getName.(string)

	res := d.kubernetesClient.DeleteDeployment(deploymentName)

	if res != nil {
		ctx.JSON(422, gin.H{
			"error":   "Deployment deleting error",
			"name":    deploymentName,
			"details": res.Error(),
		})
		return
	}
	ctx.JSON(204, gin.H{
		"name": deploymentName,
	})
	return
}
