BINARY_NAME=ytui
BUILD_DIR=bin
CMD_DIR=cmd/ytui
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build run install clean lint test fmt vet

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

install:
	go install $(LDFLAGS) ./$(CMD_DIR)

clean:
	rm -rf $(BUILD_DIR)
	go clean

lint:
	@which golangci-lint > /dev/null 2>&1 || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"
	golangci-lint run ./...

test:
	go test -v -race ./...

fmt:
	go fmt ./...

vet:
	go vet ./...
