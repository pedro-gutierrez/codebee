package main

import (
	. "github.com/flootic/generator/sql"
)

// CreateSQLSchema generates a SQL file that contains all the DDL
// statements that represent the whole database schema
func CreateSQLSchema(p *Package) error {
	f := NewFile()

	AddModelTables(p.Model.Entities, f)

	// put here more functions to generate indices, constraints, etc..

	return f.Save(p.Filename)
}

// AddModelTables generates all the CREATE TABLE statement for the given set
// of tables and adds them to the given file
func AddModelTables(entities []*Entity, f *File) {
	for _, e := range entities {
		AddEntityTable(e, f)
	}
}

// AddEntityTables generates all the necessary tables for the given
// entity to the given file
func AddEntityTable(e *Entity, f *File) {
	f.Table(TableName(e), func(t *Table) {
		for _, a := range e.Attributes {
			t.Column(AttributeColumnName(a)).Type(AttributeSqlType(a))
		}
		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
				t.Column(RelationColumnName(r)).Type(RelationSqlType(r))
			}
		}
	})
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
