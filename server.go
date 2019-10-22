// This file contains all the generator functions that we need in order
// to build the Flootic GraphQL server
package generator

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
)

func CreateServer(p *Package) error {
	f := NewFile(p.Name)

	f.PackageComment(fmt.Sprintf("%s contains all the library code for the Flootic platform", p.Name))
	f.PackageComment("This file contains all the functions that implement the GraphQL server for the persistent layer")
	f.PackageComment(" ** THIS CODE IS MACHINE GENERATED. DO NOT EDIT MANUALLY ** ")

	AddSetupServerFun(f)
	return f.Save(p.Filename)
}

func AddSetupServerFun(f *File) {
	funName := "SetupServer"

	f.Comment(fmt.Sprintf("%s configures the http server", funName))
	f.Func().Id(funName).Params(
		Id("db").Op("*").Qual("database/sql", "DB"),
		Id("schema").Qual("github.com/graphql-go/graphql", "Schema"),
	).BlockFunc(func(g *Group) {

		g.Id("h").Op(":=").Qual(
			"github.com/graphql-go/handler",
			"New",
		).Call(Op("&").Id("handler").Dot("Config").Values(Dict{
			Id("Schema"):     Op("&").Id("schema"),
			Id("Pretty"):     Id("true"),
			Id("GraphiQL"):   Id("false"),
			Id("Playground"): Id("true"),
		}))

		g.Id("static").Op(":=").Qual("net/http", "FileServer").Call(Id("http").Dot("Dir").Call(Lit("static")))

		g.Id("http").Dot("Handle").Call(Lit("/"), Id("static"))
		g.Id("http").Dot("Handle").Call(Lit("/graphql"), Id("h"))
	})
}
