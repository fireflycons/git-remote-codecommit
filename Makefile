# Local build and test for git-remote-codecommit.
#
# Requires GNU Make 4.3+ and Go. Works on Linux, macOS and Windows.
# On Windows, native GNU Make (e.g. from Chocolatey) runs recipes with
# cmd.exe; MSYS2/Cygwin make runs them with sh.
#
# All build output goes under ./bin.

GO       ?= go
BINARY   := git-remote-codecommit
BIN_DIR  := bin
PKGS     := ./...

GOEXE    := $(shell $(GO) env GOEXE)
LOCAL_BIN := $(BIN_DIR)/$(BINARY)$(GOEXE)
COVER_OUT := $(BIN_DIR)/coverage.out

BUILD_FLAGS ?= -trimpath
LDFLAGS     ?= -s -w

# Platforms built by `make cross` (same matrix as the release workflow).
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# --- Shell portability -------------------------------------------------------
# Native Windows make (MAKE_HOST=Windows32) may pick sh.exe or cmd.exe depending
# on PATH. Pin cmd.exe so behaviour is deterministic.
ifeq ($(MAKE_HOST),Windows32)
SHELL := cmd.exe
RMDIR  = if exist $(subst /,\,$(1)) rmdir /s /q $(subst /,\,$(1))
MKDIR  = if not exist $(subst /,\,$(1)) mkdir $(subst /,\,$(1))
RMFILE = if exist "$(1)" del /f /q "$(1)"
else
RMDIR  = rm -rf $(1)
MKDIR  = mkdir -p $(1)
RMFILE = rm -f "$(1)"
endif

# Where `go install` puts the binary (GOBIN, else first GOPATH entry's bin).
# Asking Go avoids parsing GOPATH lists ourselves. Lazily evaluated so only
# `uninstall` pays for it.
INSTALL_TARGET = $(shell $(GO) list -f {{.Target}} .)

# Exe suffix for a target GOOS.
exe = $(if $(filter windows,$(1)),.exe)

CROSS_BINS := $(foreach p,$(PLATFORMS),$(BIN_DIR)/$(subst /,_,$(p))/$(BINARY)$(call exe,$(word 1,$(subst /, ,$(p)))))

.DEFAULT_GOAL := all

.PHONY: all build test vet cover cross install uninstall tidy clean help FORCE

all: vet test build

build: $(LOCAL_BIN)

$(LOCAL_BIN): FORCE
	$(GO) build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o $@ .

test:
	$(GO) test $(PKGS)

vet:
	$(GO) vet $(PKGS)

cover:
	@$(call MKDIR,$(BIN_DIR))
	$(GO) test -coverprofile=$(COVER_OUT) $(PKGS)
	$(GO) tool cover -func=$(COVER_OUT)

cross: $(CROSS_BINS)

# One rule per platform. GOOS/GOARCH/CGO_ENABLED are exported as target-specific
# variables so no shell-specific `VAR=x cmd` syntax is needed.
define CROSS_RULE
$(BIN_DIR)/$(1)_$(2)/$(BINARY)$(call exe,$(1)): export GOOS := $(1)
$(BIN_DIR)/$(1)_$(2)/$(BINARY)$(call exe,$(1)): export GOARCH := $(2)
$(BIN_DIR)/$(1)_$(2)/$(BINARY)$(call exe,$(1)): export CGO_ENABLED := 0
$(BIN_DIR)/$(1)_$(2)/$(BINARY)$(call exe,$(1)): FORCE
	$$(GO) build $$(BUILD_FLAGS) -ldflags="$$(LDFLAGS)" -o $$@ .
endef
$(foreach p,$(PLATFORMS),$(eval $(call CROSS_RULE,$(word 1,$(subst /, ,$(p))),$(word 2,$(subst /, ,$(p))))))

install:
	$(GO) install $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" .

uninstall:
	$(call RMFILE,$(INSTALL_TARGET))

tidy:
	$(GO) mod tidy

clean:
	$(call RMDIR,$(BIN_DIR))

# Always hand off to `go build`; Go's build cache decides what is stale.
FORCE:

# $(info) is used rather than echo so output is identical under sh and cmd.exe.
help:
	$(info Targets:)
	$(info $()  all      Vet, test and build for the host platform (default))
	$(info $()  build    Build for the host platform into $(LOCAL_BIN))
	$(info $()  test     Run unit tests)
	$(info $()  vet      Run go vet)
	$(info $()  cover    Run tests with coverage, profile in $(COVER_OUT))
	$(info $()  cross    Cross-compile into $(BIN_DIR)/<os>_<arch>/ for: $(PLATFORMS))
	$(info $()  install  go install the helper into GOBIN or GOPATH/bin)
	$(info $()  uninstall Remove the binary that install put in GOBIN or GOPATH/bin)
	$(info $()  tidy     go mod tidy)
	$(info $()  clean    Remove $(BIN_DIR))
	@:
