// This file contains the abstract model the generator works with. This
// model is useful to decouple the generator from whatever data
// modelling source (eg. GraphQL)
package generator

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

// Model describes the Flootic application model
type Model struct {
	Entities []*Entity
}

// ReadModelFromFile reads a model from a yaml file in the local
// filesystem
func ReadModelFromFile(path string) (*Model, error) {

	m := &Model{}
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return m, err
	}
	err = yaml.Unmarshal(yamlFile, m)
	m.ImplementTraits()
	return m, err
}

// Package holds all the metadata that describes a package. Generator
// functions can then act on a package and add specific behavior
type Package struct {
	Name     string
	Filename string
	Model    *Model
}

// ImplementTraits traverses all entities in the model, and for each
// entity, it inspects the traits, and translates them into the
// appropiate attributes and relations.
func (m *Model) ImplementTraits() {
	for _, e := range m.Entities {
		e.ImplementTraits()
	}
}

// Entity represents a persisted datatype, such as an Organization, a
// User or a Poi. An entity has relations to other entities.
//
// An entity is defined by its attributes and relations.
//
// An entity can also have traits. Traits are syntax sugar, they enable
// us to inyect pre-defined, well known attributes and relations, in
// a consistent manner, so that we keep the design DRY.
//
// Supported traits are:
// - keys: adds an "ID" and "Name" attributes, of type required, string
//		   and unique
// - timestamps: adds [created|updaed]At attributes
// - authors: adds [created|updaed]By attributes
// - owner: adds an Onwer relation
//
type Entity struct {
	Name       string
	Attributes []*Attribute
	Relations  []*Relation
	Traits     []string
}

// VarName returns the variable name representation for the
// relation
func (e *Entity) VarName() string {
	return strings.ToLower(e.Name)
}

// ImplementTraits translates the entity traits into the appropriate
// attributes and relations
func (e *Entity) ImplementTraits() {
	for _, t := range e.Traits {
		switch t {
		case "id":
			e.AddAttribute("ID", "ID", []string{
				"required",
				"unique",
				"indexed",
			})

		case "keys":
			e.AddAttribute("ID", "ID", []string{
				"required",
				"unique",
				"indexed",
			})
			e.AddAttribute("Name", "String", []string{
				"required",
				"unique",
				"indexed",
			})

		case "timestamps":
			e.AddAttribute("CreatedAt", "Time", []string{
				"required",
				"generated",
			})
			e.AddAttribute("UpdatedAt", "Time", []string{
				"required",
				"generated",
			})

		case "authors":
			e.AddRelation("CreatedBy", "User", []string{
				"required",
				"hasOne",
				"generated",
			})
			e.AddRelation("UpdatedBy", "User", []string{
				"required",
				"hasOne",
				"generated",
			})

		case "owner":
			e.AddRelation("Owner", "User", []string{
				"required",
				"hasOne",
			})
		}
	}
}

// AddAttribute is a convenience function that adds a new attribute to
// the given entity
func (e *Entity) AddAttribute(n string, t string, m []string) {
	e.Attribute(n).WithType(t).WithModifiers(m)
}

// AddAttribute is a convenience function that adds an aliased relation
// to the given entity
func (e *Entity) AddRelation(a string, entity string, m []string) {
	e.AliasedRelation(a).WithEntity(entity).WithModifiers(m)
}

// Adds a new attribute with the given name to the entity
func (e *Entity) Attribute(name string) *Attribute {
	a := &Attribute{
		Name: name,
	}

	e.Attributes = append(e.Attributes, a)
	return a
}

// AliasedRelation adds a new relation with the given alias to the entity
func (e *Entity) AliasedRelation(name string) *Relation {
	r := &Relation{
		Alias: name,
	}

	e.Relations = append(e.Relations, r)
	return r
}

// Relation adds a new relation with the given entity to the entity
func (e *Entity) Relation(name string) *Relation {
	r := &Relation{
		Entity: name,
	}

	e.Relations = append(e.Relations, r)
	return r
}

// Attribute represents a simple entity field. A field has a data type
// and a set of modifiers that help to customize the code generation.
//
// Modifiers can be:
// - unique: indicate the value for the attribute needs to be unique
//			 across all instances of the entity
// - required: indicate the attribute is not nullable
// - indexed: indicate an database index should be created on this field
type Attribute struct {
	Name      string
	Type      string
	Modifiers []string
}

// WithType defines the datatype for the given attribute
func (a *Attribute) WithType(t string) *Attribute {
	a.Type = t
	return a
}

// WithModifiers defines the given set of modifiers on the given attribute
func (a *Attribute) WithModifiers(mods []string) *Attribute {
	a.Modifiers = mods
	return a
}

// HasModifier returns true, if the attribute has the given modifier
func (a *Attribute) HasModifier(m string) bool {
	for _, mod := range a.Modifiers {
		if m == mod {
			return true
		}
	}

	return false
}

// VarName returns a variable name representation for the attribute
func (a *Attribute) VarName() string {
	return strings.ToLower(a.Name)
}

// Relation represents a relation to a foreign entity.
//
// The relation can be aliased with a user defined name,
// otherwise the name of the relation will be the target entity itself.
//
// A relation might have a list of modifiers. The supported list of
// modifiers are:
//
// - belongsTo
// - hasMany
// - hasOne
//
type Relation struct {
	Alias     string
	Entity    string
	Modifiers []string
}

// has a name, then return it, otherwise use the Entity name as the name
// of the relation
func (r *Relation) Name() string {
	if r.Alias != "" {
		return r.Alias
	} else {
		return r.Entity
	}
}

// WithAlias defines the alias on the receiver relation
func (r *Relation) WithAlias(name string) *Relation {
	r.Alias = name
	return r
}

// WithEntity defines the entity on the receiver relation
func (r *Relation) WithEntity(name string) *Relation {
	r.Entity = name
	return r
}

// WithModifiers defines the set of modifiers on the receiver relation
func (r *Relation) WithModifiers(mods []string) *Relation {
	r.Modifiers = mods
	return r
}

// VarName returns the variable name representation for the
// relation
func (r *Relation) VarName() string {
	return fmt.Sprintf("%sRel", strings.ToLower(r.Name()))
}

// HasModifier returns true, if the relation has the given modifier
func (r *Relation) HasModifier(m string) bool {
	for _, mod := range r.Modifiers {
		if m == mod {
			return true
		}
	}

	return false
}

// Named is a generic interface to be implemented
// by structs that are meant to have a variable name representation
type Named interface {
	VarName() string
}
