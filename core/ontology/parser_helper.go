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

func filterDataProperties(dataProps []axioms.DataPropertyDomain, test func(axioms.DataPropertyDomain) bool) (ret []axioms.DataPropertyDomain) {
	for _, dataProp := range dataProps {
		if test(dataProp) {
			ret = append(ret, dataProp)
		}
	}
	return
}

func filterObjectPropertyDomains(objProps []axioms.ObjectPropertyDomain, test func(domain axioms.ObjectPropertyDomain) bool) (ret []axioms.ObjectPropertyDomain) {
	for _, objProp := range objProps {
		if test(objProp) {
			ret = append(ret, objProp)
		}
	}
	return
}

func filterObjectPropertyRanges(objProps []axioms.ObjectPropertyRange, test func(domain axioms.ObjectPropertyRange) bool) (ret []axioms.ObjectPropertyRange) {
	for _, objProp := range objProps{
		if test(objProp) {
			ret = append(ret, objProp)
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
