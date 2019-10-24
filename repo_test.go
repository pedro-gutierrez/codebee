package generator

import (
	"database/sql"
	"fmt"
	"github.com/flootic/generator/generated"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
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
	err = CreateGraphql(&Package{
		Name:     "flootic",
		Filename: "./generated/graphql.go",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating GraphQL schema: %v", err)
		t.FailNow()
	}

	// Generate SQL schema
	err = CreateSQLSchema(&Package{
		Name:     "flootic",
		Filename: "./generated/schema.sql",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating SQL schema: %v", err)
		t.FailNow()
	}

	dbString := "file::memory:?cache=shared"

	db, err := sql.Open("sqlite3", dbString)
	if err != nil {
		t.Errorf("Unable to open database: %v", err)
		t.FailNow()
	}

	log.Printf(fmt.Sprintf("%v", db))

	err = flootic.RunDBScript(db, "./generated/schema.sql")
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

	_, err = flootic.InsertOrganization(db, org)
	if err != nil {
		t.Errorf("Error inserting organization: %v", err)
		t.FailNow()
	}

	_, err = flootic.InsertUser(db, user)
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

	// Generate server code
	err = CreateServer(&Package{
		Name:     "flootic",
		Filename: "./generated/server.go",
		Model:    model,
	})

	if err != nil {
		t.Errorf("Error creating server: %v", err)
		t.FailNow()
	}

	//Generate a new GraphQL schema
	schema, err := flootic.Graphql(db)
	if err != nil {
		t.Errorf("Error creating GraphQL schema: %v", err)
		t.FailNow()
	}

	log.Printf(fmt.Sprintf("%v", schema))

	// Start a new server with the database and schema
	flootic.SetupServer(schema)
	log.Fatal(http.ListenAndServe(":2999", nil))

}
