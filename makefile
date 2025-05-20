# Project name
PROJECT_NAME := go-api-import-csv

# Run API in development mode
dev-api:
	docker-compose up api

# Run RabbitMQ worker to treat uploaded file
dev-worker:
	docker-compose up worker

# Run both services
dev:
	docker-compose up

# Build all services
build:
	docker-compose build

# Build and run all services
dev-build:
	docker-compose up --build

# Clean docker
clean:
	docker-compose down -v --remove-orphans

# Run local tests
test-local:
	go test -v ./...

# Run test into Docker
test:
	docker-compose run --rm test

# Lint the code
lint:
	go vet ./...

# Format the code
fmt:
	go fmt ./...

# Build API executable
build-binary-api:
	go build -o bin/api ./cmd/api

# Build Worker executable
build-binary-worker:
	go build -o bin/worker ./cmd/worker

# Build all executables
build-binary: build-api build-worker

# Run API locally
run-api:
	go run ./cmd/api

# Run Worker locally
run-worker:
	go run ./cmd/worker

.PHONY: dev dev-api dev-worker build build-all build-api build-worker clean test test-local lint fmt run-api run-worker