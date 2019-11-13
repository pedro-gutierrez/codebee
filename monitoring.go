package main

import (
	"fmt"
	. "github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

// CreateMonitoring generates all variables and functions required to
// support monitoring
func CreateMonitoring(p *Package) error {
	f := NewFile(p.Name)
	AddMetrics(p, f)
	AddMetricsRegistrations(p, f)
	return f.Save(p.Filename)
}

// AddMetrics declares all metrics that will be used by the resulting
// application, as vars
func AddMetrics(p *Package, f *File) {
	f.Var().DefsFunc(func(vars *Group) {

		for _, e := range p.Model.Entities {

			DefineMetricsForCreateMutation(e, vars)
			DefineMetricsForUpdateMutation(e, vars)
			DefineMetricsForDeleteMutation(e, vars)

			for _, a := range e.Attributes {
				if a.HasModifier("indexed") && a.HasModifier("unique") {
					DefineMetricsForFinderByAttribute(e, a, vars)
				}
			}

			for _, r := range e.Relations {
				if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
					DefineMetricsForFinderByRelation(e, r, vars)
				}
			}

		}
	})
}

// AddMetricsRegistrations registers all metrics that will be used by the resulting
// application, as vars
func AddMetricsRegistrations(p *Package, f *File) {

	funName := "init"

	f.Func().Id(funName).Params().BlockFunc(func(g *Group) {

		for _, e := range p.Model.Entities {

			RegisterMetricsForCreateMutation(e, g)
			RegisterMetricsForUpdateMutation(e, g)
			RegisterMetricsForDeleteMutation(e, g)

			for _, a := range e.Attributes {
				if a.HasModifier("indexed") && a.HasModifier("unique") {
					RegisterMetricsForFinderByAttribute(e, a, g)
				}
			}

			for _, r := range e.Relations {
				if r.HasModifier("belongsTo") || r.HasModifier("hasOne") {
					RegisterMetricsForFinderByRelation(e, r, g)
				}
			}

		}

	})
}

// DefineMetricsForCreateMutation defines the histograms and counters
// that will hold metrics when creating instances of the given entity
func DefineMetricsForCreateMutation(e *Entity, vars *Group) {

	// an histogram, to track latencies
	vars.Id(CreateMutationHistogramName(e)).Op("=").Add(
		HistogramDefinition(
			CreateMutationHistogramName(e),
			CreateMutationHistogramHelp(e),
		),
	)

	// a counter, to track errors
	vars.Id(CreateMutationErrorCounterName(e)).Op("=").Add(
		CounterDefinition(
			CreateMutationErrorCounterName(e),
			CreateMutationErrorCounterHelp(e),
		),
	)
}

// RegisterMetricsForCreateMutation registers the histograms and counters
// that will hold metrics when creating instances of the given entity
func RegisterMetricsForCreateMutation(e *Entity, g *Group) {
	RegisterMetric(CreateMutationHistogramName(e), g)
	RegisterMetric(CreateMutationErrorCounterName(e), g)
}

// DefineMetricsForUpdateMutation defines the histograms and counters
// that will hold metrics when updating instances of the given entity
func DefineMetricsForUpdateMutation(e *Entity, vars *Group) {

	// an histogram, to track latencies
	vars.Id(UpdateMutationHistogramName(e)).Op("=").Add(
		HistogramDefinition(
			UpdateMutationHistogramName(e),
			UpdateMutationHistogramHelp(e),
		),
	)

	// a counter, to track errors
	vars.Id(UpdateMutationErrorCounterName(e)).Op("=").Add(
		CounterDefinition(
			UpdateMutationErrorCounterName(e),
			UpdateMutationErrorCounterHelp(e),
		),
	)
}

// RegisterMetricsForUpdateMutation registers the histograms and counters
// that will hold metrics when updating instances of the given entity
func RegisterMetricsForUpdateMutation(e *Entity, g *Group) {
	RegisterMetric(UpdateMutationHistogramName(e), g)
	RegisterMetric(UpdateMutationErrorCounterName(e), g)
}

// DefineMetricsForDeleteMutation defines the histograms and counters
// that will hold metrics when deleting instances of the given entity
func DefineMetricsForDeleteMutation(e *Entity, vars *Group) {

	// an histogram, to track latencies
	vars.Id(DeleteMutationHistogramName(e)).Op("=").Add(
		HistogramDefinition(
			DeleteMutationHistogramName(e),
			DeleteMutationHistogramHelp(e),
		),
	)

	// a counter, to track errors
	vars.Id(DeleteMutationErrorCounterName(e)).Op("=").Add(
		CounterDefinition(
			DeleteMutationErrorCounterName(e),
			DeleteMutationErrorCounterHelp(e),
		),
	)
}

