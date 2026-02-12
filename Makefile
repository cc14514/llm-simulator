.PHONY: help docker-build docker-push docker-buildx

# Usage examples:
#   make docker-push REGISTRY=ghcr.io/your-org
#   make docker-push REGISTRY=ghcr.io/your-org TAG=v0.1.0

DOCKER ?= docker

IMAGE_NAME ?= llm-simulator
TAG ?= latest
PLATFORMS ?= linux/amd64,linux/arm64
BUILDER ?= llm-simulator-builder

# REGISTRY should be like: ghcr.io/your-org
REGISTRY ?=

# Fail fast when pushing without a registry.
# This triggers during Makefile parsing (so it also works with `make -n`).
ifneq ($(filter docker-push,$(MAKECMDGOALS)),)
ifeq ($(strip $(REGISTRY)),)
$(error REGISTRY is required. Example: make docker-push REGISTRY=ghcr.io/your-org)
endif
endif

help:
	@echo "Targets:"
	@echo "  docker-push   Build & push multi-arch image (requires REGISTRY=...)"
	@echo "  docker-build  Build image locally for current arch"
	@echo ""
	@echo "Variables:"
	@echo "  REGISTRY=ghcr.io/your-org   (required for docker-push)"
	@echo "  IMAGE_NAME=llm-simulator    (default: llm-simulator)"
	@echo "  TAG=latest                  (default: latest)"
	@echo "  PLATFORMS=linux/amd64,linux/arm64"

# Create/activate a buildx builder (safe to run repeatedly).
# Note: docker buildx requires a recent Docker (Docker Desktop on macOS is fine).
docker-buildx:
	@$(DOCKER) buildx inspect $(BUILDER) >/dev/null 2>&1 || $(DOCKER) buildx create --name $(BUILDER) --use
	@$(DOCKER) buildx use $(BUILDER) >/dev/null
	@$(DOCKER) buildx inspect --bootstrap >/dev/null

docker-build:
	@$(DOCKER) build -t $(IMAGE_NAME):$(TAG) .

# Push as: $(REGISTRY)/$(IMAGE_NAME):$(TAG)
docker-push: docker-buildx
	@if [ -z "$(REGISTRY)" ]; then \
		echo "ERROR: REGISTRY is required. Example: make docker-push REGISTRY=ghcr.io/your-org"; \
		exit 2; \
	fi
	@$(DOCKER) buildx build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/$(IMAGE_NAME):$(TAG) \
		--push \
		.
