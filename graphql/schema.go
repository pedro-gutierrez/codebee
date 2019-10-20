// graphql provides with a Jennifer-like api that helps to generate
// GraphQL schema AST.
package graphql

import (
	"bytes"
	"io"
	"io/ioutil"
)

// File represents a GraphQL schema document
type File struct {
	Types []*Type
}

// NewFile creates a new GraphQL schema
func NewFile() *File {
	return &File{}
}

// Save renders the GraphQL schema and writes it to disk
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

// Render renders the whole GraphQL schema to the specified writer. By
// default we use a byte buffer but it could also be serialized to
// standard output or to a network socket
func (f *File) Render(w io.Writer) error {
	chunks := []string{}
	for _, t := range f.Types {
		for _, c := range t.Render() {
			chunks = append(chunks, c)
		}
		chunks = append(chunks, "\n")
	}

	return Write(w, chunks)
}

// Type adds a new object type to the schema. The provided function
// performs extra customization on the new added type
func (f *File) Type(name string, fun func(t *Type)) *File {
	t := &Type{
		Name:   name,
		Fields: []*Field{},
	}
	fun(t)
	f.Types = append(f.Types, t)
	return f
}

// Type represents a GraphQL object type
type Type struct {
	Name   string
	Fields []*Field
}

// Field defines a new field on the given type.
//func (t *Type) Field(name string, dataType string) *Type {
//	t.Fields = append(t.Fields, &Field{
//		Name:     name,
//		DataType: dataType,
//	})
//	return t
//}

// Field defines a new field on the given type.
func (t *Type) Field(name string) *Field {
	f := &Field{
		Name: name,
	}

	t.Fields = append(t.Fields, f)
	return f
}

// Render renders the type as a chunk of strings. It will also render
// all fields belonging to the field.
func (t *Type) Render() []string {
	chunks := []string{}
	chunks = append(chunks, "type ")
	chunks = append(chunks, t.Name)
	chunks = append(chunks, " {\n")
	for _, f := range t.Fields {
		for _, c := range f.Render() {
			chunks = append(chunks, c)
		}
	}
	chunks = append(chunks, "}\n")
	return chunks
}

// Field represents a field that can be part of an object type, or an
// input type
type Field struct {
	Name     string
	DataType string
}

// Type defines the type for the field
func (f *Field) Type(t string) *Field {
	f.DataType = t
	return f
}

// Render renders the field as an array of string chunks
func (f *Field) Render() []string {
	chunks := []string{
		"    ",
		f.Name,
		": ",
		f.DataType,
		"!",
		"\n",
	}

	return chunks
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
