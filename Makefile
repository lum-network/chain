PACKAGES=$(shell go list ./... | grep -v '/simulation')

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf:1.7.0
HTTPS_GIT := https://github.com/lum-network/chain.git#branch=master

LEDGER_ENABLED ?= true
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=lum \
	-X github.com/cosmos/cosmos-sdk/version.AppName=lumd \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

### Base commands

.PHONY: build

build:
	@echo "--> Building lumd"
	@mkdir -p $(BUILDDIR)/
	@go build -mod=readonly $(BUILD_FLAGS) -trimpath -o $(BUILDDIR) ./...;

install: go.sum
	@echo "--> Installing lumd"
	@go install -mod=readonly $(BUILD_FLAGS) ./cmd/lumd

clean:
	@echo "--> Cleaning ${BUILDDIR}"
	@rm -rf $(BUILDDIR)/*

### Dev commends

format:
	@echo "--> Formating code and ordering imports"
	@goimports -local github.com/lum-network -w .
	@gofmt -w .

### CI commands

sec:
	@gosec -severity=high -exclude=G404 ./...

lint:
	@golangci-lint run --skip-dirs='(x/beam|x/dfract)'

format-check:
	@echo "--> Checking formatting issues"
	@format_output=$$(gofmt -l .); \
	if [ -z "$$format_output" ]; then \
		echo "No formatting issues found"; \
	else \
		echo "Formatting issues found in files:"; \
		echo "$$format_output"; \
		exit_code=1; \
	fi; \
	echo "--> Checking import ordering issues"; \
	import_output=$$(goimports -local github.com/lum-network -l .); \
	if [ -z "$$import_output" ]; then \
		echo "No import ordering issues found"; \
	else \
		echo "Import ordering issues found in files:"; \
		echo "$$import_output"; \
		exit_code=1; \
	fi; \
	exit $$exit_code

test:
	@go test -mod=readonly ./x/... ./app/...

### Proto commands

containerProtoVer=0.13.0
containerProtoImage=ghcr.io/cosmos/proto-builder:$(containerProtoVer)

proto-all: proto-format proto-lint proto-gen proto-format-gencode

proto-gen:
	@echo "--> Generating Protobuf files"
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(containerProtoImage) \
		sh ./scripts/protocgen.sh;

proto-format:
	@echo "--> Formatting Protobuf files"
	@$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace tendermintdev/docker-build-proto \
		find ./proto -name "*.proto" -exec clang-format -i {} \;

proto-format-gencode:
	@echo "--> Formatting Proto generated files"
	@goimports -local github.com/lum-network -w $$(find . -type f \( -iname \*.pb.go -o -iname \*.gw.go \))
	@gofmt -w $$(find . -type f \( -iname \*.pb.go -o -iname \*.gw.go \))

proto-lint:
	@echo "--> Linting Protobuf files"
	@$(DOCKER_BUF) lint --error-format=text

proto-check-breaking:
	@echo "$(HTTPS_GIT)"
	@$(DOCKER_BUF) breaking --against $(HTTPS_GIT)

### Other commands

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify
