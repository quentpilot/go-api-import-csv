# Import and read a CSV file using Gin and RabbitMQ
This project is a sandbox to test the power of Go.

I wish I could compare performances between a Go and a PHP application.

# Installation
## Makefile
Run command `make run` to build docker and run all services.

## Docker
Run command `docker-compose up --build`

# Roadmap
1. Implement testing
2. Benchmark performances with time and memory consumption
    1. Tests with empty datatable to insert 1M rows
    2. Tests with already filled datatable of 1M rows or more
3. Add PHP/Symfony Docker service to create the same API/Test/Benchmark scenario
