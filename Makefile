SHELL := /usr/bin/env bash
SCRIPTS_DIR := scripts/js
.PHONY: help build build-api build-frontend generate-icons generate-types generate-types-install fmt lint test docker-build docker-push docker-up docker-down migrate initdb clean image-api image-viz dev run

# Simple Makefile for common tasks across the imagine repository.
# Targets included:
#  - build: builds backend and frontend
#  - build-api: builds Go API binary
#  - build-frontend: builds the Viz frontend (pnpm in `viz`)
#  - generate-icons: run the icon generator for the frontend
#  - docker-* helpers: build/push images and bring up compose
#  - fmt/lint/test: formatting, linting, running tests
#  - migrate/initdb: helpers for database migration/init steps (best-effort)

VIZ_DIR := viz
GO_CMD ?= go
PNPM ?= pnpm
DOCKER_COMPOSE ?= docker compose
BUILDX_CACHE_DIR ?= $(CURDIR)/.buildcache

# Local buildx cache dirs (api/viz)
BUILDX_CACHE_API_DIR := $(BUILDX_CACHE_DIR)/api
BUILDX_CACHE_VIZ_DIR := $(BUILDX_CACHE_DIR)/viz

# Cache control: when `USE_HOST_CACHE=1` Make will use host-side cache dirs
# for Go module cache, Go build cache and pnpm store. This preserves fast
# incremental builds when running `make` locally. Set `USE_HOST_CACHE=0`
# to disable host caches (useful inside Docker build where BuildKit mounts
# are preferred).
USE_HOST_CACHE ?= 1
GO_MOD_CACHE_DIR ?= $(CURDIR)/.cache/go-mods
GO_BUILD_CACHE_DIR ?= $(CURDIR)/.cache/go-build
PNPM_STORE_DIR ?= $(CURDIR)/$(VIZ_DIR)/.pnpm-store

# Image registry and tag; override with `make REGISTRY=... TAG=...`
REGISTRY_HOST ?= ghcr.io
# Try to auto-detect a likely registry user/owner in this order:
# 1. GITHUB_REPOSITORY env (CI) -> owner/repo, take owner
# 2. Git remote origin URL (parse owner from URL)
# 3. Fall back to current shell user
REGISTRY_USER ?= $(shell \
	if [ -n "$$GITHUB_REPOSITORY" ]; then \
		echo "$$GITHUB_REPOSITORY" | awk -F/ '{print $$1}'; \
	else \
		url=$$(git remote get-url origin 2>/dev/null || git config --get remote.origin.url 2>/dev/null || echo); \
		if [ -n "$$url" ]; then \
			echo "$$url" | sed -E 's#.*[:/]{1,2}([^/]+)/.*#\1#' || true; \
		else \
			echo "$(USER)"; \
		fi; \
	fi)
REGISTRY ?= $(REGISTRY_HOST)/$(REGISTRY_USER)
TAG ?= $(shell cat version.txt 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo local)

### Helpers
help:
	@printf "Usage: make <target>\n\n"
	@printf "Targets:\n"
	@printf "  build                     Build backend and frontend\n"
	@printf "  build-api                 Build Go API binary\n"
	@printf "  build-frontend            Build the Viz frontend (pnpm)\n"
	@printf "  generate-icons            Run icon generator in $(VIZ_DIR)\n"
	@printf "  generate-types            Generate API types (Go DTOs + TS client)\n"
	@printf "  generate-types-install    Install type generation tools and run generation\n"
	@printf "  fmt                       Run formatting (Go and frontend if available)\n"
	@printf "  lint                      Run linters (if available)\n"
	@printf "  test                      Run Go tests and frontend tests if present\n"
	@printf "  docker-build              Build Docker images via docker compose\n"
	@printf "  docker-up                 Start services via docker compose\n"
	@printf "  docker-down               Stop services\n"
	@printf "  migrate                   Run migrations (best-effort helper)\n"
	@printf "  initdb                    Run DB init script (best-effort)\n"
	@printf "  clean                     Remove build artifacts and generated icons\n"
	@printf "  bump-version VERSION=x.y.z  Update project version (uses scripts/js/updateProjectVersion.js)\n"
	@printf "  check-env                Verify required tools are installed on host\n"
	@printf "  ci-build                 CI-optimised build (uses buildx and no host caches)\n"

### Build
build: build-api build-frontend

build-api:
	@echo "Building Go API..."
	@if [ -d "cmd/api" ]; then \
		mkdir -p build; \
		if [ "$(USE_HOST_CACHE)" = "1" ]; then \
			mkdir -p "$(GO_MOD_CACHE_DIR)" "$(GO_BUILD_CACHE_DIR)"; \
			GOMODCACHE="$(GO_MOD_CACHE_DIR)" GOCACHE="$(GO_BUILD_CACHE_DIR)" $(GO_CMD) build -o build/api ./cmd/api; \
		else \
			$(GO_CMD) build -o build/api ./cmd/api; \
		fi; \
	else \
		if [ "$(USE_HOST_CACHE)" = "1" ]; then \
			mkdir -p "$(GO_MOD_CACHE_DIR)" "$(GO_BUILD_CACHE_DIR)"; \
			GOMODCACHE="$(GO_MOD_CACHE_DIR)" GOCACHE="$(GO_BUILD_CACHE_DIR)" $(GO_CMD) build -o build/api ./...; \
		else \
			$(GO_CMD) build -o build/api ./...; \
		fi; \
	fi

