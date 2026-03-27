# ============================================================================
# Makefile for servora-platform
# ============================================================================

ifeq ($(OS),Windows_NT)
    IS_WINDOWS := 1
endif

ifneq (,$(wildcard .env))
    include .env
    export
endif

# ============================================================================
# VARIABLES & CONFIGURATION
# ============================================================================

CURRENT_DIR := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
ROOT_DIR    := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

BUF_GO_GEN_TEMPLATE := buf.go.gen.yaml
BUF_TS_GEN_TEMPLATE := buf.typescript.gen.yaml

SRCS_MK := $(foreach dir, app, $(wildcard $(dir)/*/*/Makefile))
SERVICE_DIRS := $(dir $(realpath $(SRCS_MK)))
BUF_TS_SERVICE_TEMPLATES := $(wildcard $(addsuffix api/buf.typescript.gen.yaml,$(SERVICE_DIRS)))

GOPATH := $(shell go env GOPATH)
GOVERSION := $(shell go version)

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.GitBranch=$(GIT_BRANCH)

RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
CYAN := \033[0;36m
RESET := \033[0m

COMPOSE := docker compose
COMPOSE_FILES := -f docker-compose.yaml
COMPOSE_DEV_FILES := -f docker-compose.yaml -f docker-compose.dev.yaml
MICROSERVICES := audit

WEB_APPS :=
WEB_DEV_APP ?=

GO_WORKSPACE_MODULES := app/audit/service

WEB_PNPM_FILTERS := $(foreach app,$(WEB_APPS),--filter "./web/$(app)")
INFRA_SERVICES := consul db redis openfga kafka clickhouse otel-collector jaeger loki prometheus grafana traefik
COMPOSE_STACK_SERVICES := $(INFRA_SERVICES) $(MICROSERVICES)
COMPOSE_STACK_DOWN := $(COMPOSE) $(COMPOSE_DEV_FILES) down --remove-orphans
COMPOSE_STACK_RESET := $(COMPOSE) $(COMPOSE_DEV_FILES) down --remove-orphans --volumes

SERVORA_PKG := github.com/Servora-Kit/servora

define run-in-service-dirs
	@$(foreach dir,$(SERVICE_DIRS),cd $(dir) && $(MAKE) $(1);)
endef

# ============================================================================
# MAIN TARGETS
# ============================================================================

.PHONY: help env init plugin cli dep tidy test cover vet lint lint.go lint.proto lint.ts web.dev buf-update buf-push tag
.PHONY: wire ent gen api api-go api-ts openapi build clean
.PHONY: compose.build compose.up compose.rebuild compose.stop compose.down compose.reset compose.ps compose.logs
.PHONY: compose.dev compose.dev.build compose.dev.up compose.dev.restart compose.dev.ps compose.dev.stop compose.dev.down compose.dev.reset compose.dev.logs
.PHONY: openfga.init openfga.model.validate openfga.model.test openfga.model.apply

# show environment variables
env:
	@echo "CURRENT_DIR: $(CURRENT_DIR)"
	@echo "ROOT_DIR: $(ROOT_DIR)"
	@echo "SRCS_MK: $(SRCS_MK)"
	@echo "MICROSERVICES: $(MICROSERVICES)"
	@echo "WEB_APPS: $(WEB_APPS)"
	@echo "GO_WORKSPACE_MODULES: $(GO_WORKSPACE_MODULES)"
	@echo "VERSION: $(VERSION)"
	@echo "GOVERSION: $(GOVERSION)"

# initialize develop environment
init: plugin cli
	@echo "$(GREEN)✓ Development environment initialized$(RESET)"

# install protoc plugins
plugin:
	@echo "$(CYAN)Installing protoc plugins...$(RESET)"
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	@go install github.com/go-kratos/protoc-gen-typescript-http@latest
	@go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest
	@go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	@go install github.com/envoyproxy/protoc-gen-validate@latest
	@go install github.com/menta2k/protoc-gen-redact/v3@latest
	@go install $(SERVORA_PKG)/cmd/protoc-gen-servora-authz@latest
	@go install $(SERVORA_PKG)/cmd/protoc-gen-servora-audit@latest
	@go install $(SERVORA_PKG)/cmd/protoc-gen-servora-mapper@latest
	@echo "$(GREEN)✓ Protoc plugins installed$(RESET)"

# install cli tools
cli:
	@echo "$(CYAN)Installing CLI tools...$(RESET)"
	@go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	@go install github.com/google/gnostic@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/google/wire/cmd/wire@latest
	@go install entgo.io/ent/cmd/ent@latest
	@go install $(SERVORA_PKG)/cmd/svr@latest
	@echo "$(GREEN)✓ CLI tools installed$(RESET)"

# download dependencies of module
dep:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),echo "  $(mod)" && (cd $(ROOT_DIR)$(mod) && go mod download) && ) true

