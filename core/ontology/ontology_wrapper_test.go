package ontology

import (
	"path/filepath"
	"testing"
)

func haveSlicesTheSameElems(x []string, y []string) bool {
	xMap := make(map[string]int)
	yMap := make(map[string]int)

	for _, xElem := range x {
		xMap[xElem]++
	}
	for _, yElem := range y {
		yMap[yElem]++
	}

	for xMapKey, xMapVal := range xMap {
		if yMap[xMapKey] != xMapVal {
			return false
		}
	}
	return true
}

func dataPropsToStr(dataProps []string) string {
	dataPropsStr := ""
	for _, dp := range dataProps {
		dataPropsStr += dp + ", "
	}
	return dataPropsStr[:len(dataPropsStr)-2]
}

func TestOntologyWrapper(t *testing.T) {
	sut := NewOntologyWrapper(filepath.Join("assets", "CNOnt.owl"))

	t.Run("Test getting data property names by class", func(t *testing.T) {
		testParams := []struct {
			className string
			dataProps []string
		}{
			{clusterClassName, []string{":name"}},
			{nodesClassName, []string{":name"}},
			{podsClassName, []string{":name", ":app", ":replicas"}},
			{containersClassName, []string{":name", ":image", ":port"}},
		}

		for _, param := range testParams {
			dataProps, err := sut.DataPropertyNamesByClass(param.className)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !haveSlicesTheSameElems(dataProps, param.dataProps) {
				t.Errorf("Error DataPropertyNamesByClass\n"+
					"Expected data props for class %s are : %s\n"+
					"The result was: %s", param.className, dataPropsToStr(param.dataProps), dataPropsToStr(dataProps))
			}
		}
	})
}