build-frontend:
	@echo "Building frontend in $(VIZ_DIR)..."
	@cd $(VIZ_DIR) && if [ "$(USE_HOST_CACHE)" = "1" ]; then \
		mkdir -p "$(PNPM_STORE_DIR)"; \
		$(PNPM) install --frozen-lockfile --store-dir "$(PNPM_STORE_DIR)" || true; \
		$(PNPM) run build; \
	else \
		$(PNPM) install --frozen-lockfile || true; \
		$(PNPM) run build; \
	fi

generate-icons:
	@echo "Generating icons for frontend..."
	@cd $(VIZ_DIR) && $(PNPM) run generate:icons

generate-types:
	@echo "Generating API types (Go DTOs + TS client) using scripts in $(SCRIPTS_DIR)..."
	@if [ -d "$(SCRIPTS_DIR)" ]; then \
		cd $(SCRIPTS_DIR); \
		if command -v $(PNPM) >/dev/null 2>&1; then \
			$(PNPM) install --frozen-lockfile || true; \
			$(PNPM) run gen:api:build || $(PNPM) run gen:api; \
		else \
			echo "pnpm not found, attempting npx tsx fallback..."; \
			npx tsx gen-api.ts --build || npx tsx gen-api.ts; \
		fi; \
	else \
		echo "No scripts found at $(SCRIPTS_DIR)"; \
	fi

generate-types-install:
	@echo "Install type generation tools and run the generator (scripts in $(SCRIPTS_DIR))..."
	@if [ -d "$(SCRIPTS_DIR)" ]; then \
		cd $(SCRIPTS_DIR); \
		if command -v $(PNPM) >/dev/null 2>&1; then \
			$(PNPM) install --frozen-lockfile || true; \
			$(PNPM) run gen:api:install || $(PNPM) run gen:api:build || $(PNPM) run gen:api; \
		else \
			echo "pnpm not found, attempting npx tsx fallback..."; \
			npx tsx gen-api.ts --install-tools || npx tsx gen-api.ts --build || npx tsx gen-api.ts; \
		fi; \
	else \
		echo "No scripts found at $(SCRIPTS_DIR)"; \
	fi

### Code hygiene
fmt:
	@echo "Formatting Go sources..."
	@$(GO_CMD) fmt ./...
	@if [ -f "$(VIZ_DIR)/package.json" ]; then \
		echo "Formatting frontend (if script exists)..."; \
		cd $(VIZ_DIR) && $(PNPM) run format || true; \
	fi

lint:
	@echo "Running linters (golangci-lint if available)..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, skipping Go lint. Install it or add a linter."; \
	fi

test:
	@echo "Running Go tests..."
	@$(GO_CMD) test ./...
	@if [ -f "$(VIZ_DIR)/package.json" ]; then \
		cd $(VIZ_DIR) && $(PNPM) run test || true; \
	fi

### Docker
docker-build:
	@echo "Building docker images using docker-compose..."
	@$(DOCKER_COMPOSE) build --parallel

### Buildx targets (useful to persist cache locally and speed up Go module downloads)
.PHONY: buildx-api buildx-viz buildx-build
buildx-api:
	@echo "Building API image with buildx and local cache ($(BUILDX_CACHE_API_DIR))"
	@mkdir -p $(BUILDX_CACHE_API_DIR)
	@docker buildx build --progress=plain \
		--cache-to=type=local,dest=$(BUILDX_CACHE_API_DIR) \
		--cache-from=type=local,src=$(BUILDX_CACHE_API_DIR) \
		-f Dockerfile.api -t imagine-api:local --load .

buildx-viz:
	@echo "Building Viz image with buildx and local cache ($(BUILDX_CACHE_VIZ_DIR))"
	@mkdir -p $(BUILDX_CACHE_VIZ_DIR)
	@docker buildx build --progress=plain \
		--cache-to=type=local,dest=$(BUILDX_CACHE_VIZ_DIR) \
		--cache-from=type=local,src=$(BUILDX_CACHE_VIZ_DIR) \
		-f viz/Dockerfile -t imagine-viz:local viz --load

buildx-build: buildx-api buildx-viz
	@echo "buildx build complete. Use the images 'imagine-api:local' and 'imagine-viz:local' or tag/push as needed."

docker-up:
	@echo "Starting services with docker-compose..."
	@$(DOCKER_COMPOSE) up -d --build

docker-down:
	@echo "Stopping services with docker-compose..."
	@$(DOCKER_COMPOSE) down

image-api:
	@echo "Building API image"
	@docker build -f Dockerfile.api -t $(REGISTRY)/imagine-api:$(TAG) .

image-viz:
	@echo "Building Viz image"
	@docker build -f $(VIZ_DIR)/Dockerfile -t $(REGISTRY)/imagine-viz:$(TAG) $(VIZ_DIR)

