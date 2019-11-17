package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

// Model describes the Flootic application model
type Model struct {
	Types    []*UDType
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
	m.ResolveTypes()
	m.ResolveOperations()
	return m, err
}

// Package holds all the metadata that describes a package. Generator
// functions can then act on a package and add specific behavior
type Package struct {
	Name     string
	Filename string
	Model    *Model
	Database string
}

// ImplementTraits traverses all entities in the model, and for each
// entity, it inspects the traits, and translates them into the
// appropiate attributes and relations.
func (m *Model) ImplementTraits() {
	for _, e := range m.Entities {
		e.ImplementTraits()
	}
}

// ResolveOperations traverses all entities in the model, and for each
// entity, it inspects the operations. If no operations are defined,
// then by default we assign create, update, delete and find.
func (m *Model) ResolveOperations() {
	for _, e := range m.Entities {
		e.ResolveOperations()
	}
}

// EntityForName returns the entity of the given name, in the model, or
// nil if no such entity is found
func (m *Model) EntityForName(n string) *Entity {
	for _, e := range m.Entities {
		if e.Name == n {
			return e
		}
	}

	return nil
}

// EntityForNameOrPanic returns the entity of the given name, in the
// model. If the entity is not found, then this function will panic
func (m *Model) EntityForNameOrPanic(n string) *Entity {
	e := m.EntityForName(n)

	if e == nil {
		panic(fmt.Sprintf("Entity %s not found in model", n))
	}

	return e

}

// TypeForName returns the user defined of the given name, in the model, or
// nil if no such entity is found
func (m *Model) TypeForName(n string) *UDType {
	for _, t := range m.Types {
		if t.Name == n {
			return t
		}
	}

	return nil
}

// ResolveTypes resolves complex user defined types.
func (m *Model) ResolveTypes() {
	for _, t := range m.Types {
		if t.Type == "Union" {
			newValues := []string{}
			newType := t.Type
			for _, v := range t.Values {
				if t2 := m.TypeForName(v); t2 != nil {
					for _, v2 := range t2.Values {
						newValues = append(newValues, v2)
					}
					newType = "String"
				} else {
					newValues = append(newValues, v)
				}
			}
			t.Type = newType
			t.Values = newValues
		}
	}
}

// VarName converts the given name, into a golang variable name. The
// convention is to convert all to lowercase.
func VarName(name string) string {
	return strings.ToLower(name)
}

// UDType represents a user defined type. This will allow for
// extensibility
type UDType struct {
	Name   string
	Type   string
	Values []string
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
	Hooks      map[string][]string
	Operations []string
}

// VarName returns the variable name representation for the
// relation
func (e *Entity) VarName() string {
	return VarName(e.Name)
}

// Plural returns the plural name for the given entity. This function
// appends an 's' to the entity name, unless the plural is overriden by
// the user (to be implemented)
func (e *Entity) Plural() string {
	return fmt.Sprintf("%ss", e.Name)
}

var entityOps = []string{
	"create", "update", "delete", "find",
}

// ResolveOperations ensures that the entity has a valid set of
// operations defined
func (e *Entity) ResolveOperations() {
	if len(e.Operations) == 0 {
		e.Operations = entityOps
		return
	}

	// check we have configured valid operations
	ops := strings.Join(entityOps, ",")
	for _, op := range e.Operations {
		if !strings.Contains(ops, op) {
			panic(fmt.Sprintf("Invalid operation %s in %s", op, e.Name))
		}

	}
}

// SupportsOperation returns whether the given operation is supported by
// the entity
func (e *Entity) SupportsOperation(op string) bool {
	for _, o := range e.Operations {
		if o == op {
			return true
		}
	}

	return false
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

// PreferredSort returns the default attribute to be used for
// sorting items of this entity
func (e *Entity) PreferredSort() *Attribute {
	for _, a := range e.Attributes {
		if a.Name != "ID" && (a.HasModifier("unique") || a.HasModifier("indexed")) {
			return a
		}
	}

	// if no indexed attributes are defined,
	// then use the ID attribute
	return &Attribute{Name: "ID"}
}

// HasGenerators returns whether or not this entity has generated
// attributes or relations
func (e *Entity) HasGenerators() bool {

	for _, a := range e.Attributes {
		if a.HasModifier("generated") {
			return true
		}
	}

	for _, r := range e.Relations {
		if r.HasModifier("generated") {
			return true
		}
	}

	return false
}

// EntityInitialization builds the initialization of a new entity struct
// pointer for the given entity
func EntityInitialization(e *Entity) *Statement {
	return Op("&").Id(e.Name).Values(DictFunc(func(d Dict) {
		for _, a := range e.Attributes {
			d[Id(a.Name)] = Id(a.VarName())
		}

		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
				d[Id(r.Name())] = Op("&").Id(r.Entity).Values(Dict{
					Id("ID"): Id(r.VarName()),
				})
			}
		}
	}))
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
	return fmt.Sprintf("%s", strings.ToLower(r.Name()))
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
