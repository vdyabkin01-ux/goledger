.PHONY: run test fmt vet check

## run: start the API server locally (PORT defaults to 8080)
run:
	go run ./cmd/server

## test: run all unit tests with the race detector
test:
	go test -race -count=1 ./...

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
