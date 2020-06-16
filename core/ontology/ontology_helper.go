package ontology

import (
	"strings"

	"github.com/shful/gofp/owlfunctional/assertions"
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
