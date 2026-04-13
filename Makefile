.PHONY: build install test lint clean release-dry-run

build:
	go build -o bin/ccs ./cmd/ccs/

install:
	go install ./cmd/ccs/

test:
	go test ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || { echo "Files need gofmt:"; gofmt -l .; exit 1; }

clean:
	rm -rf bin/

release-dry-run:
	goreleaser release --snapshot --clean
