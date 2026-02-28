BINARY_NAME := tmux-prayer-times
CMD_PATH    := ./cmd/tmux-prayer-times
BIN_DIR     := bin
DIST_DIR    := dist

VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS     := -s -w -X main.version=$(VERSION)

PLATFORMS   := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

.PHONY: build test vet clean release install

## build: compile for the current platform
build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY_NAME) $(CMD_PATH)
	@echo "Built $(BIN_DIR)/$(BINARY_NAME)"

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
		tar -czf $(DIST_DIR)/$${output}.tar.gz -C $(DIST_DIR) $(BINARY_NAME); \
		rm $(DIST_DIR)/$(BINARY_NAME); \
	done
	@cd $(DIST_DIR) && shasum -a 256 *.tar.gz > checksums.txt
	@echo "Release artifacts in $(DIST_DIR)/:"
	@ls -lh $(DIST_DIR)/

## install: build and install to plugin bin/ for local testing
install: build
	@echo "Binary ready at $(BIN_DIR)/$(BINARY_NAME)"
	@echo "To test in tmux, run:  ./prayer-times.tmux"

## help: show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
