# Using make as a task runner.

.DEFAULT_GOAL := help

SHELL := /bin/bash
PWD := $(shell pwd)
CI_CONTAINER_IMAGE_NAME ?= nmfr/sqlc-template

# By default -count=1 for no cache.
# -p number of paralel processes allowed
#  -cover to include coverage %
GO_TEST_FLAGS?=-count=1 -p=4 -cover

# make help # Display available commands.
# Only comments starting with "# make " will be printed.
.PHONY: help
help:
	@egrep "^# make " [Mm]akefile | cut -c 3-

# make tidy # Tidy the go module files.
.PHONY: tidy
tidy:
	go mod tidy

# make fmt # Format and fix (if possible) all .go files.
.PHONY: fmt
fmt:
	go fmt ./...

# make lint # Lint the code base searching for formatting or known bad patterns.
.PHONY: lint
lint:
	GOFLAGS=-buildvcs=false golangci-lint run

# make test # Run tests.
.PHONY: test
test:
	go test $(GO_TEST_FLAGS) ./...

# Generate protobuf code for every "*.proto" file in this repository.
.PHONY: generate-protobuf
generate-protobuf:
	find . -type f -name "*.proto" -exec protoc --go_out=. --go_opt=paths=source_relative "{}" \;

# make clean # Clean up the previous build artifacts.
.PHONY: clean
clean:
	rm -rf ./bin

# make build # Build the go binary.
.PHONY: build
build: clean generate-protobuf
	mkdir -p bin
	GOOS=wasip1 GOARCH=wasm go build -o bin/sqlc-template.wasm cmd/sqlc-template/main.go

# make container run="<command>" # Run a command from inside the container. Examples: `make container run="make spell-check"`.
.PHONY: container
container:
# If caching is enabled attempt to pull the container from the registry to fill the cache before the build.
	[[ "$$USE_CONTAINER_CACHE" == "true" ]] && (docker pull $(CI_CONTAINER_IMAGE_NAME)) || true
	docker build --target ci --tag $(CI_CONTAINER_IMAGE_NAME) --cache-from=$(CI_CONTAINER_IMAGE_NAME) --build-arg BUILDKIT_INLINE_CACHE=1 .
	docker run --init -v "$(CURDIR):/opt/app" $(CI_CONTAINER_IMAGE_NAME) $(run)
