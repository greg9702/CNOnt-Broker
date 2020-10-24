package ontology

import (
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
)

type DataProperties = map[string]string

// e.g {"pod1": {"name": "name1", "replicas": "3"}, "pod2": {"name": "name2", "replicas": "2"}, ...}
type DataPropertiesByInstanceName = map[string]DataProperties
type ClassName = string

// TODO generate basing on ontology file and use in OntologyBuilder.GenerateCollection
//const classNamesList[] := {":Cluster", ":Node"}

const clusterClassName string = "KubernetesCluster"
const containersClassName string = "DockerContainer"
const podsClassName string = "Pod"
const nodesClassName string = "Node"

type DataPropertiesFinder struct {
	command    string
	filterFunc func(list *v1.PodList) DataPropertiesByInstanceName // TODO make PodList generic!!!
}

func podNameFromPodInstanceName(podInstanceName string) string {
	return podInstanceName[:strings.IndexByte(podInstanceName, '-')]
}

func podsDataProperties(podList *v1.PodList) DataPropertiesByInstanceName {
	res := DataPropertiesByInstanceName{}

	//TODO maybe could be simplified to one loop
	replicasByPodName := make(map[string]int)
	for _, pod := range podList.Items {
		name := podNameFromPodInstanceName(pod.ObjectMeta.Name)
		if _, podNameRepeated := replicasByPodName[name]; podNameRepeated {
			replicasByPodName[name]++
		} else {
			replicasByPodName[name] = 1
		}
	}

	podNum := 1
	for key := range replicasByPodName {
		res["Pod"+strconv.Itoa(podNum)] = map[string]string{
			"name": key, "replicas": strconv.Itoa(replicasByPodName[key]),
		}
		podNum++
	}

	return res
}

//map[className](command, filterFunction)
var commandAndFilterFunctionByClass = map[ClassName]DataPropertiesFinder{
	"Pod": {"GET_PODS", podsDataProperties},
}
