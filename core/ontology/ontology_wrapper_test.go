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

	if len(xMap) != len(yMap) {
		return false
	}

	for xMapKey, xMapVal := range xMap {
		if yMap[xMapKey] != xMapVal {
			return false
		}
	}
	return true
}

func haveObjPropSlicesTheSameElems(x []*ObjectPropertyTuple, y []*ObjectPropertyTuple) bool {
	xMap := make(map[ObjectPropertyTuple]int)
	yMap := make(map[ObjectPropertyTuple]int)

	for _, xElem := range x {
		xMap[*xElem]++
	}
	for _, yElem := range y {
		yMap[*yElem]++
	}

	if len(xMap) != len(yMap) {
		return false
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

func objPropsToStr(dataProps []*ObjectPropertyTuple) string {
	objPropsStr := ""
	if len(dataProps) == 0 {
		return ""
	}
	for _, op := range dataProps {
		objPropsStr += "{" + op.PropertyName + ", " + op.RelatedClassName + "}; "
	}
	return objPropsStr[:len(objPropsStr)-2]
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

	t.Run("Test getting object property names by class", func(t *testing.T) {
		testParams := []struct {
			className string
			dataProps []*ObjectPropertyTuple
		}{
			{clusterClassName, []*ObjectPropertyTuple{
				{":contains_node", nodesClassName}}},
			{nodesClassName, []*ObjectPropertyTuple{
				{":belongs_to_cluster", clusterClassName},
				{":contains_pod", podsClassName}}},
			{podsClassName, []*ObjectPropertyTuple{
				{":belongs_to_node", nodesClassName},
				{":contains_container", containersClassName},
				{":runs_inside", microservicesClassName}}},
			{containersClassName, []*ObjectPropertyTuple{
				{":has_limits", hardwareClassName},
				{":requests", hardwareClassName},
				{":is_managed_by", containerEngineClassName},
				{":belongs_to_group", podsClassName}}},
		}

		for _, param := range testParams {
			objProps, err := sut.ObjectPropertiesByClass(param.className)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !haveObjPropSlicesTheSameElems(objProps, param.dataProps) {
				t.Errorf("Error ObjectPropertyNamesByClass\n"+
					"Expected data props for class %s are : %s\n"+
					"The result was: %s", param.className, objPropsToStr(param.dataProps), objPropsToStr(objProps))
			}
		}
	})
}
