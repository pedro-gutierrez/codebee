package main

import (
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
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

	for _, t := range m.Types {
		if t.Type == "String" && len(t.Values) > 0 {
			s.Enums = append(s.Enums, GraphqlEnumFromUDType(t))
		}

		if t.Type == "Union" && len(t.Values) > 0 {
			s.Unions = append(s.Unions, GraphqlUnionFromUDType(t))
		}

	}

	for _, e := range m.Entities {
		s.Types = append(s.Types, GraphqlSchemaTypeFromEntity(e))

		if e.SupportsOperation("create") {
			s.Mutations = append(s.Mutations, GraphqlCreateMutationFromEntity(e))
		}

		if e.SupportsOperation("update") {
			s.Mutations = append(s.Mutations, GraphqlUpdateMutationFromEntity(e))
		}

		if e.SupportsOperation("delete") {
			s.Mutations = append(s.Mutations, GraphqlDeleteMutationFromEntity(e))

		}

		if e.SupportsOperation("find") {
			for _, a := range e.Attributes {
				if a.HasModifier("indexed") && a.HasModifier("unique") {
					s.Queries = append(s.Queries, GraphqlFinderQueryFromAttribute(e, a))
				}
			}
			for _, r := range e.Relations {
				if r.HasModifier("hasOne") || r.HasModifier("belongsTo") {
					s.Queries = append(s.Queries, GraphqlFinderQueryFromRelation(e, r))
				}
			}

			s.Queries = append(s.Queries, GraphqlFinderQueryForAll(e))

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
		Name: GraphqlCreateMutationName(e),
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
		if !r.HasModifier("generated") && !r.HasModifier("hasMany") {
			f := GraphqlFieldFromRelation(r)
			f.DataType = "ID"

			m.Args = append(m.Args, f)
		}
	}

	return m
}

// GraphqlUpdateMutationFromEntity returns a mutation that updates
// instances of the given entity
func GraphqlUpdateMutationFromEntity(e *Entity) *GraphqlFun {
	m := &GraphqlFun{
		Name: GraphqlUpdateMutationName(e),
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
		if !r.HasModifier("generated") && !r.HasModifier("hasMany") {
			f := GraphqlFieldFromRelation(r)
			f.DataType = "ID"

			m.Args = append(m.Args, f)
		}
	}

	return m
}

// GraphqlDeleteMutationFromEntity returns a mutation that deletes
// instances of the given entity
func GraphqlDeleteMutationFromEntity(e *Entity) *GraphqlFun {
	m := &GraphqlFun{
		Name: GraphqlDeleteMutationName(e),
		Returns: &GraphqlField{
			DataType: e.Name,
			Required: true,
			Many:     false,
		},
	}

	m.Args = append(m.Args, &GraphqlField{
		Name:     "id",
		DataType: "ID",
		Required: true,
		Many:     false,
	})

	return m
}

// GraphqlFinderQueryForAll returns a query that finds
// all instances of an entity.
func GraphqlFinderQueryForAll(e *Entity) *GraphqlFun {
	m := &GraphqlFun{
		Name: GraphqlFindAllQueryName(e),
		Returns: &GraphqlField{
			DataType: e.Name,
			Required: false,
			Many:     true,
		},
	}

	m.Args = append(m.Args, &GraphqlField{
		Name:     "Limit",
		DataType: "Int",
		Required: true,
		Many:     false,
	})

	m.Args = append(m.Args, &GraphqlField{
		Name:     "Offset",
		DataType: "Int",
		Required: true,
		Many:     false,
	})

	return m
}

// GraphqlFinderQueryFromAttribute returns a query that finds
// a single instance of entity by an indexed and unique attribute
func GraphqlFinderQueryFromAttribute(e *Entity, a *Attribute) *GraphqlFun {
	m := &GraphqlFun{
		Name: GraphqlFindByAttributeQueryName(e, a),
		Returns: &GraphqlField{
			DataType: e.Name,
			Required: true,
			Many:     false,
		},
	}

	m.Args = append(m.Args, GraphqlFieldFromAttribute(a))
	return m
}

// GraphqlFinderQueryFromRelation returns a query that finds
// a list of instances of entity by the ID of the related entity
func GraphqlFinderQueryFromRelation(e *Entity, r *Relation) *GraphqlFun {
	m := &GraphqlFun{
		Name: GraphqlFindByRelationQueryName(e, r),
		Returns: &GraphqlField{
			DataType: e.Name,
			Required: false,
			Many:     true,
		},
	}

	m.Args = append(m.Args, &GraphqlField{
		Name:     r.Alias(),
		DataType: "ID",
		Required: true,
		Many:     false,
	})

	m.Args = append(m.Args, &GraphqlField{
		Name:     "Limit",
		DataType: "Int",
		Required: true,
		Many:     false,
	})

	m.Args = append(m.Args, &GraphqlField{
		Name:     "Offset",
		DataType: "Int",
		Required: true,
		Many:     false,
	})

	return m
}

// GraphqlCreateMutationName returns the name of the mutation that
// creates new instances of the given entity
func GraphqlCreateMutationName(e *Entity) string {
	return fmt.Sprintf("create%s", e.Name)
}

// GraphqlUpdateMutationName returns the name of the mutation that
// updates instances of the given entity
func GraphqlUpdateMutationName(e *Entity) string {
	return fmt.Sprintf("update%s", e.Name)
}

// GraphqlDeleteMutationName returns the name of the mutation that
// deletes instances of the given entity
func GraphqlDeleteMutationName(e *Entity) string {
	return fmt.Sprintf("delete%s", e.Name)
}

// GraphqlFindAllQueryName returns the name of the query
// that finds all instances of the given entity
func GraphqlFindAllQueryName(e *Entity) string {
	return fmt.Sprintf("findAll%s", e.PluralName())
}

// GraphqlFindByAttributeQueryName returns the name of the query that
// find instances of the given entity by the given attribute
func GraphqlFindByAttributeQueryName(e *Entity, a *Attribute) string {
	return fmt.Sprintf("find%sBy%s", e.Name, a.Name)
}

// GraphqlFindByRelationQueryName returns the name of the query that
// find instances of the given entity by the given attribute
func GraphqlFindByRelationQueryName(e *Entity, r *Relation) string {
	return fmt.Sprintf("find%sBy%s", e.PluralName(), r.Alias())
}

// GraphqlFinderQueryByID is a convenience function representation that
// models a lookup of an entity by its id
func GraphqlFinderQueryByID(e *Entity) *GraphqlFun {
	return GraphqlFinderQueryFromAttribute(e, &Attribute{
		Name: "ID",
		Type: "ID",
	})
}

// GraphqlFinderQueryByParent is a convenience function representation
// that models a lookup of a collection of entities by a parent
// id
func GraphqlFinderQueryByParent(e *Entity, r *Relation) *GraphqlFun {
	return GraphqlFinderQueryFromAttribute(e, &Attribute{
		Name: r.Alias(),
		Type: "ID",
	})

}

// GraphqlSchema is an internal simplified Graphql model
type GraphqlSchema struct {
	Enums     []*GraphqlEnum
	Unions    []*GraphqlUnion
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

	for _, e := range s.Enums {
		chunks = append(chunks, fmt.Sprintf("%s\n", e.String()))
	}
	chunks = append(chunks, "\n\n")
	for _, u := range s.Unions {
		chunks = append(chunks, fmt.Sprintf("%s\n", u.String()))
	}
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
		args = append(args, fmt.Sprintf("%s:%s", strcase.ToLowerCamel(a.Name), a.DataTypeString()))
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

// GraphqlUnion represents a Graphql Union type.
type GraphqlUnion struct {
	Name   string
	Values []string
}

// Generates the text representation of the Union type
func (t *GraphqlUnion) String() string {
	return fmt.Sprintf("union %s = %s", t.Name, strings.Join(t.Values, " | "))
}

// GraphqlUnionFromUDType converts the given user defined type into a
// Graphql union type
func GraphqlUnionFromUDType(t *UDType) *GraphqlUnion {
	return &GraphqlUnion{
		Name:   t.Name,
		Values: t.Values,
	}
}

// GraphqlEnum represents a Graphql Enum.
type GraphqlEnum struct {
	Name   string
	Values []string
}

func (t *GraphqlEnum) String() string {
	chunks := []string{}
	chunks = append(chunks, fmt.Sprintf("enum %s {\n", t.Name))
	for _, v := range t.Values {
		chunks = append(chunks, fmt.Sprintf("  %s\n", v))
	}
	chunks = append(chunks, "}\n")
	return strings.Join(chunks, "")
}

// GraphqlEnumFromUDType converts the given user defined type into a
// Graphql enum
func GraphqlEnumFromUDType(t *UDType) *GraphqlEnum {
	return &GraphqlEnum{
		Name:   t.Name,
		Values: t.Values,
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
	switch a.Type {
	case "ID":
		return "ID"
	default:

		return a.Type
	}

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

// GraphqlInputFieldFromRelation converts a model relation into a more
// convenient Graphql Field. This function deals with input types, where
// the datatype is a graphql ID, not an entity.
func GraphqlInputFieldFromRelation(r *Relation) *GraphqlField {
	f := GraphqlFieldFromRelation(r)
	f.DataType = "ID"
	return f
}

// RelationGraphqlFieldName returns the Graphql field name for the
// given relation
func RelationGraphqlFieldName(r *Relation) string {
	return strcase.ToLowerCamel(r.Alias())
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
