package main

import (
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

// CreateSql generates the Golang module that produces the necessary SQL
// statements that build the database
func CreateSql(p *Package) error {
	f := NewFile(p.Name)

	AddNewDbFun(p.Database, f)
	AddSqlSchemaFun(p.Model, p.Database, f)

	return f.Save(p.Filename)
}

// AddDbFun builds the function that initializes the database
func AddNewDbFun(db string, f *File) {

	funName := "NewDb"

	f.Anon(DatabaseImport(db))

	f.Comment(fmt.Sprintf("%s initializes a new database handle", funName))
	f.Func().Id(funName).Params(
		Id("conn").String(),
	).Parens(List(
		Op("*").Qual("database/sql", "DB"),
		Error(),
	)).Block(

		// For now we support only Sqlite3, but here we can
		// add parameters, flags and adopt different strategies and
		// the rest of the application should not be aware
		Return(
			Id("sql").Dot("Open").Call(
				Lit(DatabaseDriver(db)),
				Id("conn"),
			),
		),
	)
}

// AddDatabaseImport returns the most appropiate database import package,
// according to the given db type. Supported types are: sqlite,
// postgres
func DatabaseImport(db string) string {
	switch db {
	case "postgres":
		return "github.com/lib/pq"
	default:
		return "github.com/mattn/go-sqlite3"
	}
}

// DatabaseDriver translates the given database into the real database
// driver
func DatabaseDriver(db string) string {
	switch db {
	case "postgres":
		return "postgres"
	default:
		return "sqlite3"
	}
}

// AddSqlSchemaFun builds the function that returns the list of SQL
// statements that initialize the database
func AddSqlSchemaFun(m *Model, db string, f *File) {
	funName := "SqlSchema"
	f.Comment(fmt.Sprintf("%s returns the database Sql schema, as a list of statements", funName))
	f.Func().Id(funName).Params().Op("[]").Id("string").Block(
		Return(Op("[]").Id("string").ValuesFunc(func(g *Group) {

			for _, e := range m.Entities {
				AddEntityDropTable(e, db, g)
				AddEntityCreateTable(e, m, db, g)
			}

			for _, e := range m.Entities {
				AddEntityIndices(e, db, g)
				AddEntityForeignKeyConstraints(e, m, db, g)
			}

			AddExtraSqlInitialization(db, g)
		}),
		))
}

// AddEntityDropTable adds a DROP TABLE statement to the schema, for the
// given entity
func AddEntityDropTable(e *Entity, db string, g *Group) {
	g.Lit(DropTableStatement(e, db))
}

// DropTableStatement builds a drop table statement for the given entity
func DropTableStatement(e *Entity, db string) string {
	str := fmt.Sprintf("DROP TABLE IF EXISTS %s", TableName(e))
	if db != "sqlite3" {
		str = fmt.Sprintf("%s CASCADE", str)
	}
	return str
}

// AddEntityCreateTable adds a CREATE TABLE statement to the schema, for
// the given entity
func AddEntityCreateTable(e *Entity, m *Model, db string, g *Group) {
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

	// sqlite3 does not support ALTER table statements,
	// so we need to inline forein keys inside the table definition
	if db == "sqlite3" {
		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
				colsChunks = append(colsChunks, ForeignKeyConstraintFromRelation(r, m))
			}
		}
	}

	chunks = append(chunks, strings.Join(colsChunks, ", "))
	chunks = append(chunks, ")")
	g.Lit(strings.Join(chunks, ""))
}

// AddEntityIndices generates the database indices for the given table
func AddEntityIndices(e *Entity, db string, g *Group) {
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

// AddForeignConstraints builds foreign contraints for the given entity
func AddEntityForeignKeyConstraints(e *Entity, m *Model, db string, g *Group) {

	tableName := TableName(e)

	if db != "sqlite3" {
		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
				constraintName := ForeignKeyContraintName(e, r)
				g.Lit(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s %s", tableName, constraintName, ForeignKeyConstraintFromRelation(r, m)))
			}
		}
	}

}

// AddExtraSqlInitialization adds extra database initialization steps
func AddExtraSqlInitialization(db string, g *Group) {
	switch db {
	case "sqlite3":
		g.Lit("PRAGMA foreign_keys = ON")
		return

	default:
		return

	}
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

// ForeignKeyContraintName returns the name of the foreign key for the
// given entity and relation
func ForeignKeyContraintName(e *Entity, r *Relation) string {
	return fmt.Sprintf("%s_%s",
		TableName(e),
		RelationColumnName(r),
	)

}

// ForeignKeyConstraintFromRelation builds the foreign key specification for the
// given relation
func ForeignKeyConstraintFromRelation(r *Relation, m *Model) string {
	e := m.EntityForNameOrPanic(r.Entity)

	return fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s)", RelationColumnName(r), TableName(e), "id")
}

// TablePrefix returns a prefix for all sql tables. For now
// we don't set a prefix.
func TablePrefix() string {
	return ""
}

// TableName builds a SQL table name, for the given entity
func TableName(e *Entity) string {
	return strings.ToLower(strcase.ToSnake(fmt.Sprintf("%s%s", TablePrefix(), e.Plural())))
}

// ColumnName
func AttributeColumnName(a *Attribute) string {
	return strings.ToLower(strcase.ToSnake(a.Name))
}

// RelationColumnName returns the column name for a given
// relation. The name of the relation is convereted to snake case
// and we append the _id suffix.
func RelationColumnName(r *Relation) string {
	return fmt.Sprintf("%s_id", strings.ToLower(strcase.ToSnake(r.Name())))
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
