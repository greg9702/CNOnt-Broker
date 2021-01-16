package ontology

import (
	"CNOnt-Broker/core/kubernetes/client"
	"fmt"
	"io/ioutil"
	"strconv"

	logger "CNOnt-Broker/core/common"

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
	logger.BaseLog().Debug("[ObjectsToDumpCollection] add: Added new element: " + object.objectName)
}

// clears ObjectsToDumpCollection buffer
func (oc *ObjectsToDumpCollection) clear() {
	oc.collection = nil
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
	templatePath  string
}

// NewOntologyBuilder creates new OntologyBuilder instance
func NewOntologyBuilder(kubernetesClient *client.KubernetesClient, wrapper *OntologyWrapper, path string) *OntologyBuilder {
	objectCollection := ObjectsToDumpCollection{}
	apiData := make(map[string][]interface{})
	ob := OntologyBuilder{kubernetesClient, &objectCollection, apiData, wrapper, path}
	return &ob
}

// fetchDataFromAPI obtains all required data from API and saves it in
// ow.apiData, this approach  will significantly reduce API calls
func (ow *OntologyBuilder) fetchDataFromAPI() error {
	// this are only 3-4 calls, if some of them will be redundant it wont change much
	// that's why we do not ask what data we exactly need, we just fetch everything

	var tempList []interface{}

	ow.apiData = make(map[string][]interface{})

	// cluster
	cs := ClusterStruct{"k8s-cluster"}
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

	logger.BaseLog().Debug(fmt.Sprintf("[OntologyBuilder] fetchDataFromAPI: Added %d nodes", len(tempList)))

	tempList = nil

	namespaces, err := ow.k8sClient.AllNamespaces()

	if err != nil {
		return err
	}

	for it := range namespaces.Items {
		namespace := namespaces.Items[it].Name
		logger.BaseLog().Debug("for namespce: " + namespace)

		// replicasets
		replicaSets, err := ow.k8sClient.AllReplicaSets(namespace)
		if err != nil {
			return err
		}

		for ix := range replicaSets.Items {
			rs := &replicaSets.Items[ix]
			tempList = append(tempList, rs)
		}

		ow.apiData[replicaSetClassName] = append(ow.apiData[replicaSetClassName], tempList...)

		logger.BaseLog().Debug(fmt.Sprintf("[OntologyBuilder] fetchDataFromAPI: Added %d replica sets", len(tempList)))

		tempList = nil

		// pods
		pods, err := ow.k8sClient.AllPods(namespace)
		if err != nil {
			return err
		}

		for ix := range pods.Items {
			pod := &pods.Items[ix]
			tempList = append(tempList, pod)
		}

		ow.apiData[podsClassName] = append(ow.apiData[podsClassName], tempList...)

		logger.BaseLog().Debug(fmt.Sprintf("[OntologyBuilder] fetchDataFromAPI: Added %d pods", len(tempList)))

		thisNamespacePods := tempList

		tempList = nil

		// containers
		for ix := range thisNamespacePods {
			pod := thisNamespacePods[ix].(*v1.Pod)
			containers := pod.Spec.Containers
			for i := range containers {
				cs := ContainerStruct{pod.Name, &containers[i]}

				tempList = append(tempList, &cs)
			}
		}

		ow.apiData[containersClassName] = append(ow.apiData[containersClassName], tempList...)

		logger.BaseLog().Debug(fmt.Sprintf("[OntologyBuilder] fetchDataFromAPI: Added %d containers", len(tempList)))

		tempList = nil
	}
	return nil
}

// GenerateCollection creates objects for all individuals and
// fills their Class- and DataProperty- Assertions
func (ow *OntologyBuilder) GenerateCollection() (string, error) {

	err := ow.fetchDataFromAPI()

	if err != nil {
		logger.BaseLog().Error("[OntologyBuilder] GenerateCollection: fetchDataFromAPI error: " + err.Error())
		return "", err
	}

	// based on ow.apiData we have to fill ow.objectsToDump collection

	ow.objectsToDump.clear()

	// first we create all ObjectToDump objects with its className, objectName
	// and its dataPropertyAssertions

	for ix := range allClassesKeys {
		className := allClassesKeys[ix]

		allObjects := ow.apiData[className]

		if len(allObjects) == 0 {
			logger.BaseLog().Debug("[OntologyBuilder] GenerateCollection: No data found for classname " + className)
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
				logger.BaseLog().Debug("[OntologyBuilder] GenerateCollection: " + err.Error())
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

				generatedValue := fn(object)

				// if data property function returns "", ignore this property
				if generatedValue != "" {
					dataProperties[property] = generatedValue
				}
			}

			obj := &ObjectToDump{className, objectName, dataProperties, make(map[string][]string), object}
			ow.objectsToDump.add(obj)
			logger.BaseLog().Debug("Added object to the collection: " + obj.objectName)
		}
	}

	// we have to link all objects with each other setting proper objectPropertyAssertions
	// for every of them

	logger.BaseLog().Debug("--------------OBJECT PROPERTIES--------------")

	for ix := range allClassesKeys {
		className := allClassesKeys[ix]

		// [runs_on_node, NODE], [contains_container, CONTAINER]
		allObjectPropertiesForClass, err := ow.objectPropertiesList(className)

		if err != nil {
			logger.BaseLog().Error("[OntologyBuilder] GenerateCollection: objectPropertiesList error: " + err.Error())
			return "", err
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

	logger.BaseLog().Debug("---------------DUMPING DATA---------------")
	savedFile, err := ow.dumpData()

	if err != nil {
		logger.BaseLog().Error("[OntologyBuilder] GenerateCollection: dumpData error: " + err.Error())
		return "", err
	}

	return savedFile, nil
}

// saveToFile saves passed data stream into the target file
func (ow *OntologyBuilder) saveToFile(stream string) (string, error) {

	// read ontology template file
	b, err := ioutil.ReadFile(ow.templatePath)
	if err != nil {
		logger.BaseLog().Error("Template file not found, " + err.Error())
		return "", err
	}

	var dataToWrite []byte

	dataToWrite = append(dataToWrite, b[:len(b)-1]...)
	dataToWrite = append(dataToWrite, []byte(stream)...)
	dataToWrite = append(dataToWrite, b[len(b)-1:]...)

	err = ioutil.WriteFile("/tmp/cluster_mapping.owl", dataToWrite, 0644)
	if err != nil {
		logger.BaseLog().Error("Writing file error, " + err.Error())
		return "", err
	}
	// TODO make it class member!!!!!!!!!!1
	return "/tmp/cluster_mapping.owl", nil
}

// dumpData prepare all objects to dump and convert them into
// final form stream of data
func (ow *OntologyBuilder) dumpData() (string, error) {

	var finalStream string

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
		finalStream += individualHeader + classAssertion + objectPropertyAssertions + dataPropertyAssertions + "\n"
	}
	path, err := ow.saveToFile(finalStream)

	if err != nil {
		return "", err
	}

	return path, nil
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
