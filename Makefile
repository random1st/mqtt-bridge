# Makefile for github.com/random1st/mqtt-bridge

# Go parameters
GO       := go
GOCMD    := $(GO)
GOBUILD  := $(GOCMD) build
GOTEST   := $(GOCMD) test
GOCLEAN  := $(GOCMD) clean
GOMODTIDY:= $(GOCMD) mod tidy
BIN_DIR  := bin
BIN_NAME := mqtt-bridge

MAIN_PKG := ./cmd/bridge

GOLANGCI_LINT := golangci-lint

.PHONY: all build  test lint runclean install-lint help docker-build

all: build build-test

help:
	@echo "Usage:"
	@echo "  make build        - Build the main application."
	@echo "  make test         - Run 'go test' on all packages."
	@echo "  make lint         - Run golangci-lint."
	@echo "  make run          - Build and run the main app."
	@echo "  make clean        - Remove temporary files and binaries."
	@echo "  make install-lint - Install golangci-lint."
	@echo "  make help         - Show this help message."

build:
	@echo "==> Building the main binary..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BIN_NAME) $(MAIN_PKG)

test:
	@echo "==> Running go test..."
	$(GOTEST) ./... -v

lint:
	@echo "==> Running golangci-lint..."
	$(GOLANGCI_LINT) run ./...

install-lint:
	@echo "==> Installing golangci-lint..."
	GO111MODULE=on $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

run: build
	@echo "==> Running the main binary..."
	@./$(BIN_DIR)/$(BIN_NAME)

clean:
	@echo "==> Cleaning..."
	$(GOCLEAN)
	rm -rf $(BIN_DIR)

docker-build:
	@echo "==> Building Docker image..."
	docker buildx build   --platform linux/amd64  -t mqtt-bridge:latest .
	@echo "==> Docker image built successfully."
	@echo "==> To run the Docker container, use:"
	@echo "    docker run -it --rm mqtt-bridge:latest"