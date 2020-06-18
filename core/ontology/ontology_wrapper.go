package ontology

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/shful/gofp"
	"github.com/shful/gofp/owlfunctional"
	"github.com/shful/gofp/owlfunctional/assertions"
	"github.com/shful/gofp/owlfunctional/axioms"
	"github.com/shful/gofp/owlfunctional/decl"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const podClass = ":Pod"
const podNameAssertion = ":name"
const replicasAssertion = ":replicas"
const belongsToNodeAssertion = ":belongs_to_node"

type OntologyWrapper struct {
	ontology *owlfunctional.Ontology
}

func NewOntologyWrapper(path string) *OntologyWrapper {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	o, err := gofp.OntologyFromReader(f, "source was "+path)
	if err != nil {
		log.Fatal(gofp.ErrorMsgWithPosition(err))
	}
	ow := OntologyWrapper{o}
	return &ow
}

func (ow *OntologyWrapper) PrintClasses() {
	log.Println("That's what we parsed: ", ow.ontology.About())

	for _, declaration := range ow.ontology.K.AllClassDecls() {
		fmt.Println(declaration.IRI)
	}
}

// objectPropertyAssertionValue returns string value of particular ObjectPropertyAssertion about passed individual
func (ow *OntologyWrapper) objectPropertyAssertionValue(assertionName string, individual string) (string, error) {
	allObjectPropertyAssertions := ow.ontology.K.AllObjectPropertyAssertions()

	isAssertionAboutIndividual := func(s assertions.ObjectPropertyAssertion) bool {
		return s.A1.Name == individual && convertIRI2Name(s.PN) == assertionName
	}

	filteredAssrtions := filterObjPropAssertions(allObjectPropertyAssertions, isAssertionAboutIndividual)
	if len(filteredAssrtions) == 0 {
		return "", errors.New("No '" + assertionName + "' assertions found for " + individual)
	} else if len(filteredAssrtions) > 1 {
		return "", errors.New("Multiple '" + assertionName + "' assertions found for " + individual)
	}

	return filteredAssrtions[0].A2.Name, nil
}

// dataPropertyAssertionValue returns string value of particular DataPropertyAssertion about passed individual
func (ow *OntologyWrapper) dataPropertyAssertionValue(assertionName string, individual string) (string, error) {
	allDataPropertyAssertions := ow.ontology.K.AllDataPropertyAssertions()

	isAssertionAboutIndividual := func(s axioms.DataPropertyAssertion) bool {
		return s.A.Name == individual && convertIRI2Name(s.R.(*decl.DataPropertyDecl).IRI) == assertionName
	}

	filteredAssrtions := filterDataPropAssertions(allDataPropertyAssertions, isAssertionAboutIndividual)
	if len(filteredAssrtions) == 0 {
		return "", errors.New("No '" + assertionName + "' assertions found for " + individual)
	} else if len(filteredAssrtions) > 1 {
		return "", errors.New("Multiple '" + assertionName + "' assertions found for " + individual)
	}

	return filteredAssrtions[0].V.Value, nil
}

// individualsByClass returns all individuals from a given class
func (ow *OntologyWrapper) individualsByClass(className string) (individuals []string) {
	allClassAssertions := ow.ontology.K.AllClassAssertions()

	isAssertionAboutClass := func(s axioms.ClassAssertion) bool {
		if !s.C.IsNamedClass() {
			return false
		}
		return convertIRI2Name((s.C).(*decl.ClassDecl).IRI) == className
	}
	classAssertions := filterClassAssertions(allClassAssertions, isAssertionAboutClass)

	for _, classAssertion := range classAssertions {
		individuals = append(individuals, classAssertion.A.Name)
	}

	return individuals
}

// pods returns all pod individuals
func (ow *OntologyWrapper) pods() []string {
	return ow.individualsByClass(podClass)
}

// podName returns name of a given pod
func (ow *OntologyWrapper) podName(pod string) (string, error) {
	podName, err := ow.dataPropertyAssertionValue(podNameAssertion, pod)
	if err != nil {
		return "", err
	} else {
		return podName, nil
	}
}

// podReplicas returns replicas of a given pod
func (ow *OntologyWrapper) podReplicas(pod string) (*int32, error) {
	replicas, err := ow.dataPropertyAssertionValue(replicasAssertion, pod)
	if err != nil {
		return nil, err
	} else {
		r, err := strconv.Atoi(replicas)
		if err != nil {
			return nil, err
		} else {
			return int32Ptr(int32(r)), nil // TODO make sure it's ok returning pointer here
		}
	}
}

// BuildDeploymentConfiguration returns Kubernetes Deployment basing on parsed ontology
func (ow *OntologyWrapper) BuildDeploymentConfiguration() *appsv1.Deployment {

	podName := ""

	// override default values if something was found in ontology
	for _, pod := range ow.pods() {
		
		fmt.Println("Pod: " + pod)
		n, err := ow.podName(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			podName = n
		}

		r, err := ow.podReplicas(pod)
		if err != nil {
			fmt.Println(err)
		} else {
			replicas = r
		}
	}

	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "",
			"kind":       "",
			"metadata": map[string]interface{}{
				"name": "",
			},
			"spec": map[string]interface{}{},
		},
	}

	fmt.Println(*deployment)

	panic(2)
}
