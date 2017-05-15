# Makefile used to build XDS daemon Web Server

# Application Version
VERSION := 0.0.1

# Syncthing version to install
SYNCTHING_VERSION = 0.14.27
SYNCTHING_INOTIFY_VERSION = 0.8.5



# Retrieve git tag/commit to set sub-version string
ifeq ($(origin SUB_VERSION), undefined)
	SUB_VERSION := $(shell git describe --tags --always | sed 's/^v//')
	ifeq ($(SUB_VERSION), )
		SUB_VERSION=unknown-dev
	endif
endif

# Configurable variables for installation (default /usr/local/...)
ifeq ($(origin INSTALL_DIR), undefined)
	INSTALL_DIR := /usr/local/bin
endif

HOST_GOOS=$(shell go env GOOS)
HOST_GOARCH=$(shell go env GOARCH)
ARCH=$(HOST_GOOS)-$(HOST_GOARCH)
REPOPATH=github.com/iotbzh/xds-agent

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_SRCDIR := $(patsubst %/,%,$(dir $(mkfile_path)))
ROOT_GOPRJ := $(abspath $(ROOT_SRCDIR)/../../../..)
LOCAL_BINDIR := $(ROOT_SRCDIR)/bin
PACKAGE_DIR := $(ROOT_SRCDIR)/package

export GOPATH := $(shell go env GOPATH):$(ROOT_GOPRJ)
export PATH := $(PATH):$(ROOT_SRCDIR)/tools

VERBOSE_1 := -v
VERBOSE_2 := -v -x


all: build

build: vendor tools/syncthing
	@echo "### Build XDS agent (version $(VERSION), subversion $(SUB_VERSION))";
	@cd $(ROOT_SRCDIR); $(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o $(LOCAL_BINDIR)/xds-agent -ldflags "-X main.AppVersion=$(VERSION) -X main.AppSubVersion=$(SUB_VERSION)" .

package: clean build
	@mkdir -p $(PACKAGE_DIR)/xds-agent
	@cp agent-config.json.in $(PACKAGE_DIR)/xds-agent/agent-config.json
	@cp -a $(LOCAL_BINDIR)/* $(PACKAGE_DIR)/xds-agent
	cd $(PACKAGE_DIR) && zip -r $(LOCAL_BINDIR)/xds-agent_$(ARCH)-v$(VERSION)_$(SUB_VERSION).zip ./xds-agent

test: tools/glide
	go test --race $(shell ./tools/glide novendor)

vet: tools/glide
	go vet $(shell ./tools/glide novendor)

fmt: tools/glide
	go fmt $(shell ./tools/glide novendor)

run: build/xds tools/syncthing
	$(LOCAL_BINDIR)/xds-agent --log info -c agent-config.json.in

debug: build/xds tools/syncthing
	$(LOCAL_BINDIR)/xds-agent --log debug -c agent-config.json.in

.PHONY: clean
clean:
	rm -rf $(LOCAL_BINDIR)/* debug $(ROOT_GOPRJ)/pkg/*/$(REPOPATH) $(PACKAGE_DIR)

.PHONY: distclean
distclean: clean
	rm -rf $(LOCAL_BINDIR) tools glide.lock vendor

.PHONY: install
install: all
	mkdir -p $(INSTALL_DIR) && cp $(LOCAL_BINDIR)/* $(INSTALL_DIR)

vendor: tools/glide glide.yaml
	./tools/glide install --strip-vendor

tools/glide:
	@echo "Downloading glide"
	mkdir -p tools
	curl --silent -L https://glide.sh/get | GOBIN=./tools  sh

.PHONY: tools/syncthing
tools/syncthing:
	@(test -s $(LOCAL_BINDIR)/syncthing || \
	DESTDIR=$(LOCAL_BINDIR) \
	SYNCTHING_VERSION=$(SYNCTHING_VERSION) \
	SYNCTHING_INOTIFY_VERSION=$(SYNCTHING_INOTIFY_VERSION) \
	./scripts/get-syncthing.sh)

.PHONY: help
help:
	@echo "Main supported rules:"
	@echo "  build               (default)"
	@echo "  package"
	@echo "  install"
	@echo "  clean"
	@echo "  distclean"
	@echo ""
	@echo "Influential make variables:"
	@echo "  V                 - Build verbosity {0,1,2}."
	@echo "  BUILD_ENV_FLAGS   - Environment added to 'go build'."