docker-push: image-api image-viz
	@echo "Pushing images to $(REGISTRY)"
	@docker push $(REGISTRY)/imagine-api:$(TAG)
	@docker push $(REGISTRY)/imagine-viz:$(TAG)

### Database / Migrations (best-effort)
migrate:
	@echo "Migrations helper: uses your local migrate tool or docker-compose migration service if configured."
	@if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "${DATABASE_URL}" up || true; \
	else \
		echo "No migrate tool found. Consider running migrations in a DB container or install golang-migrate."; \
	fi

initdb:
	@echo "Running DB init script (if present)"
	@if [ -f ./docker/initdb/01-create-superuser.sh ]; then \
		bash ./docker/initdb/01-create-superuser.sh || true; \
	else \
		echo "No initdb script found at ./docker/initdb/01-create-superuser.sh"; \
	fi

clean:
	@echo "Cleaning build artifacts and generated icons..."
	@rm -rf build
	@rm -rf $(VIZ_DIR)/src/lib/components/icons/generated || true

dev:
	@echo "Developer: start API (if you have a dev script) and frontend dev server"
	@echo " - Run backend dev in one terminal: $(GO_CMD) run ./cmd/api (or main.go)"
	@echo " - Run frontend dev in another: cd $(VIZ_DIR) && $(PNPM) run dev"

run: build
	@echo "Run completed build artifacts (see README for local run instructions)."

check-env:
	@echo "Checking required host tools..."
	@set -e; \
	for cmd in go node pnpm make docker git; do \
		if ! command -v $$cmd >/dev/null 2>&1; then \
			echo "ERROR: $$cmd not found in PATH"; exit 1; \
		else \
			echo "OK: $$cmd"; \
		fi; \
	done; \
	# Check docker buildx (warning only)
	if docker buildx version >/dev/null 2>&1; then \
		echo "OK: docker buildx available"; \
	else \
		echo "WARN: docker buildx not found; ci-build may behave differently"; \
	fi

ci-build: check-env
	@echo "Starting CI-optimised build (no host caches, using buildx where available)"
	@set -e; \
	# Ensure we do not use host caches during CI build — invoke the target
	# via $(MAKE) and pass the variable override so the same make binary and
	# flags are preserved in recursive invocation.
	$(MAKE) generate-types-install USE_HOST_CACHE=0; \
	# Build API image with buildx (load into local daemon). Adjust cache flags to your CI.
	if docker buildx version >/dev/null 2>&1; then \
		docker buildx build --progress=plain --tag $(REGISTRY)/imagine-api:$(TAG) -f Dockerfile.api --load .; \
		docker buildx build --progress=plain --tag $(REGISTRY)/imagine-viz:$(TAG) -f viz/Dockerfile --load viz; \
	else \
		docker build -f Dockerfile.api -t $(REGISTRY)/imagine-api:$(TAG) .; \
		docker build -f viz/Dockerfile -t $(REGISTRY)/imagine-viz:$(TAG) viz; \
	fi


bump-version:
	@echo "Bumping project version to $(VERSION)..."
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: supply VERSION=x.y.z"; exit 1; \
	fi; \
	cd $(SCRIPTS_DIR); \
	if command -v $(PNPM) >/dev/null 2>&1; then \
		$(PNPM) exec tsx updateProjectVersion.js $(VERSION) || $(PNPM) exec node updateProjectVersion.js $(VERSION); \
	else \
		npx tsx updateProjectVersion.js $(VERSION) || node updateProjectVersion.js $(VERSION); \
	fi; \
	echo "Version updated to $(VERSION)"

release: 
	@echo "Creating release for version $(VERSION)"
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: supply VERSION=x.y.z"; exit 1; \
	fi; \
	# Run the bump-version helper to update files
	$(MAKE) bump-version VERSION=$(VERSION); \
	# Stage updated files (version.txt and package.json files if present)
	git add version.txt || true; \
	if [ -f package.json ]; then git add package.json; fi; \
	if [ -f viz/package.json ]; then git add viz/package.json; fi; \
	if [ -f scripts/js/package.json ]; then git add scripts/js/package.json; fi; \
	# Commit changes (no-op if nothing changed)
	if git diff --staged --quiet; then \
		echo "No changes to commit"; \
	else \
		git commit -m "chore(release): v$(VERSION)"; \
	fi; \
	# Prevent tagging over an existing tag
	if git rev-parse -q --verify "refs/tags/v$(VERSION)" >/dev/null 2>&1; then \
		echo "ERROR: tag v$(VERSION) already exists"; exit 1; \
	fi; \
	# Create annotated tag
	git tag -a "v$(VERSION)" -m "Release v$(VERSION)"; \
	@echo "Created tag v$(VERSION)"; \
	# Optionally push commit and tag when PUSH=1
	if [ "$(PUSH)" = "1" ]; then \
		git push origin HEAD && git push origin "v$(VERSION)"; \
	else \
		echo "Not pushing. To push, run: make release VERSION=$(VERSION) PUSH=1"; \
	fi
