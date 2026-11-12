.PHONY: test fmt vet run check

test:
	go test -count=1 ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

run:
	go run .


check:
	@[ -z "$$(gofmt -l .)" ] || (echo "run 'make fmt':" && gofmt -l . && exit 1)
	go vet ./...
	go test -count=1 ./...