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
