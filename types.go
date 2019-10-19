// This file contains the abstract model the generator works with. This
// model is useful to decouple the generator from whatever data
// modelling source (eg. GraphQL)
package generator

import (
	"strings"
)

// Package holds all the metadata that describes a package. Generator
// functions can then act on a package and add specific behavior
type Package struct {
	Name     string
	Filename string
	Entities []*Entity
}

// Entity represents a persisted datatype, such as an Organization, a
// User or a Poi. An entity has relations to other entities. Entities
// can be of different nature, such as one-to-one, one-to-many,
// many-to-many, and we use this info in order to generate the
// repository code
type Entity struct {
	Name      string
	Relations []*Relation
}

// VarName returns the variable name representation for the
// relation
func (e *Entity) VarName() string {
	return strings.ToLower(e.Name)
}

// Relation represents a relation to a foreign entity. The Name
// is the Golang struct that represents the database relation. Entity
// is the foreign type we point to. And Prop is the field that
// contains the ID of the foreign entity instance
type Relation struct {
	Name   string
	Entity string
	Prop   string
	Kind   string
}

// VarName returns the variable name representation for the
// relation
func (r *Relation) VarName() string {
	return strings.ToLower(r.Name)
}

// Named is a generic interface to be implemented
// by structs that are meant to have a variable name representation
type Named interface {
	VarName() string
}
