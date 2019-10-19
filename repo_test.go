package generator

import "testing"

func TestGenerateRepo(t *testing.T) {

	// This is a test model. The goal is to parse the GraphQL schema,
	// and translate it into this abstract model.
	entities := []*Entity{
		&Entity{
			Name: "Organization",
			Relations: []*Relation{
				&Relation{
					Alias:  "Owner",
					Entity: "User",
					Kind:   "has-one",
				},
			},
		},

		&Entity{
			Name: "User",
			Relations: []*Relation{
				&Relation{
					Entity: "Organization",
					Kind:   "has-one",
				},
			},
		},
		&Entity{
			Name: "Poi",
			Relations: []*Relation{
				&Relation{
					Alias:  "Owner",
					Entity: "User",
					Kind:   "has-one",
				},
			},
		},
	}

	// Take that model, and put it inside a package
	p := &Package{
		Name:     "flootic",
		Filename: "./generated/repo.go",
		Entities: entities,
	}

	// Generate repository code
	CreateRepo(p)
}
