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
const appAssertion = ":app"
const kindAssertion = ":kind"
const replicasAssertion = ":replicas"
const imageAssertion = ":image"
const portAssertion = ":port"

const belongsToNodeAssertion = ":belongs_to_node"
const containsContainerAssertion = ":contains_container"

// OntologyWrapper functional OWL ontology wrapper
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

// PrintClasses prints to console all classes in ontology
func (ow *OntologyWrapper) PrintClasses() {
	log.Println("That's what we parsed: ", ow.ontology.About())

	for _, declaration := range ow.ontology.K.AllClassDecls() {
		fmt.Println(declaration.IRI)
	}
}

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

// isC1C2OrSubclassOfC2 returns true if c1 is c2 or its child class, false otherwise
func (ow *OntologyWrapper) isC1C2OrSubclassOfC2(c1 string, c2 string) bool {
	allInheritanceRelations := ow.ontology.K.AllSubClassOfs()
	if c1 == c2 {
		return true
	}
	for _, rel := range allInheritanceRelations {
		if ((convertIRI2Name(rel.C1.(*decl.ClassDecl).IRI)) == c1) &&
			(convertIRI2Name(rel.C2.(*decl.ClassDecl).IRI)) == c2 {
			return true
		}
	}
	return false
}

// DataPropertyNamesByClass returns string slice with data property names of a given class
func (ow *OntologyWrapper) DataPropertyNamesByClass(className string) ([]string, error) {
	allDataProperties := ow.ontology.K.AllDataPropertyDomains()

	isClassDomainOfProp := func(dataProp axioms.DataPropertyDomain) bool {
		return ow.isC1C2OrSubclassOfC2(className, convertIRI2Name(dataProp.C.(*decl.ClassDecl).IRI))
	}

	filteredDataProperties := filterDataProperties(allDataProperties, isClassDomainOfProp)

	if len(filteredDataProperties) == 0 {
		return []string{}, errors.New("No data properties found for " + className)
	}

	dataPropertyNamesSet := make(map[string]struct{})

	for _, dp := range filteredDataProperties {
		dataPropertyNamesSet[convertIRI2Name(dp.R.(*decl.DataPropertyDecl).IRI)] = struct{}{}
	}
	dataPropertyNames := make([]string, 0, len(dataPropertyNamesSet))
	for dp := range dataPropertyNamesSet {
		dataPropertyNames = append(dataPropertyNames, dp)
	}
	return dataPropertyNames, nil
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

// app returns app of a given pod
func (ow *OntologyWrapper) app(pod string) (string, error) {
	return ow.dataPropertyAssertionValue(appAssertion, pod)
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

	// we will call this base deployment structure in the future
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": "",
			},
			"spec": map[string]interface{}{
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "",
						},
					},
				},
			},
		},
	}

	// TODO first get names from all Pods - this will specify how many different deployment would we need to create
	// now we create only one deployment and if all Pods have the same value set it makes no problem

	for _, pod := range ow.pods() {

		podName, err := ow.name(pod)
		if err != nil {
			fmt.Println("Could not get 'name' for Pod, error " + err.Error())
			return nil, errors.New("Could not get 'name' for Pod, error " + err.Error())
		}

		// set deployment name, but all Pods from this deployment will use this name too
		deployment.Object["metadata"].(map[string]interface{})["name"] = podName

		app, err := ow.app(pod)
		if err != nil {
			fmt.Println("Could not get 'app' for Pod " + podName + ", error: " + err.Error())
			return nil, errors.New("Could not get 'app' for Pod " + podName + ", error: " + err.Error())
		}

		deployment.Object["spec"].(map[string]interface{})["selector"].(map[string]interface{})["matchLabels"].(map[string]interface{})["app"] = app
		deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["app"] = app

		replicas, err := ow.replicas(pod)
		if err != nil {
			fmt.Println("Could not get 'replicas' for Pod " + podName + ", error: " + err.Error())
			return nil, errors.New("Could not get 'replicas' for Pod " + podName + ", error: " + err.Error())
		}

		deployment.Object["spec"].(map[string]interface{})["replicas"] = replicas

		containers, err := ow.containers(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"] = map[string]interface{}{}
			deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"] = []map[string]interface{}{}

			for _, container := range containers {

				containerSpec := map[string]interface{}{}

				containerName, err := ow.name(container)
				if err != nil {
					fmt.Println("Could not get 'name' for container in Pod " + podName + ", error: " + err.Error())
					return nil, errors.New("Could not get 'name' for container in Pod " + podName + ", error: " + err.Error())
				}
				containerSpec["name"] = containerName

				containerImage, err := ow.image(container)
				if err != nil {
					fmt.Println("Could not get 'image' for container in Pod " + podName + ", error: " + err.Error())
					return nil, errors.New("Could not get 'image' for container in Pod " + podName + ", error: " + err.Error())
				}
				containerSpec["image"] = containerImage

				deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"] =
					append(deployment.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]map[string]interface{}), containerSpec)
			}
		}
	}

	return deployment, nil
}
