.PHONY: build format lint run test test-race help
[[- if modeIs "http" ]]
.PHONY: build-[[.ServiceName]] run-[[.ServiceName]]
[[- end ]]
[[- if modeIs "cli" "cli-library" ]]
.PHONY: build-[[.CliName]] run-[[.CliName]]
[[- end ]]

SOURCE_CODE ?= ./internal/... ./test/unit/...[[- if modeIs "cli" "cli-library" "http" ]] ./cmd/...[[- end ]][[- if modeIs "library" "cli-library" ]] ./pkg/...[[- end ]]
REV := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_OUTPUT_DIR := ./dist

help: ## Show available make targets
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "%-24s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build selected binaries
build:[[- if modeIs "http" ]] build-[[.ServiceName]][[- end ]][[- if modeIs "cli" "cli-library" ]] build-[[.CliName]][[- end ]]
[[- if modeIs "library" ]]
	@echo "no binaries selected (library mode)"
[[- end ]]

[[- if modeIs "http" ]]
build-[[.ServiceName]]: ## Build [[.ServiceName]] binary from ./cmd/[[.ServiceName]]
	mkdir -p $(BUILD_OUTPUT_DIR)
	go build -trimpath -ldflags "-s -w -X [[.ModulePath]]/cmd/[[.ServiceName]]/version.Revision=$(REV) -X [[.ModulePath]]/cmd/[[.ServiceName]]/version.BuildTime=$(BUILD_TIME)" -o $(BUILD_OUTPUT_DIR)/[[.ServiceName]] ./cmd/[[.ServiceName]]

run-[[.ServiceName]]: ## Run [[.ServiceName]] from ./cmd/[[.ServiceName]]
	go run ./cmd/[[.ServiceName]]
[[- end ]]
[[- if modeIs "cli" "cli-library" ]]
build-[[.CliName]]: ## Build [[.CliName]] binary from ./cmd/[[.CliName]]
	mkdir -p $(BUILD_OUTPUT_DIR)
	go build -trimpath -ldflags "-s -w -X [[.ModulePath]]/cmd/[[.CliName]]/version.Revision=$(REV) -X [[.ModulePath]]/cmd/[[.CliName]]/version.BuildTime=$(BUILD_TIME)" -o $(BUILD_OUTPUT_DIR)/[[.CliName]] ./cmd/[[.CliName]]

run-[[.CliName]]: ## Run [[.CliName]] from ./cmd/[[.CliName]]
	go run ./cmd/[[.CliName]]
[[- end ]]
[[- if modeIs "http" ]]
run: ## Run [[.ServiceName]] (default)
	go run ./cmd/[[.ServiceName]]
[[- else if modeIs "cli" "cli-library" ]]
run: ## Run [[.CliName]] (default)
	go run ./cmd/[[.CliName]]
[[- else ]]
run: ## No runnable binary (library mode)
	@echo "no runnable binary (library mode)"
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
