// This file contains all the generator functions that we need in order
// to build the Flootic GraphQL schema and resolvers.
package generator

import (
	. "github.com/flootic/generator/graphql"
)

// CreateGraphqlSchema creates a new GraphQL schema from the given
// package. The schema is then saved to disk.
func CreateGraphqlSchema(p *Package) error {
	f := NewFile()

	for _, e := range p.Model.Entities {
		CreateType(e, f)
	}

	return f.Save(p.Filename)
}

// CreateType adds a new object type for the given entity to the schema.
func CreateType(e *Entity, f *File) {
	f.Type(e.Name, func(t *Type) {
		for _, a := range e.Attributes {
			t.Field(a.Name).Type(a.Type)
		}
		for _, r := range e.Relations {
			t.Field(r.Name()).Type(r.Entity)
		}
	})
}
