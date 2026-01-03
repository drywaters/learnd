.DEFAULT_GOAL := help
.PHONY: help run build test docker-buildx tail-watch tail-prod migrate migrate-down migrate-status gen-api-key

# Include local.mk for local environment variables (API keys, DATABASE_URL, etc.)
-include local.mk

# Local development (assumes tailwindcss binary is installed)
run: tail-prod ## Build Tailwind and run the app (go run ./cmd/learnd)
	go run ./cmd/learnd

build: tail-prod ## Build production binary to bin/learnd
	go build -o bin/learnd ./cmd/learnd

# Tailwind (using standalone CLI binary)
tail-watch: ## Build Tailwind in watch mode (requires tailwindcss CLI)
	tailwindcss -i ./tailwind/styles.css -o ./static/styles.css --watch

tail-prod: ## Build minified Tailwind output to static/styles.css
	tailwindcss -i ./tailwind/styles.css -o ./static/styles.css --minify

# Database migrations
migrate: ## Apply database migrations
	goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down: ## Roll back the last migration
	goose -dir migrations postgres "$$DATABASE_URL" down

migrate-status: ## Show migration status
	goose -dir migrations postgres "$$DATABASE_URL" status

# Testing
test: ## Run Go tests
	go test -v ./...

# Docker (production)
docker-buildx: ## Build and push multi-arch Docker image using buildx
	docker buildx build \
		--platform $(PLATFORMS) \
		--tag $(REGISTRY)/$(IMAGE_REPO):$(TAG) \
		--tag $(REGISTRY)/$(IMAGE_REPO):latest \
		--push \
		.

# Generate API key hash
gen-api-key: ## Generate bcrypt hash for API_KEY_HASH
	@go run ./scripts/hashkey.go

help: ## Show this help menu
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-20s %s\n", $$1, $$2} END {printf "\n"}' $(MAKEFILE_LIST)
