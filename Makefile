PREFIX ?= /usr/local
BINDIR  = $(PREFIX)/bin

.PHONY: build install install-user uninstall test lint clean release-dry-run

build:
	go build -o bin/ccs ./cmd/ccs/

# Install to $(PREFIX)/bin (default /usr/local/bin — usually needs sudo).
# Override with PREFIX=, e.g. `PREFIX=$$HOME/.local make install`.
install: build
	install -d "$(BINDIR)"
	install -m 0755 bin/ccs "$(BINDIR)/ccs"
	@echo "Installed $(BINDIR)/ccs"

# User-local install — no sudo needed; assumes ~/.local/bin is on PATH.
install-user:
	$(MAKE) install PREFIX=$(HOME)/.local

uninstall:
	rm -f "$(BINDIR)/ccs"
	@echo "Removed $(BINDIR)/ccs"

test:
	go test ./...

lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || { echo "Files need gofmt:"; gofmt -l .; exit 1; }

clean:
	rm -rf bin/

release-dry-run:
	goreleaser release --snapshot --clean