# tidy all workspace modules
tidy:
	@echo "$(CYAN)Tidying Go modules...$(RESET)"
	@$(foreach mod,$(GO_WORKSPACE_MODULES),echo "  $(mod)" && (cd $(ROOT_DIR)$(mod) && go mod tidy) && ) true
	@echo "$(GREEN)✓ All modules tidied$(RESET)"

# run tests
test:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),echo "$(CYAN)Testing $(mod)...$(RESET)" && (cd $(ROOT_DIR)$(mod) && go test ./...) && ) true

# run coverage tests
cover:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),(cd $(ROOT_DIR)$(mod) && go test -v ./... -coverprofile=coverage.out) && ) true

# run static analysis
vet:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),(cd $(ROOT_DIR)$(mod) && go vet ./...) && ) true

# run all configured linters
lint: lint.go lint.ts
	@echo "$(GREEN)✓ lint complete$(RESET)"

# run golang lint
lint.go:
	@$(foreach mod,$(GO_WORKSPACE_MODULES),echo "$(CYAN)Linting Go ($(mod))...$(RESET)" && (cd $(ROOT_DIR)$(mod) && golangci-lint run) && ) true
	@echo "$(GREEN)✓ Go lint complete$(RESET)"

# lint TypeScript
lint.ts:
ifneq (,$(WEB_APPS))
	@echo "$(CYAN)Type checking TypeScript...$(RESET)"
	@pnpm $(WEB_PNPM_FILTERS) run --if-present typecheck
	@echo "$(CYAN)Linting TypeScript...$(RESET)"
	@pnpm $(WEB_PNPM_FILTERS) run --if-present lint
	@echo "$(GREEN)✓ TypeScript lint complete$(RESET)"
else
	@echo "$(YELLOW)No WEB_APPS configured, skipping TypeScript lint$(RESET)"
endif

# start web dev server
web.dev:
ifneq (,$(WEB_DEV_APP))
	@echo "$(CYAN)Starting web dev server ($(WEB_DEV_APP))...$(RESET)"
	@pnpm --filter "./web/$(WEB_DEV_APP)" run dev
else
	@echo "$(YELLOW)No WEB_DEV_APP configured$(RESET)"
endif

# generate wire code for all services
wire:
	@echo "$(CYAN)Generating wire code for all services...$(RESET)"
	$(call run-in-service-dirs,wire)
	@echo "$(GREEN)✓ Wire code generated$(RESET)"

# generate ent code for services that define data/generate.go
ent:
	@echo "$(CYAN)Generating ent code for all services...$(RESET)"
	$(call run-in-service-dirs,gen.ent)
	@echo "$(GREEN)✓ Ent code generated$(RESET)"

# generate all code
gen: api openapi wire ent
	@echo "$(GREEN)✓ All code generated$(RESET)"

# generate protobuf api code (go + ts)
api: api-go api-ts
	@echo "$(GREEN)✓ Protobuf code generated$(RESET)"

# generate protobuf api go code
api-go:
	@echo "$(CYAN)Generating protobuf Go code via $(BUF_GO_GEN_TEMPLATE)...$(RESET)"
	@buf generate --template $(BUF_GO_GEN_TEMPLATE)

# generate protobuf api typescript code
api-ts:
ifneq (,$(wildcard $(BUF_TS_GEN_TEMPLATE)))
	@echo "$(CYAN)Generating shared TypeScript via $(BUF_TS_GEN_TEMPLATE)...$(RESET)"
	@buf generate --template $(BUF_TS_GEN_TEMPLATE)
	@$(foreach tpl,$(BUF_TS_SERVICE_TEMPLATES),echo "$(CYAN)Generating TypeScript via $(tpl)...$(RESET)" && buf generate --template $(tpl) &&) true
