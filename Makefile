.PHONY: test fmt vet run check lint

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
