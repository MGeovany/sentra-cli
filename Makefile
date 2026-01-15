.PHONY: scan build

scan:
	go run ./cmd/sentra scan

build:
	go build -o bin/sentra ./cmd/sentra