// RegisterMetricsForDeleteMutation registers the histograms and counters
// that will hold metrics when deleting instances of the given entity
func RegisterMetricsForDeleteMutation(e *Entity, g *Group) {
	RegisterMetric(DeleteMutationHistogramName(e), g)
	RegisterMetric(DeleteMutationErrorCounterName(e), g)
}

// DefineMetricsForFinderByAttribute defines the histograms and counters
// that will hold metrics when finding instances of the given entity by
// the given attribute
func DefineMetricsForFinderByAttribute(e *Entity, a *Attribute, vars *Group) {

	// an histogram, to track latencies
	vars.Id(FindByAttributeQueryHistogramName(e, a)).Op("=").Add(
		HistogramDefinition(
			FindByAttributeQueryHistogramName(e, a),
			FindByAttributeQueryHistogramHelp(e, a),
		),
	)

	// a counter, to track errors
	vars.Id(FindByAttributeQueryErrorCounterName(e, a)).Op("=").Add(
		CounterDefinition(
			FindByAttributeQueryErrorCounterName(e, a),
			FindByAttributeQueryErrorCounterHelp(e, a),
		),
	)

}

// RegisterMetricsForFinderByAttribute defines the histograms and counters
// that will hold metrics when finding instances of the given entity by
// the given attribute
func RegisterMetricsForFinderByAttribute(e *Entity, a *Attribute, g *Group) {
	RegisterMetric(FindByAttributeQueryHistogramName(e, a), g)
	RegisterMetric(FindByAttributeQueryErrorCounterName(e, a), g)
}

// DefineMetricsForFinderByRelation defines the histograms and counters
// that will hold metrics when finding instances of the given entity by
// the given relation
func DefineMetricsForFinderByRelation(e *Entity, r *Relation, vars *Group) {

	// an histogram, to track latencies
	vars.Id(FindByRelationQueryHistogramName(e, r)).Op("=").Add(
		HistogramDefinition(
			FindByRelationQueryHistogramName(e, r),
			FindByRelationQueryHistogramHelp(e, r),
		),
	)

	// a counter, to track errors
	vars.Id(FindByRelationQueryErrorCounterName(e, r)).Op("=").Add(
		CounterDefinition(
			FindByRelationQueryErrorCounterName(e, r),
			FindByRelationQueryErrorCounterHelp(e, r),
		),
	)
}

// RegisterMetricsForFinderByRelation defines the histograms and counters
// that will hold metrics when finding instances of the given entity by
// the given relation
func RegisterMetricsForFinderByRelation(e *Entity, r *Relation, g *Group) {
	RegisterMetric(FindByRelationQueryHistogramName(e, r), g)
	RegisterMetric(FindByRelationQueryErrorCounterName(e, r), g)
}

// CreateMutationHistogramName returns the variable name of the metric that
// observes latencies for the create mutation for the given entity
func CreateMutationHistogramName(e *Entity) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlCreateMutationName(e),
			"Latencies",
		),
	)
}

// CreateMutationHistogramHelp returns the help for the metric that
// keeps track of latencies for the create mutation for the given entity
func CreateMutationHistogramHelp(e *Entity) string {
	return fmt.Sprintf("Elapsed time in milliseconds to create entities of type %s", e.Name)
}

// CreateMutationErrorCounterName returns the name of the metric that
// counts errors for the create mutation for the given entity
func CreateMutationErrorCounterName(e *Entity) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlCreateMutationName(e),
			"Errors",
		),
	)
}

// CreateMutationErrorCounterHelp returns the help for the metric that
// counts errors for the create mutation for the given entity
func CreateMutationErrorCounterHelp(e *Entity) string {
	return fmt.Sprintf("Errors when creating entities of type %s", e.Name)
}

// UpdateMutationHistogramVar returns the variable name of the metric that
// observes latencies for the update mutation for the given entity
func UpdateMutationHistogramName(e *Entity) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlUpdateMutationName(e),
			"Latencies",
		),
	)
}

// UpdateMutationHistogramHelp returns the help for the metric that
// keeps track of latencies for the update mutation for the given entity
func UpdateMutationHistogramHelp(e *Entity) string {
	return fmt.Sprintf("Elapsed time in milliseconds to update entities of type %s", e.Name)
}

// UpdateMutationErrorCounterName returns the name of the metric that
// counts errors for the create mutation for the given entity
func UpdateMutationErrorCounterName(e *Entity) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlUpdateMutationName(e),
			"Errors",
		),
	)
}

// UpdateMutationErrorCounterHelp returns the help for the metric that
// counts errors for the update mutation for the given entity
func UpdateMutationErrorCounterHelp(e *Entity) string {
	return fmt.Sprintf("Errors when updating entities of type %s", e.Name)
}

