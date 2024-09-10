# Variables
BINARY_NAME=telemetry
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build flags
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

# Platforms to build for
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: all build build-all docker-local clean test

all: build

# Build for the current platform
build:
	@echo "Building for host platform..."
	go build ${LDFLAGS} -o ${BINARY_NAME}

# Build for all platforms
build-all:
	@for platform in ${PLATFORMS}; do \
		echo "Building for $$platform..."; \
		GOOS=$$(echo $$platform | cut -d/ -f1) \
		GOARCH=$$(echo $$platform | cut -d/ -f2) \
		go build ${LDFLAGS} -o ${BINARY_NAME}_$${GOOS}_$${GOARCH}; \
	done

# Build and run local Docker container
docker-local:
	@echo "Building and running local Docker container..."
	docker build -t ${BINARY_NAME}:local \
		--build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIME=${BUILD_TIME} .
	docker run -p 8080:8080 ${BINARY_NAME}:local

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean up built binaries
clean:
	@echo "Cleaning up..."
	rm -f ${BINARY_NAME}*

# Help target
help:
	@echo "Available targets:"
	@echo "  build       : Build for host platform"
	@echo "  build-all   : Build for all defined platforms"
	@echo "  docker-local: Build and run local Docker container"
	@echo "  test        : Run tests"
	@echo "  clean       : Remove built binaries"
	@echo "  help        : Show this help message"
