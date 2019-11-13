# Flootic GraphDB generator

## Build

```
go build .
chmod +x ./generator
```
## Generate

```
mkdir -p $GOPATH/src/github.com/flootic/graphdb
./generator --model=flootic.yml --output=$GOPATH/src/github.com/flootic/graphdb
```

## Run your server

```
cd $GOPATH/src/github.com/flootic/graphdb
go get
go run .
```

GraphiQL should be available at: `http://localhost:8080/`
