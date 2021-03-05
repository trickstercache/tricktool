# Copyright 2018 Comcast Cable Communications Management, LLC
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DEFAULT: build

PROJECT_DIR    := $(shell pwd)
GO             ?= go
GOFMT          ?= $(GO)fmt
FIRST_GOPATH   := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
TRICKTOOL      := $(FIRST_GOPATH)/bin/tricktool
PROGVER        := $(shell grep 'applicationVersion = ' ./main.go | awk '{print $$4}' | sed -e 's/\"//g')
BUILD_TIME     := $(shell date -u +%FT%T%z)
GIT_LATEST_COMMIT_ID     := $(shell git rev-parse HEAD)
GO_VER         := $(shell go version | awk '{print $$3}')
IMAGE_TAG      ?= latest
IMAGE_ARCH     ?= amd64
GOARCH         ?= amd64
TAGVER         ?= unspecified
LDFLAGS         =-ldflags "-extldflags '-static' -w -s"
BUILD_SUBDIR   := OPATH
PACKAGE_DIR    := ./$(BUILD_SUBDIR)/tricktool-$(PROGVER)
BIN_DIR        := $(PACKAGE_DIR)/bin
CGO_ENABLED    ?= 0

.PHONY: spelling
spelling:
	@codespell --skip="vendor,*.git,*.png,*.pdf,*.tiff,*.plist,*.pem,rangesim*.go,*.gz" --ignore-words="./testdata/ignore_words.txt"  -q 3

.PHONY: validate-app-version
validate-app-version:
	@if [ "$(PROGVER)" != $(TAGVER) ]; then\
		(echo "mismatch between TAGVER '$(TAGVER)' and applicationVersion '$(PROGVER)'"; exit 1);\
	fi

.PHONY: go-mod-vendor
go-mod-vendor:
	$(GO) mod vendor

.PHONY: go-mod-tidy
go-mod-tidy:
	$(GO) mod tidy

.PHONY: test-go-mod
test-go-mod:
	@git diff --quiet --exit-code go.mod go.sum || echo "There are changes to go.mod and go.sum which needs to be committed"

.PHONY: build
build: go-mod-tidy go-mod-vendor
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o ./$(BUILD_SUBDIR)/tricktool -a -v ./*.go

.PHONY: release
release: validate-app-version clean go-mod-tidy go-mod-vendor release-artifacts

.PHONY: release-artifacts
release-artifacts: clean

	mkdir -p $(PACKAGE_DIR)
	mkdir -p $(BIN_DIR)

	cp ./README.md $(PACKAGE_DIR)
	cp ./CONTRIBUTING.md $(PACKAGE_DIR)
	cp ./LICENSE $(PACKAGE_DIR)
	
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BIN_DIR)/tricktool-$(PROGVER).darwin-amd64  -a -v ./*.go
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BIN_DIR)/tricktool-$(PROGVER).linux-amd64   -a -v ./*.go
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BIN_DIR)/tricktool-$(PROGVER).linux-arm64   -a -v ./*.go
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(LDFLAGS) -o $(BIN_DIR)/tricktool-$(PROGVER).windows-amd64 -a -v ./*.go

	cd ./$(BUILD_SUBDIR) && tar cvfz ./tricktool-$(PROGVER).tar.gz ./tricktool-$(PROGVER)/*

.PHONY: style
style:
	! gofmt -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

.PHONY: test
test:
	@go test -v ./... 

.PHONY: clean
clean:
	rm -rf ./tricktool ./$(BUILD_SUBDIR)
