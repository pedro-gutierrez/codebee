// This file contains all the generator functions that we need in order
// to build the Flootic repository layer.
package generator

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
	"strings"
)

// CreateRepo generates a Golang file that contains all the necessary
// source code required to persist the given set of entities to the
// database
func CreateRepo(p *Package) error {
	f := NewFile(p.Name)

	f.PackageComment(fmt.Sprintf("%s contains all the library code for the Flootic platform", p.Name))
	f.PackageComment("This file contains all the functions that implement the Database repository")
	f.PackageComment(" ** THIS CODE IS MACHINE GENERATED. DO NOT EDIT MANUALLY ** ")

	f.ImportAlias("database/sql", "sql")
	f.ImportAlias("io/ioutil", "ioutil")

	AddRunDBScriptFun(f)

	AddRepoFuns(p.Model.Entities, f)

	return f.Save(p.Filename)
}

// AddRunScriptFun generates a convenience function that accepts the
// path to a SQL file, and runs all statements in that file one by one
func AddRunDBScriptFun(f *File) {
	funName := "RunDBScript"

	f.Comment(fmt.Sprintf("%s runs all SQL statements found in the given file", funName))
	f.Func().Id(funName).Params(Id("db").Op("*").Qual("database/sql", "DB"), Id("path").Id("string")).Error().BlockFunc(func(g *Group) {
		ReadFile(g)
		IfErrorReturn(g)
		g.Id("stmts").Op(":=").Qual("strings", "Split").Call(Id("string").Call(Id("file")), Lit(";"))
		g.For(List(Id("_"), Id("stmt")).Op(":=").Id("range").Id("stmts")).BlockFunc(func(g2 *Group) {
			g2.List(Id("_"), Err()).Op("=").Id("db").Dot("Exec").Call(Id("stmt"))
			IfErrorReturn(g2)
		})

		g.Return(Nil())
	})
}

// ReadFile returns the code required to read a file
func ReadFile(g *Group) {
	g.List(Id("file"), Err()).Op(":=").Qual("io/ioutil", "ReadFile").Call(Id("path"))
}

// AddRepoFun generates all the repository functions for the Flootic
// mode, and adds them to the given file
func AddRepoFuns(entities []*Entity, f *File) {
	for _, e := range entities {

		// Add functions to insert data
		AddInsertFun(e, f)

		// TODO: add code to find, update and delete data
	}
}

// AddInsertFun produces the function that inserts the given
// entity to the database.
func AddInsertFun(e *Entity, f *File) {
	funName := fmt.Sprintf("Insert%s", e.Name)

	f.Comment(fmt.Sprintf("%s inserts an entity of type %s to the database", funName, e.Name))
	f.Comment("This function also persists its relations to other linked entities")
	f.Func().Id(funName).Params(Id("db").Op("*").Qual("database/sql", "DB"), Id(e.VarName()).Op("*").Id(e.Name)).Error().BlockFunc(func(g *Group) {

		// Open a transaction
		BeginTransaction(g)
		IfErrorReturn(g)

		DeferRollbackTransaction(g)
		PrepareStatement(InsertStatement(e), g)
		IfErrorReturn(g)

		DeferCloseStatement(g)

		ExecuteStatement(g, func(g2 *Group) {
			InsertStatementValues(e, g2)
		})
		IfErrorReturn(g)

		ReturnCommitTransaction(g)
	})
}

// TableName builds a SQL table name, for the given entity
func TableName(e *Entity) string {
	return strings.ToLower(strcase.ToSnake(fmt.Sprintf("%s%s%s", "Fl", e.Name, "s")))
}

// ColumnName
func AttributeColumnName(a *Attribute) string {
	return strings.ToLower(strcase.ToSnake(a.Name))
}

func RelationColumnName(r *Relation) string {
	return strings.ToLower(strcase.ToSnake(r.Name()))
}

// Quoted
func SingleQuoted(str string) string {
	return fmt.Sprintf("'%v'", str)
}

// InsertStatement generates a sql INSERT statement for the given entity
func InsertStatement(e *Entity) string {
	chunks := []string{}
	chunks = append(chunks, "INSERT INTO")
	chunks = append(chunks, TableName(e))

	columns := []string{}
	for _, a := range e.Attributes {
		columns = append(columns, SingleQuoted(AttributeColumnName(a)))
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			columns = append(columns, SingleQuoted(RelationColumnName(r)))
		}
	}

	chunks = append(chunks, fmt.Sprintf("(%s)", strings.Join(columns, ",")))
	chunks = append(chunks, "VALUES")

	placeholders := []string{}
	for _, _ = range e.Attributes {
		placeholders = append(placeholders, "?")
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			placeholders = append(placeholders, "?")
		}
	}

	chunks = append(chunks, fmt.Sprintf("(%s)", strings.Join(placeholders, ",")))
	return strings.Join(chunks, " ")
}

// InsertStatementValues generates the Golang code that populates the
// values to be sent to the INSERT sql statement for the entity
func InsertStatementValues(e *Entity, g *Group) {
	for _, a := range e.Attributes {
		g.Id(e.VarName()).Dot(a.Name)
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			g.Id(e.VarName()).Dot(r.Name()).Dot("ID")
		}
	}
}

// BeginTransaction is a helper function that generates the code needed
// to start a new transaction
func BeginTransaction(g *Group) {
	g.List(Id("tx"), Err()).Op(":=").Id("db").Dot("Begin").Call()
}

// PrepareStatement produces the code required to create a new statement
func PrepareStatement(sql string, g *Group) {
	g.List(Id("stmt"), Err()).Op(":=").Id("tx").Dot("Prepare").Call(Lit(sql))
}

// ExecuteStatement produces the code required to execute a statement.
// Since not every database fully supports the Result interface, we
// simply ignore it
func ExecuteStatement(g *Group, argsFun func(g *Group)) {
	g.List(Id("_"), Err()).Op("=").Id("stmt").Dot("Exec").CallFunc(func(g2 *Group) {
		g2.ListFunc(argsFun)
	})
}

// ReturnCommitTransaction is a helper function that generates the code
// needed to commit the transaction and return its error
func ReturnCommitTransaction(g *Group) {
	g.Return(Id("tx").Dot("Commit").Call())
}

// DeferRollbackTransaction produces a default defer transaction
// rollback statement
func DeferRollbackTransaction(g *Group) {
	DeferCall("tx", "Rollback", g)
}

// DeferCloseSttement produces a default defer sql statement close
// statement
func DeferCloseStatement(g *Group) {
	DeferCall("stmt", "Close", g)
}

// Defer is a helper function that generates the code
// needed to defer a call
func DeferCall(id string, fun string, g *Group) {
	g.Defer().Id(id).Dot(fun).Call()
}

// IfErrorReturn is a helper function that checks the err variable and
// returns immediately
func IfErrorReturn(g *Group) {
	g.If(Err().Op("!=").Nil()).Block(Return(Err()))
}
