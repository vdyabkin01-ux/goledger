-include .env
DB_URL = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable

.PHONY: test fmt vet run check lint migrate-up migrate-down

test:
	go test -count=1 ./...

fmt:
	gofmt -w .

vet:
	go vet ./...	

run:
	go run .

lint:
	golangci-lint run	


check:
	@[ -z "$$(gofmt -l .)" ] || (echo "run 'make fmt':" && gofmt -l . && exit 1)
	go vet ./...
	golangci-lint run
	go test -count=1 ./...

migrate-up:
	@migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	@migrate -path migrations -database "$(DB_URL)" down 1	
