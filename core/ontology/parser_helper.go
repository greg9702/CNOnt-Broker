package ontology

import (
	"github.com/shful/gofp/owlfunctional/assertions"
	"github.com/shful/gofp/owlfunctional/axioms"
	"strings"
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
