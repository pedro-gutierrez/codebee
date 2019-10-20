# Flootic Code Generator

The goal of this proyect is to provide with a custom code generator so that we can:

* Define our business model in an abstract way (YAML)
* Generate a GraphQL schema out of it, and then use it in a gqlgen proyect.
* Genereate resolvers for gqlgen.
* Generate the flootic repo layer out of it.
* Generate the SQL schema for various platforms (Cockroach, Sqlite, etc...)
* Generate seed data for the database
* Generate the model as Golang structs
* Implement our code style conventions.
* Instrument code

This will save us a lot of time in maintenance and will dramatically increase consistency. We will be able to have our very own FaunaDB clone running in premises with the right features we want (sorting, pagination, referential integrity).

# Flootic model

Everything should be derived from a technology agnostic model and YAML is probably the best format for this.

The suggested reference at the moment is `flootic.yml`. 

# Supported generators

| Generator       | Status        |
| --------------- |:-------------:|
| Repo            | in progress   |
| GraphQL schema  | in progress   |

# Example

Have a look at `flootic.yml`. Then run `go test` and you should see the following files being created:

* `generated/flootic.graphql`
* `generated/repo.go`

# Libraries

We use [Jennifer](https://github.com/dave/jennifer) as our main Golang code generator. The API is fluent and well documented. 

If we need to write our custom generators (eg. GraphQL schema), it is advised to model them after Jennifer's API.
