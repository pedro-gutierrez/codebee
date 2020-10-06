# Codebee

A GraphQL API and Database generator written in Go.


## Build

Build the `codebee` binary:

```
go build .
chmod +x ./codebee
```

## Write a model

See `examples/betting.yml` to get an idea of what a model looks like

## Generate

```
mkdir -p ~/Projects/betting
./codebee --db=postgres --model=examples/betting.yml --output=~/Projects/betting
```

If you omit the `db` option, then the app will be optimized for
`sqlite3`

Selecting `postgres` also makes the generated app compatible with CockroachDB.

## Create your database

Assuming you have Postgres up and running:

```
createuser betting
createdb betting -O betting
```

## Run your server

```
cd ~/Projects/betting
go get
go run . -db=postgres://betting@localhost/betting?sslmode=disable
```

If you omit the `db` option, then the app will attempt to start an in
memory sqlite3 database. The project must be built for sqlite3.

## Test

GraphiQL should be available at: `http://localhost:8080/`


## Hooks

It is possible to add custom logic via user defined hooks. 

This is useful to plugin application behavior in, before and/or after entities are created, updated or
deleted.

For example:

```yaml
- name: Login
  attributes:
    ...
  hooks:
    create:
       - after
```

This will **force** you to implement a function named `AfterCreateLogin`.
The following example instructs the server to perform user credentials verification,
and issue a token:

```go
package main

import (
    "database/sql"
    "errors"
)

// AfterCreateLogin is a user defined hook that creates a token, or
// returns a login error. In an after hook, the entity involved has
// already been persisted to the database, so it is possible to link
// extra resources to it.
//
// TODO: make this part of a database transaction
func AfterCreateLogin(db *sql.DB, l *Login) error {

    // Look for the user. If no user was found, or the password
    // does not match, then return an error
    // TODO: check hashed passwords
    creds, err := FindCredentialsByUsername(db, l.Username)
    if err != nil || creds == nil || creds.Password != l.Password {
        return errors.New("Invalid login")
    }

    // Create a token and persist into the Database
    _, err = CreateToken(db, &Token{
        Expires:     3600,
        Permissions: "*",
        ID:          l.ID,
        Login:       l,
        Owner:       creds.Owner,
    })

    return err
}

```

You will need to implement your hooks in the `main` package. This has
the advantage of easier pluggability, and in return, you get access to
all repository functions in the entire model.

A `before` hook could also be defined, for example, in order to check of the current load in the system, and deny the login request for that user, or all users. If the hook returns an error, the flow is interrumpted and returned immediately. 

This hooks feature opens the door for many features, such as congestion control, back pressure, security, traceability and real, loosely coupled microservice architectures based on streams, by publishing to NATS using `after` hooks.
