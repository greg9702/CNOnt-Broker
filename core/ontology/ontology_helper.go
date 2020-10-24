package ontology

import (
	"errors"
	"strings"
	"sync"

	"github.com/shful/gofp/owlfunctional/assertions"
	"github.com/shful/gofp/owlfunctional/axioms"
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

	// TODO init dataPropertyFunctions here
	bh := builderHelper{}
	return &bh
}

// GetDataPropertyFunction return function for retreiving dataProperty value
// for given className and dataPropertyName
func (bh *builderHelper) GetDataPropertyFunction(className string, dataPropertyName string) (func(interface{}) string, error) {

	if m, ok := bh.dataPropertyFunctions[className]; ok {
		if fn, ok := m[dataPropertyName]; ok {
			return fn, nil
		}
	}
	return nil, errors.New("Classname " + className + " or data property name" + dataPropertyName + " not found")
}
