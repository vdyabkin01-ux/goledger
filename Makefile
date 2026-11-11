.PHONY: run test test-integration fmt vet check db-up db-down

# Load local settings from .env (if it exists) and pass them to commands.
-include .env
export

## run: start the API server locally (needs .env and a running database)
run:
	go run ./cmd/server

## db-up: start local PostgreSQL in Docker and wait until it is ready
db-up:
	docker compose up -d --wait db

## db-down: stop local PostgreSQL (data is kept in a Docker volume)
db-down:
	docker compose down

## test: run all unit tests with the race detector
test:
	go test -race -count=1 ./...

## test-integration: run integration tests against PostgreSQL in Docker (needs Docker running)
test-integration:
	go test -race -count=1 -tags=integration -v ./internal/storage/...

## fmt: format all Go files
fmt:
	gofmt -w .

## vet: run static checks
vet:
	go vet ./...

## check: everything CI runs, before you push
check:
	@test -z "$$(gofmt -l .)" || (echo "run 'make fmt':" && gofmt -l . && exit 1)
	go vet ./...
	go test -race -count=1 ./...
