BINARY_NAME  := prayer-times
ALIAS_NAME   := pt
CMD_PATH     := ./cmd/prayer-times
ALIAS_PATH   := ./cmd/pt
BIN_DIR      := bin
DIST_DIR     := dist

VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS      := -s -w -X main.version=$(VERSION)

PLATFORMS    := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: build test vet clean release install help

## build: compile both binaries for the current platform
build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(ALIAS_NAME) $(ALIAS_PATH)
	@echo "Built $(BIN_DIR)/$(BINARY_NAME) and $(BIN_DIR)/$(ALIAS_NAME)"

## test: run all tests with race detector
test:
	go test -v -race ./...

## vet: run go vet
vet:
	go vet ./...

## clean: remove build artifacts
clean:
	rm -rf $(BIN_DIR) $(DIST_DIR)

## release: cross-compile for all platforms and create tarballs
release: clean
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		output="$(BINARY_NAME)_$${GOOS}_$${GOARCH}"; \
		echo "Building $${output}..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 \
			go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME) $(CMD_PATH); \
		GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 \
			go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(ALIAS_NAME) $(ALIAS_PATH); \
		tar -czf $(DIST_DIR)/$${output}.tar.gz -C $(DIST_DIR) $(BINARY_NAME) $(ALIAS_NAME); \
		rm $(DIST_DIR)/$(BINARY_NAME) $(DIST_DIR)/$(ALIAS_NAME); \
	done
	@cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt
	@echo "Release artifacts in $(DIST_DIR)/:"
	@ls -lh $(DIST_DIR)/

## install: install both binaries to $GOPATH/bin (or $HOME/go/bin)
install:
	go install -ldflags "$(LDFLAGS)" $(CMD_PATH)
	go install -ldflags "$(LDFLAGS)" $(ALIAS_PATH)
	@echo "Installed $(BINARY_NAME) and $(ALIAS_NAME) to $$(go env GOPATH)/bin/"

## help: show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
