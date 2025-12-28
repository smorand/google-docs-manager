.PHONY: build check clean fmt help install rebuild test uninstall vet

# Binary name
BINARY_NAME=google-docs-manager
BUILD_DIR=bin

# Build settings
CMD_PATH=./cmd/$(BINARY_NAME)

# Build target
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME):
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Build complete! Binary: $(BUILD_DIR)/$(BINARY_NAME)"

# Rebuild from scratch
rebuild: clean build

# Install binary
install: build
ifndef TARGET
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installation complete!"
else
	@echo "Installing $(BINARY_NAME) to $(TARGET)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(TARGET)/ 2>/dev/null || sudo cp $(BUILD_DIR)/$(BINARY_NAME) $(TARGET)/
	@echo "Installation complete!"
endif

# Uninstall binary
uninstall:
	@echo "Looking for $(BINARY_NAME) in system..."
	@BINARY_PATH=$$(which $(BINARY_NAME) 2>/dev/null); \
	if [ -z "$$BINARY_PATH" ]; then \
		echo "$(BINARY_NAME) not found in PATH"; \
		exit 0; \
	fi; \
	if [ -f "$$BINARY_PATH" ]; then \
		if [ "$$(basename $$(dirname $$BINARY_PATH))" = "bin" ]; then \
			echo "Found $(BINARY_NAME) at $$BINARY_PATH"; \
			echo "Removing..."; \
			sudo rm -f "$$BINARY_PATH"; \
			echo "Uninstallation complete!"; \
		else \
			echo "$(BINARY_NAME) found at $$BINARY_PATH but not in a standard bin directory"; \
			echo "Please remove it manually if needed"; \
		fi; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Clean complete!"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Format complete!"

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete!"

# Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  rebuild    - Clean and rebuild from scratch"
	@echo "  install    - Build and install to /usr/local/bin (or TARGET env variable)"
	@echo "  uninstall  - Remove installed binary"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  check      - Run fmt, vet, and test"
	@echo "  help       - Show this help message"
