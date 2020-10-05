package main

import (
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"
)

// CreateModel generates a Golang file that contains all the necessary
// Golang structs that represent the model
func CreateModel(p *Package) error {
	f := NewFile(p.Name)

	f.PackageComment(fmt.Sprintf("%s contains all the library code for the platform", p.Name))
	f.PackageComment("This file contains all the functions that implement the model")
	f.PackageComment(" ** THIS CODE IS MACHINE GENERATED. DO NOT EDIT MANUALLY ** ")

	AddModelStructs(p.Model.Entities, f)

	return f.Save(p.Filename)
}

// AddModelStructs generates all the structs from the
// model, and adds them to the given file
func AddModelStructs(entities []*Entity, f *File) {
	for _, e := range entities {
		AddModelStruct(f, e)
	}
}

// AddModelStruct is a helper function that generates the Golang struct
// that represents the model for the given entity
func AddModelStruct(f *File, e *Entity) {
	f.Type().Id(e.Name).StructFunc(func(g *Group) {

		// Add a struct field for each entity attribute
		for _, a := range e.Attributes {
			TypedFromAttribute(g.Id(a.Name), a)
		}

		// Add a struct field for each relation. We we built a pointer
		// type for each entity we point at
		for _, r := range e.Relations {
			g.Id(r.Name()).Op("*").Id(r.Entity)
		}
	})
}

// TypedFromAttribute appends the appropiate Golang type to the given
// statement, according to the type of the given attribute
func TypedFromAttribute(s *Statement, a *Attribute) *Statement {
	return TypedFromDataType(s, AttributeDatatype(a))
}

// TypedFromDataType appends the appropiate Golang type to the given
// statement, according to the type of the given data type
func TypedFromDataType(s *Statement, dataType string) *Statement {
	return s.Add(TypeFromDataType(dataType))

}

// TypeFromAttribute returns the Golang type statement for the given
// attribute
func TypeFromAttribute(a *Attribute) *Statement {
	return TypeFromDataType(AttributeDatatype(a))
}

// TypeFromDataType returns the Golang type statement for the given
// data type string representation
func TypeFromDataType(dataType string) *Statement {

	switch dataType {
	case "int":
		return Int()

	case "float":
		return Float64()

	case "boolean":
		return Bool()

	default:
		return String()

	}
}

// AttributeDatatype normalizes the attribute datatype so that we can
// safely transform it into a Golang type
func AttributeDatatype(a *Attribute) string {
	return strings.ToLower(a.Type)
}

// TypeFromRelation returns the Golang type statement for the given
// relation
func TypeFromRelation(r *Relation) *Statement {
	return TypeFromDataType(RelationDatatype(r))
}

// Relation normalizes the attribute datatype so that we can
// safely transform it into a Golang type. A relation is an Id, which is
// implemented as a string in Go.
func RelationDatatype(r *Relation) string {
	return "string"
}
