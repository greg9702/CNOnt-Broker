package main

import (
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"CNOnt-Broker/core/kubernetes/client"
	"CNOnt-Broker/core/ontology"
)

func main() {

	for {
		_, err := http.Get("http://127.0.0.1:8001")

		if err == nil {
			fmt.Println("Cluster ready!")
			break
		}

		fmt.Println("Waiting for cluster...")
		time.Sleep(5 * time.Second)
	}

	var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	kubernetesClient := client.NewKubernetesClient(kubeconfig)
	kubernetesClient.Init()

	var ontologyWrapper = ontology.NewOntologyWrapper(filepath.Join("ontology", "assets", "CNOnt.owl"))
	var ontologyBuilder = ontology.NewOntologyBuilder(kubernetesClient, ontologyWrapper)
	// deploymentController := controllers.NewDeploymentController(kubernetesClient, ontologyWrapper, ontologyBuilder)

	ontologyBuilder.GenerateCollection()

	// router := gin.Default()
	// router.Use(cors.Default())

	// v1Router := router.Group("/api/v1")

	// v1Router.GET("/create-deployment", deploymentController.CreateDeployment)
	// v1Router.GET("/delete-deployment", deploymentController.DeleteDeployment)
	// v1Router.GET("/preview-deployment", deploymentController.PreviewDeployment)
	// v1Router.GET("/serialize-cluster-conf", deploymentController.SerializeClusterConf)

	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080"
	// 	log.Printf("Defaulting to port %s", port)
	// }

	// router.Run(":" + port)
}
