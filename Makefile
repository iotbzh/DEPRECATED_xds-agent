# Makefile used to build XDS Agent

# Syncthing version to install
SYNCTHING_VERSION = 0.14.38
SYNCTHING_INOTIFY_VERSION = 0.8.7


# Retrieve git tag/commit to set version & sub-version strings
GIT_DESC := $(shell git describe --always --tags)
VERSION := $(firstword $(subst -, ,$(GIT_DESC)))
SUB_VERSION := $(subst $(VERSION)-,,$(GIT_DESC))
ifeq ($(VERSION), )
	VERSION := unknown-dev
endif

# Configurable variables for installation (default /opt/AGL/...)
ifeq ($(origin DESTDIR), undefined)
	DESTDIR := /opt/AGL/xds/agent
endif
ifeq ($(origin DESTDIR_WWW), undefined)
	DESTDIR_WWW := $(DESTDIR)/www
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
	GO_LDFLAGS=
	# disable compiler optimizations and inlining
	GO_GCFLAGS=-N -l
	BUILD_MODE="Debug mode"
else
	# optimized code without debug info
	GO_LDFLAGS=-s -w
	GO_GCFLAGS=
	BUILD_MODE="Release mode"
endif

ifeq ($(SUB_VERSION), )
	PACKAGE_ZIPFILE := xds-agent_$(ARCH)-v$(VERSION).zip
else
	PACKAGE_ZIPFILE := xds-agent_$(ARCH)-v$(VERSION)_$(SUB_VERSION).zip
endif


all: tools/syncthing build

.PHONY: build
build: vendor xds webapp

xds: scripts tools/syncthing/copytobin
	@echo "### Build XDS agent (version $(VERSION), subversion $(SUB_VERSION)) - $(BUILD_MODE)";
	@cd $(ROOT_SRCDIR); $(BUILD_ENV_FLAGS) go build $(VERBOSE_$(V)) -i -o $(LOCAL_BINDIR)/xds-agent$(EXT) -ldflags "$(GO_LDFLAGS) -X main.AppVersion=$(VERSION) -X main.AppSubVersion=$(SUB_VERSION)" -gcflags "$(GO_GCFLAGS)" .

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
	rm -rf $(LOCAL_BINDIR)/* $(ROOT_SRCDIR)/debug $(ROOT_GOPRJ)/pkg/*/$(REPOPATH) $(PACKAGE_DIR)

.PHONY: distclean
distclean: clean
	cd $(ROOT_SRCDIR) && rm -rf $(LOCAL_BINDIR) ./tools ./glide.lock ./vendor ./*.zip ./webapp/node_modules ./webapp/dist

webapp: webapp/install
	(cd webapp && gulp build)

webapp/debug:
	(cd webapp && gulp watch &)

webapp/install:
	(cd webapp && npm install)
	@if [ -d ${DESTDIR}/usr/local/etc ]; then rm -rf ${DESTDIR}/usr; fi

.PHONY: install
install:
	@test -e $(LOCAL_BINDIR)/xds-agent$(EXT) || { echo "Please execute first: make all\n"; exit 1; }
	@test -e $(LOCAL_BINDIR)/syncthing$(EXT) -a -e $(LOCAL_BINDIR)/syncthing-inotify$(EXT) || { echo "Please execute first: make all\n"; exit 1; }
	export DESTDIR=$(DESTDIR) && export DESTDIR_WWW=$(DESTDIR_WWW) && $(ROOT_SRCDIR)/scripts/install.sh

.PHONY: uninstall
uninstall:
	export DESTDIR=$(DESTDIR) && export DESTDIR_WWW=$(DESTDIR_WWW) && $(ROOT_SRCDIR)/scripts/install.sh uninstall

package: clean tools/syncthing vendor build
	@mkdir -p $(PACKAGE_DIR)/xds-agent $(PACKAGE_DIR)/scripts
	@cp -a $(LOCAL_BINDIR)/* $(PACKAGE_DIR)/xds-agent
	@cp -r $(ROOT_SRCDIR)/conf.d $(ROOT_SRCDIR)/scripts $(PACKAGE_DIR)/xds-agent
	cd $(PACKAGE_DIR) && zip -r $(ROOT_SRCDIR)/$(PACKAGE_ZIPFILE) ./xds-agent

.PHONY: package-all
package-all:
	@echo "# Build linux amd64..."
	GOOS=linux GOARCH=amd64 RELEASE=1 make -f $(ROOT_SRCDIR)/Makefile package
	@echo "# Build windows amd64..."
	GOOS=windows GOARCH=amd64 RELEASE=1 make -f $(ROOT_SRCDIR)/Makefile package
	@echo "# Build darwin amd64..."
	GOOS=darwin GOARCH=amd64 RELEASE=1 make -f $(ROOT_SRCDIR)/Makefile package
	make -f $(ROOT_SRCDIR)/Makefile clean

vendor: tools/glide glide.yaml
	$(LOCAL_TOOLSDIR)/glide install --strip-vendor

vendor/debug: vendor
	(cd vendor/github.com/iotbzh && \
		rm -rf xds-common && ln -s ../../../../xds-common )

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

.PHONY: tools/syncthing/copytobin
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
	@echo "  uninstall"
	@echo "  clean"
	@echo "  distclean"
	@echo ""
	@echo "Influential make variables:"
	@echo "  V                 - Build verbosity {0,1,2}."
	@echo "  BUILD_ENV_FLAGS   - Environment added to 'go build'."
