package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
)

func CreateServer(p *Package) error {
	f := NewFile(p.Name)
	AddSetupServerFun(f)
	AddHtmlHandlerFun(f)
	AddHtml(f)

	return f.Save(p.Filename)
}

func AddSetupServerFun(f *File) {

	funName := "SetupServer"

	f.Comment(fmt.Sprintf("%s glues the given schema string, and the given resolver interface and configures the http handlers that serve both the GraphQL api and the GraphQi UI", funName))
	f.Func().Id(funName).Params(
		Id("s").Id("string"),
		Id("r").Id("interface").Values(Dict{}),
	).BlockFunc(func(g *Group) {

		g.Id("schema").Op(":=").Qual("github.com/graph-gophers/graphql-go", "MustParseSchema").Call(
			Id("s"),
			Id("r"),
		)

		g.Qual("net/http", "Handle").Call(
			Lit("/graphql"),
			Op("&").Qual("github.com/graph-gophers/graphql-go/relay", "Handler").Values(Dict{
				Id("Schema"): Id("schema"),
			}))

		g.Qual("net/http", "Handle").Call(
			Lit("/"),
			Qual("net/http", "HandlerFunc").Call(Id("htmlHandlerFun")),
		)

		g.Qual("net/http", "Handle").Call(
			Lit("/metrics"),
			g.Qual("github.com/prometheus/client_golang/prometheus/promhttp", "Handler").Call(),
		)

	})
}

func AddHtmlHandlerFun(f *File) {
	f.Var().Id("htmlHandlerFun").Op("=").Func().Params(
		Id("w").Id("http").Dot("ResponseWriter"),
		Id("r").Op("*").Id("http").Dot("Request"),
	).Block(
		Id("w").Dot("Write").Call(Id("html")),
	)
}

//func AddSetupServerFun(f *File) {
//	funName := "SetupServer"
//
//	f.Comment(fmt.Sprintf("%s configures the http server", funName))
//	f.Func().Id(funName).Params(
//		Id("schema").Qual("github.com/graphql-go/graphql", "Schema"),
//	).BlockFunc(func(g *Group) {
//
//		g.Id("h").Op(":=").Qual(
//			"github.com/graphql-go/handler",
//			"New",
//		).Call(Op("&").Id("handler").Dot("Config").Values(Dict{
//			Id("Schema"):     Op("&").Id("schema"),
//			Id("Pretty"):     Id("true"),
//			Id("GraphiQL"):   Id("false"),
//			Id("Playground"): Id("true"),
//		}))
//
//		g.Qual("net/http", "Handle").Call(Lit("/"), Id("h"))
//	})
//}

func AddHtml(f *File) {
	f.Var().Id("html").Op("=").Op("[]").Byte().Call(Lit(`<!DOCTYPE html><html><head><link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.css" /><script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script><script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script><script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script><script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.10.2/graphiql.js"></script></head><body style="width: 100%; height: 100%; margin: 0; overflow: hidden;"><div id="graphiql" style="height: 100vh;">Loading...</div><script>function graphQLFetcher(graphQLParams) {return fetch("/graphql", {method: "post",body: JSON.stringify(graphQLParams),credentials: "include",}).then(function (response) {return response.text();}).then(function (responseBody) {try {return JSON.parse(responseBody);} catch (error) {return responseBody;}});}ReactDOM.render(React.createElement(GraphiQL, {fetcher: graphQLFetcher}),document.getElementById("graphiql"));</script></body></html>`))

}
