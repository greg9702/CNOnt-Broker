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
)

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

// ObjectPropertyAssertions returns ObjectPropertyAssertions that are about individual which declaration was passed
func (ow *OntologyWrapper) ObjectPropertyAssertions(individualDecl *decl.NamedIndividualDecl) []assertions.ObjectPropertyAssertion {
	allObjectPropertyAssertions := ow.ontology.K.AllObjectPropertyAssertions()

	isAssertionAboutIndividual := func(s assertions.ObjectPropertyAssertion) bool {
		return s.A1.Name == convertIRI2Name(individualDecl.IRI)
	}

	return filterObjPropAssertions(allObjectPropertyAssertions, isAssertionAboutIndividual)
}

func (ow *OntologyWrapper) IndividualsOfClass(classDecl *decl.ClassDecl) (individuals []string) {
	allClassAssertions := ow.ontology.K.AllClassAssertions()

	isAssertionAboutClass := func(s axioms.ClassAssertion) bool {
		if !s.C.IsNamedClass() {
			return false
		}
		return (s.C).(*decl.ClassDecl).IRI == classDecl.IRI
	}

	classAssertions := filterClassAssertions(allClassAssertions, isAssertionAboutClass)

	for _, classAssertion := range classAssertions {
		individuals = append(individuals, classAssertion.A.Name)
	}

	return individuals
}

func (ow *OntologyWrapper) BuildDeploymentConfiguration() {
	log.Println("That's what we parsed: ", ow.ontology.About())

	for _, individual := range ow.ontology.K.AllNamedIndividualDecls() {
		fmt.Println(individual)
	}

	for _, declaration := range ow.ontology.K.AllClassDecls() {
		classDecl := declaration
		fmt.Println(convertIRI2Name(classDecl.IRI))
		fmt.Println("\t", ow.IndividualsOfClass(classDecl))
	}
}
