package generator

import (
	"database/sql"
	"fmt"
	"github.com/flootic/generator/generated"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"testing"
)

func TestGenerateRepo(t *testing.T) {

	// This is a test model. The goal is to parse the GraphQL schema,
	// and translate it into this abstract model.
	model, err := ReadModelFromFile("flootic.yml")
	if err != nil {
		t.Errorf("Error reading model from yaml: %v", err)
	}

	// Generate model code
	err = CreateModel(&Package{
		Name:     "flootic",
		Filename: "./generated/model.go",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating model: %v", err)
	}

	// Generate repository code
	err = CreateRepo(&Package{
		Name:     "flootic",
		Filename: "./generated/repo.go",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating repo: %v", err)
	}

	// Generate GraphQL schema
	err = CreateGraphqlSchema(&Package{
		Name:     "flootic",
		Filename: "./generated/flootic.graphql",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating GraphQL schema: %v", err)
	}

	// Generate SQL schema
	err = CreateSQLSchema(&Package{
		Name:     "flootic",
		Filename: "./generated/flootic.sql",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating SQL schema: %v", err)
	}

	// Start an empty database. For our tests, we use in memory Sqlite
	// so we know the database is totally empty and we need to build it
	// from scratch
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		t.Errorf("Unable to open database: %v", err)
	}

	log.Printf(fmt.Sprintf("%v", db))

	err = flootic.RunDBScript(db, "./generated/flootic.sql")
	if err != nil {
		t.Errorf("Unable to iniatialize database: %v", err)
	}

	// Test the generated model and repo
	org := &flootic.Organization{
		ID:   "1",
		Name: "flootic",
	}

	user := &flootic.User{
		Email: "admin@flootic.com",
		ID:    "1",
	}

	org.Owner = user
	user.Organization = org

	err = flootic.InsertOrganization(db, org)
	if err != nil {
		t.Errorf("Error inserting organization: %v", err)
	}

}
