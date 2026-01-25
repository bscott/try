.PHONY: build test clean install fmt vet

# Build variables
BINARY_NAME=try
CMD_DIR=./cmd/try
INSTALL_PATH=/usr/local/bin

# Build the binary
build:
	go build -o $(BINARY_NAME) $(CMD_DIR)

# Run tests
test:
	go test -v -race ./...

# Run tests with coverage
coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Install binary to system path
install: build
	sudo mv $(BINARY_NAME) $(INSTALL_PATH)/

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Run all checks
check: fmt vet test

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	GOOS=linux GOARCH=amd64 go build -o dist/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	GOOS=linux GOARCH=arm64 go build -o dist/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build -o dist/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

# Help target
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  test       - Run tests"
	@echo "  coverage   - Run tests with coverage report"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  install    - Install binary to system path"
	@echo "  clean      - Clean build artifacts"
	@echo "  check      - Run fmt, vet, and test"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  help       - Show this help message"
