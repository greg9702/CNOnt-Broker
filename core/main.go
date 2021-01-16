package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"CNOnt-Broker/core/api/controllers"
	logger "CNOnt-Broker/core/common"
	"CNOnt-Broker/core/kubernetes/client"
	"CNOnt-Broker/core/ontology"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func usage() {
	fmt.Println("usage: go run main.go --kubeconfig <PATH_TO_KUBE_CONFIG> --logLevel <LOGLEVEL>")
}

func main() {

	// when using docker-compose with local cluster, uncomment this line
	// for {
	// 	_, err := http.Get("http://127.0.0.1:8001")

	// 	if err == nil {
	// 		fmt.Println("Cluster ready!")
	// 		break
	// 	}

	// 	fmt.Println("Waiting for cluster...")
	// 	time.Sleep(5 * time.Second)
	// }

	var kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	logLevel := flag.Int("logLevel", -1, "specify log level")
	flag.Parse()

	if *logLevel == -1 {
		usage()
		return
	}

	logger.BaseLog().InitLogger(*logLevel)

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
		logger.BaseLog().Info(fmt.Sprintf("Defaulting to port %s", port))
	}

	router.Run(":" + port)
}
