# ============================================================================
# Makefile for servora-example
# ============================================================================

ifeq ($(OS),Windows_NT)
    IS_WINDOWS := 1
endif

ifneq (,$(wildcard .env))
    include .env
    export
endif

# ============================================================================
# VARIABLES
# ============================================================================

CURRENT_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
ROOT_DIR    := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

BUF_GO_GEN_TEMPLATE := buf.go.gen.yaml
BUF_TS_GEN_TEMPLATE := buf.typescript.gen.yaml

SERVICE_MODULES      := app/master/service app/worker/service
GO_WORKSPACE_MODULES := api/gen $(SERVICE_MODULES)
GO_LINT_MODULES      ?= $(SERVICE_MODULES)
SRCS_MK              := $(foreach mod,$(SERVICE_MODULES),$(wildcard $(mod)/Makefile))
SERVICE_DIRS         := $(sort $(dir $(realpath $(SRCS_MK))))
BUF_TS_SERVICE_TEMPLATES := $(wildcard $(addsuffix api/buf.typescript.gen.yaml,$(SERVICE_DIRS)))
LINT_GOWORK          ?= auto

GOPATH := $(shell go env GOPATH)
GOVERSION := $(shell go version)

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
# Docker image tags disallow "/" and many symbols. Normalize git-derived version for image tags.
DOCKER_TAG_VERSION_RAW := $(shell printf '%s' "$(VERSION)" | sed -E 's/[^[:alnum:]_.-]+/-/g; s/^[.-]+//; s/-+/-/g; s/[.-]+$$//')
DOCKER_TAG_VERSION := $(if $(DOCKER_TAG_VERSION_RAW),$(DOCKER_TAG_VERSION_RAW),dev)

LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.GitBranch=$(GIT_BRANCH)

COMPOSE := docker compose
COMPOSE_FILES := -f docker-compose.yaml
COMPOSE_APPS_FILES := -f docker-compose.yaml -f docker-compose.apps.yaml
MICROSERVICES := master worker

WEB_APPS :=
WEB_DEV_APP ?=

WEB_PNPM_FILTERS := $(foreach app,$(WEB_APPS),--filter "./web/$(app)")
INFRA_SERVICES := consul jaeger otel-collector
COMPOSE_STACK_SERVICES := $(INFRA_SERVICES) $(MICROSERVICES)
COMPOSE_STACK_DOWN := $(COMPOSE) $(COMPOSE_APPS_FILES) down --remove-orphans
COMPOSE_STACK_RESET := $(COMPOSE) $(COMPOSE_APPS_FILES) down --remove-orphans --volumes

SERVORA_PKG := github.com/Servora-Kit/servora

# Tool versions - override to pin a specific version.
PROTOC_GEN_GO_VERSION              := latest
PROTOC_GEN_GO_GRPC_VERSION         := latest
PROTOC_GEN_GO_HTTP_VERSION         := latest
PROTOC_GEN_TYPESCRIPT_HTTP_VERSION := latest
PROTOC_GEN_GO_ERRORS_VERSION       := latest
PROTOC_GEN_OPENAPI_VERSION         := latest
PROTOC_GEN_VALIDATE_VERSION        := latest
PROTOC_GEN_REDACT_VERSION          := latest
SERVORA_VERSION                    := latest
KRATOS_VERSION                     := latest
GNOSTIC_VERSION                    := latest
BUF_VERSION                        := latest
GOLANGCI_LINT_VERSION              := latest
WIRE_VERSION                       := latest
ENT_VERSION                        := latest
AIR_VERSION                        := latest

define run-in-service-dirs
	@$(foreach dir,$(SERVICE_DIRS),echo "==> $(1) $(dir)" && (cd $(dir) && $(MAKE) $(1)) && ) true
endef

# ============================================================================
# MAIN TARGETS
# ============================================================================

