# ============================================================================
# web-app.mk — common Makefile for frontend apps (TanStack Start + shadcn/ui)
# ============================================================================
# Include from a web app directory (e.g. web/iam/Makefile):
#   WEB_APP_NAME ?= IAM
#   include ../../web-app.mk
# ============================================================================

# Load .env.local from the app directory (recommended: keep per-web-app, not monorepo root)
WEB_APP_ENV_FILE ?= $(CURDIR)/.env.local
ifneq (,$(wildcard $(WEB_APP_ENV_FILE)))
    include $(WEB_APP_ENV_FILE)
    export
endif

# Package manager: prefer pnpm if available, else npm
WEB_APP_PNPM := $(shell command -v pnpm 2>/dev/null)
ifeq ($(WEB_APP_PNPM),)
    PKG_RUN := npm run
    PKG_CI  := npm ci
    PKG_I   := npm install
else
    PKG_RUN := pnpm run
    PKG_CI  := pnpm install --frozen-lockfile
    PKG_I   := pnpm install
endif

WEB_APP_NAME ?= frontend
CYAN  := \033[0;36m
GREEN := \033[0;32m
RESET := \033[0m

.PHONY: help install ci dev build preview start test lint lint.ts typecheck format check clean

# ----------------------------------------------------------------------------
# Help
# ----------------------------------------------------------------------------
help:
	@echo "$(CYAN)$(WEB_APP_NAME) frontend — Makefile targets$(RESET)"
	@echo ""
	@echo "  $(GREEN)install$(RESET)   Install dependencies ($(PKG_I))"
	@echo "  $(GREEN)ci$(RESET)         Install with frozen lockfile (CI)"
	@echo "  $(GREEN)dev$(RESET)        Start dev server (port 3000)"
	@echo "  $(GREEN)build$(RESET)     Production build"
	@echo "  $(GREEN)preview$(RESET)   Preview production build"
	@echo "  $(GREEN)start$(RESET)     Run production server (after build)"
	@echo "  $(GREEN)test$(RESET)      Run tests (vitest)"
	@echo "  $(GREEN)lint$(RESET)      Run ESLint"
	@echo "  $(GREEN)typecheck$(RESET) TypeScript (tsc --noEmit, if script exists)"
	@echo "  $(GREEN)lint.ts$(RESET)   typecheck then ESLint (if scripts exist)"
	@echo "  $(GREEN)format$(RESET)    Check formatting (Prettier)"
	@echo "  $(GREEN)check$(RESET)     Format + lint fix (Prettier --write + ESLint --fix)"
	@echo "  $(GREEN)clean$(RESET)     Remove build artifacts and node_modules"
	@echo ""

# ----------------------------------------------------------------------------
# Dependencies
# ----------------------------------------------------------------------------
install:
	$(PKG_I)

ci:
	$(PKG_CI)

# ----------------------------------------------------------------------------
# Develop & Build
# ----------------------------------------------------------------------------
dev:
	$(PKG_RUN) dev

build:
	$(PKG_RUN) build

preview:
	$(PKG_RUN) preview

start:
	$(PKG_RUN) start

# ----------------------------------------------------------------------------
# Test & Lint
# ----------------------------------------------------------------------------
test:
	$(PKG_RUN) test

lint:
	$(PKG_RUN) lint

typecheck:
ifeq ($(WEB_APP_PNPM),)
	@npm run typecheck --if-present
else
	@pnpm run --if-present typecheck
endif

lint.ts: typecheck
ifeq ($(WEB_APP_PNPM),)
	@npm run lint --if-present
else
	@pnpm run --if-present lint
endif

format:
	$(PKG_RUN) format

check:
	$(PKG_RUN) check

# ----------------------------------------------------------------------------
# Clean
# ----------------------------------------------------------------------------
clean:
	rm -rf node_modules .output dist .vinxi
	@echo "$(GREEN)Cleaned build artifacts and node_modules$(RESET)"
