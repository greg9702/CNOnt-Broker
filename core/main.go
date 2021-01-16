package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"CNOnt-Broker/core/api/controllers"
	"CNOnt-Broker/core/kubernetes/client"
	"CNOnt-Broker/core/ontology"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	kubernetesClient := client.NewKubernetesClient(kubeconfig)
	kubernetesClient.Init()

	ontologyTemplateFile := filepath.Join("ontology", "assets", "CNOnt_template.owl")

	var ontologyWrapper = ontology.NewOntologyWrapper(ontologyTemplateFile)
	var ontologyBuilder = ontology.NewOntologyBuilder(kubernetesClient, ontologyWrapper, ontologyTemplateFile)
	deploymentController := controllers.NewDeploymentController(kubernetesClient, ontologyWrapper, ontologyBuilder)

	router := gin.Default()
	router.Use(cors.Default())

	v1Router := router.Group("/api/v1")

	v1Router.GET("/create-deployment", deploymentController.CreateDeployments)
	v1Router.GET("/delete-deployment", deploymentController.DeleteDeployments)
	v1Router.GET("/preview-deployment", deploymentController.PreviewDeployments)
	v1Router.GET("/serialize-cluster-conf", deploymentController.SerializeClusterConf)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router.Run(":" + port)
}
