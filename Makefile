.DEFAULT_GOAL := help

PROJECT_DIR := $(shell pwd)
BUILD_DIR := $(PROJECT_DIR)/build

GOARCH ?= amd64

.PHONY: all
all: clean build

.PHONY: clean
clean: ## Clean build artifacts
clean:
	$(info "Cleaning project...")
	$(info "Removing 'build' directory")
	@rm -rf $(BUILD_DIR)

.PHONY: build
build: ## Build executable
build:
	$(info "Building project... $(GOARCH)")
	@go build -o $(BUILD_DIR)/camjam \
		-ldflags='-s -w'

.PHONY: help
help: ALIGN=14
help: ## Print this message
	@awk -F ': ## ' -- "/^[^':]+: ## /"' { printf "'$$(tput bold)'%-$(ALIGN)s'$$(tput sgr0)' %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
