BIN = volume

all: clean build

build: deps
	go build -o build/$(BIN) ./cmd/$(BIN)

install: deps
	go install ./...

deps:
	go get -d -v ./...

test: testdeps build
	go test -v .
	go test -v ./cmd/volume

testdeps:
	go get -d -v -t ./...

lint: lintdeps
	go vet
	golint -set_exit_status ./...

lintdeps:
	go get -d -v -t .
	go get -u github.com/golang/lint/golint

clean:
	rm -rf build
	go clean

.PHONY: build install deps test testdeps lint lintdeps clean
