package ontology

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/shful/gofp"
	"github.com/shful/gofp/owlfunctional"
	"github.com/shful/gofp/owlfunctional/assertions"
	"github.com/shful/gofp/owlfunctional/axioms"
	"github.com/shful/gofp/owlfunctional/decl"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const podClass = ":Pod"

const nameAssertion = ":name"
const apiVersionAssertion = ":apiVersion"
const kindAssertion = ":kind"
const replicasAssertion = ":replicas"
const imageAssertion = ":image"
const portAssertion = ":port"

const belongsToNodeAssertion = ":belongs_to_node"
const containsContainerAssertion = ":contains_container"

// OntologyWrapper functional OWL ontology wraper
type OntologyWrapper struct {
	ontology *owlfunctional.Ontology
}

// NewOntologyWrapper creates new OntologyWrapper instance
func NewOntologyWrapper(path string) *OntologyWrapper {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	o, err := gofp.OntologyFromReader(f, "source was "+path)
	if err != nil {
		log.Fatal(gofp.ErrorMsgWithPosition(err))
	}
	ow := OntologyWrapper{o}
	return &ow
}

// func (ow *OntologyWrapper) PrintClasses() {
// 	log.Println("That's what we parsed: ", ow.ontology.About())

// 	for _, declaration := range ow.ontology.K.AllClassDecls() {
// 		fmt.Println(declaration.IRI)
// 	}
// }

// objectPropertyAssertionValue returns string value of particular ObjectPropertyAssertion about passed individual
func (ow *OntologyWrapper) objectPropertyAssertionValue(assertionName string, individual string) (string, error) {
	allObjectPropertyAssertions := ow.ontology.K.AllObjectPropertyAssertions()

	isAssertionAboutIndividual := func(s assertions.ObjectPropertyAssertion) bool {
		return s.A1.Name == individual && convertIRI2Name(s.PN) == assertionName
	}

	filteredAssertions := filterObjPropAssertions(allObjectPropertyAssertions, isAssertionAboutIndividual)
	if len(filteredAssertions) == 0 {
		return "", errors.New("No '" + assertionName + "' assertions found for " + individual)
	} else if len(filteredAssertions) > 1 {
		return "", errors.New("Multiple '" + assertionName + "' assertions found for " + individual)
	}

	return filteredAssertions[0].A2.Name, nil
}

// objectPropertyAssertionValues returns slice of strings with values of particular ObjectPropertyAssertion about passed individual
func (ow *OntologyWrapper) objectPropertyAssertionValues(assertionName string, individual string) ([]string, error) {
	allObjectPropertyAssertions := ow.ontology.K.AllObjectPropertyAssertions()

	isAssertionAboutIndividual := func(s assertions.ObjectPropertyAssertion) bool {
		return s.A1.Name == individual && convertIRI2Name(s.PN) == assertionName
	}

	filteredAssertions := filterObjPropAssertions(allObjectPropertyAssertions, isAssertionAboutIndividual)
	if len(filteredAssertions) == 0 {
		return nil, errors.New("No '" + assertionName + "' assertions found for " + individual)
	}

	return objectAssertions2String(filteredAssertions), nil
}

// dataPropertyAssertionValue returns string value of particular DataPropertyAssertion about passed individual
func (ow *OntologyWrapper) dataPropertyAssertionValue(assertionName string, individual string) (string, error) {
	allDataPropertyAssertions := ow.ontology.K.AllDataPropertyAssertions()

	isAssertionAboutIndividual := func(s axioms.DataPropertyAssertion) bool {
		return s.A.Name == individual && convertIRI2Name(s.R.(*decl.DataPropertyDecl).IRI) == assertionName
	}

	filteredAssertions := filterDataPropAssertions(allDataPropertyAssertions, isAssertionAboutIndividual)
	if len(filteredAssertions) == 0 {
		return "", errors.New("No '" + assertionName + "' assertions found for " + individual)
	} else if len(filteredAssertions) > 1 {
		return "", errors.New("Multiple '" + assertionName + "' assertions found for " + individual)
	}

	return filteredAssertions[0].V.Value, nil
}

