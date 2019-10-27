package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	//"github.com/iancoleman/strcase"
	"strings"
)

// CreateSql generates the Golang module that produces the necessary SQL
// statements that build the database
func CreateSql(p *Package) error {
	f := NewFile(p.Name)

	AddNewDbFun(f)
	AddSqlSchemaFun(p.Model, f)

	return f.Save(p.Filename)
}

// AddDbFun builds the function that initializes the database
func AddNewDbFun(f *File) {

	funName := "NewDb"

	// For now this code is sqlite3 specific.
	f.Anon("github.com/mattn/go-sqlite3")

	f.Comment(fmt.Sprintf("%s a new database handle", funName))
	f.Func().Id(funName).Params().Parens(List(
		Op("*").Qual("database/sql", "DB"),
		Error(),
	)).Block(

		// For now we support only Sqlite3, but here we can
		// add parameters, flags and adopt different strategies and
		// the rest of the application should not be aware
		Return(
			Id("sql").Dot("Open").Call(
				Lit("sqlite3"),
				Lit("file::memory:?cache=shared"),
			),
		),
	)
}

// AddSqlSchemaFun builds the function that returns the list of SQL
// statements that initialize the database
func AddSqlSchemaFun(m *Model, f *File) {
	funName := "SqlSchema"
	f.Comment(fmt.Sprintf("%s returns the database Sql schema, as a list of statements", funName))
	f.Func().Id(funName).Params().Op("[]").Id("string").Block(
		Return(Op("[]").Id("string").ValuesFunc(func(g *Group) {

			for _, e := range m.Entities {
				AddEntityDropTable(e, g)
				AddEntityCreateTable(e, g)
			}
		}),
		))
}

// AddEntityDropTable adds a DROP TABLE statement to the schema, for the
// given entity
func AddEntityDropTable(e *Entity, g *Group) {
	g.Lit(fmt.Sprintf("DROP TABLE IF EXISTS %s", TableName(e)))
}

// AddEntityCreateTable adds a CREATE TABLE statement to the schema, for
// the given entity
func AddEntityCreateTable(e *Entity, g *Group) {
	chunks := []string{}
	chunks = append(chunks, fmt.Sprintf("CREATE TABLE %s (", TableName(e)))

	colsChunks := []string{}
	for _, a := range e.Attributes {
		colsChunks = append(colsChunks, fmt.Sprintf("%s %s", AttributeColumnName(a), AttributeSqlType(a)))
	}
	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			colsChunks = append(colsChunks, fmt.Sprintf("%s %s", RelationColumnName(r), RelationSqlType(r)))
		}
	}

	chunks = append(chunks, strings.Join(colsChunks, ", "))
	chunks = append(chunks, ")")
	g.Lit(strings.Join(chunks, ""))
}

// AttributeSqlType returns the SQL datatype for an attribute.
func AttributeSqlType(a *Attribute) string {
	switch a.Type {

	case "Int":
		return "integer"

	default:
		return "varchar"
	}
}

// RelationSqlType returns the SQL datatype for a relation. We will
// always use the varchar data type for references between entities.
func RelationSqlType(r *Relation) string {
	return "varchar"
}