endif

# generate protobuf api OpenAPI v3 docs for all services
openapi:
	@echo "$(CYAN)Generating OpenAPI documentation for all services...$(RESET)"
	$(call run-in-service-dirs,openapi)
	@echo "$(GREEN)✓ OpenAPI documentation generated$(RESET)"

# lint protobuf files
lint.proto:
	@echo "$(CYAN)Linting protobuf files...$(RESET)"
	@buf lint
	@echo "$(GREEN)✓ Proto lint complete$(RESET)"

# update buf dependencies
buf-update:
	@echo "$(CYAN)Updating buf dependencies...$(RESET)"
	@buf dep update
	@echo "$(GREEN)✓ Buf dependencies updated$(RESET)"

# build all service applications
build: gen
	@echo "$(CYAN)Building all services...$(RESET)"
	$(call run-in-service-dirs,_build)
	@echo "$(GREEN)✓ All services built$(RESET)"

# Tag root module.
# Usage: make tag TAG=v0.2.0
tag:
ifndef TAG
	$(error TAG is required. Usage: make tag TAG=v0.2.0)
endif
	@echo "$(CYAN)Tagging $(TAG)...$(RESET)"
	@git tag $(TAG)
	@echo "$(GREEN)✓ Tagged: $(TAG)$(RESET)"
	@echo "  Run 'git push --tags' to push"

# Tag api/gen sub-module when proto/gen changes require it.
# Usage: make tag.api TAG=v0.2.0
tag.api:
ifndef TAG
	$(error TAG is required. Usage: make tag.api TAG=v0.2.0)
endif
	@echo "$(CYAN)Tagging api/gen/$(TAG)...$(RESET)"
	@git tag api/gen/$(TAG)
	@echo "$(GREEN)✓ Tagged: api/gen/$(TAG)$(RESET)"
	@echo "  Run 'git push --tags' to push"

# Push proto to BSR, auto-labeling with current Git tag if available
buf-push:
	@echo "$(CYAN)Pushing proto to BSR...$(RESET)"
	@GIT_TAG=$$(git tag --points-at HEAD 2>/dev/null | grep -E '^v[0-9]' | head -1); \
	if [ -n "$$GIT_TAG" ]; then \
		echo "  Using Git tag as BSR label: $$GIT_TAG"; \
		buf push --exclude-unnamed --label "$$GIT_TAG"; \
	else \
		echo "  $(YELLOW)No Git version tag on HEAD, pushing without label$(RESET)"; \
		buf push --exclude-unnamed; \
	fi
	@echo "$(GREEN)✓ Proto pushed to BSR$(RESET)"

# ============================================================================
# COMPOSE TARGETS
# ============================================================================

# build production images for microservices
compose.build:
	@echo "$(CYAN)Build production images: $(MICROSERVICES) (version: $(VERSION))$(RESET)"
	@$(foreach svc,$(MICROSERVICES),docker build --build-arg SERVICE_NAME=$(svc) --build-arg VERSION=$(VERSION) -t servora-$(svc):$(VERSION) . &&) true
	@echo "$(GREEN)✓ Production images built$(RESET)"

# start infrastructure compose stack
compose.up:
	@echo "$(CYAN)Compose infra up: $(INFRA_SERVICES)$(RESET)"
	@$(COMPOSE) $(COMPOSE_FILES) up -d $(INFRA_SERVICES)
	@echo "$(GREEN)✓ Infrastructure services started$(RESET)"

# rebuild production images and ensure infrastructure is running
compose.rebuild:
	@$(MAKE) compose.build
	@$(MAKE) compose.up
	@echo "$(GREEN)✓ Production images rebuilt and infrastructure started$(RESET)"

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

# build Air-based development images for microservices
compose.dev.build:
	@echo "$(CYAN)Compose dev build: $(MICROSERVICES)$(RESET)"
	@$(COMPOSE) $(COMPOSE_DEV_FILES) build $(MICROSERVICES)
	@echo "$(GREEN)✓ Compose dev images built$(RESET)"

