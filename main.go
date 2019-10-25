package main

import (
	"flag"
	"fmt"
	. "github.com/dave/jennifer/jen"
	"golang.org/x/sys/unix"
	"log"
	"path"
)

var (
	model   *string
	output  *string
	db      *string
	metrics *bool
)

func init() {
	model = flag.String("model", "", "the input model, in yaml format")
	output = flag.String("output", "", "the output folder")
	db = flag.String("db", "sqlite3", "the target database")
	metrics = flag.Bool("metrics", false, "add Prometheus instrumentation")
}

func main() {
	flag.Parse()

	if *output == "" {
		log.Fatal("Please specific an output directory")
	}

	if unix.Access(*output, unix.W_OK) != nil {
		log.Fatal(fmt.Sprintf("Path %v is not writable", *output))
	}

	model, err := ReadModelFromFile(*model)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error reading model from yaml: %v", err))
	}

	packageName := "main"

	err = CreateModel(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "model.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating model: %v", err))
	}

	err = CreateRepo(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "repo.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating repo: %v", err))
	}

	err = CreateSchema(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "schema.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating schema: %v", err))
	}

	err = CreateResolver(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "resolver.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating resolver: %v", err))
	}

	err = CreateServer(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "server.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating server: %v", err))
	}

	err = CreateMain(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "main.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating main: %v", err))
	}
}

// CreateMain generates the main.go in the target directory. This will
// be the file that will glue things
// together and bootstrap the whole system.
func CreateMain(p *Package) error {
	f := NewFile(p.Name)

	AddMainFun(f)
	return f.Save(p.Filename)
}

func AddMainFun(f *File) {
	funName := "main"

	f.Func().Id(funName).Params().BlockFunc(func(g *Group) {

		g.Id("schema").Op(":=").Id("Schema").Call()
		g.Id("resolver").Op(":=").Op("&").Id("Resolver").Values(Dict{})

		g.Id("SetupServer").Call(
			Id("schema"),
			Id("resolver"),
		)

		g.Qual("net/http", "ListenAndServe").Call(Lit(":8080"), Nil())
	})
}