.PHONY: help env init plugin cli dep tidy test cover vet lint lint.go lint.proto lint.ts web.dev
.PHONY: wire ent gen gen.fresh api api-go api-ts openapi build clean
.PHONY: bsr.update bsr.push buf-update buf-push tag tag.api
.PHONY: compose.build compose.up compose.up.infra compose.up.all compose.rebuild compose.stop compose.down compose.reset compose.ps compose.logs
.PHONY: openfga.init openfga.model.validate openfga.model.test openfga.model.apply

env:
	@echo "CURRENT_DIR: $(CURRENT_DIR)"
	@echo "ROOT_DIR: $(ROOT_DIR)"
	@echo "SRCS_MK: $(SRCS_MK)"
	@echo "MICROSERVICES: $(MICROSERVICES)"
	@echo "WEB_APPS: $(WEB_APPS)"
	@echo "GO_WORKSPACE_MODULES: $(GO_WORKSPACE_MODULES)"
	@echo "VERSION: $(VERSION)"
	@echo "GOVERSION: $(GOVERSION)"

init: plugin cli
	@echo "✓ Development environment initialized"

plugin:
	@echo "==> Installing protoc plugins..."
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	@go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@$(PROTOC_GEN_GO_HTTP_VERSION)
	@go install github.com/go-kratos/protoc-gen-typescript-http@$(PROTOC_GEN_TYPESCRIPT_HTTP_VERSION)
	@go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@$(PROTOC_GEN_GO_ERRORS_VERSION)
	@go install github.com/google/gnostic/cmd/protoc-gen-openapi@$(PROTOC_GEN_OPENAPI_VERSION)
	@go install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_GEN_VALIDATE_VERSION)
	@go install github.com/menta2k/protoc-gen-redact/v3@$(PROTOC_GEN_REDACT_VERSION)
	@go install $(SERVORA_PKG)/cmd/protoc-gen-servora-authz@$(SERVORA_VERSION)
	@go install $(SERVORA_PKG)/cmd/protoc-gen-servora-audit@$(SERVORA_VERSION)
	@go install $(SERVORA_PKG)/cmd/protoc-gen-servora-authn@$(SERVORA_VERSION)
	@go install $(SERVORA_PKG)/cmd/protoc-gen-servora-conf@$(SERVORA_VERSION)
	@go install $(SERVORA_PKG)/cmd/protoc-gen-servora-mapper@$(SERVORA_VERSION)
	@echo "✓ Protoc plugins installed"

cli:
	@echo "==> Installing CLI tools..."
	@go install github.com/go-kratos/kratos/cmd/kratos/v2@$(KRATOS_VERSION)
	@go install github.com/google/gnostic@$(GNOSTIC_VERSION)
	@go install github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION)
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@go install github.com/google/wire/cmd/wire@$(WIRE_VERSION)
	@go install entgo.io/ent/cmd/ent@$(ENT_VERSION)
	@go install github.com/air-verse/air@$(AIR_VERSION)
	@go install $(SERVORA_PKG)/cmd/svr@$(SERVORA_VERSION)
	@echo "✓ CLI tools installed"

dep:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),echo "  $(mod)" && (cd $(ROOT_DIR)$(mod) && go mod download) && ) true

tidy:
	@echo "==> Tidying Go modules..."
	@$(foreach mod,$(GO_WORKSPACE_MODULES),echo "  $(mod)" && (cd $(ROOT_DIR)$(mod) && go mod tidy) && ) true
	@go work sync
	@echo "✓ All modules tidied"

test:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),echo "==> Testing $(mod)..." && (cd $(ROOT_DIR)$(mod) && go test -short ./...) && ) true

cover:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),(cd $(ROOT_DIR)$(mod) && go test -v ./... -coverprofile=coverage.out) && ) true

vet:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),(cd $(ROOT_DIR)$(mod) && go vet ./...) && ) true

lint: lint.go lint.ts
	@echo "✓ lint complete"

