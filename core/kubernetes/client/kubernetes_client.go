package client

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesClient main k8s client module
type KubernetesClient struct {
	kubeconfig *string
	clientset  *kubernetes.Clientset
	deployment *appsv1.Deployment
}

// NewKubernetesClient creates new KubernetesClient instance
func NewKubernetesClient(path *string) *KubernetesClient {
	k := KubernetesClient{path, &kubernetes.Clientset{}, &appsv1.Deployment{}}
	return &k
}

func (k *KubernetesClient) SetDeployment(deployment *appsv1.Deployment) {
	k.deployment = deployment
}

// Init initialize KubernetesClient kubeconfig
func (k *KubernetesClient) Init() {

	config, err := clientcmd.BuildConfigFromFlags("", *k.kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	k.clientset = clientset
}

// GetAllPods prints pods from default namespace
func (k *KubernetesClient) GetAllPods() {
	fmt.Println(k.clientset.CoreV1().Pods("default").List(context.TODO(), v1.ListOptions{}))
}

func (k *KubernetesClient) CreateDeployment() error {
	fmt.Println("Create deployment...")

	result, err := k.clientset.AppsV1().Deployments(apiv1.NamespaceDefault).Create(context.TODO(), k.deployment, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("Creating deployment error, %s", err.Error)
		return err
	}

	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
	return nil
}

func (k *KubernetesClient) DeleteDeployment() error {
	fmt.Println("Deleting deployment...")

	deletePolicy := metav1.DeletePropagationForeground
	if err := k.clientset.AppsV1().Deployments(apiv1.NamespaceDefault).Delete(context.TODO(), "demo-deployment", metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Printf("Deleting deployment error, %s", err.Error)
		return err
	}
	fmt.Printf("Deleted deployment")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
