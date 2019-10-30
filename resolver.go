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
		AddUpdateMutationResolverFun(e, f)
		AddDeleteMutationResolverFun(e, f)

		for _, a := range e.Attributes {
			if a.HasModifier("indexed") && a.HasModifier("unique") {
				AddFinderByAttributeQueryResolverFun(e, a, f)
			}
		}

		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
				AddFinderByRelationQueryResolverFun(e, r, f)
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
	).Add(returnType).BlockFunc(func(g *Group) {
		value := Id("r").Dot("Data").Dot(a.Name)
		g.Return(value)
	})
}

// AddRelationResolver builds a resolver function for the given entity
// and relation
func AddRelationResolver(e *Entity, r *Relation, f *File) {

	fun := GraphqlFinderQueryByID(&Entity{Name: r.Entity})
	res := GraphqlResolverResult(fun)
	resolver := GraphqlResolverForEntity(e)
	returnType := GraphqlResolverDataTypeFromRelation(r)

	f.Func().Parens(Id("r").Op("*").Id(resolver)).Id(strings.Title(r.Name())).Params(
		Id("ctx").Qual("context", "Context"),
	).Parens(List(
		returnType,
		Error(),
	)).BlockFunc(func(g *Group) {
		g.List(
			Id(VarName(r.Entity)),
			Err(),
		).Op(":=").Id(fmt.Sprintf("Find%sByID", r.Entity)).Call(
			Id("r").Dot("Db"),
			Id("r").Dot("Data").Dot(r.Name()).Dot("ID"),
		)

		MaybeReturnWrappedError(fmt.Sprintf("Error resolving %s as %s of %s", r.Entity, r.Name(), e.Name), g)

		g.Return(
			Op("&").Add(Id(res)).Values(Dict{
				Id("Db"):   Id("r").Dot("Db"),
				Id("Data"): Id(VarName(r.Entity)),
			}),
			Nil(),
		)

	})
}

// AddCreateResolverFun defines a create resolver function for the given
// entity
func AddCreateMutationResolverFun(e *Entity, f *File) {
	fun := GraphqlCreateMutationFromEntity(e)
	res := GraphqlResolverResult(fun)
	ResolverFun(fun, func(g *Group) {
		g.List(
			Id(e.VarName()),
			Err(),
		).Op(":=").Id(fmt.Sprintf("Insert%s", e.Name)).Call(
			Id("r").Dot("Db"),
			Op("&").Id(e.Name).Values(DictFunc(func(d Dict) {
				// build a input for the entity, taking values
				// from the resolver args
				for _, a := range e.Attributes {
					value := Id("args").Dot(strings.Title(AttributeGraphqlFieldName(a)))
					d[Id(a.Name)] = value
				}

				for _, r := range e.Relations {
					if r.HasModifier("hasOne") || r.HasModifier("belongsTo") {
						d[Id(r.Name())] = Op("&").Id(r.Entity).Values(Dict{
							Id("ID"): Id("args").Dot(strings.Title(r.Name())),
						})
					}
				}
			})),
		)

		MaybeReturnWrappedError(fmt.Sprintf("Error inserting %s", e.Name), g)
		g.Return(
			Op("&").Add(Id(res)).Values(Dict{
				Id("Db"):   Id("r").Dot("Db"),
				Id("Data"): Id(e.VarName()),
			}),
			Nil(),
		)
	}, f)
}

// AddUpdateResolverFun defines an update resolver function for the given
// entity
func AddUpdateMutationResolverFun(e *Entity, f *File) {
	fun := GraphqlUpdateMutationFromEntity(e)
	res := GraphqlResolverResult(fun)
	ResolverFun(fun, func(g *Group) {
		g.List(
			Id(e.VarName()),
			Err(),
		).Op(":=").Id(fmt.Sprintf("Update%s", e.Name)).Call(
			Id("r").Dot("Db"),
			Op("&").Id(e.Name).Values(DictFunc(func(d Dict) {
				// build a input for the entity, taking values
				// from the resolver args
				for _, a := range e.Attributes {
					value := Id("args").Dot(strings.Title(AttributeGraphqlFieldName(a)))
					d[Id(a.Name)] = value
				}

				for _, r := range e.Relations {
					if r.HasModifier("hasOne") || r.HasModifier("belongsTo") {
						d[Id(r.Name())] = Op("&").Id(r.Entity).Values(Dict{
							Id("ID"): Id("args").Dot(strings.Title(r.Name())),
						})
					}
				}
			})),
		)

		MaybeReturnWrappedError(fmt.Sprintf("Error updating %s", e.Name), g)
		g.Return(
			Op("&").Add(Id(res)).Values(Dict{
				Id("Db"):   Id("r").Dot("Db"),
				Id("Data"): Id(e.VarName()),
			}),
			Nil(),
		)
	}, f)
}

