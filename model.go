// This file contains all the generator functions that we need in order
// to build the Flootic model layer.
package generator

import (
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"
)

// CreateModel generates a Golang file that contains all the necessary
// Golang structs that represent the model
func CreateModel(p *Package) error {
	f := NewFile(p.Name)

	f.PackageComment(fmt.Sprintf("%s contains all the library code for the Flootic platform", p.Name))
	f.PackageComment("This file contains all the functions that implement the model")
	f.PackageComment(" ** THIS CODE IS MACHINE GENERATED. DO NOT EDIT MANUALLY ** ")

	AddModelStructs(p.Model.Entities, f)

	return f.Save(p.Filename)
}

// AddModelStructs generates all the structs of the Flootic
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
	return TypedFromDataType(s, attributeDatatype(a))
}

// TypedFromDataType appends the appropiate Golang type to the given
// statement, according to the type of the given data type
func TypedFromDataType(s *Statement, dataType string) *Statement {
	switch dataType {
	case "int":
		return s.Int()

	case "float":
		return s.Float64()

	case "boolean":
		return s.Bool()

	default:
		return s.String()

	}
}

// attributeDatatype normalizes the attribute datatype so that we can
// safely transform it into a Golang type
func attributeDatatype(a *Attribute) string {
	return strings.ToLower(a.Type)
}
