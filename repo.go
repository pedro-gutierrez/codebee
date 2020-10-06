package main

import (
	"fmt"
	"strings"

	. "github.com/dave/jennifer/jen"
)

// CreateRepo generates a Golang file that contains all the necessary
// source code required to persist the given set of entities to the
// database
func CreateRepo(p *Package) error {
	f := NewFile(p.Name)

	f.PackageComment(fmt.Sprintf("%s contains all the library code for the platform", p.Name))
	f.PackageComment("This file contains all the functions that implement the Database repository")
	f.PackageComment(" ** THIS CODE IS MACHINE GENERATED. DO NOT EDIT MANUALLY ** ")

	f.ImportAlias("database/sql", "sql")
	f.ImportAlias("io/ioutil", "ioutil")

	AddExecStatementsFun(f)

	AddRepoFuns(p.Model, f)

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

// AddRepoFun generates all the repository functions and adds them to the given file
func AddRepoFuns(m *Model, f *File) {
	for _, e := range m.Entities {

		if e.SupportsOperation("create") {

			AddInsertFun(m, e, f)

		}

		if e.SupportsOperation("update") {

			AddUpdateFun(e, f)
		}

		if e.SupportsOperation("delete") {

			AddDeleteFun(e, f)
		}

		if e.SupportsOperation("find") {
			AddFindFuns(e, f)
		}

	}
}

// InsertEntityFunName returns the name of the insert function for the
// entity
func InsertEntityFunName(e *Entity) string {
	return fmt.Sprintf("Create%s", e.Name)
}

// AddInsertFun produces the function that inserts the given
// entity to the database.
func AddInsertFun(m *Model, e *Entity, f *File) {
	funName := InsertEntityFunName(e)

	f.Comment(fmt.Sprintf("%s inserts an entity of type %s to the database", funName, e.Name))
	f.Comment("This function also persists its relations to other linked entities")
	f.Func().Id(funName).Params(Id("db").Op("*").Qual("database/sql", "DB"), Id(e.VarName()).Op("*").Id(e.Name)).Parens(List(Op("*").Id(e.Name), Error())).BlockFunc(func(g *Group) {

		// Open a transaction
		BeginTransaction(g)
		IfErrorReturnEntityAndError(e, g)

		DeferRollbackTransaction(g)

		// insert statement for the entity
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

// UpdateEntityFunName returns the name of the update function for the
// entity
func UpdateEntityFunName(e *Entity) string {
	return fmt.Sprintf("Update%s", e.Name)
}

// AddInsertFun produces the function that inserts the given
// entity to the database.
func AddUpdateFun(e *Entity, f *File) {
	funName := UpdateEntityFunName(e)

	f.Comment(fmt.Sprintf("%s updates an existing entity of type %s into the database", funName, e.Name))
	f.Func().Id(funName).Params(Id("db").Op("*").Qual("database/sql", "DB"), Id(e.VarName()).Op("*").Id(e.Name)).Parens(List(Op("*").Id(e.Name), Error())).BlockFunc(func(g *Group) {

		// Open a transaction
		BeginTransaction(g)
		IfErrorReturnEntityAndError(e, g)

		DeferRollbackTransaction(g)
		PrepareTransactionStatement(UpdateStatement(e), g)
		IfErrorReturnEntityAndError(e, g)

		DeferCloseStatement(g)

		ExecuteStatement(g, func(g2 *Group) {
			UpdateStatementValues(e, g2)
		})
		IfErrorReturnEntityAndError(e, g)

		CommitTransaction(g)
		IfErrorReturnEntityAndError(e, g)
		ReturnEntityAndNil(e, g)
	})
}

// DeleteEntityFunName returns the name of the delete function for the
// entity
func DeleteEntityFunName(e *Entity) string {
	return fmt.Sprintf("Delete%s", e.Name)
}

// AddDeleteFun produces the function that deletes the given
// entity to the database, by its ID.
func AddDeleteFun(e *Entity, f *File) {
	funName := DeleteEntityFunName(e)

	f.Comment(fmt.Sprintf("%s deletes an existing entity of type %s from the database, by its id", funName, e.Name))
	f.Func().Id(funName).Params(
		Id("db").Op("*").Qual("database/sql", "DB"),
		Id("id").String(),
	).Parens(
		List(Op("*").Id(e.Name),
			Error(),
		)).BlockFunc(func(g *Group) {

		g.Add(EmptyStructForEntity(e))

		BeginTransaction(g)
		IfErrorReturnEntityAndError(e, g)

		DeferRollbackTransaction(g)
		PrepareTransactionStatement(DeleteStatement(e), g)
		IfErrorReturnEntityAndError(e, g)

		DeferCloseStatement(g)

		ExecuteStatement(g, func(g2 *Group) {
			DeleteStatementValues(e, g2)
		})
		IfErrorReturnEntityAndError(e, g)

		CommitTransaction(g)
		ReturnEntityAndNil(e, g)
	})
}

// AddFindFuns produces functions that perform lookups by key on the
// given entity
func AddFindFuns(e *Entity, f *File) {
	for _, a := range e.Attributes {
		if a.HasModifier("unique") && a.HasModifier("indexed") {
			AddFindByAttributeFun(e, a, f)
		}
	}

	for _, r := range e.Relations {
		if r.HasModifier("hasOne") || r.HasModifier("belongsTo") {
			AddFindByRelationFun(e, r, f)
		}
	}

	AddFindAllFun(e, f)
}

// FindEntityByAttributeFunName returns the name of the finder function for the given
// entity and attribute
func FindEntityByAttributeFunName(e *Entity, a *Attribute) string {
	return fmt.Sprintf("Find%sBy%s", e.Name, a.Name)
}

// AddFindAllFun produces a finder function that returns instances
// of the given entity
func AddFindAllFun(e *Entity, f *File) {

	// error handling code to be used in different points of this
	// function body
	ifErrReturn := If(Err().Op("!=").Nil()).Block(
		Return(
			Id(VarName(e.PluralName())),
			Err(),
		),
	)

	funName := FindAllFunName(e)
	f.Comment(fmt.Sprintf("%s finds all instances of type %s. If no row matches, then this function returns an empty slice", funName, e.Name))
	f.Func().Id(funName).Params(
		Id("db").Op("*").Qual("database/sql", "DB"),
		Id("limit").Int32(),
		Id("offset").Int32(),
	).Parens(List(
		Op("[]").Op("*").Id(e.Name),
		Error(),
	)).BlockFunc(func(g *Group) {

		g.Id(VarName(e.PluralName())).Op(":=").Op("[]").Op("*").Id(e.Name).Values(Dict{})

		g.List(
			Id("stmt"),
			Err(),
		).Op(":=").Id("db").Dot("Prepare").Call(
			Qual("fmt", "Sprintf").Call(
				Lit(fmt.Sprintf(
					"%s ORDER BY %s ASC LIMIT %%v OFFSET %%v",
					SelectAllStatement(e),
					AttributeColumnName(e.PreferredSort()),
				)),
				Id("limit"),
				Id("offset"),
			),
		)

		g.Add(ifErrReturn)
		DeferCall("stmt", "Close", g)

		g.List(
			Id("rows"),
			Err(),
		).Op(":=").Id("stmt").Dot("Query").Call()
		g.Add(ifErrReturn)

		DeferCall("rows", "Close", g)

		g.For(
			Id("rows").Dot("Next").Call(),
		).BlockFunc(func(g2 *Group) {

			g2.Add(EmptyStructForEntity(e))
			g2.Err().Op(":=").Id("rows").Dot("Scan").Call(ListFunc(
				ScanRowIntoEntityStruct(e),
			))

			g2.Add(ifErrReturn)
			g2.Id(VarName(e.PluralName())).Op("=").Append(Id(VarName(e.PluralName())), Id(e.VarName()))
		})

		g.Return(List(
			Id(VarName(e.PluralName())),
			Nil(),
		))
	})
}

// FindAllFunName returns the name of the finder function for the given
// entity
func FindAllFunName(e *Entity) string {
	return fmt.Sprintf("FindAll%s", e.PluralName())
}

// AddFindByAttributeFun produces a finder function for the given entity and
// attribute
func AddFindByAttributeFun(e *Entity, a *Attribute, f *File) {
	funName := FindEntityByAttributeFunName(e, a)
	f.Comment(fmt.Sprintf("%s finds an instance of type %s by %s. If no row matches, then this function returns an error", funName, e.Name, a.Name))
	f.Func().Id(funName).Params(Id("db").Op("*").Qual("database/sql", "DB"), TypedFromAttribute(Id(a.VarName()), a)).Parens(List(Op("*").Id(e.Name), Error())).BlockFunc(func(g *Group) {

		g.Add(EmptyStructForEntity(e))
		PrepareDbStatement(SelectByColumnFromAttributeStatement(e, a), g)
		IfErrorReturnWithEntity(e, g)
		DeferCloseStatement(g)

		g.Err().Op("=").Id("stmt").Dot("QueryRow").Call(Id(a.VarName())).Dot("Scan").Call(ListFunc(
			ScanRowIntoEntityStruct(e),
		))
		g.Return(List(
			Id(e.VarName()),
			Err(),
		))
	})
}

// FindEntityByRelationFunName returns the name of the finder function for the given
// entity and relation
func FindEntityByRelationFunName(e *Entity, r *Relation) string {
	return fmt.Sprintf("Find%sBy%s", e.PluralName(), r.Alias())
}

// AddFindByRelationFun produces a finder function for the given entity
// and relation. This function will return a list of instances of the
// given entity
func AddFindByRelationFun(e *Entity, r *Relation, f *File) {
	funName := FindEntityByRelationFunName(e, r)

	// error handling code to be used in different points of this
	// function body
	ifErrReturn := If(Err().Op("!=").Nil()).Block(
		Return(
			Id(VarName(e.PluralName())),
			Err(),
		),
	)

	f.Comment(fmt.Sprintf("%s finds a list of instances of type %s by %s. If no rows match, then this function returns an empty slice. Results are sorted and paginated.", funName, e.Name, r.Alias()))
	f.Func().Id(funName).Params(
		Id("db").Op("*").Qual("database/sql", "DB"),
		Id(r.VarName()).String(),
		Id("limit").Int32(),
		Id("offset").Int32(),
	).Parens(List(
		Op("[]").Op("*").Id(e.Name),
		Error(),
	)).BlockFunc(func(g *Group) {
		g.Id(VarName(e.PluralName())).Op(":=").Op("[]").Op("*").Id(e.Name).Values(Dict{})

		g.List(
			Id("stmt"),
			Err(),
		).Op(":=").Id("db").Dot("Prepare").Call(
			Qual("fmt", "Sprintf").Call(
				Lit(fmt.Sprintf(
					"%s ORDER BY %s ASC LIMIT %%v OFFSET %%v",
					SelectByColumnFromRelationStatement(e, r),
					AttributeColumnName(e.PreferredSort()),
				)),
				Id("limit"),
				Id("offset"),
			),
		)

		g.Add(ifErrReturn)
		DeferCall("stmt", "Close", g)

		g.List(
			Id("rows"),
			Err(),
		).Op(":=").Id("stmt").Dot("Query").Call(
			Id(r.VarName()),
		)
		g.Add(ifErrReturn)

		DeferCall("rows", "Close", g)

		g.For(
			Id("rows").Dot("Next").Call(),
		).BlockFunc(func(g2 *Group) {

			g2.Add(EmptyStructForEntity(e))
			g2.Err().Op(":=").Id("rows").Dot("Scan").Call(ListFunc(
				ScanRowIntoEntityStruct(e),
			))

			g2.Add(ifErrReturn)
			g2.Id(VarName(e.PluralName())).Op("=").Append(Id(VarName(e.PluralName())), Id(e.VarName()))
		})

		g.Return(List(
			Id(VarName(e.PluralName())),
			Nil(),
		))
	})
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
		columns = append(columns, AttributeColumnName(a))
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			columns = append(columns, RelationColumnName(r))
		}
	}

	chunks = append(chunks, fmt.Sprintf("(%s)", strings.Join(columns, ",")))
	chunks = append(chunks, "VALUES")

	placeholders := []string{}
	i := 1
	for _, _ = range e.Attributes {
		placeholders = append(placeholders, placeholder(i))
		i++
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			placeholders = append(placeholders, placeholder(i))
			i++
		}
	}

	chunks = append(chunks, fmt.Sprintf("(%s)", strings.Join(placeholders, ",")))
	return strings.Join(chunks, " ")
}

// placeholder returns a postgres style placeholder
func placeholder(i int) string {
	return fmt.Sprintf("$%v", i)
}

// InsertStatementValues generates the Golang code that populates the
// values to be sent to the INSERT sql statement for the entity
func InsertStatementValues(e *Entity, g *Group) {
	for _, a := range e.Attributes {
		g.Id(e.VarName()).Dot(a.Name)
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			g.Id(e.VarName()).Dot(r.Alias()).Dot("ID")
		}
	}
}