// individualsByClass returns all individuals from a given class
func (ow *OntologyWrapper) individualsByClass(className string) (individuals []string) {
	allClassAssertions := ow.ontology.K.AllClassAssertions()

	isAssertionAboutClass := func(s axioms.ClassAssertion) bool {
		if !s.C.IsNamedClass() {
			return false
		}
		return convertIRI2Name((s.C).(*decl.ClassDecl).IRI) == className
	}
	classAssertions := filterClassAssertions(allClassAssertions, isAssertionAboutClass)

	for _, classAssertion := range classAssertions {
		individuals = append(individuals, classAssertion.A.Name)
	}

	return individuals
}

// pods returns all pod individuals
func (ow *OntologyWrapper) pods() []string {
	return ow.individualsByClass(podClass)
}

// name returns name of a given individual
func (ow *OntologyWrapper) name(individual string) (string, error) {
	return ow.dataPropertyAssertionValue(nameAssertion, individual)
}

// apiVersion returns apiVersion of a given pod
func (ow *OntologyWrapper) apiVersion(pod string) (string, error) {
	return ow.dataPropertyAssertionValue(apiVersionAssertion, pod)
}

// kind returns kind for a given pod (deployment, server etc.)
func (ow *OntologyWrapper) kind(pod string) (string, error) {
	return ow.dataPropertyAssertionValue(kindAssertion, pod)
}

// replicas returns replicas of a given pod
func (ow *OntologyWrapper) replicas(pod string) (int32, error) {
	replicas, err := ow.dataPropertyAssertionValue(replicasAssertion, pod)
	if err != nil {
		return -1, err
	}
	r, err := strconv.Atoi(replicas)
	if err != nil {
		return -1, err
	}
	return int32(r), nil
}

// image returns image for a given container (deployment, server etc.)
func (ow *OntologyWrapper) image(container string) (string, error) {
	return ow.dataPropertyAssertionValue(imageAssertion, container)
}

// port returns port for a given container (deployment, server etc.)
func (ow *OntologyWrapper) port(container string) (string, error) {
	return ow.dataPropertyAssertionValue(portAssertion, container)
}

// containers returns containers for a given pod
func (ow *OntologyWrapper) containers(pod string) ([]string, error) {
	return ow.objectPropertyAssertionValues(containsContainerAssertion, pod)
}

// BuildDeploymentConfiguration returns Kubernetes Deployment basing on parsed ontology
func (ow *OntologyWrapper) BuildDeploymentConfiguration() (*unstructured.Unstructured, error) {

	// we will call this base structure
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "",
			"kind":       "",
			"metadata": map[string]interface{}{
				"name": "",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "demo", // TODO take from ontology
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "demo", // TODO take from ontology
						},
					},
				},
			},
		},
	}

	for _, pod := range ow.pods() {

		// fmt.Println("Pod: " + pod)

		apiVersion, err := ow.apiVersion(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			deployment.Object["apiVersion"] = apiVersion
		}

		kind, err := ow.kind(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			deployment.Object["kind"] = kind
		}

		podName, err := ow.name(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			deployment.Object["metadata"].(map[string]interface{})["name"] = podName
		}

		replicas, err := ow.replicas(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			deployment.Object["spec"].(map[string]interface{})["replicas"] = replicas
		}

		containers, err := ow.containers(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"] = map[string]interface{}{}
			deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"] = []map[string]interface{}{}

			for _, container := range containers {
				// fmt.Println(container)

				containerSpec := map[string]interface{}{}
				containerName, err := ow.name(container)
				if err != nil {
					fmt.Println(err)
				} else {
					containerSpec["name"] = containerName
				}

				containerImage, err := ow.image(container)
				if err != nil {
					fmt.Println(err)
				} else {
					containerSpec["image"] = containerImage
				}
				deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"] =
					append(deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]map[string]interface{}), containerSpec)
			}
		}
	}

	return deployment, nil
}
