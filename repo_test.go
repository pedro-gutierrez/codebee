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
		t.FailNow()
	}

	// Generate model code
	err = CreateModel(&Package{
		Name:     "flootic",
		Filename: "./generated/model.go",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating model: %v", err)
		t.FailNow()
	}

	// Generate repository code
	err = CreateRepo(&Package{
		Name:     "flootic",
		Filename: "./generated/repo.go",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating repo: %v", err)
		t.FailNow()
	}

	// Generate GraphQL schema
	err = CreateGraphqlSchema(&Package{
		Name:     "flootic",
		Filename: "./generated/flootic.graphql",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating GraphQL schema: %v", err)
		t.FailNow()
	}

	// Generate SQL schema
	err = CreateSQLSchema(&Package{
		Name:     "flootic",
		Filename: "./generated/flootic.sql",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating SQL schema: %v", err)
		t.FailNow()
	}

	//dbString := "./flootic.db"
	dbString := "file::memory:?cache=shared"

	db, err := sql.Open("sqlite3", dbString)
	if err != nil {
		t.Errorf("Unable to open database: %v", err)
		t.FailNow()
	}

	log.Printf(fmt.Sprintf("%v", db))

	err = flootic.RunDBScript(db, "./generated/flootic.sql")
	if err != nil {
		t.Errorf("Unable to iniatialize database: %v", err)
		t.FailNow()
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
		t.FailNow()
	}

	err = flootic.InsertUser(db, user)
	if err != nil {
		t.Errorf("Error inserting user: %v", err)
		t.FailNow()
	}

	userFromDb, err := flootic.FindUserByID(db, "1")
	if err != nil {
		t.Errorf("Error in user lookup by ID: %v", err)
		t.FailNow()
	}

	if userFromDb.Email != user.Email {
		t.Errorf("Email mismatch %v (memory) vs %v (db)", user.Email, userFromDb.Email)
		t.FailNow()
	}

	userFromDb, err = flootic.FindUserByEmail(db, "admin@flootic.com")
	if err != nil {
		t.Errorf("Error in user lookup by ID: %v", err)
		t.FailNow()
	}

	if userFromDb.Email != user.Email {
		t.Errorf("Email mismatch %v (memory) vs %v (db)", user.Email, userFromDb.Email)
		t.FailNow()
	}

	userFromDb, err = flootic.FindUserByID(db, "2")
	expectedError := "sql: no rows in result set"
	if err == nil || err.Error() != expectedError {
		t.Errorf("Expected error %v, got: %v", expectedError, err)
		t.FailNow()
	}

}