// InsertStatementValuesForGeneratedRelation generates the Golang code that populates the
// values to be sent to the INSERT sql statement for the generated
// relation of the given entity
func InsertStatementValuesForGeneratedRelation(e *Entity, r *Relation, e2 *Entity, g *Group) {
	for _, a := range e2.Attributes {
		g.Id(e.VarName()).Dot(r.Alias()).Dot(a.Name)
	}

	for _, r2 := range e2.Relations {
		if r2.HasModifier("belongsTo") || r2.HasModifier("hasOne") {
			g.Id(e.VarName()).Dot(r.Alias()).Dot(r.Alias()).Dot("ID")
		}
	}
}

// UpdateStatement generates a sql INSERT statement for the given entity
func UpdateStatement(e *Entity) string {
	chunks := []string{}
	chunks = append(chunks, "UPDATE")
	chunks = append(chunks, TableName(e))
	chunks = append(chunks, "SET")

	columns := []string{}
	i := 1
	for _, a := range e.Attributes {
		if a.Name != "ID" {
			col := fmt.Sprintf("%s=%s", AttributeColumnName(a), placeholder(i))
			i++
			columns = append(columns, col)
		}
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			col := fmt.Sprintf("%s=%s", RelationColumnName(r), placeholder(i))
			i++
			columns = append(columns, col)
		}
	}

	chunks = append(chunks, strings.Join(columns, ","))
	chunks = append(chunks, fmt.Sprintf("WHERE id=%s", placeholder(i)))
	return strings.Join(chunks, " ")
}

