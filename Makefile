.PHONY: all build clean run test

BINARY_NAME=multicore-sim
BUILD_DIR=./build
CONFIG_PATH=./configs/default.yaml

all: clean build

build:
	@echo "Building multicore processor simulator..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/simulator
	@echo "Build complete!"

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@echo "Cleanup complete!"

run: build
	@echo "Running multicore processor simulator..."
	@$(BUILD_DIR)/$(BINARY_NAME) --config $(CONFIG_PATH)

run-verbose: build
	@echo "Running multicore processor simulator with verbose output..."
	@$(BUILD_DIR)/$(BINARY_NAME) --config $(CONFIG_PATH) --v

test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests complete!"