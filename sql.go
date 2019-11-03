package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
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

	f.Comment(fmt.Sprintf("%s initializes a new database handle", funName))
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
				AddEntityCreateTable(e, m, g)
				AddEntityIndices(e, g)
			}

			AddExtraSqlInitialization(g)
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
func AddEntityCreateTable(e *Entity, m *Model, g *Group) {
	chunks := []string{}
	chunks = append(chunks, fmt.Sprintf("CREATE TABLE %s (", TableName(e)))

	colsChunks := []string{}
	for _, a := range e.Attributes {
		colsChunks = append(colsChunks, TableColumnFromAttribute(a))
	}
	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			colsChunks = append(colsChunks, TableColumnFromRelation(r))
		}
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			colsChunks = append(colsChunks, ForeignKeyConstraintFromRelation(r, m))
		}
	}

	chunks = append(chunks, strings.Join(colsChunks, ", "))
	chunks = append(chunks, ")")
	g.Lit(strings.Join(chunks, ""))
}

// AddEntityIndices generates the database indices for the given table
func AddEntityIndices(e *Entity, g *Group) {
	for _, a := range e.Attributes {
		if a.Name != "ID" && (a.HasModifier("unique") || a.HasModifier("indexed")) {

			tableName := TableName(e)
			columnName := AttributeColumnName(a)

			chunks := []string{}
			chunks = append(chunks, "CREATE")

			if a.HasModifier("unique") {
				chunks = append(chunks, "UNIQUE")
			}

			chunks = append(chunks, "INDEX")
			chunks = append(chunks, fmt.Sprintf("%s_%s", tableName, columnName))
			chunks = append(chunks, "ON")
			chunks = append(chunks, fmt.Sprintf("%s(%s)", tableName, columnName))

			g.Lit(strings.Join(chunks, " "))
		}
	}
}

// AddExtraSqlInitialization adds extra database initialization steps
func AddExtraSqlInitialization(g *Group) {
	g.Lit("PRAGMA foreign_keys = ON")

}

// TableColumnFromAttribute builds the column specification for the
// given attribute.
func TableColumnFromAttribute(a *Attribute) string {
	dataType := AttributeSqlType(a)
	spec := fmt.Sprintf("%s %s", AttributeColumnName(a), dataType)
	if dataType == "varchar" {
		spec = fmt.Sprintf("%s NOT NULL", spec)
	}

	if a.Name == "ID" {
		spec = fmt.Sprintf("%s PRIMARY KEY", spec)
	}

	return spec
}

// TableColumnFromRelation builds the column specification for the given
// relation.
func TableColumnFromRelation(r *Relation) string {
	return fmt.Sprintf("%s %s", RelationColumnName(r), RelationSqlType(r))
}

// ForeignKeyFromRelation builds the foreign key specification for the
// given relation
func ForeignKeyConstraintFromRelation(r *Relation, m *Model) string {
	e := m.EntityForNameOrPanic(r.Entity)

	return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", RelationColumnName(r), TableName(e), "id")
}

// TableName builds a SQL table name, for the given entity
func TableName(e *Entity) string {
	return strings.ToLower(strcase.ToSnake(fmt.Sprintf("%s%s", "Fl", e.Plural())))
}

// ColumnName
func AttributeColumnName(a *Attribute) string {
	return strings.ToLower(strcase.ToSnake(a.Name))
}

func RelationColumnName(r *Relation) string {
	return strings.ToLower(strcase.ToSnake(r.Name()))
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
