package controllers

import (
	"CNOnt-Broker/core/kubernetes/client"
	"CNOnt-Broker/core/ontology"
	"fmt"

	"github.com/gin-gonic/gin"
)

// DeploymentController is controller used for deployment requests
type DeploymentController struct {
	kubernetesClient *client.KubernetesClient
	ontologyWrapper  *ontology.OntologyWrapper
	ontologyBuilder  *ontology.OntologyBuilder
}

// NewDeploymentController creates new DeploymentController instance
func NewDeploymentController(client *client.KubernetesClient, ontology *ontology.OntologyWrapper, ontologyBuilder *ontology.OntologyBuilder) *DeploymentController {
	d := DeploymentController{client, ontology, ontologyBuilder}
	return &d
}

// PreviewDeployments used for preview current ontology
func (d *DeploymentController) PreviewDeployments(ctx *gin.Context) {
	deployments, err := d.ontologyWrapper.BuildDeploymentConfiguration()

	if err != nil {
		ctx.JSON(422, gin.H{
			"error":   "Creating deployment configuration error",
			"details": err.Error(),
		})
		return
	}
	ctx.JSON(200, gin.H{
		"deployment": deployments,
	})
}

// CreateDeployments used for create deployment
func (d *DeploymentController) CreateDeployments(ctx *gin.Context) {

	deployments, err := d.ontologyWrapper.BuildDeploymentConfiguration()

	if err != nil {
		ctx.JSON(422, gin.H{
			"error":   "Creating deployment configuration error",
			"details": err.Error(),
		})
		return
	}
	for _, deployment := range deployments {
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
}

// DeleteDeployments used for deleting current deployment
func (d *DeploymentController) DeleteDeployments(ctx *gin.Context) {

	deployments, err := d.ontologyWrapper.BuildDeploymentConfiguration()

	if err != nil {
		ctx.JSON(422, gin.H{
			"error":   "Creating deployment configuration error",
			"details": err.Error(),
		})
		return
	}

	for _, deployment := range deployments {
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
	}
	return
}

// SerializeClusterConf create a cluster mapping and return it as response
// 304 status code is returned when attempt was sucesfully
func (d *DeploymentController) SerializeClusterConf(ctx *gin.Context) {

	pathFile, err := d.ontologyBuilder.GenerateCollection()

	if err != nil {
		ctx.JSON(200, gin.H{
			"error":   "Could not create a mapping",
			"details": err,
		})
		return
	}

	ctx.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", "cluster_mapping.owl"))
	ctx.Writer.Header().Add("Content-Type", "application/octet-stream")
	ctx.File(pathFile) // automatically respond with 404 if wrong file passed
}
