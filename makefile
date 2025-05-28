# Project name
PROJECT_NAME := go-api-import-csv
SERVICE_NAME_API := api
SERVICE_NAME_WORKER := worker
SERVICE_NAME_AMQP := rabbitmq

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
run:
	docker-compose up --build

start-api:
	docker-compose up -d $(SERVICE_NAME_API)

start-worker:
	docker-compose up -d $(SERVICE_NAME_WORKER)

stop-api:
	docker-compose stop $(SERVICE_NAME_API)

stop-worker:
	docker-compose stop $(SERVICE_NAME_WORKER)

restart-api: stop-api start-api

restart-worker: stop-worker start-worker

start: start-api start-worker

stop: stop-api stop-worker

# Clean docker
clean:
	docker-compose down -v --remove-orphans

# Run local tests
test-local:
	go test ./...

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

reload-conf-api:
	@echo "Send SIGHUP to service $(SERVICE_NAME_API)"; \
	docker-compose exec $(SERVICE_NAME_API) pkill -SIGHUP $(SERVICE_NAME_API)

reload-conf-worker:
	@echo "Send SIGHUP to service $(SERVICE_NAME_WORKER)"; \
	docker-compose exec $(SERVICE_NAME_WORKER) pkill -SIGHUP $(SERVICE_NAME_WORKER)

reload: reload-conf-api reload-conf-worker

interrupt:
	@echo "Send SIGINT to service $(SERVICE_NAME_WORKER)"; \
	docker-compose exec $(SERVICE_NAME_WORKER) pkill -SIGINT $(SERVICE_NAME_WORKER)


.PHONY: dev dev-api dev-worker build build-all build-api build-worker clean test test-local lint fmt run-api run-worker reload-conf-api reload-conf-worker reload