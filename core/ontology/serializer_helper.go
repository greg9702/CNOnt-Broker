package ontology

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

const clusterClassName string = ":KubernetesCluster"
const containersClassName string = ":DockerContainer"
const applicationContainersClassName string = ":ApplicationContainer"
const applicationContainerGroupClassName string = ":ApplicationContainerGroup"
const podsClassName string = ":Pod"
const nodesClassName string = ":Node"
const microservicesClassName string = ":Microservice"
const hardwareClassName string = ":Hardware"
const containerEngineClassName string = ":ContainerEngine"
const replicaSetClassName string = ":ReplicaSet"

// TODO get this from ontology
var allClassesKeys = []string{clusterClassName, containersClassName, podsClassName, replicaSetClassName, nodesClassName}

// ClusterStruct used to describe cluster object for our requirements,
// k8s do not define structure like this, so we do
type ClusterStruct struct {
	Name string
}

var buildHelperInstance *builderHelper
var once sync.Once

// BuilderHelpersInstance return instance of builderHelper singleton class
func BuilderHelpersInstance() *builderHelper {
	once.Do(func() {
		buildHelperInstance = newBuilderHelpers()
	})
	return buildHelperInstance
}

type builderHelper struct {
	dataPropertyFunctions map[string]map[string]func(interface{}) string
}

func newBuilderHelpers() *builderHelper {

	dataPropertyFunctions := make(map[string]map[string]func(interface{}) string)

	// node
	tmpMap := make(map[string]func(interface{}) string)

	tmpMap[":name"] = func(object interface{}) string {
		nodeObject := object.(*apiv1.Node)
		return nodeObject.Name
	}

	dataPropertyFunctions[nodesClassName] = tmpMap
	tmpMap = nil

	// pod
	tmpMap = make(map[string]func(interface{}) string)

	tmpMap[":name"] = func(object interface{}) string {
		rsObject := object.(*appsv1.ReplicaSet)
		lastInd := strings.LastIndex(rsObject.Name, "-")
		if lastInd != -1 {
			return rsObject.Name[:lastInd]
		}
		return rsObject.Name
	}
	tmpMap[":app"] = func(object interface{}) string {
		rsObject := object.(*appsv1.ReplicaSet)
		if x, found := rsObject.Spec.Selector.MatchLabels["app"]; found {
			return x
		}
		return ""
	}
	tmpMap[":replicas"] = func(object interface{}) string {
		rsObject := object.(*appsv1.ReplicaSet)
		return strconv.Itoa(int(*rsObject.Spec.Replicas))
	}

	dataPropertyFunctions[podsClassName] = tmpMap
	tmpMap = nil

	// containers
	tmpMap = make(map[string]func(interface{}) string)

	tmpMap[":name"] = func(object interface{}) string {
		containerObject := object.(*apiv1.Container)
		return containerObject.Name
	}
	tmpMap[":image"] = func(object interface{}) string {
		containerObject := object.(*apiv1.Container)
		return containerObject.Image
	}
	tmpMap[":port"] = func(object interface{}) string {
		containerObject := object.(*apiv1.Container)
		if len(containerObject.Ports) != 0 {
			fmt.Println(containerObject.Ports)
			return strconv.Itoa(int(containerObject.Ports[0].ContainerPort))
		}
		return ""
	}

	dataPropertyFunctions[containersClassName] = tmpMap
	tmpMap = nil

	// cluster
	tmpMap = make(map[string]func(interface{}) string)

	tmpMap[":name"] = func(object interface{}) string {
		return "kind"
	}

	dataPropertyFunctions[clusterClassName] = tmpMap
	tmpMap = nil

	bh := builderHelper{dataPropertyFunctions}
	return &bh
}

// DataPropertyFunction return function for retrieving dataProperty value
// for given className and dataPropertyName
func (bh *builderHelper) DataPropertyFunction(className string, dataPropertyName string) (func(interface{}) string, error) {

	var errorMessage string
	if m, ok := bh.dataPropertyFunctions[className]; ok {
		if fn, ok := m[dataPropertyName]; ok {
			return fn, nil
		}
		errorMessage = "For classname " + className + " data property " + dataPropertyName + " not found"
	} else {
		errorMessage = "Classname " + className + " not found"
	}
	return nil, errors.New(errorMessage)
}
