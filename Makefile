.PHONY: default build test run

default: build
test:
	go test ./...

build:
	go build -o dist/mixtape ./cmd/mixtape

run:
	./cmd/mixtape
