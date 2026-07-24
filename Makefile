SERVORA_CONTEXT := root

PROJECT_NAME := servora-example
SERVICE_MODULES := app/master/service app/worker/service
GO_WORKSPACE_MODULES := api/gen $(SERVICE_MODULES)
GO_LINT_MODULES ?= $(SERVICE_MODULES)
LINT_GOWORK ?= auto

MICROSERVICES := master worker
COMPOSE_FILES := -f docker-compose.yaml
COMPOSE_SERVICES ?=

ENV_FILE ?= .env
OPENFGA_ENV_PREFIX ?= EXAMPLE_
OPENFGA_API_URL ?= http://localhost:18080

include make/core.mk
