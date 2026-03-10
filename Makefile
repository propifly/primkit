.PHONY: all build build-pi test lint fmt tidy clean docs docs-check check-registration install-hooks sync-agent-docs

# Build both binaries for the current platform.
all: tidy fmt lint test build docs-check check-registration

VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS_TASK   := -ldflags "-X github.com/propifly/primkit/taskprim/internal/cli.Version=$(VERSION)"
LDFLAGS_STATE  := -ldflags "-X github.com/propifly/primkit/stateprim/internal/cli.Version=$(VERSION)"
LDFLAGS_KNOW   := -ldflags "-X github.com/propifly/primkit/knowledgeprim/internal/cli.Version=$(VERSION)"
LDFLAGS_QUEUE  := -ldflags "-X github.com/propifly/primkit/queueprim/internal/cli.Version=$(VERSION)"

build:
	cd taskprim && go build $(LDFLAGS_TASK) -o ../bin/taskprim ./cmd/taskprim
	cd stateprim && go build $(LDFLAGS_STATE) -o ../bin/stateprim ./cmd/stateprim
	cd knowledgeprim && go build $(LDFLAGS_KNOW) -o ../bin/knowledgeprim ./cmd/knowledgeprim
	cd queueprim && go build $(LDFLAGS_QUEUE) -o ../bin/queueprim ./cmd/queueprim

# Cross-compile for Raspberry Pi (ARM64 Linux).
build-pi:
	cd taskprim && GOOS=linux GOARCH=arm64 go build $(LDFLAGS_TASK) -o ../bin/taskprim-linux-arm64 ./cmd/taskprim
	cd stateprim && GOOS=linux GOARCH=arm64 go build $(LDFLAGS_STATE) -o ../bin/stateprim-linux-arm64 ./cmd/stateprim
	cd knowledgeprim && GOOS=linux GOARCH=arm64 go build $(LDFLAGS_KNOW) -o ../bin/knowledgeprim-linux-arm64 ./cmd/knowledgeprim
	cd queueprim && GOOS=linux GOARCH=arm64 go build $(LDFLAGS_QUEUE) -o ../bin/queueprim-linux-arm64 ./cmd/queueprim

test:
	cd primkit && go test -v -race -count=1 ./...
	cd taskprim && go test -v -race -count=1 ./...
	cd stateprim && go test -v -race -count=1 ./...
	cd knowledgeprim && go test -v -race -count=1 ./...
	cd queueprim && go test -v -race -count=1 ./...

lint:
	cd primkit && go vet ./...
	cd taskprim && go vet ./...
	cd stateprim && go vet ./...
	cd knowledgeprim && go vet ./...
	cd queueprim && go vet ./...

fmt:
	cd primkit && gofmt -s -w .
	cd taskprim && gofmt -s -w .
	cd stateprim && gofmt -s -w .
	cd knowledgeprim && gofmt -s -w .
	cd queueprim && gofmt -s -w .

tidy:
	cd primkit && go mod tidy
	# go mod tidy adds primkit (workspace-local private module) to go.mod in each prim;
	# strip it so CI doesn't try to download/verify the workspace-local module version.
	for prim in taskprim stateprim knowledgeprim queueprim; do \
		cd $$prim && go mod tidy && go mod edit -droprequire=github.com/propifly/primkit/primkit && grep -v "propifly/primkit/primkit" go.sum > go.sum.tmp && mv go.sum.tmp go.sum || true && cd ..; \
	done

clean:
	rm -rf bin/

# Documentation generation
docs:
	bash scripts/docgen.sh

docs-check:
	bash scripts/docgen.sh --check

# Registration validation
check-registration:
	bash scripts/check-registration.sh

# Install git hooks (run once after cloning)
install-hooks:
	bash scripts/install-hooks.sh

# Copy agent-facing docs to clawson-config (run after 'make docs' when docs change)
sync-agent-docs:
	bash scripts/sync-agent-docs.sh
