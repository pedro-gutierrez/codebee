package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

// CreateGraphql generates a Golang file that produces the Graphql
// schema.
func CreateGraphql(p *Package) error {
	f := NewFile(p.Name)

	f.PackageComment(fmt.Sprintf("%s contains all the library code for the Flootic platform", p.Name))
	f.PackageComment("This file contains all the functions that implement the Graphql schema")
	f.PackageComment(" ** THIS CODE IS MACHINE GENERATED. DO NOT EDIT MANUALLY ** ")

	AddGraphqlFun(p.Model, f)

	return f.Save(p.Filename)
}

// AddGraphqlFun generates a convenience function that accepts the
// path to a SQL file, and runs all statements in that file one by one
func AddGraphqlFun(m *Model, f *File) {
	funName := "Graphql"

	f.Comment(fmt.Sprintf("%s returns the Graphql schema", funName))
	f.Func().Id(funName).Params(
		Id("db").Op("*").Qual("database/sql", "DB"),
	).Parens(List(
		Qual("github.com/graphql-go/graphql", "Schema"),
		Error(),
	)).BlockFunc(func(g *Group) {

		GraphqlTypes(m, g)

		g.Return(Id("graphql").Dot("NewSchema").Call(
			Id("graphql").Dot("SchemaConfig").Values(Dict{
				Id("Query"):    GraphqlQuery(m),
				Id("Mutation"): GraphqlMutation(m),
			}),
		))
	})
}

// GraphqlQuery returns the code that builds the Graphql queries of
// the schema
func GraphqlQuery(m *Model) *Statement {
	return NewGraphqlObject("Queries", GraphqlQueryFields(m))
}

// GraphqlMutation returns the code that builds the Graphql mutations of
// the schema
func GraphqlMutation(m *Model) *Statement {
	return NewGraphqlObject("Mutations", GraphqlMutationFields(m))
}

// NewGraphqlObject returns the code that builds Graphql object
// (a set of queries, or mutations) with the given name and fields
func NewGraphqlObject(name string, fun func(Dict)) *Statement {
	return Id("graphql").Dot("NewObject").Call(
		Id("graphql").Dot("ObjectConfig").Values(Dict{
			Id("Name"):   Lit(name),
			Id("Fields"): Id("graphql").Dot("Fields").Values(DictFunc(fun)),
		}))
}

// GraphqlTypes produces the code that defines all the types used in
// the graphql schema
func GraphqlTypes(m *Model, g *Group) {
	for _, e := range m.Entities {
		g.Id(GraphqlTypeVarForEntity(e)).Op(":=").Add(
			NewGraphqlObject(e.Name, func(d Dict) {
				for _, a := range e.Attributes {
					d[Lit(GraphqlAttributeName(a))] = Op("&").Id("graphql").Dot("Field").Values(Dict{
						Id("Type"): GraphqlAttributeType(a),
					})
				}
			}),
		)
	}
}

// GraphqlQueryFields builds the code that define the queries for the
// graphql schema, for the model provided
func GraphqlQueryFields(m *Model) func(Dict) {
	return func(d Dict) {

		// iterate all entities, and for each entity, add finder
		// queries for each indexed attribute
		for _, e := range m.Entities {
			for _, a := range e.Attributes {
				if a.HasModifier("unique") && a.HasModifier("indexed") {
					queryName := fmt.Sprintf("find%sBy%s", e.Name, a.Name)
					d[Lit(queryName)] = GraphqlObjectField(
						GraphqlTypeVarForEntity(e),
						fmt.Sprintf("Find a single entry of type %s by %s", e.Name, a.Name),
						func(d Dict) {
							d[Lit(GraphqlAttributeName(a))] = GraphqlAttributeArgument(a)
						},
						func(g *Group) {

							GraphqlAttributeValidation(a, g)

							g.Return(Id(
								FindEntityByAttributeFunName(e, a)).Call(Id("db"), Id(a.VarName())))
						})
				}
			}
		}

	}
}

// GraphqlMutationFields builds the code that define the mutations for
// the graphql schema, for the model provided
func GraphqlMutationFields(m *Model) func(Dict) {
	return func(d Dict) {

		// iterate all entities, and for each entity, add finder
		// queries for each indexed attribute
		for _, e := range m.Entities {

			queryName := fmt.Sprintf("create%s", e.Name)
			d[Lit(queryName)] = GraphqlObjectField(
				GraphqlTypeVarForEntity(e),
				fmt.Sprintf("Create a new entry of type %s", e.Name),
				func(d Dict) {
					for _, a := range e.Attributes {
						d[Lit(GraphqlAttributeName(a))] = GraphqlAttributeArgument(a)
					}

					for _, r := range e.Relations {
						d[Lit(GraphqlRelationName(r))] = GraphqlRelationArgument(r)
					}

				},
				func(g *Group) {
					// add code to validate fields
					for _, a := range e.Attributes {
						GraphqlAttributeValidation(a, g)
					}

					for _, r := range e.Relations {
						GraphqlRelationValidation(r, g)
					}

					g.Return(Id(InsertEntityFunName(e)).Call(
						Id("db"),
						EntityInitialization(e),
					))
				})
		}
	}
}

