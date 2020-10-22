package ontology

import (
	"CNOnt-Broker/core/kubernetes/client"
	"fmt"
)

//TODO create objects of this struct in GenerateCollection
type ObjectToDump struct {
	className              string
	objectName             string
	dataPropertyAssertions map[string]string
	//objectPropertyAssertions map[string]string //TODO
}

//TODO should be used to store all ObjectsToDump
type ObjectToDumpCollection struct {
	collection []*ObjectToDump
}

// TODO implementation
// OntologyBuilder offers serialization utility (cluster configuration -> ontology)
type OntologyBuilder struct {
	k8sClient *client.KubernetesClient
}

// NewOntologyBuilder creates new NewOntologyBuilder instance
func NewOntologyBuilder(kubernetesClient *client.KubernetesClient) *OntologyBuilder {
	ob := OntologyBuilder{kubernetesClient}
	return &ob
}

// askAPI obtains Data Properties for each instance of a given class
func (ow *OntologyBuilder) askAPI(className string) (DataPropertiesByInstanceName, error) {
	// 1. obtain info about command and filter function
	dataFinder := commandAndFilterFunctionByClass[className]

	// 2. Call API
	commandOutput, err := ow.k8sClient.ExecuteCommand(dataFinder.command)

	fmt.Println(commandOutput)
	if err != nil {
		fmt.Printf("Error in executing command, %s", err.Error())
		return DataPropertiesByInstanceName{}, err
	}
	fmt.Println(dataFinder.filterFunc(commandOutput))

	// 3. Filter command output and return map (keys - instance names, values - data properties)
	return dataFinder.filterFunc(commandOutput), nil
}

// GenerateCollection creates objects for all individuals and
// fills their Class- and DataProperty- Assertions
func (ow *OntologyBuilder) GenerateCollection() {
	// iterate through all elements in cluste
	// create and save cluster object using clasterClassName
	// create and save nodes objects
	// ...

	// dump cluster
	// - get data from api
	// - create ObjectToDump
	// - add ref to o.collection
	className := "Pod"
	fmt.Println("GenerateCollection")
	ow.askAPI(className)

}