// UpdateStatementValues generates the Golang code that populates the
// values to be sent to the UPDATE sql statement for the entity
func UpdateStatementValues(e *Entity, g *Group) {

	// bindings for the columns to update
	for _, a := range e.Attributes {
		if a.Name != "ID" {
			g.Id(e.VarName()).Dot(a.Name)
		}
	}

	for _, r := range e.Relations {
		if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
			g.Id(e.VarName()).Dot(r.Alias()).Dot("ID")
		}
	}

	// binding for the where clause
	g.Id(e.VarName()).Dot("ID")
}

// DeleteStatement generates a sql DELETE statement for the given entity
func DeleteStatement(e *Entity) string {
	chunks := []string{}
	chunks = append(chunks, "DELETE FROM")
	chunks = append(chunks, TableName(e))
	chunks = append(chunks, "WHERE id=$1")
	return strings.Join(chunks, " ")
}

// DeleteStatementValues generates the Golang code that populates the
// values to be sent to the DELETE sql statement for the entity
func DeleteStatementValues(e *Entity, g *Group) {

	g.Id("id")
}

// SelectByColumnFromAttributeStatement generates a SELECT statement that performs a
// query for an entity by a single column. The column is inferred from the given attribute
func SelectByColumnFromAttributeStatement(e *Entity, a *Attribute) string {
	return SelectByColumnFromStatement(e, AttributeColumnName(a))
}

