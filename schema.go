package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
	"strings"
)

// CreateSchema generates a Golang file that produces the Graphql schema
// in text format
func CreateSchema(p *Package) error {
	f := NewFile(p.Name)

	AddNewSchemaFun(p.Model, f)

	return f.Save(p.Filename)
}

// AddNewSchemaFun generates the function that builds the schema.
func AddNewSchemaFun(m *Model, f *File) {
	funName := "Schema"

	f.Comment(fmt.Sprintf("%s returns the Graphql schema", funName))
	f.Func().Id(funName).Params().Id("string").BlockFunc(func(g *Group) {
		schema := BuildSchema(m)
		g.Return(Lit(schema))
	})
}

// BuildSchema parses the model and generates the text representation of
// the Graphql schema for the given model
func BuildSchema(m *Model) string {
	s := &GraphqlSchema{}

	for _, e := range m.Entities {
		s.Types = append(s.Types, GraphqlSchemaTypeFromEntity(e))

		s.Mutations = append(s.Mutations, GraphqlCreateMutationFromEntity(e))

		for _, a := range e.Attributes {
			if a.HasModifier("indexed") && a.HasModifier("unique") {
				s.Queries = append(s.Queries, GraphqlFinderQueryFromAttribute(e, a))
			}
		}
	}

	return s.String()
}

// GraphqlSchemaTypeFromEntity converts the given entity to the more
// convenient GraphqlType
func GraphqlSchemaTypeFromEntity(e *Entity) *GraphqlType {
	t := &GraphqlType{
		Name: e.Name,
	}

	for _, a := range e.Attributes {
		t.Fields = append(t.Fields, GraphqlFieldFromAttribute(a))
	}

	for _, r := range e.Relations {
		t.Fields = append(t.Fields, GraphqlFieldFromRelation(r))
	}

	return t
}

// GraphqlCreateMutationFromEntity returns a mutation that creates
// instances of the given entity
func GraphqlCreateMutationFromEntity(e *Entity) *GraphqlFun {
	m := &GraphqlFun{
		Name: fmt.Sprintf("create%s", e.Name),
		Returns: &GraphqlField{
			DataType: e.Name,
			Required: true,
			Many:     false,
		},
	}

	for _, a := range e.Attributes {
		m.Args = append(m.Args, GraphqlFieldFromAttribute(a))
	}

	for _, r := range e.Relations {
		f := GraphqlFieldFromRelation(r)
		f.DataType = "ID"

		m.Args = append(m.Args, f)
	}

	return m
}

// GraphqlFinderQueryFromAttribute returns a query that finds
// a single instance of entity by an indexed and unique attribute
func GraphqlFinderQueryFromAttribute(e *Entity, a *Attribute) *GraphqlFun {
	m := &GraphqlFun{
		Name: fmt.Sprintf("find%sBy%s", e.Name, a.Name),
		Returns: &GraphqlField{
			DataType: e.Name,
			Required: true,
			Many:     false,
		},
	}

	m.Args = append(m.Args, GraphqlFieldFromAttribute(a))
	return m
}

// GraphqlSchema is an internal simplified Graphql model
type GraphqlSchema struct {
	Types     []*GraphqlType
	Queries   []*GraphqlFun
	Mutations []*GraphqlFun
}

func (s *GraphqlSchema) String() string {
	chunks := []string{}
	chunks = append(chunks, `
	schema {
        query: Query
        mutation: Mutation
    }`)

	chunks = append(chunks, "\n\n")
	for _, t := range s.Types {
		chunks = append(chunks, fmt.Sprintf("%s\n", t.String()))
	}
	chunks = append(chunks, "\n\n")
	chunks = append(chunks, "type Mutation {\n")
	for _, m := range s.Mutations {
		chunks = append(chunks, fmt.Sprintf("  %s\n", m.String()))
	}
	chunks = append(chunks, "}\n")
	chunks = append(chunks, "\n\n")
	chunks = append(chunks, "type Query {\n")
	for _, q := range s.Queries {
		chunks = append(chunks, fmt.Sprintf("  %s\n", q.String()))
	}
	chunks = append(chunks, "}\n")
	return strings.Join(chunks, "")
}

// GraphqlType is an internal simplified Graphql model
type GraphqlType struct {
	Name   string
	Desc   string
	Fields []*GraphqlField
}

func (t *GraphqlType) String() string {
	chunks := []string{}
	chunks = append(chunks, fmt.Sprintf("type %s {\n", t.Name))
	for _, f := range t.Fields {
		chunks = append(chunks, fmt.Sprintf("  %s: %s\n", f.Name, f.DataTypeString()))
	}
	chunks = append(chunks, "}\n")
	return strings.Join(chunks, "")
}

// GraphqlFun is an internal simplified Graphql model
type GraphqlFun struct {
	Name    string
	Desc    string
	Args    []*GraphqlField
	Returns *GraphqlField
}

func (o *GraphqlFun) String() string {
	args := []string{}
	for _, a := range o.Args {
		args = append(args, fmt.Sprintf("%s:%s", a.Name, a.DataTypeString()))
	}

	return fmt.Sprintf("%s(%s): %s",
		o.Name,
		strings.Join(args, ", "),
		o.Returns.DataTypeString(),
	)
}

// GraphqlField is an internal simplified Graphql model
type GraphqlField struct {
	Name     string
	DataType string
	Required bool
	Many     bool
}

// GraphqlFieldFromAttribute converts a model attribute into a more
// convenient Graphql Field
func GraphqlFieldFromAttribute(a *Attribute) *GraphqlField {
	return &GraphqlField{
		Name:     AttributeGraphqlFieldName(a),
		DataType: AttributeGraphqlFieldDataType(a),
		Required: true,
		Many:     false,
	}
}

// AttributeGraphqlFieldName returns the Graphql field name for the
// given attribute
func AttributeGraphqlFieldName(a *Attribute) string {
	switch a.Name {
	case "ID":
		return "id"

	default:
		return strcase.ToLowerCamel(a.Name)
	}
}

// AttributeGraphqlFieldDataType returns the Graphql field type for the
// given attribute
func AttributeGraphqlFieldDataType(a *Attribute) string {
	return a.Type
}

// GraphqlFieldFromRelation converts a model relation into a more
// convenient Graphql Field
func GraphqlFieldFromRelation(r *Relation) *GraphqlField {
	return &GraphqlField{
		Name:     RelationGraphqlFieldName(r),
		DataType: RelationGraphqlFieldDataType(r),
		Required: true,
		Many:     r.HasModifier("hasMany"),
	}
}

// RelationGraphqlFieldName returns the Graphql field name for the
// given relation
func RelationGraphqlFieldName(r *Relation) string {
	return strcase.ToLowerCamel(r.Name())
}

// RelationGraphqlFieldDatatype returns the Graphql field type for the
// given relation
func RelationGraphqlFieldDataType(r *Relation) string {
	return r.Entity
}

func (f *GraphqlField) DataTypeString() string {
	s := f.DataType
	if f.Required {
		s = fmt.Sprintf("%s!", s)
	}

	if f.Many {
		s = fmt.Sprintf("[%s]", s)
	}

	return s
}
