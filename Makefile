.PHONY: default build test run

default: build

include Protobuf.mk

generate:
	go install github.com/tinylib/msgp@latest
	go generate ./...
test:
	go test ./...

build:
	go build -o dist/mixtape ./cmd/mixtape

build-prototypes:
	go build -o prototypes/thenet ./prototypes/thenet

run:
	./cmd/mixtape
