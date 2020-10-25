package ontology

import (
	"errors"
	appsv1 "k8s.io/api/apps/v1"
	"strconv"
	"strings"
	"sync"

	"github.com/shful/gofp/owlfunctional/assertions"
	"github.com/shful/gofp/owlfunctional/axioms"

	apiv1 "k8s.io/api/core/v1"
)

func convertIRI2Name(IRI string) string {
	return ":" + strings.Split(IRI, "#")[1]
}

func filterObjPropAssertions(assertions []assertions.ObjectPropertyAssertion, test func(assertions.ObjectPropertyAssertion) bool) (ret []assertions.ObjectPropertyAssertion) {
	for _, assertion := range assertions {
		if test(assertion) {
			ret = append(ret, assertion)
		}
	}
	return
}

func filterClassAssertions(assertions []axioms.ClassAssertion, test func(axioms.ClassAssertion) bool) (ret []axioms.ClassAssertion) {
	for _, assertion := range assertions {
		if test(assertion) {
			ret = append(ret, assertion)
		}
	}
	return
}

func filterDataPropAssertions(assertions []axioms.DataPropertyAssertion, test func(axioms.DataPropertyAssertion) bool) (ret []axioms.DataPropertyAssertion) {
	for _, assertion := range assertions {
		if test(assertion) {
			ret = append(ret, assertion)
		}
	}
	return
}

func objectAssertions2String(assertions []assertions.ObjectPropertyAssertion) (ret []string) {
	for _, assertion := range assertions {
		ret = append(ret, assertion.A2.Name)
	}
	return
}

func int32Ptr(i int32) *int32 { return &i }

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

	tmpMap["name"] = func(object interface{}) string {
		nodeObject := object.(*apiv1.Node)
		return nodeObject.Name
	}

	dataPropertyFunctions[nodesClassName] = tmpMap
	tmpMap = nil

	// pod
	tmpMap = make(map[string]func(interface{}) string)

	tmpMap["name"] = func(object interface{}) string {
		rsObject := object.(*appsv1.ReplicaSetSpec)
		return rsObject.Template.Name
	}
	tmpMap["app"] = func(object interface{}) string {
		return "MOCK APP"
	}
	tmpMap["replicas"] = func(object interface{}) string {
		rsObject := object.(*appsv1.ReplicaSetSpec)

		return strconv.Itoa(int(*rsObject.Replicas))
	}

	dataPropertyFunctions[podsClassName] = tmpMap
	tmpMap = nil

	// containers
	tmpMap = make(map[string]func(interface{}) string)

	tmpMap["name"] = func(object interface{}) string {
		containerObject := object.(*apiv1.Container)
		return containerObject.Name
	}
	tmpMap["image"] = func(object interface{}) string {
		containerObject := object.(*apiv1.Container)
		return containerObject.Image
	}
	tmpMap["port"] = func(object interface{}) string {
		containerObject := object.(*apiv1.Container)
		// TODO we assume we have only 1 port...
		if len(containerObject.Ports) != 0 {
			return containerObject.Ports[0].Name
		}
		return ""
	}

	dataPropertyFunctions[containersClassName] = tmpMap
	tmpMap = nil

	// cluster
	tmpMap = make(map[string]func(interface{}) string)

	tmpMap["name"] = func(object interface{}) string {
		return "MOCKCLUSTERNAME"
	}

	dataPropertyFunctions[clusterClassName] = tmpMap
	tmpMap = nil

	bh := builderHelper{dataPropertyFunctions}
	return &bh
}

// GetDataPropertyFunction return function for retreiving dataProperty value
// for given className and dataPropertyName
func (bh *builderHelper) GetDataPropertyFunction(className string, dataPropertyName string) (func(interface{}) string, error) {

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
