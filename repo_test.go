package generator

import "testing"

func TestGenerateRepo(t *testing.T) {

	// This is a test model. The goal is to parse the GraphQL schema,
	// and translate it into this abstract model.
	model, err := ReadModelFromFile("flootic.yml")
	if err != nil {
		t.Errorf("Error reading model from yaml: %v", err)
	}

	// Generate repository code
	CreateRepo(&Package{
		Name:     "flootic",
		Filename: "./generated/repo.go",
		Model:    model,
	})

	// Generate GraphQL schema
	err = CreateGraphqlSchema(&Package{
		Name:     "flootic",
		Filename: "./generated/flootic.graphql",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating schema: %v", err)
	}
}
