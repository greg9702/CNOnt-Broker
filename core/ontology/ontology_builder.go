package ontology

import (
	"CNOnt-Broker/core/kubernetes/client"
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
)

// ObjectToDump contains all required data for dumping object
// in functional OWL ontology format
type ObjectToDump struct {
	className                string
	objectName               string
	dataPropertyAssertions   map[string]string
	objectPropertyAssertions map[string][]string
	originalObjectPointer    interface{}
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

// ObjectsByClassNameFiltered returns list of objects of className given type, related to one
// specific object
func (oc *ObjectsToDumpCollection) ObjectsByClassNameFiltered(className string, object interface{}, function func(interface{}, interface{}) bool) []*ObjectToDump {

	var tempList []*ObjectToDump

	for ix := range oc.collection {
		if oc.collection[ix].className == className {
			if function(object, oc.collection[ix].originalObjectPointer) == true {
				tempList = append(tempList, oc.collection[ix])
			}
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

	// replicasets
	replicaSets, err := ow.k8sClient.AllReplicaSets("default")
	if err != nil {
		return err
	}

	for ix := range replicaSets.Items {
		rs := &replicaSets.Items[ix]
		tempList = append(tempList, rs)
	}

	ow.apiData[replicaSetClassName] = tempList

	fmt.Printf("[OntologyBuilder] fetchDataFromAPI: Added %d replica sets\n", len(tempList))

	tempList = nil

	// pods
	pods, err := ow.k8sClient.AllPods("default")
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
	for ix := range ow.apiData[podsClassName] {
		pod := ow.apiData[podsClassName][ix].(*v1.Pod)
		containers := pod.Spec.Containers
		for i := range containers {
			cs := ContainerStruct{pod.Name, &containers[i]}

			tempList = append(tempList, &cs)
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

		allObjects := ow.apiData[className]

		if len(allObjects) == 0 {
			fmt.Printf("[OntologyBuilder] GenerateCollection: No data found for classname %s\n", className)
			continue
		}

		var objectCounter int = 1

		for ix := range allObjects {

			// this should never happen, but if we exceed integer limit
			// everything will blow up
			objectName := className + strconv.Itoa(objectCounter)
			objectCounter++

			object := allObjects[ix]

			propertiesList, err := ow.dataPropertiesList(className)
			if err != nil {
				fmt.Printf("[OntologyBuilder] GenerateCollection: %s\n", err.Error())
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

			obj := &ObjectToDump{className, objectName, dataProperties, make(map[string][]string), object}
			ow.objectsToDump.add(obj)
			// fmt.Println(obj)
		}
	}

	// we have to link all objects with each other setting proper objectPropertyAssertions
	// for every of them

	for ix := range allClassesKeys {
		className := allClassesKeys[ix]

		// [runs_on_node, NODE], [contains_container, CONTAINER]
		allObjectPropertiesForClass, err := ow.objectPropertiesList(className)
		// _ = err
		// fmt.Println(className)

		// for abc := range allObjectPropertiesForClass {
		// 	fmt.Println(allObjectPropertiesForClass[abc])
		// }

		if err != nil {
			fmt.Printf("[OntologyBuilder] GenerateCollection: objectPropertiesList error: %s", err.Error())
			return err
		}
		objectsToSet := ow.objectsToDump.ObjectsByClassName(className)

		for i := range objectsToSet {
			object := objectsToSet[i]

			for it := range allObjectPropertiesForClass {

				singlePropertyTuple := allObjectPropertiesForClass[it]
				// [Node1, Node2] it gives all Nodes here, need only specific one
				relatedObjects := ow.objectsToDump.ObjectsByClassNameFiltered(singlePropertyTuple.RelatedClassName, object.originalObjectPointer, singlePropertyTuple.FilterFunction)

				for iter := range relatedObjects {
					relatedObject := relatedObjects[iter]
					object.objectPropertyAssertions[singlePropertyTuple.PropertyName] = append(object.objectPropertyAssertions[singlePropertyTuple.PropertyName], relatedObject.objectName)
				}
			}
		}
	}

	fmt.Println("---------------DUMPING DATA---------------")
	err = ow.dumpData()

	if err != nil {
		fmt.Printf("[OntologyBuilder] GenerateCollection: dumpData error: %s", err.Error())
		return err
	}

	return nil
}

// saveToFile saves passed data stream into the target file
func (ow *OntologyBuilder) saveToFile(stream []string) error {
	for ix := range stream {
		fmt.Printf(stream[ix])
	}
	fmt.Printf("\n")
	// b := []byte{}

	// for ix := range stream {
	// 	b = append(b, (stream[ix]).([]byte))
	// 	fmt.Println(b) // [65 66 67 226 130 172]
	// }

	// // write the whole body at once
	// _ = ioutil.WriteFile("/tmp/onto.txt", stream, 0644)
	// if err != nil {
	// 	panic(err)
	// }
	return nil
}

// TODO dumpData dumps all objects from ObjectsToDumpCollection
// into the file
func (ow *OntologyBuilder) dumpData() error {

	for ix := range ow.objectsToDump.collection {
		ind := ow.objectsToDump.collection[ix]

		individualHeader := fmt.Sprintf("# Individual: %s (%s)\n\n", ind.objectName, ind.objectName)
		classAssertion := fmt.Sprintf("ClassAssertion(%s %s)\n", ind.className, ind.objectName)

		// object properties
		var objectPropertyAssertions string
		for key := range ind.objectPropertyAssertions {
			oaList := ind.objectPropertyAssertions[key]
			for id := range oaList {
				objectPropertyAssertions += fmt.Sprintf("ObjectPropertyAssertion(%s %s %s)\n", key, ind.objectName, oaList[id])
			}
		}

		// data properties
		var dataPropertyAssertions string
		for key := range ind.dataPropertyAssertions {
			dataPropertyAssertions += fmt.Sprintf("DataPropertyAssertion(%s %s \"%s\")\n", key, ind.objectName, ind.dataPropertyAssertions[key])
		}
		ow.saveToFile([]string{individualHeader, classAssertion, objectPropertyAssertions, dataPropertyAssertions})
	}
	return nil
}

//dataPropertiesList returns list of data properties for given class
func (ow *OntologyBuilder) dataPropertiesList(className string) ([]string, error) {
	return ow.wrapper.DataPropertyNamesByClass(className)
}

//objectPropertiesList returns list of data properties for given class
func (ow *OntologyBuilder) objectPropertiesList(className string) ([]*ObjectPropertyTuple, error) {
	return ow.wrapper.ObjectPropertiesByClass(className)
}

// ObjectPropertyTuple contains pair of object property name
// and related class to it
type ObjectPropertyTuple struct {
	PropertyName     string
	RelatedClassName string                                        // range
	FilterFunction   func(obj1 interface{}, obj2 interface{}) bool // checks if given object is related to
}
