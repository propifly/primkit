.PHONY: all build build-pi test lint fmt tidy clean

# Build both binaries for the current platform.
all: tidy fmt lint test build

build:
	cd taskprim && go build -o ../bin/taskprim ./cmd/taskprim
	cd stateprim && go build -o ../bin/stateprim ./cmd/stateprim

# Cross-compile for Raspberry Pi (ARM64 Linux).
build-pi:
	cd taskprim && GOOS=linux GOARCH=arm64 go build -o ../bin/taskprim-linux-arm64 ./cmd/taskprim
	cd stateprim && GOOS=linux GOARCH=arm64 go build -o ../bin/stateprim-linux-arm64 ./cmd/stateprim

test:
	cd primkit && go test -v -race -count=1 ./...
	cd taskprim && go test -v -race -count=1 ./...
	cd stateprim && go test -v -race -count=1 ./...

lint:
	cd primkit && go vet ./...
	cd taskprim && go vet ./...
	cd stateprim && go vet ./...

fmt:
	cd primkit && gofmt -s -w .
	cd taskprim && gofmt -s -w .
	cd stateprim && gofmt -s -w .

tidy:
	cd primkit && go mod tidy
	cd taskprim && go mod tidy
	cd stateprim && go mod tidy

clean:
	rm -rf bin/
