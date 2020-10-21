package ontology

import (
	"fmt"
	apiv1 "k8s.io/api/core/v1"
)

// TODO implementation
// OntologyBuilder offers serialization utility (cluster configuration -> ontology)
type OntologyBuilder struct {
	ontology string
}

// NewOntologyBuilder creates new NewOntologyBuilder instance
func NewOntologyBuilder() *OntologyBuilder {
	ob := OntologyBuilder{}
	return &ob
}

// WithPods appends info about pods to ontology
func (ow *OntologyBuilder) WithPods(pods *apiv1.PodList) {
	fmt.Print(pods)
	for _, pod := range pods.Items {
		fmt.Println("Pod name: ", pod.ObjectMeta.Name)
		fmt.Println("Containers: ")
		for _, container := range pod.Spec.Containers {
			fmt.Println("\t", container.Name)
		}
	}
}