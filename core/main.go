package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"CNOnt-Broker/core/api/controllers"
	client "CNOnt-Broker/core/kubernetes/client"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/util/homedir"
)

func main() {

	var kubeconfig *string

	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	kuberenetesClient := client.NewKubernetesClient(kubeconfig)
	kuberenetesClient.Init()

	helloController := controllers.NewHelloController(kuberenetesClient)

	router := gin.Default()
	router.Use(cors.Default())

	v1 := router.Group("/api/v1")

	v1.GET("/hello", helloController.GetHello)
	v1.GET("/hello/:id", helloController.EchoNumber)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	router.Run(":" + port)
}
