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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (ow *OntologyWrapper) Pods() []string {
	return ow.individualsByClass(podClass)
}

func (ow *OntologyWrapper) BuildDeploymentConfiguration() *appsv1.Deployment {
	deploymentName := "demo-deployment"

	// default values
	replicas := int32Ptr(2)
	podName := "demo"

	containerName := "web"
	image := "nginx:1.12"
	protocol := apiv1.ProtocolTCP
	var containerPort int32 = 80

	// override default values if something was found in ontology
	for _, pod := range ow.Pods() {
		fmt.Println("Pod: " + pod)
		p, err := ow.dataPropertyAssertionValue(podNameAssertion, pod)
		if err != nil {
			fmt.Println(err)
		} else {
			podName = p
		}

		r, err := ow.dataPropertyAssertionValue(replicasAssertion, pod)
		if err != nil {
			fmt.Println(err)
		} else {
			i, err := strconv.Atoi(r)
			if err != nil {
				fmt.Println(err)
			} else {
				replicas = int32Ptr(int32(i))
				print(*replicas)
			}
		}
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": podName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": podName,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  containerName,
							Image: image,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      protocol,
									ContainerPort: containerPort,
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Println(*deployment)

	return deployment
}
