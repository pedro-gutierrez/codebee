// This file contains all the generator functions that we need in order
// to build the Flootic repository layer.
package generator

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
)

// CreateRepo generates a Golang file that contains all the necessary
// source code required to persist the given set of entities to the
// database
func CreateRepo(p *Package) error {
	f := NewFile(p.Name)

	f.PackageComment(fmt.Sprintf("%s contains all the library code for the Flootic platform", p.Name))
	f.PackageComment("This file contains all the functions that implement the Database repository")
	f.PackageComment(" ** THIS CODE IS MACHINE GENERATED. DO NOT EDIT MANUALLY ** ")

	for _, e := range p.Entities {
		CreateEntityInDbFun(e, f)
	}

	return f.Save(p.Filename)
}

// AddRepoFun generates all the repository functions for the Flootic
// mode, and adds them to the given file
func AddRepoFuns(entities []*Entity, f *File) {
	for _, e := range entities {

		// Add functions to insert data
		CreateEntityInDbFun(e, f)

		// TODO: add code to find, update and delete data
	}
}

// CreateEntityInDbFun produces the function that persists the given
// entity to the database.
func CreateEntityInDbFun(e *Entity, f *File) {
	funName := fmt.Sprintf("Create%sInDb", e.Name)

	f.Comment(fmt.Sprintf("%s persists an entity of type %s to the database, as well as its relations to other linked entities", funName, e.Name))
	f.Func().Id(funName).Params(Id("db").Op("*").Id("DB"), Id(e.VarName()).Op("*").Id(e.Name)).Error().BlockFunc(func(g *Group) {
		PersistEntityFunBody(g, e)
	})
}

// PersistEntity generates the function code that persists the given
// entity to the database.
//
// This function also generates the necessary
// code in order to persist relations to other entities.
//
// The code relies on Gorm models.
//
func PersistEntityFunBody(g *Group, e *Entity) {
	g.Id("trx").Op(":=").Id("db").Dot("Begin").Call()
	CreateOrRollbackExpression(g, e)

	// If this entity has relations to other entities
	for _, r := range e.Relations {

		// Persist many to one relations
		if "many-to-one" == r.Kind {
			g.Comment(fmt.Sprintf("entities of type %s need a %s. Make sure they are linked them propertly inside this transaction", e.Name, r.Entity))
			RelationStruct(g, e, r)
			CreateOrRollbackExpression(g, r)
		}
	}

	// Commit the transaction and return
	g.Return(Id("trx").Dot("Commit").Dot("Error"))
}

// RelationStruct is a helper function that generates the Golang
// struct that represents a relation to be persisted in the database.
func RelationStruct(g *Group, e *Entity, r *Relation) {
	g.Id(r.VarName()).Op(":=").Op("&").Id(r.Name).Values(Dict{
		Id(fmt.Sprintf(r.Entity)): Id(e.VarName()).Dot(r.Prop),
		Id(fmt.Sprintf(e.Name)):   Id(e.VarName()).Dot("ID"),
	})
}

// CreateOrRollbackExpression is a helper function that generates
// the necessary code to trigger a database create interaction, check
// for errors, and rollback the transaction if necessary.
//
// This function can be used on anything that can be named.
func CreateOrRollbackExpression(g *Group, n Named) {
	g.If(Err().Op(":=").Id("trx").Dot("Create").Call(Id(n.VarName())).Dot("Error"),
		Err().Op("!=").Nil(),
	).Block(
		Id("trx").Dot("Rollback").Call(),
		Return(Err()),
	)
}
