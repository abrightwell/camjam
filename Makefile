.PHONY: all build resolve clean

PROJECT_DIR := $(shell pwd)
BUILD_DIR := $(PROJECT_DIR)/build
VENDOR_DIR := $(PROJECT_DIR)/vendor

all: clean resolve build

clean:
	@echo "Cleaning project..."
	@echo "Removing 'build' directory"
	@rm -rf $(BUILD_DIR)
	@echo "Removing 'vendor' directory"
	@rm -rf $(VENDOR_DIR)

resolve:
	@echo "Resolving dependencies..."
	@dep ensure -v

build:
	@echo "Building project... ${GOARCH}"
	@go build -o $(BUILD_DIR)/camjam \
		-ldflags='-s -w'
