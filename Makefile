
module: module.tar.gz

bin/viamseakeepermodule: go.mod *.go cmd/module/*.go
	go build -o bin/viamseakeepermodule cmd/module/cmd.go

lint:
	gofmt -s -w .

sample: bin/viamseakeeper
	./bin/viamseakeeper data/sample.json

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test ./...


module.tar.gz: bin/viamseakeepermodule
	tar czf $@ $^

all: test bin/viamseakeeper module 


