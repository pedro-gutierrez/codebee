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
