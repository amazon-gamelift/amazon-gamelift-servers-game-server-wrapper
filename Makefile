# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

.PHONY: clean download-gamelift-servers-server-sdk build test
default: test

64_PLATFORMS ?= darwin/amd64 darwin/arm64 linux/amd64 linux/arm64
COMPUTE_TYPES ?= managed-ec2 anywhere managed-containers

MAJOR ?= 1
MINOR ?= 1
PATCH ?= 0
APP_NAME ?= amazon-gamelift-servers-game-server-wrapper
APP_PACKAGE ?= github.com/amazon-gamelift/$(APP_NAME)
SERVER_SDK_URL=https://github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/releases/download/v5.4.0/GameLift-Go-ServerSDK-5.4.0.zip
SERVER_SDK_FILE_NAME=gamelift-servers-server-sdk.zip
SERVER_SDK_EXTRACT_DIR=src/ext/gamelift-servers-server-sdk

OUT?=out
BUILD_PLAT_DIR_SEP=/

GOOS ?= $(shell uname | tr '[:upper:]' '[:lower:]'| tr -d '[:space:]')# linux | darwin
BUILDARCH = $(shell uname -m)

ifeq ($(BUILDARCH),arm64)
	GOARCH?=arm64
else ifeq ($(BUILDARCH),aarch64)
	GOARCH?=arm64
else ifeq ($(BUILDARCH),x86_64)
	GOARCH?=amd64
endif

OUT_FOLDER=$(OUT)$(BUILD_PLAT_DIR_SEP)$(GOOS)$(BUILD_PLAT_DIR_SEP)$(GOARCH)$(BUILD_PLAT_DIR_SEP)
COMPUTE_TYPE_OUT_FOLDER=$(OUT_FOLDER)gamelift-servers-$(COMPUTE_TYPE)$(BUILD_PLAT_DIR_SEP)

download-server-sdk:
	if [ ! -f $(SERVER_SDK_FILE_NAME) ]; then \
		rm -rf $(SERVER_SDK_EXTRACT_DIR); \
		curl -L $(SERVER_SDK_URL) -o $(SERVER_SDK_FILE_NAME); \
	fi
	if [ ! -d $(SERVER_SDK_EXTRACT_DIR) ]; then \
		mkdir -p $(SERVER_SDK_EXTRACT_DIR); \
		unzip $(SERVER_SDK_FILE_NAME) -d $(SERVER_SDK_EXTRACT_DIR); \
	fi

build: download-server-sdk
	CGO_ENABLED=0 \
	GOARCH=$(GOARCH) \
	GOOS=$(GOOS) \
	go build \
	-C src \
	-trimpath \
	-v \
	-ldflags="-X '$(APP_PACKAGE)/internal.version=$(MAJOR).$(MINOR).$(PATCH)'" \
	-o ..$(BUILD_PLAT_DIR_SEP)$(OUT_FOLDER)$(APP_NAME)$(BUILD_PLAT_DIR_SEP)$(APP_NAME) \
	.

	@- $(foreach computeType,$(COMPUTE_TYPES),\
			$(MAKE) COMPUTE_TYPE=$(computeType) computeType; \
		)
	rm -rf $(OUT_FOLDER)$(APP_NAME)

computeType:
	mkdir -p $(COMPUTE_TYPE_OUT_FOLDER)

	cp $(OUT_FOLDER)$(APP_NAME)$(BUILD_PLAT_DIR_SEP)$(APP_NAME) \
		$(COMPUTE_TYPE_OUT_FOLDER)

	cp src$(BUILD_PLAT_DIR_SEP)template$(BUILD_PLAT_DIR_SEP)template-$(COMPUTE_TYPE)-config.yaml \
		$(COMPUTE_TYPE_OUT_FOLDER)config.yaml

ifeq ($(COMPUTE_TYPE),managed-containers)
	# Files specific to managed-containers
	cp src$(BUILD_PLAT_DIR_SEP)template$(BUILD_PLAT_DIR_SEP)Dockerfile $(COMPUTE_TYPE_OUT_FOLDER)
endif

rebuild: clean build

build-all:
	@- $(foreach platform,$(64_PLATFORMS),\
			$(eval os = $(word 1,$(subst /, ,$(platform)))) \
			$(eval arch = $(word 2,$(subst /, ,$(platform)))) \
			GOARCH=$(arch) GOOS=$(os) \
			$(MAKE) build; \
		)

clean:
	rm -rf $(OUT)
	rm -rf $(SERVER_SDK_EXTRACT_DIR)
	rm -f $(SERVER_SDK_FILE_NAME)

test: build
	go test \
	-C src \
	./...
