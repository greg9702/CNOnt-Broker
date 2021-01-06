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

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const nameAssertion = ":name"
const apiVersionAssertion = ":apiVersion"
const namespaceAssertion = ":namespace"
const kindAssertion = ":kind"
const replicasAssertion = ":replicas"
const imageAssertion = ":image"
const portAssertion = ":port"

const containsContainerAssertion = ":contains_container"
const ownsAssertion = ":owns"

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
	if c1 == c2 {
		return true
	}

	allInheritanceRelations := ow.ontology.K.AllSubClassOfs()
	var parents []string
	for _, rel := range allInheritanceRelations {
		if convertIRI2Name(rel.C1.(*decl.ClassDecl).IRI) == c1 {
			parent := convertIRI2Name(rel.C2.(*decl.ClassDecl).IRI)
			if parent == c2 {
				return true
			}
			parents = append(parents, convertIRI2Name(rel.C2.(*decl.ClassDecl).IRI))
		}
	}
	for _, p := range parents {
		if ow.isC1C2OrSubclassOfC2(p, c2) {
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

// generateFilterFunction generates filter function for given property
func (ow *OntologyWrapper) generateFilterFunction(objPropName string) func(interface{}, interface{}) bool {

	// default
	fn := func(interface{}, interface{}) bool {
		return false
	}
	// we assume that all nodes belong to one cluster, so :contains_node and
	// :belongs_to_cluster filter functions always return true

	if objPropName == ":belongs_to_cluster" {
		fn = func(interface{}, interface{}) bool {
			return true
		}
	} else if objPropName == ":contains_node" {
		fn = func(interface{}, interface{}) bool {
			return true
		}
	} else if objPropName == ":belongs_to_node" {
		fn = func(obj1 interface{}, obj2 interface{}) bool {
			podObj := obj1.(*v1.Pod)
			nodeObj := obj2.(*v1.Node)

			if podObj.Spec.NodeName == nodeObj.Name {
				return true
			}
			return false
		}
	} else if objPropName == ":is_owned_by" {
		fn = func(obj1 interface{}, obj2 interface{}) bool {
			podObj := obj1.(*v1.Pod)
			rsObj := obj2.(*appsv1.ReplicaSet)

			var podHash string
			var rsHash string

			if x, found := podObj.Labels["pod-template-hash"]; found {
				podHash = x
			} else {
				return false
			}

			if x, found := rsObj.Labels["pod-template-hash"]; found {
				rsHash = x
			} else {
				return false
			}

			if podHash == rsHash {
				return true
			}
			return false
		}
	} else if objPropName == ":owns" {
		fn = func(obj1 interface{}, obj2 interface{}) bool {
			rsObj := obj1.(*appsv1.ReplicaSet)
			podObj := obj2.(*v1.Pod)

			var podHash string
			var rsHash string

			if x, found := podObj.Labels["pod-template-hash"]; found {
				podHash = x
			} else {
				return false
			}

			if x, found := rsObj.Labels["pod-template-hash"]; found {
				rsHash = x
			} else {
				return false
			}

			if podHash == rsHash {
				return true
			}
			return false
		}
	} else if objPropName == ":contains_pod" {
		fn = func(obj1 interface{}, obj2 interface{}) bool {
			nodeObj := obj1.(*v1.Node)
			podObj := obj2.(*v1.Pod)

			if podObj.Spec.NodeName == nodeObj.Name {
				return true
			}
			return false
		}
	} else if objPropName == ":belongs_to_group" {
		fn = func(obj1 interface{}, obj2 interface{}) bool {
			containerStruct := obj1.(*ContainerStruct)
			podObj := obj2.(*v1.Pod)

			return containerStruct.PodName == podObj.Name
		}
	} else if objPropName == ":contains_container" {
		fn = func(obj1 interface{}, obj2 interface{}) bool {
			podObj := obj1.(*v1.Pod)
			containerStruct := obj2.(*ContainerStruct)

			return containerStruct.PodName == podObj.Name
		}
	}
	return fn
}

func (ow *OntologyWrapper) objectPropertyByName(name string) (ObjectPropertyTuple, error) {
	allObjectPropertyRanges := ow.ontology.K.AllObjectPropertyRanges()

	isClassRangeOfProp := func(objProp axioms.ObjectPropertyRange) bool {
		return convertIRI2Name(objProp.P.(*decl.ObjectPropertyDecl).IRI) == name
	}

	filteredObjectProperties := filterObjectPropertyRanges(allObjectPropertyRanges, isClassRangeOfProp)

	if len(filteredObjectProperties) == 0 {
		return ObjectPropertyTuple{}, errors.New("No '" + name + "' object properties found")
	} else if len(filteredObjectProperties) > 1 {
		return ObjectPropertyTuple{}, errors.New("Multiple '" + name + "' object properties found")
	}

	fn := ow.generateFilterFunction(name)
	return ObjectPropertyTuple{name, convertIRI2Name(filteredObjectProperties[0].C.(*decl.ClassDecl).IRI), fn}, nil
}

// ObjectPropertiesByClass returns ObjectPropertyTuple slice with object properties of a given class
func (ow *OntologyWrapper) ObjectPropertiesByClass(className string) ([]*ObjectPropertyTuple, error) {

	allObjectProperties := ow.ontology.K.AllObjectPropertyDomains()

	isClassDomainOfProp := func(objProp axioms.ObjectPropertyDomain) bool {
		return ow.isC1C2OrSubclassOfC2(className, convertIRI2Name(objProp.C.(*decl.ClassDecl).IRI))
	}

	filteredObjProperties := filterObjectPropertyDomains(allObjectProperties, isClassDomainOfProp)

	objectPropertyNamesSet := make(map[string]struct{})
	for _, dp := range filteredObjProperties {
		objectPropertyNamesSet[convertIRI2Name(dp.P.(*decl.ObjectPropertyDecl).IRI)] = struct{}{}
	}

	objectProperties := make([]*ObjectPropertyTuple, 0, len(objectPropertyNamesSet))
	for name := range objectPropertyNamesSet {
		op, err := ow.objectPropertyByName(name)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		objectProperties = append(objectProperties, &op)
	}

	return objectProperties, nil
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

// replicaSets returns all rs individuals
func (ow *OntologyWrapper) replicaSets() []string {
	return ow.individualsByClass(replicaSetClassName)
}

// podOwnedByReplicaSet returns pod individual that is owned by rs individual
func (ow *OntologyWrapper) podOwnedByReplicaSet(rs string) (string, error) {
	pod, err := ow.objectPropertyAssertionValue(ownsAssertion, rs)
	if err != nil {
		return "", err
	}
	rsAppLabel, err := ow.dataPropertyAssertionValue(appAssertion, rs)
	podAppLabel, err := ow.dataPropertyAssertionValue(appAssertion, pod)
	if err != nil {
		return "", err
	}
	if rsAppLabel != podAppLabel {
		return "", errors.New("Ontology not consistent: 'app' label in ReplicaSet " + rs + " selector ('app'='" + rsAppLabel +
			"') differs from the app label of pod " + pod + " owned by the ReplicaSet ('app'='" + podAppLabel + "').")
	}
	return pod, nil
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
func (ow *OntologyWrapper) namespace(pod string) (string, error) {
	return ow.dataPropertyAssertionValue(namespaceAssertion, pod)
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

// buildContainerResources returns container resources (requests and limits)
func (ow *OntologyWrapper) buildContainerResources(container string) (map[string]interface{}, error) {
	var resources = map[string]interface{}{}

	memReq, memReqErr := ow.dataPropertyAssertionValue(":memory_requests", container)
	cpuReq, cpuReqErr := ow.dataPropertyAssertionValue(":cpu_requests", container)
	if memReqErr == nil || cpuReqErr == nil {
		resources["requests"] = map[string]interface{}{}

		if memReqErr == nil && memReq != "0" {
			resources["requests"].(map[string]interface{})["memory"] = memReq
		}
		if cpuReqErr == nil && cpuReq != "0" {
			resources["requests"].(map[string]interface{})["cpu"] = cpuReq
		}
	}

	memLim, memLimErr := ow.dataPropertyAssertionValue(":memory_limits", container)
	cpuLim, cpuLimErr := ow.dataPropertyAssertionValue(":cpu_limits", container)
	if memLimErr == nil || cpuLimErr == nil {
		resources["limits"] = map[string]interface{}{}

		// 'memLim != "0"' etc. is a workaround in case the ontology gives us inconsistent entries (lim > req)
		// F.e. :memory_limits = 0 is returned by serializer when there are no memory limits specified
		// TODO implement handling of this case in serializer and write a better consistency check here
		if memLimErr == nil && memLim != "0" {
			resources["limits"].(map[string]interface{})["memory"] = memLim
		}
		if cpuLimErr == nil && cpuLim != "0" {
			resources["limits"].(map[string]interface{})["cpu"] = cpuLim
		}
	}

	if len(resources) == 0 {
		return map[string]interface{}{}, errors.New("No resources data found in ontology for container: " + container)
	}

	return resources, nil
}

// BuildDeploymentConfiguration returns Kubernetes Deployments basing on parsed ontology
func (ow *OntologyWrapper) BuildDeploymentConfiguration() ([]*unstructured.Unstructured, error) {
	var deployments []*unstructured.Unstructured

	for i, rs := range ow.replicaSets() {
		deployments = append(deployments, &unstructured.Unstructured{
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
		})

		rsName, err := ow.name(rs)
		if err != nil {
			fmt.Println("Could not get 'name' for ReplicaSet, error " + err.Error())
			return nil, errors.New("Could not get 'name' for ReplicaSet, error " + err.Error())
		}

		// set deployment name, but all Pods from this deployment will use this name too
		deployments[i].Object["metadata"].(map[string]interface{})["name"] = rsName

		app, err := ow.app(rs)
		if err != nil {
			fmt.Println("Could not get 'app' label for ReplicaSet " + rsName + ", error: " + err.Error())
			return nil, errors.New("Could not get 'app' for ReplicaSet " + rsName + ", error: " + err.Error())
		}

		deployments[i].Object["spec"].(map[string]interface{})["selector"].(map[string]interface{})["matchLabels"].(map[string]interface{})["app"] = app
		deployments[i].Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["metadata"].(map[string]interface{})["labels"].(map[string]interface{})["app"] = app

		replicas, err := ow.replicas(rs)
		if err != nil {
			fmt.Println("Could not get 'replicas' for ReplicaSet " + rsName + ", error: " + err.Error())
			return nil, errors.New("Could not get 'replicas' for ReplicaSet " + rsName + ", error: " + err.Error())
		}

		deployments[i].Object["spec"].(map[string]interface{})["replicas"] = replicas

		pod, err := ow.podOwnedByReplicaSet(rs)

		if err != nil {
			fmt.Println("Could not get any pods owned by ReplicaSet " + rsName + ", error: " + err.Error())
			return nil, errors.New("Could not get any pods owned by ReplicaSet " + rsName + ", error: " + err.Error())
		}
		containers, err := ow.containers(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			deployments[i].Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"] = map[string]interface{}{}
			deployments[i].Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"] = []map[string]interface{}{}

			for _, container := range containers {

				containerSpec := map[string]interface{}{}

				containerName, err := ow.name(container)
				if err != nil {
					fmt.Println("Could not get 'name' for container in Pod from template of ReplicaSet " + rsName + ", error: " + err.Error())
					return nil, errors.New("Could not get 'name' for container in Pod from template of ReplicaSet " + rsName + ", error: " + err.Error())
				}
				containerSpec["name"] = containerName

				containerImage, err := ow.image(container)
				if err != nil {
					fmt.Println("Could not get 'image' for container in Pod from template of ReplicaSet " + rsName + ", error: " + err.Error())
					return nil, errors.New("Could not get 'image' for container in Pod from template of ReplicaSet " + rsName + ", error: " + err.Error())
				}
				containerSpec["image"] = containerImage

				resources, err := ow.buildContainerResources(container)
				if err != nil {
					fmt.Println(err)
				} else {
					containerSpec["resources"] = resources
				}

				deployments[i].Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"] =
					append(deployments[i].Object["spec"].(map[string]interface{})["template"].(map[string]interface{})["spec"].(map[string]interface{})["containers"].([]map[string]interface{}), containerSpec)
			}
		}
	}

	return deployments, nil
}
