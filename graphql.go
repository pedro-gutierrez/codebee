// This file contains all the generator functions that we need in order
// to build the Flootic GraphQL schema.
package generator

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
)

// CreateGraphql generates a Golang file that produces the Graphql
// schema.
func CreateGraphql(p *Package) error {
	f := NewFile(p.Name)

	f.PackageComment(fmt.Sprintf("%s contains all the library code for the Flootic platform", p.Name))
	f.PackageComment("This file contains all the functions that implement the Graphql schema")
	f.PackageComment(" ** THIS CODE IS MACHINE GENERATED. DO NOT EDIT MANUALLY ** ")

	AddGraphqlFun(f)

	return f.Save(p.Filename)
}

// AddGraphqlFun generates a convenience function that accepts the
// path to a SQL file, and runs all statements in that file one by one
func AddGraphqlFun(f *File) {
	funName := "Graphql"

	f.Comment(fmt.Sprintf("%s returns the Graphql schema", funName))
	f.Func().Id(funName).Params().Parens(List(
		Qual("github.com/graphql-go/graphql", "Schema"),
		Error(),
	)).BlockFunc(func(g *Group) {
		g.Var().Id("schema").Id("graphql").Dot("Schema")
		g.Return(List(Id("schema"), Nil()))
	})
}
