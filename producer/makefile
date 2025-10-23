# Simple Makefile for Kafka Go Producer

APP_NAME=go_producer

run:
	@echo "🚀 Running $(APP_NAME)..."
	go run .

build:
	@echo "🏗️  Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) .

tidy:
	go mod tidy

lint:
	golangci-lint run || true

proto:
	protoc --go_out=. --go_opt=paths=source_relative user.proto

clean:
	rm -rf bin
