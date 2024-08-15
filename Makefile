.DEFAULT_GOAL := help

GO ?= go
container_tool ?= docker

image_repository ?= quay.io/morvencao/event-based-transport-demo
image_tag ?= latest

# Prints a list of useful targets.
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "  verify       Verifies that source passes standard checks."
	@echo "  build        Builds the binary."
	@echo "  test         Runs tests."
	@echo "  image        Builds the container image."
	@echo "  push         Pushes the container image to the registry."
	@echo "  e2e          Runs end-to-end tests."
	@echo "  e2e/setup    Set up the e2e environment."
	@echo "  e2e/teardown Tear down the e2e environment."
	@echo ""
.PHONY: help

# Verifies that source passes standard checks.
verify:
	${GO} vet \
		./cmd/... \
		./pkg/...
.PHONY: verify

# Build binaries
build:
	${GO} build -ldflags="$(ldflags)" \
		-o event-based-transport-demo \
		./cmd/main.go
.PHONY: build

# Runs tests.
test:
	${GO} test \
		./pkg/... \
		./cmd/...
.PHONY: test

# Builds the container image.
image:
	$(container_tool) build -t "$(image_repository):$(image_tag)" .
.PHONY: image

# Pushes the container image to the registry.
push: image
	$(container_tool) push "$(image_repository):$(image_tag)"
.PHONY: push

# Set up the environment
setup:
	./test/setup.sh
.PHONY: setup

# Tear down the environment
teardown:
	./test/teardown.sh
.PHONY: teardown

run: teardown setup
	./test/run.sh
.PHONY: run
