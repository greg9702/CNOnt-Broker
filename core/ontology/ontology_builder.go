package ontology

import (
	"CNOnt-Broker/core/kubernetes/client"
	"fmt"
	"strconv"
	"strings"

	apiv1 "k8s.io/api/core/v1"
)

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

	for ix := range pods.Items {
		pod := &pods.Items[ix]
		tempList = append(tempList, pod)
	}
	ow.apiData[podsClassName] = tempList

	fmt.Printf("[OntologyBuilder] fetchDataFromAPI: Added %d pods\n", len(tempList))

	tempList = nil

	// containers
	for ix := range pods.Items {
		pod := &pods.Items[ix]
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

	// how to get all object properties of

	// TODO we need to unify retreiving elements for different classes
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

				fn, err := BuilderHelpersInstance().GetDataPropertyFunction(className, "name")

				if err != nil {
					fn = func(interface{}) string {
						return ""
					}
				}

				dataProperties := make(map[string]string)
				dataProperties["name"] = fn(node)

				obj := &ObjectToDump{className, objectName, dataProperties, make(map[string]string)}
				fmt.Println(obj)
				ow.objectsToDump.Add(obj)
			}

		} else if className == podsClassName {
			allPods := ow.apiData[className]

			// we have to take all pods take only the unique one
			// all pods of the same kind (from the same deployment)
			// are just treated as next replicas

			// TODO implement logic for removing redundant Pods
			namesOfProcessedPods := make(map[string]int)

			// we use temp list to avoid writing tons of getters and setters for ObjectsToDumpCollection
			// we have local list and we can easily modify previously written elements - for example
			// incrementing its replicas number
			var tempObjectList []*ObjectToDump

			for ix := range allPods {
				pod := allPods[ix].(*apiv1.Pod)

				fn, err := BuilderHelpersInstance().GetDataPropertyFunction(className, "name")

				if err != nil {
					fn = func(interface{}) string {
						return ""
					}
				}

				podName := fn(pod)
				podName = trimPodsIDs(podName)

				// check if pod with the same name already exists
				if val, ok := namesOfProcessedPods[podName]; ok {
					//do something here
					val++
					// update existing pod
					for index := range tempObjectList {
						if tempObjectList[index].dataPropertyAssertions["name"] == podName {
							val, err := strconv.ParseInt(tempObjectList[index].dataPropertyAssertions["replicas"], 10, 64)
							if err != nil {
								panic("!!!!!!!!!")
							}
							val++
							tempObjectList[index].dataPropertyAssertions["replicas"] = strconv.Itoa(int(val))
						}
					}
					continue
				} else {
					namesOfProcessedPods[podName] = 1
				}

				// get pod data properties
				dataProperties := make(map[string]string)

				// this two are very special for Pod
				dataProperties["name"] = podName
				dataProperties["replicas"] = "1"

				// TODO we will iteratate through list of data properties retreived
				// from ontology instead of using hard coded values "app" etc

				fn, err = BuilderHelpersInstance().GetDataPropertyFunction(className, "app")

				if err != nil {
					fn = func(interface{}) string {
						return ""
					}
				}

				dataProperties["app"] = fn(pod)

				obj := &ObjectToDump{className, objectName, dataProperties, make(map[string]string)}
				fmt.Println(obj)
				tempObjectList = append(tempObjectList, obj)
			}

			// save obj from tempObjectList to ow.objectsToDump
			for ix := range tempObjectList {
				ow.objectsToDump.Add(tempObjectList[ix])
			}
		} else {
			fmt.Printf("[OntologyBuilder] GenerateCollection: Omtiting class %s\n", allTypesKeys[ix])
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
