# Root Makefile for Go Kafka Producer + Consumer

.PHONY: all build-producer build-consumer run-producer run-consumer proto-producer proto-consumer clean

# Directories
PRODUCER_DIR=producer
CONSUMER_DIR=consumer
BIN_DIR=bin

# Default target
all: build-producer build-consumer

# Ensure bin directory exists
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# ------------------------
# Producer Targets
# ------------------------

run-producer:
	@echo "🚀 Running producer..."
	cd $(PRODUCER_DIR) && go run .

build-producer: $(BIN_DIR)
	@echo "🏗️ Building producer..."
	cd $(PRODUCER_DIR) && go build -o ../$(BIN_DIR)/producer .

proto-producer:
	@echo "📦 Generating producer protobuf..."
	cd $(PRODUCER_DIR) && protoc --go_out=. --go_opt=paths=source_relative user.proto

# ------------------------
# Consumer Targets
# ------------------------

run-consumer:
	@echo "🚀 Running consumer..."
	cd $(CONSUMER_DIR) && go run .

build-consumer: $(BIN_DIR)
	@echo "🏗️ Building consumer..."
	cd $(CONSUMER_DIR) && go build -o ../$(BIN_DIR)/consumer .

proto-consumer:
	@echo "📦 Generating consumer protobuf..."
	cd $(CONSUMER_DIR) && protoc --go_out=. --go_opt=paths=source_relative user.proto

# ------------------------
# Clean all binaries
# ------------------------
clean:
	@echo "🧹 Cleaning binaries..."
	rm -rf $(BIN_DIR)
