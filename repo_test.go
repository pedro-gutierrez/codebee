package generator

import "testing"

func TestGenerateRepo(t *testing.T) {

	// This is a test model. The goal is to parse the GraphQL schema,
	// and translate it into this abstract model.
	entities := []*Entity{
		&Entity{Name: "Organization"},
		&Entity{
			Name: "User",
			Relations: []*Relation{
				&Relation{
					Name:   "UserOrganization",
					Entity: "Organization",
					Prop:   "Org",
					Kind:   "many-to-one",
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
