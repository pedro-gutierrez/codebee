// sql provides with a Jennifer-like api that helps to generate
// standard SQL schema code.
package sql

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// File represents a SQL schema document
type File struct {
	Tables []*Table
}

// NewFile creates a new SQL schema
func NewFile() *File {
	return &File{}
}

// Save renders the SQL schema and writes it to disk
func (f *File) Save(filename string) error {
	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

// Write is a general purpose helper function that writes the given
// string chunks to the given writer.
func Write(w io.Writer, chunks []string) error {
	for _, c := range chunks {
		if _, err := w.Write([]byte(c)); err != nil {
			return err
		}
	}
	return nil
}

// Render renders the whole SQL schema to the specified writer. By
// default we use a byte buffer but it could also be serialized to
// standard output or to a network socket
func (f *File) Render(w io.Writer) error {
	chunks := []string{}
	for _, t := range f.Tables {
		chunks = append(chunks, t.RenderDrop())
		chunks = append(chunks, ";\n")
		chunks = append(chunks, t.RenderCreate())
		chunks = append(chunks, ";\n")
	}

	return Write(w, chunks)
}

// Type adds a new table to the schema. The provided function performs
// further customization on the table
func (f *File) Table(name string, fun func(t *Table)) *File {
	t := &Table{
		Name:    name,
		Columns: []*Column{},
	}
	fun(t)
	f.Tables = append(f.Tables, t)
	return f
}

// Table represents a SQL table
type Table struct {
	Name    string
	Columns []*Column
}

// Field defines a new field on the given type.
func (t *Table) Column(name string) *Column {
	c := &Column{
		Name: name,
	}

	t.Columns = append(t.Columns, c)
	return c
}

// Render renders the DROP TABLE code chunks for the table
func (t *Table) RenderDrop() string {
	return strings.Join([]string{
		"DROP",
		"TABLE",
		"IF",
		"EXISTS",
		t.Name,
	}, " ")
}

// Render renders the CREATE TABLE code chunks for the table
func (t *Table) RenderCreate() string {
	return strings.Join([]string{
		"CREATE",
		"TABLE",
		t.Name,
		fmt.Sprintf("(%s)", t.RenderColumns()),
	}, " ")
}

// RenderColumns helper function that returns the table columns SQL code
// in a single string
func (t *Table) RenderColumns() string {
	chunks := []string{}
	for _, c := range t.Columns {
		chunks = append(chunks, c.Render())
	}

	return strings.Join(chunks, ", ")
}

// Column represents a SQL table column
type Column struct {
	Name     string
	DataType string
}

// Type defines the data type for the column
func (c *Column) Type(t string) *Column {
	c.DataType = t
	return c
}

// Render renders the column SQL code
func (f *Column) Render() string {
	return strings.Join([]string{
		f.Name,
		f.DataType,
	}, " ")
}
