package ontology

import (
	"CNOnt-Broker/core/kubernetes/client"
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

// ObjectToDump contains all required data for dumping object
// in functional OWL ontology format
type ObjectToDump struct {
	className                string
	objectName               string
	dataPropertyAssertions   map[string]string
	objectPropertyAssertions map[string][]string
}

// ObjectsToDumpCollection stores all ObjectToDump objects
type ObjectsToDumpCollection struct {
	collection []*ObjectToDump
}

// add inserts new element into ObjectsToDumpCollection
func (oc *ObjectsToDumpCollection) add(object *ObjectToDump) {
	oc.collection = append(oc.collection, object)
	fmt.Printf("[ObjectsToDumpCollection] add: Added new element: %s\n", object.objectName)
}

// ObjectsByClassName returns list of objects of className given type
func (oc *ObjectsToDumpCollection) ObjectsByClassName(className string) []*ObjectToDump {

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
	wrapper       *OntologyWrapper
}

// NewOntologyBuilder creates new OntologyBuilder instance
func NewOntologyBuilder(kubernetesClient *client.KubernetesClient, wrapper *OntologyWrapper) *OntologyBuilder {
	objectCollection := ObjectsToDumpCollection{}
	apiData := make(map[string][]interface{})
	ob := OntologyBuilder{kubernetesClient, &objectCollection, apiData, wrapper}
	return &ob
}

// fetchDataFromAPI obtains all required data from API and saves it in
// ow.apiData, this approach  will significantly reduce API calls
func (ow *OntologyBuilder) fetchDataFromAPI() error {
	// this are only 3-4 calls, if some of them will be redundant it wont change much
	// that's why we do not ask what data we exactly need, we just fetch everything

	var tempList []interface{}

	// cluster
	cs := ClusterStruct{"TEST_CLUSTER"}
	tempList = append(tempList, &cs)

	ow.apiData[clusterClassName] = tempList
	tempList = nil

	// nodes
	nodes, err := ow.k8sClient.AllNodes()
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
	replicaSets, err := ow.k8sClient.AllReplicaSets("default")
	if err != nil {
		return err
	}

	for _, rs := range replicaSets.Items {
		rsSpec := &rs.Spec
		tempList = append(tempList, rsSpec)
	}

	ow.apiData[podsClassName] = tempList

	fmt.Printf("[OntologyBuilder] fetchDataFromAPI: Added %d pods\n", len(tempList))

	tempList = nil

	// containers
	for ix := range ow.apiData[podsClassName] {
		pod := ow.apiData[podsClassName][ix].(*appsv1.ReplicaSetSpec)

		containers := pod.Template.Spec.Containers
		for i := range containers {
			tempList = append(tempList, &containers[i])
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
		fmt.Printf("[OntologyBuilder] GenerateCollection: fetchDataFromAPI error: %s\n", err.Error())
		return err
	}

	// based on ow.apiData we have to fill ow.objectsToDump collection

	// first we create all ObjectToDump objects with its className, objectName
	// and its dataPropertyAssertions

	for ix := range allClassesKeys {
		className := allClassesKeys[ix]
		objectName := className + "RANDOMSTRING" // TODO random string

		allObjects := ow.apiData[className]

		if len(allObjects) == 0 {
			fmt.Printf("[OntologyBuilder] GenerateCollection: No data found for classname %s\n", className)
			continue
		}

		for ix := range allObjects {
			object := allObjects[ix]

			propertiesList, err := ow.dataPropertiesList(className)
			if err != nil {
				fmt.Printf("[OntologyBuilder] GenerateCollection: %s", err.Error())
				continue
			}

			dataProperties := make(map[string]string)

			for propertyIx := range propertiesList {
				property := propertiesList[propertyIx]

				fn, err := BuilderHelpersInstance().DataPropertyFunction(className, property)

				if err != nil {
					fn = func(interface{}) string {

						return ""
					}
				}
				dataProperties[property] = fn(object)
			}

			obj := &ObjectToDump{className, objectName, dataProperties, make(map[string]string)}
			ow.objectsToDump.add(obj)
			fmt.Println(obj)
		}
	}

	// TODO second, we have to link all objects with each other setting proper objectPropertyAssertions
	// for every of them

	err = ow.dumpData()

	if err != nil {
		fmt.Printf("[OntologyBuilder] GenerateCollection: dumpData error: %s", err.Error())
		return err
	}

	return nil
}

// TODO dumpData dumps all objects from ObjectsToDumpCollection
// into the file
func (ow *OntologyBuilder) dumpData() error {

	return nil
}

//dataPropertiesList returns list of data properties for given class
func (ow *OntologyBuilder) dataPropertiesList(className string) ([]string, error) {
	return ow.wrapper.DataPropertyNamesByClass(className)
}