lint.go:
	@$(foreach mod,$(GO_LINT_MODULES),echo "==> Linting Go ($(mod), GOWORK=$(LINT_GOWORK))..." && (cd $(ROOT_DIR)$(mod) && GOWORK=$(LINT_GOWORK) golangci-lint run) && ) true
	@echo "✓ Go lint complete"

lint.ts:
ifneq (,$(WEB_APPS))
	@echo "Type checking TypeScript..."
	@pnpm $(WEB_PNPM_FILTERS) run --if-present typecheck
	@echo "Linting TypeScript..."
	@pnpm $(WEB_PNPM_FILTERS) run --if-present lint
	@echo "✓ TypeScript lint complete"
else
	@echo "No WEB_APPS configured, skipping TypeScript lint"
endif

web.dev:
ifneq (,$(WEB_DEV_APP))
	@echo "Starting web dev server ($(WEB_DEV_APP))..."
	@pnpm --filter "./web/$(WEB_DEV_APP)" run dev
else
	@echo "No WEB_DEV_APP configured"
endif

wire:
	@echo "Generating wire code for all services..."
	$(call run-in-service-dirs,wire)
	@echo "✓ Wire code generated"

ent:
	@echo "Generating ent code for all services..."
	$(call run-in-service-dirs,gen.ent)
	@echo "✓ Ent code generated"

gen: api openapi wire ent
	@echo "✓ All code generated"

gen.fresh: clean gen

api: api-go api-ts
	@echo "✓ Protobuf code generated"

api-go:
	@echo "Generating protobuf Go code via $(BUF_GO_GEN_TEMPLATE)..."
	@buf generate --template $(BUF_GO_GEN_TEMPLATE)

api-ts:
ifneq (,$(wildcard $(BUF_TS_GEN_TEMPLATE)))
	@echo "Generating shared TypeScript via $(BUF_TS_GEN_TEMPLATE)..."
	@buf generate --template $(BUF_TS_GEN_TEMPLATE)
	@$(foreach tpl,$(BUF_TS_SERVICE_TEMPLATES),echo "Generating TypeScript via $(tpl)..." && buf generate --template $(tpl) &&) true
endif

openapi:
	@echo "Generating OpenAPI documentation for all services..."
	$(call run-in-service-dirs,openapi)
	@echo "✓ OpenAPI documentation generated"

lint.proto:
	@echo "Linting protobuf files..."
	@buf lint
	@echo "✓ Proto lint complete"

bsr.update:
	@echo "Updating BSR dependencies..."
	@buf dep update
	@echo "✓ BSR dependencies updated"

buf-update: bsr.update

build: gen
	@echo "Building all services..."
	$(call run-in-service-dirs,_build)
	@echo "✓ All services built"

# Tag root module.
# Usage: make tag TAG=v0.2.0
tag:
ifndef TAG
	$(error TAG is required. Usage: make tag TAG=v0.2.0)
endif
	@echo "Tagging $(TAG)..."
	@git tag $(TAG)
	@echo "✓ Tagged: $(TAG)"
	@echo "  Run 'git push --tags' to push"

# Tag api/gen sub-module when proto/gen changes require it.
# Usage: make tag.api TAG=v0.2.0
tag.api:
ifndef TAG
	$(error TAG is required. Usage: make tag.api TAG=v0.2.0)
endif
	@echo "Tagging api/gen/$(TAG)..."
	@git tag api/gen/$(TAG)
	@echo "✓ Tagged: api/gen/$(TAG)"
	@echo "  Run 'git push --tags' to push"

# Push proto to BSR, auto-labeling with current Git tag if available.
bsr.push:
	@GIT_TAG=$$(git tag --points-at HEAD 2>/dev/null | grep -E '^v[0-9]' | head -1); \
	if [ -n "$$GIT_TAG" ]; then \
		echo "==> Pushing to BSR with labels: $$GIT_TAG, main"; \
		buf push --exclude-unnamed --label "$$GIT_TAG" --label main; \
	else \
		echo "==> No Git version tag on HEAD, pushing with label: main"; \
		buf push --exclude-unnamed --label main; \
	fi
	@echo "✓ Proto pushed to BSR"

