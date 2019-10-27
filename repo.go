package main

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

	AddExecStatementsFun(f)

	AddRepoFuns(p.Model.Entities, f)

	return f.Save(p.Filename)
}

// AddExecStatementsFun generates a convenience function that accepts a
// slice of sql statements, and runs them all, one by one
func AddExecStatementsFun(f *File) {
	funName := "ExecStatements"

	f.Comment(fmt.Sprintf("%s runs all SQL statements in the given slice. No transaction is opened. This function is designed to run DLL such as DROP/CREATE table.", funName))
	f.Func().Id(funName).Params(
		Id("db").Op("*").Qual("database/sql", "DB"),
		Id("stmts").Op("[]").Id("string")).Error().BlockFunc(func(g *Group) {
		g.For(List(Id("_"), Id("stmt")).Op(":=").Id("range").Id("stmts")).BlockFunc(func(g2 *Group) {
			g2.List(Id("_"), Err()).Op(":=").Id("db").Dot("Exec").Call(Id("stmt"))
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

		AddInsertFun(e, f)
		AddFindFuns(e, f)
	}
}

// InsertEntityFunName returns the name of the insert function for the
// entity
func InsertEntityFunName(e *Entity) string {
	return fmt.Sprintf("Insert%s", e.Name)
}

// AddInsertFun produces the function that inserts the given
// entity to the database.
func AddInsertFun(e *Entity, f *File) {
	funName := InsertEntityFunName(e)

	f.Comment(fmt.Sprintf("%s inserts an entity of type %s to the database", funName, e.Name))
	f.Comment("This function also persists its relations to other linked entities")
	f.Func().Id(funName).Params(Id("db").Op("*").Qual("database/sql", "DB"), Id(e.VarName()).Op("*").Id(e.Name)).Parens(List(Op("*").Id(e.Name), Error())).BlockFunc(func(g *Group) {

		// Open a transaction
		BeginTransaction(g)
		IfErrorReturnEntityAndError(e, g)

		DeferRollbackTransaction(g)
		PrepareTransactionStatement(InsertStatement(e), g)
		IfErrorReturnEntityAndError(e, g)

		DeferCloseStatement(g)

		ExecuteStatement(g, func(g2 *Group) {
			InsertStatementValues(e, g2)
		})
		IfErrorReturnEntityAndError(e, g)

		CommitTransaction(g)
		IfErrorReturnEntityAndError(e, g)
		ReturnEntityAndNil(e, g)
	})
}

// AddFindFuns produces functions that perform lookups by key on the
// given entity
func AddFindFuns(e *Entity, f *File) {
	for _, a := range e.Attributes {
		if a.HasModifier("unique") && a.HasModifier("indexed") {
			AddFindFun(e, a, f)
		}
	}
}

// FindEntityByAttributeFunName returns the name of the finder function for the given
// entity and attribute
func FindEntityByAttributeFunName(e *Entity, a *Attribute) string {
	return fmt.Sprintf("Find%sBy%s", e.Name, a.Name)
}

// AddFindFun produces a finder function for the given entity and
// attribute
func AddFindFun(e *Entity, a *Attribute, f *File) {
	funName := FindEntityByAttributeFunName(e, a)
	f.Comment(fmt.Sprintf("%s finds an instance of type %s by %s. If no row matches, then this function returns an error", funName, e.Name, a.Name))
	f.Func().Id(funName).Params(Id("db").Op("*").Qual("database/sql", "DB"), TypedFromAttribute(Id("v"), a)).Parens(List(Op("*").Id(e.Name), Error())).BlockFunc(func(g *Group) {

		VarForEntity(e, g)
		PrepareDbStatement(SelectByColumnStatement(e, a), g)
		IfErrorReturnWithEntity(e, g)
		DeferCloseStatement(g)
		VarNamesForEntity(e, g)
		ScanRow(e, g)
		IfErrorReturnWithEntity(e, g)
		ReturnRow(e, g)
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

// SelectByColumnStatement generates a SELECT statement that performs a
// query for an entity by a single column
func SelectByColumnStatement(e *Entity, a *Attribute) string {
	chunks := []string{}
	chunks = append(chunks, "SELECT")

	columns := []string{}
	for _, a := range e.Attributes {
		columns = append(columns, AttributeColumnName(a))
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			columns = append(columns, RelationColumnName(r))
		}
	}

	chunks = append(chunks, strings.Join(columns, ","))
	chunks = append(chunks, "FROM")
	chunks = append(chunks, TableName(e))
	chunks = append(chunks, "WHERE")
	chunks = append(chunks, AttributeColumnName(a))
	chunks = append(chunks, "=")
	chunks = append(chunks, "?")
	return strings.Join(chunks, " ")
}

// VarNameForEntity produces a variable for the entity.
func VarForEntity(e *Entity, g *Group) {
	g.Var().Id(e.VarName()).Id(e.Name)
}

// VarNamesForEntity produces a variable for each attribute and relation
// in the given entity. This is used when scanning rows returned from
// the database
func VarNamesForEntity(e *Entity, g *Group) {

	// use the Golang type for the attribute
	for _, a := range e.Attributes {
		TypedFromAttribute(g.Var().Id(a.VarName()), a)
	}

	// IDs to other tables are modelled as strings
	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			TypedFromDataType(g.Var().Id(r.VarName()), "string")
		}
	}
}

// ScanRow produces the code required to scan a database row for the
// given entity
func ScanRow(e *Entity, g *Group) {
	g.Err().Op("=").Id("stmt").Dot("QueryRow").Call(Id("v")).Dot("Scan").Call(ListFunc(func(g *Group) {
		// use the Golang type for the attribute
		for _, a := range e.Attributes {
			g.Op("&").Id(a.VarName())
		}

		// IDs to other tables are modelled as strings
		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
				g.Op("&").Id(r.VarName())
			}
		}
	}))
}

// ReturnRow produces the code required to return an entity populated
// from scanned variables, and a nil error
func ReturnRow(e *Entity, g *Group) {
	g.Return().List(Op("&").Id(e.Name).Values(DictFunc(func(d Dict) {

		for _, a := range e.Attributes {
			d[Id(a.Name)] = Id(a.VarName())
		}

		// For each relation, wrap the id in an struct of the appropiate
		// tyoe
		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {

				d[Id(r.Name())] = Op("&").Id(r.Entity).Values(Dict{
					Id("ID"): Id(r.VarName()),
				})
			}
		}

	})), Nil())
}

