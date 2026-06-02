SHELL := /bin/bash

BINARY ?= talkkonnect
OUTDIR ?= dist
PKG ?= ./cmd/talkkonnect

# Build metadata (optional)
VERSION ?= dev
GIT_SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo "nogit")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(GIT_SHA)

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make deps-debian      Install build deps on Debian/Raspbian (NO dist-upgrade)"
	@echo "  make deps-opus        Install libopus >= 1.6.1 from source if needed"
	@echo "  make build            Build $(BINARY) into $(OUTDIR)/"
	@echo "  make test             Run tests"
	@echo "  make install          Install into PREFIX (default: /usr/local)"
	@echo "  make clean            Remove build output"
	@echo "  make version          Print version vars"

.PHONY: deps deps-arch
deps:
	./scripts/deps/install.sh

deps-arch:
	./scripts/deps/arch.sh

.PHONY: deps-debian deps-opus
deps-debian:
	sudo ./scripts/deps/debian.sh

deps-opus:
	sudo ./scripts/deps/opus.sh

PKG_CONFIG_PATH := /usr/local/lib/pkgconfig:$(PKG_CONFIG_PATH)
export PKG_CONFIG_PATH
CGO_LDFLAGS ?= -Wl,-rpath,/usr/local/lib
export CGO_LDFLAGS

.PHONY: build
build:
	mkdir -p $(OUTDIR)
	go mod download
	CGO_ENABLED=1 go build -trimpath -ldflags '$(LDFLAGS)' -o $(OUTDIR)/$(BINARY) $(PKG)

.PHONY: test
test:
	go test ./...

PREFIX ?= /usr/local
DESTDIR ?=

.PHONY: install
install: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 $(OUTDIR)/$(BINARY) $(DESTDIR)$(PREFIX)/bin/$(BINARY)

.PHONY: clean
clean:
	rm -rf $(OUTDIR)

.PHONY: version
version:
	@echo "VERSION=$(VERSION)"
	@echo "GIT_SHA=$(GIT_SHA)"