// GraphqlObjectField generates all the code required to represent a
// graphql object field. A field can be part of a query, or a mutation,
// they are structurally identical. This function offers some
// convenience, so that we don´t have to repeat the same AST over and
// over
func GraphqlObjectField(graphqlType string, desc string, args func(Dict), resolver func(g *Group)) *Statement {
	return Op("&").Id("graphql").Dot("Field").Values(Dict{
		Id("Type"):        Id(graphqlType),
		Id("Description"): Lit(desc),
		Id("Args"):        Id("graphql").Dot("FieldConfigArgument").Values(DictFunc(args)),
		Id("Resolve"): Func().Params(
			Id("p").Id("graphql").Dot("ResolveParams"),
		).Parens(List(Interface(), Error())).BlockFunc(resolver),
	})
}

// GraphqlTypeVarForEntity returns the name of the variable that holds
// the Graphql type definition for the given entity
func GraphqlTypeVarForEntity(e *Entity) string {
	return strcase.ToLowerCamel(fmt.Sprintf("%sType", e.Name))
}

// GraphqlAttributeName returns the name of the attribute within the
// entity's type definition
func GraphqlAttributeName(a *Attribute) string {
	switch a.Name {
	case "ID":
		return "id"

	default:
		return strcase.ToLowerCamel(a.Name)
	}

}

// GraphqlAttributeArgument builds the code that defines a Graphql
// attribute for the given model attribute
func GraphqlAttributeArgument(a *Attribute) *Statement {
	return Op("&").Id("graphql").Dot("ArgumentConfig").Values(Dict{
		Id("Type"): GraphqlAttributeType(a),
	})
}

// GraphqlAttributeType returns the code that defines the Graphql type
// for the given attribute. This function takes into consideration
// whether or not the attribute might be nullable or not
func GraphqlAttributeType(a *Attribute) *Statement {
	return Id("graphql").Dot("NewNonNull").Call(Id("graphql").Dot(GraphqlAttributeTypeLiteral(a)))
}

// GraphqlAttributeTypeLiteral translate the attribute type into the
// matching Graphql type literal
func GraphqlAttributeTypeLiteral(a *Attribute) string {
	return a.Type
}

// GraphqlRelationName returns the name of the relation within the
// entity's type definition
func GraphqlRelationName(r *Relation) string {
	return strcase.ToLowerCamel(r.Name())
}

// GraphqlRelationArgument builds the code that defines a Graphql
// attribute for the given model relation
func GraphqlRelationArgument(r *Relation) *Statement {
	return Op("&").Id("graphql").Dot("ArgumentConfig").Values(Dict{
		Id("Type"): GraphqlRelationType(r),
	})
}

// GraphqlRelationType returns the code that defines the Graphql type
// for the given relation. This function takes into consideration
// whether or not the attribute might be nullable or not
func GraphqlRelationType(r *Relation) *Statement {
	return Id("graphql").Dot("NewNonNull").Call(Id("graphql").Dot("ID"))
}

// GraphqlAttributeValidation adds the necessary code to validate a
// query or mutation argument for the given attribute.
func GraphqlAttributeValidation(a *Attribute, g *Group) {

	g.List(Id(a.VarName()), Id("ok")).Op(":=").Id("p").Dot("Args").Index(Lit(GraphqlAttributeName(a))).Assert(TypeFromAttribute(a))
	g.If(Op("!").Id("ok")).Block(
		Return(Nil(), Qual("errors", "New").Call(Lit(fmt.Sprintf("Invalid or missing %s argument", GraphqlAttributeName(a))))),
	)

}

// GraphqlRelationValidation adds the necessary code to validate a
// query or mutation argument for the given relation.
func GraphqlRelationValidation(r *Relation, g *Group) {

	g.List(Id(r.VarName()), Id("ok")).Op(":=").Id("p").Dot("Args").Index(Lit(GraphqlRelationName(r))).Assert(TypeFromRelation(r))

	g.If(Op("!").Id("ok")).Block(
		Return(Nil(), Qual("errors", "New").Call(Lit(fmt.Sprintf("Invalid or missing %s argument", GraphqlRelationName(r))))),
	)
}
