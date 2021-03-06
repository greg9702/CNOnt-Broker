package client

import (
	"context"
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	logger "CNOnt-Broker/core/common"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesClient main k8s client module
type KubernetesClient struct {
	kubeconfig *string
	clientset  *kubernetes.Clientset
}

// NewKubernetesClient creates new KubernetesClient instance
func NewKubernetesClient(path *string) *KubernetesClient {
	k := KubernetesClient{path, &kubernetes.Clientset{}}
	return &k
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

// AllPods returns pods from default namespace
// If empty string provided as namespace argument, "default" namespace is used
func (k *KubernetesClient) AllPods(namespace string) (*apiv1.PodList, error) {

	if namespace == "" {
		namespace = "default"
	}

	return k.clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
}

func (k *KubernetesClient) AllReplicaSets(namespace string) (*appsv1.ReplicaSetList, error) {

	if namespace == "" {
		namespace = "default"
	}

	return k.clientset.AppsV1().ReplicaSets(namespace).List(context.TODO(), v1.ListOptions{})
}

// AllNodes returns list of all nodes in cluster
func (k *KubernetesClient) AllNodes() (*apiv1.NodeList, error) {
	return k.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
}

// ContainersFromPod returns list of all nodes in cluster
func (k *KubernetesClient) ContainersFromPod(pod apiv1.Pod) ([]*apiv1.Container, error) {

	// TODO errors checks?

	containers := pod.Spec.Containers

	var containersList []*apiv1.Container

	for i := range containers {
		containersList = append(containersList, &containers[i])
	}

	return containersList, nil
}

// AllNamespaces returns list of all namespaces
func (k *KubernetesClient) AllNamespaces() (*apiv1.NamespaceList, error) {
	return k.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
}

// CreateDeployment creates deployment deployment passed in
func (k *KubernetesClient) CreateDeployment(deployment *unstructured.Unstructured) error {
	logger.BaseLog().Debug("Create deployment...")

	namespace := "default"

	config, err := clientcmd.BuildConfigFromFlags("", *k.kubeconfig)
	if err != nil {
		panic(err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	logger.BaseLog().Debug("Creating deployment...")
	result, err := client.Resource(deploymentRes).Namespace(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		logger.BaseLog().Error(err.Error())
		return err
	}

	logger.BaseLog().Info(fmt.Sprintf("Created deployment %s", result.GetName()))

	if err != nil {
		logger.BaseLog().Error(fmt.Sprintf("Creating deployment error, %s", err.Error()))
		return err
	}

	return nil
}

// DeleteDeployment deletes deployment passed in
func (k *KubernetesClient) DeleteDeployment(deploymentName string) error {
	logger.BaseLog().Debug("Deleting deployment...")

	deletePolicy := metav1.DeletePropagationForeground
	if err := k.clientset.AppsV1().Deployments(apiv1.NamespaceDefault).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		logger.BaseLog().Error(fmt.Sprintf("Deleting deployment error, %s", err.Error()))
		return err
	}
	logger.BaseLog().Info(fmt.Sprintf("Deleted deployment %s", deploymentName))
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
