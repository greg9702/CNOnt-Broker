package ontology

import (
	"fmt"
	"log"
	"os"

	"github.com/shful/gofp"
	"github.com/shful/gofp/owlfunctional"
	"github.com/shful/gofp/owlfunctional/assertions"
	"github.com/shful/gofp/owlfunctional/axioms"
	"github.com/shful/gofp/owlfunctional/decl"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const podClassName = ":Pod"

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

// ObjectPropertyAssertionsByIndividual returns ObjectPropertyAssertions that are about individual which declaration was passed
func (ow *OntologyWrapper) ObjectPropertyAssertionsByIndividual(individualDecl *decl.NamedIndividualDecl) []assertions.ObjectPropertyAssertion {
	allObjectPropertyAssertions := ow.ontology.K.AllObjectPropertyAssertions()

	isAssertionAboutIndividual := func(s assertions.ObjectPropertyAssertion) bool {
		return s.A1.Name == convertIRI2Name(individualDecl.IRI)
	}

	return filterObjPropAssertions(allObjectPropertyAssertions, isAssertionAboutIndividual)
}

// ObjectPropertyAssertionsByIndividual returns ObjectPropertyAssertions that are about individual which declaration was passed
func (ow *OntologyWrapper) DataPropertyAssertionsByIndividual(individual string) []axioms.DataPropertyAssertion {
	allDataPropertyAssertions := ow.ontology.K.AllDataPropertyAssertions()

	isAssertionAboutIndividual := func(s axioms.DataPropertyAssertion) bool {
		return s.A.Name == individual
	}

	return filterDataPropAssertions(allDataPropertyAssertions, isAssertionAboutIndividual)
}

func (ow *OntologyWrapper) IndividualsByClass(className string) (individuals []string) {
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
	return ow.IndividualsByClass(podClassName)
}

func (ow *OntologyWrapper) BuildDeploymentConfiguration() *appsv1.Deployment {
	log.Println("That's what we parsed: ", ow.ontology.About())

	for _, pod := range ow.Pods() {
		fmt.Println(pod)
		for _, dataAssertion := range ow.DataPropertyAssertionsByIndividual(pod) {
			fmt.Println(convertIRI2Name(dataAssertion.R.(*decl.DataPropertyDecl).IRI)) //TODO maybe is there a way to avoid casting...
		}
	}
	// for _, individual := range ow.ontology.K.AllNamedIndividualDecls() {
	// 	fmt.Println(individual)
	// }

	// for _, declaration := range ow.ontology.K.AllClassDecls() {
	// 	classDecl := declaration
	// 	fmt.Println(convertIRI2Name(classDecl.IRI))
	// 	fmt.Println("\t", ow.IndividualsByClass(convertIRI2Name(classDecl.IRI)))
	// }

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment
}

func int32Ptr(i int32) *int32 { return &i }
