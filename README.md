# Flootic GraphDB generator

## Build

```
go build .
chmod +x ./generator
```

## Generate

```
mkdir -p $GOPATH/src/github.com/flootic/graphdb
./generator --db=postgres --model=flootic.yml --output=$GOPATH/src/github.com/flootic/graphdb
```

If you omit the `db` option, then the app will be optimized for
`sqlite3`

Selecting `postgres` also makes the generated app compatible with cockroach.

## Run your server

```
cd $GOPATH/src/github.com/flootic/graphdb
go get
go run . -db=postgres://localhost/flootic?sslmode=disable
```

If you omit the `db` option, then the app will attempt to start an in
memory sqlite3 database. The project must be built for sqlite3.

## Test

GraphiQL should be available at: `http://localhost:8080/`


## Hooks

It is possible to define user defined hooks. For example:

```yaml
- name: Login
  attributes:
    ...
  hooks:
    create:
       - after
```

This will force you to implement a function named `AfterCreateLogin`.
This can be useful to instruct the server to perform user authentication
and issue a token:

```go
package main

import (
    "database/sql"
    "errors"
)

// AfterCreateLogin is a user defined hook that creates a token, or
// returns a login error
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

