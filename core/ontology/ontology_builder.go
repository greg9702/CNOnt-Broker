package ontology

import (
	"CNOnt-Broker/core/kubernetes/client"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
)

// ReplicasNumberForPods TODO do something with this...
var ReplicasNumberForPods map[string]int

// TODO move it to configs

// TODO get this from ontology
var allTypesKeys = []string{clusterClassName, containersClassName, podsClassName, nodesClassName}

// ObjectToDump contains all required data for dumping object
// in functionl OWL ontology format
type ObjectToDump struct {
	className                string
	objectName               string
	dataPropertyAssertions   map[string]string
	objectPropertyAssertions map[string]string
}

// ObjectsToDumpCollection stores all ObjectToDump objects
type ObjectsToDumpCollection struct {
	collection []*ObjectToDump
}

// Add insert new element into ObjectsToDumpCollection
func (oc *ObjectsToDumpCollection) Add(object *ObjectToDump) {
	oc.collection = append(oc.collection, object)
	fmt.Printf("[ObjectsToDumpCollection] Add: Added new element: %s\n", object.objectName)
}

// GetObjectsByClassName returns list of objects of className given type
func (oc *ObjectsToDumpCollection) GetObjectsByClassName(className string) []*ObjectToDump {

	var tempList []*ObjectToDump

	for ix := range oc.collection {
		if oc.collection[ix].className == className {
			tempList = append(tempList, oc.collection[ix])
		}
	}

	return tempList
}

// OntologyBuilder offers serialization utility (cluster configuration -> ontology)
type OntologyBuilder struct {
	k8sClient     *client.KubernetesClient
	objectsToDump *ObjectsToDumpCollection
	apiData       map[string][]interface{}
}

// NewOntologyBuilder creates new NewOntologyBuilder instance
func NewOntologyBuilder(kubernetesClient *client.KubernetesClient) *OntologyBuilder {
	objectCollection := ObjectsToDumpCollection{}
	apiData := make(map[string][]interface{})
	ob := OntologyBuilder{kubernetesClient, &objectCollection, apiData}
	return &ob
}

// fetchDataFromAPI obtains all required data from API and save it in
// ow.apiData, this approach  will significantly reduce API calls
func (ow *OntologyBuilder) fetchDataFromAPI() error {
	// this are only 3-4 calls, if some of them will be redundant it wont change much
	// that's why we do not ask what data we exactly need, we just fetch everything

	var tempList []interface{}

	// nodes
	nodes, err := ow.k8sClient.GetAllNodes()
	if err != nil {
		return err
	}

	for ix := range nodes.Items {
		node := &nodes.Items[ix]
		tempList = append(tempList, node)
	}
	ow.apiData[nodesClassName] = tempList

	fmt.Printf("[OntologyBuilder] fetchDataFromAPI: Added %d nodes\n", len(tempList))

	tempList = nil

	// pods
	pods, err := ow.k8sClient.GetAllPods("default")
	if err != nil {
		return err
	}

	// here we have cut out Pods which are only replicas, we will create seperate
	// map for it and save it too
	ReplicasNumberForPods = make(map[string]int)

	for ix := range pods.Items {
		pod := &pods.Items[ix]
		podName := trimPodsIDs(pod.Name)

		if val, alreadyExists := ReplicasNumberForPods[podName]; alreadyExists {
			val++
			ReplicasNumberForPods[podName] = val
			continue
		}
		ReplicasNumberForPods[podName] = 1
		pod.Name = podName // use trimmed name for this pod
		tempList = append(tempList, pod)
	}

	ow.apiData[podsClassName] = tempList

	fmt.Printf("[OntologyBuilder] fetchDataFromAPI: Added %d pods\n", len(tempList))

	tempList = nil

	// containers
	for ix := range ow.apiData[podsClassName] {
		pod := ow.apiData[podsClassName][ix].(*apiv1.Pod)

		containers := pod.Spec.Containers
		for i := range containers {
			tempList = append(tempList, containers[i])
		}
	}

	ow.apiData[containersClassName] = tempList

	fmt.Printf("[OntologyBuilder] fetchDataFromAPI: Added %d containers\n", len(tempList))

	tempList = nil

	return nil
}

// GenerateCollection creates objects for all individuals and
// fills their Class- and DataProperty- Assertions
func (ow *OntologyBuilder) GenerateCollection() error {

	err := ow.fetchDataFromAPI()

	if err != nil {
		return err
	}

	// based on ow.apiData we have to fill ow.objectsToDump collection

	// first we create all ObjectToDump objects with its className, objectName
	// and its dataPropertyAssertions

	// TODO we need to get all dataProperties names (keys) from ontology
	for ix := range allTypesKeys {

		className := allTypesKeys[ix]
		objectName := className + "RANDOMSTRING"

		if className == clusterClassName {

			dataProperties := make(map[string]string)
			dataProperties["name"] = "TOBEREMOVED"

			// can we use make in constructor?
			obj := &ObjectToDump{clusterClassName, objectName, dataProperties, make(map[string]string)}
			ow.objectsToDump.Add(obj)

		} else if className == nodesClassName {
			allNodes := ow.apiData[className]

			for ix := range allNodes {
				node := allNodes[ix].(*apiv1.Node)

				podsProperties := []string{"name"}

				for propertyIx := range podsProperties {
					property := podsProperties[propertyIx]

					fn, err := BuilderHelpersInstance().GetDataPropertyFunction(className, property)

					if err != nil {
						fn = func(interface{}) string {
							return ""
						}
					}

					dataProperties := make(map[string]string)
					dataProperties[property] = fn(node)

					obj := &ObjectToDump{className, objectName, dataProperties, make(map[string]string)}
					fmt.Println(obj)
					ow.objectsToDump.Add(obj)
				}
			}

		} else if className == podsClassName {
			allPods := ow.apiData[className]

			for ix := range allPods {
				pod := allPods[ix].(*apiv1.Pod)

				// get pod data properties
				dataProperties := make(map[string]string)

				podsProperties := []string{"name", "app", "replicas"}

				for propertyIx := range podsProperties {
					property := podsProperties[propertyIx]

					fn, err := BuilderHelpersInstance().GetDataPropertyFunction(className, property)

					if err != nil {
						fn = func(interface{}) string {
							return ""
						}
					}

					dataProperties[property] = fn(pod)
				}
				obj := &ObjectToDump{className, objectName, dataProperties, make(map[string]string)}
				fmt.Println(obj)
				ow.objectsToDump.Add(obj)
			}
		} else {
			fmt.Printf("[OntologyBuilder] GenerateCollection: Skipping class %s\n", allTypesKeys[ix])
		}
	}

	// second, we have to link all objects with each other setting proper objectPropertyAssertions
	// for every of them

	err = ow.dumpData()

	if err != nil {
		fmt.Printf("[OntologyBuilder] GenerateCollection: dumpData error: %s", err.Error())
		return err
	}

	return nil
}

// dumpData dumps all objects from ObjectsToDumpCollection
// into the file
func (ow *OntologyBuilder) dumpData() error {

	return nil
}

func trimPodsIDs(podName string) string {

	// we trim it from the beginning to the 2nd last "-" character
	// Example: example-deployment-58fd8d47cd-5ggz4
	// Output: example-deployment

	for i := 0; i < 2; i++ {
		if ix := strings.LastIndex(podName, "-"); ix != -1 {
			podName = podName[:ix]
		} else {
			break
		}
	}
	return podName
}