// AddDeleteMutationResolverFun defines a delete resolver function for the given
// entity
func AddDeleteMutationResolverFun(e *Entity, f *File) {
	fun := GraphqlDeleteMutationFromEntity(e)
	res := GraphqlResolverResult(fun)
	ResolverFun(fun, func(g *Group) {
		g.List(
			Id(e.VarName()),
			Err(),
		).Op(":=").Id(fmt.Sprintf("Delete%s", e.Name)).Call(
			Id("r").Dot("Db"),
			Id("args").Dot("Id"),
		)

		MaybeReturnWrappedError(fmt.Sprintf("Error deleting %s", e.Name), g)
		g.Return(
			Op("&").Add(Id(res)).Values(Dict{
				Id("Db"):   Id("r").Dot("Db"),
				Id("Data"): Id(e.VarName()),
			}),
			Nil(),
		)
	}, f)

}

// AddFinderByAttributeQueryResolverFun defines a resolver function for the given
// indexed and
// unique attribute of the given entity
func AddFinderByAttributeQueryResolverFun(e *Entity, a *Attribute, f *File) {
	fun := GraphqlFinderQueryFromAttribute(e, a)
	res := GraphqlResolverResult(fun)

	ResolverFun(fun, func(g *Group) {
		value := Id("args").Dot(strings.Title(AttributeGraphqlFieldName(a)))
		g.List(
			Id(e.VarName()),
			Err(),
		).Op(":=").Id(fmt.Sprintf("Find%sBy%s", e.Name, a.Name)).Call(
			Id("r").Dot("Db"),
			value,
		)

		MaybeReturnWrappedError(fmt.Sprintf("Error resolving %s by %s", e.Name, a.Name), g)
		g.Return(
			Op("&").Add(Id(res)).Values(Dict{
				Id("Db"):   Id("r").Dot("Db"),
				Id("Data"): Id(e.VarName()),
			}),
			Nil(),
		)
	}, f)
}

// AddFinderByRelationQueryResolverFun defines a resolver function for the given
// relation.
func AddFinderByRelationQueryResolverFun(e *Entity, r *Relation, f *File) {
	fun := GraphqlFinderQueryFromRelation(e, r)
	res := GraphqlResolverResult(fun)

	ResolverFun(fun, func(g *Group) {
		g.List(
			Id(VarName(e.Plural())),
			Err(),
		).Op(":=").Id(fmt.Sprintf("Find%sBy%s", e.Plural(), r.Name())).Call(
			Id("r").Dot("Db"),
			Id("args"),
		)

		MaybeReturnWrappedError(fmt.Sprintf("Error resolving %s by %s", e.Plural(), r.Name()), g)

		g.Id("resolvers").Op(":=").Id(res).Values(Dict{})

		g.For(
			List(
				Id("_"),
				Id(e.VarName()),
			).Op(":=").Range().Id(VarName(e.Plural())),
		).BlockFunc(func(g2 *Group) {

			g2.Id("resolvers").Op("=").Append(
				Id("resolvers"),
				Op("&").Id(GraphqlResolverForEntity(e)).Values(Dict{
					Id("Db"):   Id("r").Dot("Db"),
					Id("Data"): Id(e.VarName()),
				}),
			)
		})

		g.Return(
			Op("&").Id("resolvers"),
			Nil(),
		)
	}, f)
}

func ResolverFun(fun *GraphqlFun, blockFun func(*Group), f *File) {
	res := GraphqlResolverResult(fun)
	f.Func().Parens(Id("r").Op("*").Id("Resolver")).Id(strings.Title(fun.Name)).Params(
		Id("ctx").Qual("context", "Context"),
		GraphqlResolverArgs(fun),
	).Parens(List(
		Op("*").Id(res),
		Error(),
	)).BlockFunc(blockFun)
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
	res := GraphqlResolverForType(f.Returns.DataType)
	if f.Returns.Many {
		res = fmt.Sprintf("[]*%s", res)
	}
	return res
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
		return String()

	case "String":
		return String()

	case "Int":
		return Int32()

	case "Boolean":
		return Bool()
	default:
		return String()

	}

}

// MaybeReturnWrappedError produces the code that returns immediately
// and wraps the error with a message
func MaybeReturnWrappedError(msg string, g *Group) {
	g.If(
		Err().Op("!=").Nil(),
	).Block(
		Return(
			Nil(),
			Qual("github.com/pkg/errors", "Wrap").Call(
				Err(),
				Lit(msg),
			),
		),
	)
}
