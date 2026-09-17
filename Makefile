.PHONY: run test dist fmt

run:
	go run .

test:
	go test ./...

fmt:
	gofmt -w .

dist:
	./scripts/dist.sh
