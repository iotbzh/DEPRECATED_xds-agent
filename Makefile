# Makefile used to build XDS daemon Web Server

# Application Version
VERSION := 0.1.0

# Syncthing version to install
SYNCTHING_VERSION = 0.14.28
SYNCTHING_INOTIFY_VERSION = 0.8.6



# Retrieve git tag/commit to set sub-version string
ifeq ($(origin SUB_VERSION), undefined)
	SUB_VERSION := $(shell git describe --exact-match --tags 2>/dev/null | sed 's/^v//')
	ifneq ($(SUB_VERSION), )
		VERSION := $(firstword $(subst -, ,$(SUB_VERSION)))
		SUB_VERSION := $(word 2,$(subst -, ,$(SUB_VERSION)))
	endif
	ifeq ($(SUB_VERSION), )
		SUB_VERSION := $(shell git rev-parse --short HEAD)
		ifeq ($(SUB_VERSION), )
			SUB_VERSION := unknown-dev
		endif
	endif
endif

# for backward compatibility
DESTDIR := $(INSTALL_DIR)

# Configurable variables for installation (default /usr/local/...)
ifeq ($(origin DESTDIR), undefined)
	DESTDIR := /usr/local/bin
endif

HOST_GOOS=$(shell go env GOOS)
HOST_GOARCH=$(shell go env GOARCH)
ARCH=$(HOST_GOOS)-$(HOST_GOARCH)
REPOPATH=github.com/iotbzh/xds-agent

EXT=
ifeq ($(HOST_GOOS), windows)
	EXT=.exe
endif

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_SRCDIR := $(patsubst %/,%,$(dir $(mkfile_path)))
ROOT_GOPRJ := $(abspath $(ROOT_SRCDIR)/../../../..)
LOCAL_BINDIR := $(ROOT_SRCDIR)/bin
LOCAL_TOOLSDIR := $(ROOT_SRCDIR)/tools/${HOST_GOOS}
PACKAGE_DIR := $(ROOT_SRCDIR)/package

export GOPATH := $(shell go env GOPATH):$(ROOT_GOPRJ)
export PATH := $(PATH):$(LOCAL_TOOLSDIR)

VERBOSE_1 := -v
VERBOSE_2 := -v -x

# Release or Debug mode
ifeq ($(filter 1,$(RELEASE) $(REL)),)
	GORELEASE=
	BUILD_MODE="Debug mode"
else
	# optimized code without debug info
	GORELEASE= -s -w
	BUILD_MODE="Release mode"
endif

ifeq ($(SUB_VERSION), )
	PACKAGE_ZIPFILE := xds-agent_$(ARCH)-v$(VERSION).zip
else
	PACKAGE_ZIPFILE := xds-agent_$(ARCH)-v$(VERSION)_$(SUB_VERSION).zip
endif


all: tools/syncthing vendor build

build: tools/syncthing/copytobin
	@echo "### Build XDS agent (version $(VERSION), subversion $(SUB_VERSION)) - $(BUILD_MODE)";
	@cd $(ROOT_SRCDIR); $(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o $(LOCAL_BINDIR)/xds-agent$(EXT) -ldflags "$(GORELEASE) -X main.AppVersion=$(VERSION) -X main.AppSubVersion=$(SUB_VERSION)" .

package: clean tools/syncthing vendor build
	@mkdir -p $(PACKAGE_DIR)/xds-agent
	@cp agent-config.json.in $(PACKAGE_DIR)/xds-agent/agent-config.json
	@cp -a $(LOCAL_BINDIR)/* $(PACKAGE_DIR)/xds-agent
	cd $(PACKAGE_DIR) && zip -r $(ROOT_SRCDIR)/$(PACKAGE_ZIPFILE) ./xds-agent

.PHONY: package-all
package-all:
	@echo "# Build linux amd64..."
	GOOS=linux GOARCH=amd64 RELEASE=1 make -f $(ROOT_SRCDIR)/Makefile package
	@echo "# Build windows amd64..."
	GOOS=windows GOARCH=amd64 RELEASE=1 make -f $(ROOT_SRCDIR)/Makefile package
	@echo "# Build darwin amd64..."
	GOOS=darwin GOARCH=amd64 RELEASE=1 make -f $(ROOT_SRCDIR)/Makefile package

test: tools/glide
	go test --race $(shell $(LOCAL_TOOLSDIR)/glide novendor)

vet: tools/glide
	go vet $(shell $(LOCAL_TOOLSDIR)/glide novendor)

fmt: tools/glide
	go fmt $(shell $(LOCAL_TOOLSDIR)/glide novendor)

run: build/xds tools/syncthing/copytobin
	$(LOCAL_BINDIR)/xds-agent$(EXT) --log info -c agent-config.json.in

debug: build/xds tools/syncthing/copytobin
	$(LOCAL_BINDIR)/xds-agent$(EXT) --log debug -c agent-config.json.in

.PHONY: clean
clean:
	rm -rf $(LOCAL_BINDIR)/* debug $(ROOT_GOPRJ)/pkg/*/$(REPOPATH) $(PACKAGE_DIR)

.PHONY: distclean
distclean: clean
	rm -rf $(LOCAL_BINDIR) tools glide.lock vendor $(ROOT_SRCDIR)/*.zip

.PHONY: install
install: all
	mkdir -p $(DESTDIR) && cp $(LOCAL_BINDIR)/* $(DESTDIR)

vendor: tools/glide glide.yaml
	$(LOCAL_TOOLSDIR)/glide install --strip-vendor

.PHONY: tools/glide
tools/glide:
	@test -f $(LOCAL_TOOLSDIR)/glide || { \
		echo "Downloading glide"; \
		mkdir -p $(LOCAL_TOOLSDIR); \
		curl --silent -L https://glide.sh/get | GOBIN=$(LOCAL_TOOLSDIR)  sh; \
	}

.PHONY: tools/syncthing
tools/syncthing:
	@test -e $(LOCAL_TOOLSDIR)/syncthing$(EXT) -a -e $(LOCAL_TOOLSDIR)/syncthing-inotify$(EXT)  || { \
	mkdir -p $(LOCAL_TOOLSDIR); \
	DESTDIR=$(LOCAL_TOOLSDIR) \
	SYNCTHING_VERSION=$(SYNCTHING_VERSION) \
	SYNCTHING_INOTIFY_VERSION=$(SYNCTHING_INOTIFY_VERSION) \
	./scripts/get-syncthing.sh; }

.PHONY:
tools/syncthing/copytobin:
	@test -e $(LOCAL_TOOLSDIR)/syncthing$(EXT) -a -e $(LOCAL_TOOLSDIR)/syncthing-inotify$(EXT) || { echo "Please execute first: make tools/syncthing\n"; exit 1; }
	@mkdir -p $(LOCAL_BINDIR)
	@cp -f $(LOCAL_TOOLSDIR)/syncthing$(EXT) $(LOCAL_TOOLSDIR)/syncthing-inotify$(EXT) $(LOCAL_BINDIR)

.PHONY: help
help:
	@echo "Main supported rules:"
	@echo "  all               (default)"
	@echo "  build"
	@echo "  package"
	@echo "  install"
	@echo "  clean"
	@echo "  distclean"
	@echo ""
	@echo "Influential make variables:"
	@echo "  V                 - Build verbosity {0,1,2}."
	@echo "  BUILD_ENV_FLAGS   - Environment added to 'go build'."