# start full development compose stack (infra + Air microservices) and tail logs
compose.dev:
	@echo "$(CYAN)Compose dev start: $(COMPOSE_STACK_SERVICES)$(RESET)"
	@$(COMPOSE) $(COMPOSE_DEV_FILES) up -d $(COMPOSE_STACK_SERVICES)
	@echo "$(GREEN)✓ Compose dev stack started, tailing logs...$(RESET)"
	@$(COMPOSE) $(COMPOSE_DEV_FILES) logs -f $(COMPOSE_STACK_SERVICES)

# start Air-based development stack in background
compose.dev.up:
	@echo "$(CYAN)Compose dev up: $(COMPOSE_STACK_SERVICES)$(RESET)"
	@$(COMPOSE) $(COMPOSE_DEV_FILES) up -d $(COMPOSE_STACK_SERVICES)
	@echo "$(GREEN)✓ Compose dev stack started$(RESET)"

# restart Air-based development containers to force fresh startup build
compose.dev.restart:
	@echo "$(CYAN)Compose dev restart (Air): $(MICROSERVICES)$(RESET)"
	@$(COMPOSE) $(COMPOSE_DEV_FILES) restart $(MICROSERVICES)
	@echo "$(GREEN)✓ Compose dev services restarted$(RESET)"

# tail logs for Air-based development stack
compose.dev.logs:
	@$(COMPOSE) $(COMPOSE_DEV_FILES) logs -f $(MICROSERVICES)

# show Air-based development stack status
compose.dev.ps:
	@$(COMPOSE) $(COMPOSE_DEV_FILES) ps $(COMPOSE_STACK_SERVICES)

# stop dev microservice containers (infrastructure keeps running)
compose.dev.stop:
	@$(COMPOSE) $(COMPOSE_DEV_FILES) stop $(MICROSERVICES)

# remove dev microservice containers (infrastructure keeps running)
compose.dev.down:
	@$(COMPOSE) $(COMPOSE_DEV_FILES) rm -sf $(MICROSERVICES)

# remove all compose stack containers/networks/volumes (infra + dev)
compose.dev.reset:
	@$(COMPOSE_STACK_RESET)

# ============================================================================
# OPENFGA TARGETS
# ============================================================================

OPENFGA_MODEL := manifests/openfga/model/servora.fga
OPENFGA_TESTS := manifests/openfga/tests/servora.fga.yaml
OPENFGA_ENV_PREFIX ?= PLATFORM_
OPENFGA_API_URL ?= http://localhost:18080

# initialize OpenFGA store and upload model (via svr CLI)
openfga.init:
	@svr openfga init --model $(OPENFGA_MODEL) --env-prefix $(OPENFGA_ENV_PREFIX) --api-url $(OPENFGA_API_URL)

# validate OpenFGA model syntax (requires fga CLI)
openfga.model.validate:
	@echo "$(CYAN)Validating OpenFGA model...$(RESET)"
	@fga model validate --file $(OPENFGA_MODEL) --format fga
	@echo "$(GREEN)✓ OpenFGA model valid$(RESET)"

# run OpenFGA model tests (requires fga CLI)
openfga.model.test: openfga.model.validate
	@echo "$(CYAN)Testing OpenFGA model...$(RESET)"
	@fga model test --tests $(OPENFGA_TESTS)
	@echo "$(GREEN)✓ OpenFGA model tests passed$(RESET)"

# apply new model version: validate -> test -> upload (via svr CLI)
openfga.model.apply: openfga.model.test
	@svr openfga model apply --model $(OPENFGA_MODEL) --env-prefix $(OPENFGA_ENV_PREFIX) --api-url $(OPENFGA_API_URL)

# ============================================================================
# CLEANUP
# ============================================================================

# clean build artifacts
clean:
	@echo "$(CYAN)Cleaning build artifacts...$(RESET)"
	@rm -rf api/gen/go
	$(call run-in-service-dirs,clean)
	@echo "$(GREEN)✓ Clean complete$(RESET)"

# show help
help:
	@echo ""
	@echo "$(CYAN)servora-platform$(RESET)"
	@echo "$(CYAN)================$(RESET)"
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
			printf "  $(GREEN)%-20s$(RESET) %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
