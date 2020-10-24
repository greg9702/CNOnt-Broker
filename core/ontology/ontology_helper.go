package ontology

import (
	"errors"
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

	// TODO move it somwhere?
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
		podObject := object.(*apiv1.Pod)
		return podObject.Name
	}
	tmpMap["app"] = func(object interface{}) string {
		return "MOCK APP"
	}
	tmpMap["replicas"] = func(object interface{}) string {
		podObject := object.(*apiv1.Pod)

		if value, exists := ReplicasNumberForPods[podObject.Name]; exists {
			return strconv.Itoa(value)
		}
		return ""
	}

	dataPropertyFunctions[podsClassName] = tmpMap
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