// SelectByColumnFromRelationStatement generates a SELECT statement that performs a
// query for an entity by a single column. The column is inferred from the given attribute
func SelectByColumnFromRelationStatement(e *Entity, r *Relation) string {
	return SelectByColumnFromStatement(e, RelationColumnName(r))
}

// SelectAllStatement generates a SELECT statement that performs a
// query for all rows in a given table.
func SelectAllStatement(e *Entity) string {

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
	return strings.Join(chunks, " ")

}

// SelectByColumnFromStatement generates a SELECT statement that performs a
// query for an entity by a single column. The column is inferred from the given attribute
func SelectByColumnFromStatement(e *Entity, whereColumn string) string {
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
	chunks = append(chunks, whereColumn)
	chunks = append(chunks, "=")
	chunks = append(chunks, "$1")
	return strings.Join(chunks, " ")
}

// EmptyStructForEntity returns a new statement that builds an empty
// struct, for the given entity
func EmptyStructForEntity(e *Entity) *Statement {
	return Id(e.VarName()).Op(":=").Op("&").Id(e.Name).Values(DictFunc(func(d Dict) {

		// Leave default values for attributes

		// IDs to other tables are modelled as strings
		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
				d[Id(r.Alias())] = Op("&").Id(r.Entity).Values(Dict{})
			}
		}

	}))
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
// given entity, into a struct of that entity. A function that takes a
// Jennifer Group is returned, so that this helper function can be reused
// in different contexts
func ScanRowIntoEntityStruct(e *Entity) func(*Group) {
	return func(g *Group) {
		// use the Golang type for the attribute
		for _, a := range e.Attributes {
			g.Op("&").Id(e.VarName()).Dot(a.Name)
		}

		// IDs to other tables are modelled as strings
		for _, r := range e.Relations {
			if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
				g.Op("&").Id(e.VarName()).Dot(r.Alias()).Dot("ID")
			}
		}
	}
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

				d[Id(r.Alias())] = Op("&").Id(r.Entity).Values(Dict{
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
	g.If(Err().Op("!=").Nil()).Block(Return(List(Id(e.VarName()), Err())))
}
