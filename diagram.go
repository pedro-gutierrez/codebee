package main

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"io/ioutil"
	"strings"
)

// CreateDiagram generates a Graphviz diagram, a visual representation
// of the model
func CreateDiagram(p *Package) error {
	d := DotDiagramFromModel(p.Model)
	return ioutil.WriteFile(p.Filename, []byte(d.String()), 0644)
}

// DotDiagramFromModel generates a new Dot diagram representation from
// the given model
func DotDiagramFromModel(m *Model) *DotDiagram {
	d := &DotDiagram{}

	for _, e := range m.Entities {
		d.Nodes = append(d.Nodes, DotNodeFromEntity(e))

		for _, a := range e.Attributes {
			if a.Name != "ID" {
				d.Nodes = append(d.Nodes, DotNodeFromAttribute(a))
				d.Links = append(d.Links, DotLinkFromAttribute(e, a))
			}
		}

		for _, r := range e.Relations {
			if r.Name() != "ID" {
				d.Links = append(d.Links, DotLinkFromRelation(e, r))
			}
		}
	}

	return d
}

// DotNodeFromEntity builds a new Dot diagram node from the given entity
func DotNodeFromEntity(e *Entity) *DotNode {
	return DotNodeFromEntityName(e.Name)
}

// DotNodeFromEntityName builds a new Dot diagram node, from the given
// name. The node represents an entity
func DotNodeFromEntityName(n string) *DotNode {
	return &DotNode{
		Name:  strcase.ToSnake(n),
		Label: n,
		Shape: "box",
	}
}

// DotNodeFromAttribute builds a new Dot diagram node from the given
// attribute
func DotNodeFromAttribute(a *Attribute) *DotNode {
	return &DotNode{
		Name:  strcase.ToSnake(a.Name),
		Label: a.Name,
		Shape: "ellipse",
	}
}

// DotNodeFromRelation builds a new Dot diagram node from the given
// relation
func DotNodeFromRelation(r *Relation) *DotNode {
	return &DotNode{
		Name:  strcase.ToSnake(r.Name()),
		Label: r.Name(),
		Shape: "ellipse",
	}
}

// DotLinkFromAttribute builds a new Dot diagram link, between
// the given entity and attribute
func DotLinkFromAttribute(e *Entity, a *Attribute) *DotLink {
	eNode := DotNodeFromEntity(e)
	aNode := DotNodeFromAttribute(a)
	return &DotLink{
		From:  eNode,
		To:    aNode,
		Style: "dotted",
		Label: "",
	}
}

// DotLinkFromRelation builds a new Dot diagram link, between
// the given entity and relation
func DotLinkFromRelation(e *Entity, r *Relation) *DotLink {
	eNode := DotNodeFromEntity(e)
	aNode := DotNodeFromEntityName(r.Entity)
	return &DotLink{
		From:  eNode,
		To:    aNode,
		Style: DotLinkStyleFromRelation(r),
		Label: DotLinkLabelFromRelation(r),
	}
}

// DotLinkStyleFromRelation defines the style to apply to a link,
// according to the type of relation between two entities
func DotLinkStyleFromRelation(r *Relation) string {
	return "bold"
}

// DotLinkLabelFromRelation defines the style to apply to a link,
// according to the type of relation between two entities
func DotLinkLabelFromRelation(r *Relation) string {
	return r.Name()
}

// DotDiagram represents a Graphviz Diagram. It is made of nodes, and
// links
type DotDiagram struct {
	Nodes []*DotNode
	Links []*DotLink
}

// Renders the source for the diagram
func (d *DotDiagram) String() string {
	chunks := []string{}
	chunks = append(chunks, "digraph G {")

	for _, n := range d.Nodes {
		chunks = append(chunks, n.String())
	}

	for _, l := range d.Links {
		chunks = append(chunks, l.String())
	}

	chunks = append(chunks, "}")
	return strings.Join(chunks, "\n")
}

// DotNode represents a node in the diagram
type DotNode struct {
	Name  string
	Label string
	Shape string
}

// Renders the source for the node
func (n *DotNode) String() string {
	return fmt.Sprintf("    %s [shape=%s,label=\"%s\"];", n.Name, n.Shape, n.Label)
}

// DotLink represents a link between nodes
type DotLink struct {
	From  *DotNode
	To    *DotNode
	Label string
	Style string
}

// Renders the source for the link
func (l *DotLink) String() string {
	return fmt.Sprintf("    %s -> %s [style=%s,label=\"%s\"];", l.From.Name, l.To.Name, l.Style, l.Label)
}
