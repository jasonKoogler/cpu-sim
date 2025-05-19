.PHONY: all build clean run test

BINARY_NAME=multicore-sim
BUILD_DIR=./build
CONFIG_PATH=./configs/default.yaml
CYCLES=1000

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
	@$(BUILD_DIR)/$(BINARY_NAME) --config $(CONFIG_PATH) --cycles $(CYCLES)

run-verbose: build
	@echo "Running multicore processor simulator with verbose output..."
	@$(BUILD_DIR)/$(BINARY_NAME) --config $(CONFIG_PATH) --cycles $(CYCLES) --v

pipeline: build
	@echo "Showing pipeline structure..."
	@$(BUILD_DIR)/$(BINARY_NAME) --config $(CONFIG) --show-pipeline

# Run with deep pipeline configuration
deep-pipeline: build
	@echo "Running with deep pipeline configuration..."
	@$(BUILD_DIR)/$(BINARY_NAME) --config configs/deep-pipeline.yaml --cycles $(CYCLES) --show-pipeline


test:
	@echo "Running tests..."
	@go test -v ./...
	@echo "Tests complete!"