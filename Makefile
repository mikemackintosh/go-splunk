all: test build

test:

run:
	go run cmd/main.go

.PHONY: bin
bin:
	go build -o bin/go-splunk cmd/main.go
