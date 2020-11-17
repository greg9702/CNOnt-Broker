package ontology

import (
	"path/filepath"
	"testing"
)

type ObjectPropertyPair struct {
	PropertyName     string
	RelatedClassName string
}

func objectPropertyTuple2Pair(tuple *ObjectPropertyTuple) *ObjectPropertyPair {
	return &ObjectPropertyPair{tuple.PropertyName, tuple.RelatedClassName}
}

func objectPropertyTuples2Pairs(tuples []*ObjectPropertyTuple) []*ObjectPropertyPair {
	var pairs []*ObjectPropertyPair
	for _, t := range tuples {
		pairs = append(pairs, objectPropertyTuple2Pair(t))
	}
	return pairs
}

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

func haveObjPropSlicesTheSameElems(x []*ObjectPropertyTuple, y []*ObjectPropertyPair) bool {
	xMap := make(map[ObjectPropertyPair]int)
	yMap := make(map[ObjectPropertyPair]int)

	for _, xElem := range x {
		xMap[*objectPropertyTuple2Pair(xElem)]++
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

func objPropsToStr(dataProps []*ObjectPropertyPair) string {
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
			{replicaSetClassName, []string{":name", ":replicas"}},
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
			dataProps []*ObjectPropertyPair
		}{
			{clusterClassName, []*ObjectPropertyPair{
				{":contains_node", nodesClassName}}},
			{nodesClassName, []*ObjectPropertyPair{
				{":belongs_to_cluster", clusterClassName},
				{":contains_pod", podsClassName}}},
			{podsClassName, []*ObjectPropertyPair{
				{":belongs_to_node", nodesClassName},
				{":contains_container", containersClassName},
				{":runs_inside", microservicesClassName},
				{":is_owned_by", replicaSetClassName}}},
			{containersClassName, []*ObjectPropertyPair{
				{":is_managed_by", containerEngineClassName},
				{":belongs_to_group", podsClassName}}},
			{replicaSetClassName, []*ObjectPropertyPair{
				{":owns", podsClassName}}},
		}

		for _, param := range testParams {
			objProps, err := sut.ObjectPropertiesByClass(param.className)
			if err != nil {
				t.Errorf(err.Error())
			}
			if !haveObjPropSlicesTheSameElems(objProps, param.dataProps) {
				t.Errorf("Error ObjectPropertyNamesByClass\n"+
					"Expected data props for class %s are : %s\n"+
					"The result was: %s", param.className, objPropsToStr(param.dataProps), objPropsToStr(objectPropertyTuples2Pairs(objProps)))
			}
		}
	})
}