buf-push: bsr.push

# ============================================================================
# COMPOSE TARGETS
# ============================================================================

# build production images for microservices
compose.build:
	@echo "Build production images: $(MICROSERVICES) (version: $(DOCKER_TAG_VERSION))"
	@$(foreach svc,$(MICROSERVICES),docker build --build-arg SERVICE_NAME=$(svc) --build-arg VERSION=$(VERSION) -t servora-$(svc):$(DOCKER_TAG_VERSION) . &&) true
	@$(foreach svc,$(MICROSERVICES),docker tag servora-$(svc):$(DOCKER_TAG_VERSION) servora-$(svc):latest &&) true
	@echo "✓ Production images built"

# start infrastructure compose stack
compose.up: compose.up.infra

# start only infrastructure services
compose.up.infra:
	@echo "Compose infra up: $(INFRA_SERVICES)"
	@$(COMPOSE) $(COMPOSE_FILES) up -d $(INFRA_SERVICES)
	@echo "✓ Infrastructure services started"

# start infrastructure + app services
compose.up.all:
	@echo "Compose up: infra + apps"
	@$(COMPOSE) $(COMPOSE_APPS_FILES) up -d
	@echo "✓ All services started"

# rebuild production images and ensure infrastructure is running
compose.rebuild:
	@$(MAKE) compose.build
	@$(MAKE) compose.up
	@echo "✓ Production images rebuilt and infrastructure started"

# stop infrastructure compose stack
compose.stop:
	@$(COMPOSE) $(COMPOSE_FILES) stop $(INFRA_SERVICES)

# remove local compose stack containers/networks
compose.down:
	@$(COMPOSE_STACK_DOWN)

# remove local compose stack containers/networks/volumes
compose.reset:
	@$(COMPOSE_STACK_RESET)

# show infrastructure compose stack status
compose.ps:
	@$(COMPOSE) $(COMPOSE_FILES) ps $(INFRA_SERVICES)

# tail logs for infrastructure compose stack
compose.logs:
	@$(COMPOSE) $(COMPOSE_FILES) logs -f $(INFRA_SERVICES)


# ============================================================================
# OPENFGA TARGETS
# ============================================================================

OPENFGA_MODEL := manifests/openfga/model/servora.fga
OPENFGA_TESTS := manifests/openfga/tests/servora.fga.yaml
OPENFGA_ENV_PREFIX ?= EXAMPLE_
OPENFGA_API_URL ?= http://localhost:18080

openfga.init:
	@svr openfga init --model $(OPENFGA_MODEL) --env-prefix $(OPENFGA_ENV_PREFIX) --api-url $(OPENFGA_API_URL)

openfga.model.validate:
	@echo "Validating OpenFGA model..."
	@fga model validate --file $(OPENFGA_MODEL) --format fga
	@echo "✓ OpenFGA model valid"

openfga.model.test: openfga.model.validate
	@echo "Testing OpenFGA model..."
	@fga model test --tests $(OPENFGA_TESTS)
	@echo "✓ OpenFGA model tests passed"

openfga.model.apply: openfga.model.test
	@svr openfga model apply --model $(OPENFGA_MODEL) --env-prefix $(OPENFGA_ENV_PREFIX) --api-url $(OPENFGA_API_URL)

# ============================================================================
# CLEANUP
# ============================================================================

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf api/gen/go
	$(call run-in-service-dirs,clean)
	@echo "✓ Clean complete"

help:
	@echo ""
	@echo "servora-example"
	@echo "==========="
	@echo ""
	@echo "Usage:"
	@echo " make [target]"
	@echo ""
	@echo "Targets:"
	@awk '/^[a-zA-Z\-_0-9\.]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "  %-20s %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