// DeleteMutationHistogramName returns the variable name of the metric that
// observes latencies for the delete mutation for the given entity
func DeleteMutationHistogramName(e *Entity) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlDeleteMutationName(e),
			"Latencies",
		),
	)
}

// DeleteMutationHistogramHelp returns the help for the metric that
// keeps track of latencies for the delete mutation for the given entity
func DeleteMutationHistogramHelp(e *Entity) string {
	return fmt.Sprintf("Elapsed time in milliseconds to delete entities of type %s", e.Name)
}

// DeleteMutationErrorCounterName returns the name of the metric that
// counts errors for the delete mutation for the given entity
func DeleteMutationErrorCounterName(e *Entity) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlDeleteMutationName(e),
			"Errors",
		),
	)
}

// DeleteMutationErrorCounterHelp returns the help for the metric that
// counts errors for the update mutation for the given entity
func DeleteMutationErrorCounterHelp(e *Entity) string {
	return fmt.Sprintf("Errors when deleting entities of type %s", e.Name)
}

// FindByAttributeQueryHistogramName returns the variable name of the metric that
// observes latencies for the finder query for the given entity by the
// given attribute
func FindByAttributeQueryHistogramName(e *Entity, a *Attribute) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlFindByAttributeQueryName(e, a),
			"Latencies",
		),
	)
}

// FindByAttributeHistogramHelp returns the help for the metric that
// keeps track of latencies for the finder query for the given entity by
// the given attribute
func FindByAttributeQueryHistogramHelp(e *Entity, a *Attribute) string {
	return fmt.Sprintf("Elapsed time in milliseconds to find entties of type %s by %s", e.Name, a.Name)
}

// FindByAttributeErrorQueryCounterName returns the name of the metric that
// counts errors for the finder query for the given entity and attribute
func FindByAttributeQueryErrorCounterName(e *Entity, a *Attribute) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlFindByAttributeQueryName(e, a),
			"Errors",
		),
	)
}

// FindByAttributeQueryErrorCounterHelp returns the help for the metric that
// counts errors for the finder query for the given entity and attribute
func FindByAttributeQueryErrorCounterHelp(e *Entity, a *Attribute) string {
	return fmt.Sprintf("Errors when finding entities of type %s by %s", e.Name, a.Name)
}

// FindByRelationQueryHistogramName returns the variable name of the metric that
// observes latencies for the finder query for the given entity by the
// given relation
func FindByRelationQueryHistogramName(e *Entity, r *Relation) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlFindByRelationQueryName(e, r),
			"Latencies",
		),
	)
}

// FindByRelationHistogramHelp returns the help for the metric that
// keeps track of latencies for the finder query for the given entity by
// the given relation
func FindByRelationQueryHistogramHelp(e *Entity, r *Relation) string {
	return fmt.Sprintf("Elapsed time in milliseconds to find entities of type %s by %s", e.Name, r.Name())
}

// FindByRelationQueryErrorCounterName returns the name of the metric that
// counts errors for the finder query for the given entity and relation
func FindByRelationQueryErrorCounterName(e *Entity, r *Relation) string {
	return strcase.ToSnake(
		fmt.Sprintf("%s%s",
			GraphqlFindByRelationQueryName(e, r),
			"Errors",
		),
	)
}

// FindByRelationQueryErrorCounterHelp returns the help for the metric that
// counts errors for the finder query for the given entity and attribute
func FindByRelationQueryErrorCounterHelp(e *Entity, r *Relation) string {
	return fmt.Sprintf("Errors when finding entities of type %s by %s", e.Name, r.Name())
}

// HistogramDefinition returns the code required to produce a new
// histogram with the given name, help and a predefined set of latency
// buckets
func HistogramDefinition(name string, help string) *Statement {
	return Qual("github.com/prometheus/client_golang/prometheus", "NewHistogram").Call(
		Qual("github.com/prometheus/client_golang/prometheus", "HistogramOpts").Values(Dict{
			Id("Name"): Lit(name),
			Id("Help"): Lit(help),
			Id("Buckets"): Index().Float64().Values(
				Lit(0),
				Lit(5),
				Lit(10),
				Lit(50),
				Lit(100),
				Lit(250),
				Lit(500),
				Lit(1000),
			),
		}),
	)
}

// CounterDefinition returns the code required to produce a new
// counter with the given name and help
func CounterDefinition(name string, help string) *Statement {
	return Qual("github.com/prometheus/client_golang/prometheus", "NewCounter").Call(
		Qual("github.com/prometheus/client_golang/prometheus", "CounterOpts").Values(Dict{
			Id("Name"): Lit(name),
			Id("Help"): Lit(help),
		}),
	)
}

// RegisterMetric registers the given metric
func RegisterMetric(metric string, g *Group) {
	g.Qual("github.com/prometheus/client_golang/prometheus", "MustRegister").Call(Id(metric))
}
