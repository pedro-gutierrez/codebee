package main

import (
	"flag"
	"fmt"
	"log"
	"path"

	. "github.com/dave/jennifer/jen"
	"golang.org/x/sys/unix"
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
	db = flag.String("db", "sqlite3", "the target database type")
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

	err = CreateSql(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "sql.go"),
		Model:    model,
		Database: *db,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating sql: %v", err))
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

	err = CreateMonitoring(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "monitoring.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating monitoring: %v", err))
	}

	err = CreateMain(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "main.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating main: %v", err))
	}

	err = CreateDiagram(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "diagram.dot"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating diagram: %v", err))
	}
}

// CreateMain generates the main.go in the target directory. This will
// be the file that will glue things
// together and bootstrap the whole system.
func CreateMain(p *Package) error {
	f := NewFile(p.Name)

	AddVars(f)
	AddInit(f)
	AddMainFun(f)
	return f.Save(p.Filename)
}

// AddVars builds the variable initialization code for the main program
func AddVars(f *File) {
	f.Var().DefsFunc(func(vars *Group) {
		vars.Id("db").Op("*").String()
	})
}

// AddInit builds the init function for the main program
func AddInit(f *File) {
	f.Func().Id("init").Params().BlockFunc(func(g *Group) {

		InitFlag("db", "db", "String", "file::memory:?cache=shared", "the database connection string", g)
	})
}

// InitFlag builds the code that initializes a flag
func InitFlag(varName string, flag string, flatType string, defaultValue string, help string, g *Group) {
	g.Id(varName).Op("=").Qual("flag", flatType).Call(
		Lit(flag),
		Lit(defaultValue),
		Lit(help),
	)
}

func AddMainFun(f *File) {
	funName := "main"

	f.Func().Id(funName).Params().BlockFunc(func(g *Group) {
		g.Qual("flag", "Parse").Call()

		g.List(
			Id("db"),
			Err(),
		).Op(":=").Id("NewDb").Call(
			Op("*").Id("db"),
		)
		IfErrorLogFatal("Error opening database: %v", g)

		g.Err().Op("=").Id("ExecStatements").Call(
			Id("db"),
			Id("SqlSchema").Call(),
		)

		IfErrorLogFatal("Error initializing database: %v", g)

		g.Id("SetupServer").Call(
			Id("Schema").Call(),
			Op("&").Id("Resolver").Values(Dict{
				Id("Db"): Id("db"),
			}),
		)

		g.Qual("net/http", "ListenAndServe").Call(Lit(":8080"), Nil())
	})
}

// IfErrorLogFatal is a helper function that checks for an error and
// exits with a logged message
func IfErrorLogFatal(msg string, g *Group) {
	g.If(Err().Op("!=").Nil()).Block(
		Qual("log", "Fatal").Call(
			Qual("fmt", "Sprintf").Call(
				Lit(msg),
				Err(),
			),
		),
	)
}
