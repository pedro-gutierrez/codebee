package main

import (
	"flag"
	"fmt"
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

	err = CreateGraphql(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "graphql.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating graphql schema: %v", err))
	}

	err = CreateServer(&Package{
		Name:     packageName,
		Filename: path.Join(*output, "server.go"),
		Model:    model,
	})

	if err != nil {
		log.Fatal(fmt.Sprintf("Error generating server: %v", err))
	}
}