// BeginTransaction is a helper function that generates the code needed
// to start a new transaction.
func BeginTransaction(g *Group) {
	g.List(Id("tx"), Err()).Op(":=").Id("db").Dot("Begin").Call()
}

// PrepareTransactionStatement produces the code required to create a new
// statement from a transaction
func PrepareTransactionStatement(sql string, g *Group) {
	PrepareStatement("tx", sql, g)
}

// PrepareDbStatement produces the code required to create a new
// statement from a database
func PrepareDbStatement(sql string, g *Group) {
	PrepareStatement("db", sql, g)
}

// PrepareStatement produces the code required to create a new statement
func PrepareStatement(receiver string, sql string, g *Group) {
	g.List(Id("stmt"), Err()).Op(":=").Id(receiver).Dot("Prepare").Call(Lit(sql))
}

// ExecuteStatement produces the code required to execute a statement.
// Since not every database fully supports the Result interface, we
// simply ignore it
func ExecuteStatement(g *Group, argsFun func(g *Group)) {
	g.List(Id("_"), Err()).Op("=").Id("stmt").Dot("Exec").CallFunc(func(g2 *Group) {
		g2.ListFunc(argsFun)
	})
}

// CommitTransaction is a helper function that generates the code
// needed to commit the transaction and return it error
func CommitTransaction(g *Group) {
	g.Err().Op("=").Id("tx").Dot("Commit").Call()
}

// IfErrorReturnEntityAndError returns a final statement that returns the entity
// instance and the an error
func IfErrorReturnEntityAndError(e *Entity, g *Group) {
	g.If(Err().Op("!=").Nil().BlockFunc(func(g2 *Group) {
		ReturnEntityAndError(e, g2)
	}))
}

// ReturnEntityAndError returns a final statement that returns the entity
// instance and the an error
func ReturnEntityAndError(e *Entity, g *Group) {
	g.Return(List(Id(e.VarName()), Err()))
}

// ReturnEntityAndNil returns a final statement that returns an entity
// instance and nil as an error
func ReturnEntityAndNil(e *Entity, g *Group) {
	g.Return(List(Id(e.VarName()), Nil()))
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

// Return nil is a helper function that unconditonally returns
// nil
func ReturnNil(g *Group) {
	g.Return(Nil())
}

// IfErrorReturnWithEntity is a helper function that checks the err variable and
// returns immediately a tuple with a variable for the entity, and the
// error
func IfErrorReturnWithEntity(e *Entity, g *Group) {
	g.If(Err().Op("!=").Nil()).Block(Return(List(Op("&").Id(e.VarName()), Err())))
}
