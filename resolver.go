package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"strings"
)

func CreateResolver(p *Package) error {
	f := NewFile(p.Name)
	AddResolverStruct(f)

	for _, e := range p.Model.Entities {

		AddTypeResolver(e, f)

		AddCreateMutationResolverFun(e, f)

		for _, a := range e.Attributes {
			if a.HasModifier("indexed") && a.HasModifier("unique") {
				AddFinderQueryResolverFun(e, a, f)
			}
		}
	}

	return f.Save(p.Filename)
}

func AddResolverStruct(f *File) {

	f.Type().Id("Resolver").Struct(
		Id("Db").Op("*").Qual("database/sql", "DB"),
	)
}

func AddTypeResolver(e *Entity, f *File) {

	f.Type().Id(GraphqlResolverForEntity(e)).Struct(
		Id("Db").Op("*").Qual("database/sql", "DB"),
		Id("Data").Op("*").Id(e.Name),
	)

	for _, a := range e.Attributes {
		AddAttributeResolver(e, a, f)
	}

	for _, r := range e.Relations {
		AddRelationResolver(e, r, f)
	}
}

// AddAttributeResolver builds a resolver function for the given entity
// and attribute
func AddAttributeResolver(e *Entity, a *Attribute, f *File) {
	resolver := GraphqlResolverForEntity(e)
	returnType := Add(GraphqlResolverDataTypeFromAttribute(a))

	f.Func().Parens(Id("r").Op("*").Id(resolver)).Id(strings.Title(a.Name)).Params(
		Id("ctx").Qual("context", "Context"),
	).Add(returnType).Block(
		Var().Id("res").Add(returnType),
		Return(Id("res")),
	)
}

// AddRelationResolver builds a resolver function for the given entity
// and relation
func AddRelationResolver(e *Entity, r *Relation, f *File) {

	resolver := GraphqlResolverForEntity(e)
	returnType := GraphqlResolverDataTypeFromRelation(r)

	f.Func().Parens(Id("r").Op("*").Id(resolver)).Id(strings.Title(r.Name())).Params(
		Id("ctx").Qual("context", "Context"),
	).Parens(List(
		returnType,
		Error(),
	)).Block(
		Var().Id("res").Add(returnType),
		Return(List(
			Id("res"),
			Nil(),
		)),
	)
}

// AddCreateResolverFun defines a create resolver function for the given
// entity
func AddCreateMutationResolverFun(e *Entity, f *File) {
	fun := GraphqlCreateMutationFromEntity(e)
	ResolverFun(fun, func(g Group) {

	}, f)
}

// AddFinderQueryResolverFun defines a resolver function for the given
// indexed and
// unique attribute of the given entity
func AddFinderQueryResolverFun(e *Entity, a *Attribute, f *File) {
	fun := GraphqlFinderQueryFromAttribute(e, a)
	ResolverFun(fun, func(g Group) {

	}, f)
}

func ResolverFun(fun *GraphqlFun, blockFun func(Group), f *File) {
	res := GraphqlResolverResult(fun)

	f.Func().Parens(Id("r").Op("*").Id("Resolver")).Id(strings.Title(fun.Name)).Params(
		Id("ctx").Qual("context", "Context"),
		GraphqlResolverArgs(fun),
	).Parens(List(
		Op("*").Id(res),
		Error(),
	)).Block(
		Var().Id("res").Id(res),
		Return(List(
			Op("&").Id("res"),
			Nil(),
		)),
	)
}

// GraphqlResolverForEntity builds the resolver type
// the given entity
func GraphqlResolverForEntity(e *Entity) string {
	return GraphqlResolverForType(e.Name)
}

// GraphqlResolverForRelation builds the resolver type
// the given entity
func GraphqlResolverForRelation(r *Relation) string {
	return GraphqlResolverForType(r.Entity)
}

// GraphqlResolverResult builds the resolver type of the return type of
// the given graphql function
func GraphqlResolverResult(f *GraphqlFun) string {
	return GraphqlResolverForType(f.Returns.DataType)
}

// GraphqlResolverForType returns the resolver type for the given data
// type
func GraphqlResolverForType(t string) string {
	return fmt.Sprintf("%sResolver", t)
}

// GraphqlResolverArgs maps the given graphql function arguments to its
// Golang resolver arguments equivalent struct
func GraphqlResolverArgs(fun *GraphqlFun) *Statement {
	return Id("args").StructFunc(func(g *Group) {
		for _, a := range fun.Args {
			g.Id(strings.Title(a.Name)).Add(GraphqlResolverDataTypeFromGraphqlField(a))
		}
	})
}

// GraphqlResolverDataTypeFromAttribute returns the Golang data type for
// the given entity attribute
func GraphqlResolverDataTypeFromAttribute(a *Attribute) *Statement {
	return GraphqlResolverDataTypeFromDataType(a.Type)
}

// GraphqlResolverDataTypeFromRelation returns the Golang data type for
// the given entity relation. We return a resolver type that resolves to
// the entity pointed by the relation
func GraphqlResolverDataTypeFromRelation(r *Relation) *Statement {
	return Op("*").Id(GraphqlResolverForRelation(r))
}

// GraphqlResolverDataTypeFromGraphqlField returns the Golang data type
// for the given Grapqhl type
func GraphqlResolverDataTypeFromGraphqlField(a *GraphqlField) *Statement {
	return GraphqlResolverDataTypeFromDataType(a.DataType)
}

// GraphqlResolverDataTypeFromDataType returns the Golang data type
// from the given data type string representation
func GraphqlResolverDataTypeFromDataType(d string) *Statement {
	switch d {
	case "ID":
		return Qual("github.com/graph-gophers/graphql-go", "ID")

	default:
		return String()

	}

}
