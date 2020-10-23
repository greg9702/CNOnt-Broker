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

func (o *KubernetesClient) Mock() error {

	coreClient := o.clientset.CoreV1()

	nodes, err := coreClient.Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	fmt.Println("----------NODES----------")

	for ix := range nodes.Items {
		node := &nodes.Items[ix]
		fmt.Println(node)
		fmt.Println("-----------------------")
	}

	fmt.Println(nodes)

	var namespaces []string
	namespaceList, err := coreClient.Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for ix := range namespaceList.Items {
		namespaces = append(namespaces, namespaceList.Items[ix].Name)
	}

	fmt.Println("----------NAMESPACES----------")
	for ix := range namespaceList.Items {
		namespace := &namespaceList.Items[ix]
		fmt.Println(namespace)
		fmt.Println("-----------------------")
	}

	for _, namespace := range namespaces {

		pods, err := coreClient.Pods(namespace).List(context.TODO(), metav1.ListOptions{})

		if len(pods.Items) != 0 {
			if err != nil {
				return err
			}

			fmt.Println("----------PODS----------")

			for ix := range pods.Items {
				pod := &pods.Items[ix]
				fmt.Println(pod)
				fmt.Println("-----------------------")
			}

			fmt.Println("----------CONTAINERS----------")

			for ix := range pods.Items {
				pod := &pods.Items[ix]
				containers := pod.Spec.Containers

				for i := range containers {
					fmt.Println(containers[i])
					fmt.Println("-----------------------")
				}
			}
		}
	}
	return nil
}

// GetAllPods returns pods from default namespace
func (k *KubernetesClient) GetAllPods() (*apiv1.PodList, error) {
	return k.clientset.CoreV1().Pods("default").List(context.TODO(), v1.ListOptions{})
}

// TODO generalize
func (k *KubernetesClient) ExecuteCommand(command string) (*apiv1.PodList, error) {
	return k.clientset.CoreV1().Pods("default").List(context.TODO(), v1.ListOptions{})
}

// CreateDeployment creates deployment deployment passed in
func (k *KubernetesClient) CreateDeployment(deployment *unstructured.Unstructured) error {
	fmt.Println("Create deployment...")

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

	fmt.Println("Creating deployment...")
	result, err := client.Resource(deploymentRes).Namespace(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})

	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Printf("Created deployment %q.\n", result.GetName())
	if err != nil {
		fmt.Printf("Creating deployment error, %s", err.Error())
		return err
	}

	return nil
}

// DeleteDeployment deletes deployment passed in
func (k *KubernetesClient) DeleteDeployment(deploymentName string) error {
	fmt.Println("Deleting deployment...")

	deletePolicy := metav1.DeletePropagationForeground
	if err := k.clientset.AppsV1().Deployments(apiv1.NamespaceDefault).Delete(context.TODO(), deploymentName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		fmt.Printf("Deleting deployment error, %s", err.Error())
		return err
	}
	fmt.Printf("Deleted deployment")
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
