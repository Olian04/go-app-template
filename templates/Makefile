.PHONY: [[ if .Binary ]] build run [[ end ]] format lint test test-race help

# Only list trees this mode actually renders: a missing path fails the whole
# target. `internal/` is unconditional because the domain model ships in every mode.
SOURCE_CODE ?= ./internal/... ./test/unit/...[[- if modeIs "cli" "cli-library" "http" ]] ./cmd/...[[- end ]][[- if modeIs "library" "cli-library" ]] ./pkg/...[[- end ]]
REV := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_OUTPUT_DIR := ./dist

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "%-24s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

[[- if .Binary ]]
build: ## Build [[.Binary.Name]] binary from ./cmd/[[.Binary.Name]]
	mkdir -p $(BUILD_OUTPUT_DIR)
	go build -trimpath -ldflags "-s -w -X [[.Binary.VersionPackage]].Revision=$(REV) -X [[.Binary.VersionPackage]].BuildTime=$(BUILD_TIME)" -o $(BUILD_OUTPUT_DIR)/[[.Binary.Name]] ./cmd/[[.Binary.Name]]

run: ## Run [[.Binary.Name]] from ./cmd/[[.Binary.Name]]
	go run ./cmd/[[.Binary.Name]]
[[- end ]]

format: ## Run go fmt and gofmt
	go fmt ./...
	gofmt -w .

lint: ## Run go vet, module verify, vuln scan, golangci
	go vet ./...
	go mod verify
	go tool govulncheck $(SOURCE_CODE)
	go tool golangci-lint run $(SOURCE_CODE)

test: ## Run unit tests
	go test -shuffle=on -timeout 120s $(SOURCE_CODE)

test-race: ## Run unit tests with race detector
	go test -race -shuffle=on -timeout 180s $(SOURCE_CODE)
