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

// GetAllPods prints pods from default namespace
func (k *KubernetesClient) GetAllPods() {
	fmt.Println(k.clientset.CoreV1().Pods("default").List(context.TODO(), v1.ListOptions{}))
}

func (k *KubernetesClient) CreateDeployment() error {
	fmt.Println("Create deployment...")

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := k.clientset.AppsV1().Deployments(apiv1.NamespaceDefault).Create(context.TODO(), deployment, metav1.CreateOptions{})
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
