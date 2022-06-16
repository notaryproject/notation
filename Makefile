MODULE         = github.com/notaryproject/notation
DOCKER_PLUGINS = docker-generate docker-notation
COMMANDS       = notation $(DOCKER_PLUGINS)
GIT_TAG        = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
BUILD_METADATA =
ifeq ($(GIT_TAG),) # unreleased build
    GIT_COMMIT     = $(shell git rev-parse HEAD)
    GIT_STATUS     = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "unreleased")
	BUILD_METADATA = $(GIT_COMMIT).$(GIT_STATUS)
endif
LDFLAGS        = -X $(MODULE)/internal/version.BuildMetadata=$(BUILD_METADATA)
GO_BUILD_FLAGS = --ldflags="$(LDFLAGS)"

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: build

.PHONY: FORCE
FORCE:

bin/%: cmd/% FORCE
	go build $(GO_BUILD_FLAGS) -o $@ ./$<

.PHONY: download
download: ## download dependencies via go mod
	go mod download

.PHONY: build
build: $(addprefix bin/,$(COMMANDS)) ## builds binaries

.PHONY: test
test: vendor check-line-endings ## run unit tests
	./scripts/test.sh

.PHONY: clean
clean:
	git status --ignored --short | grep '^!! ' | sed 's/!! //' | xargs rm -rf

.PHONY: check-line-endings
check-line-endings: ## check line endings
	! find cmd pkg internal -name "*.go" -type f -exec file "{}" ";" | grep CRLF

.PHONY: fix-line-endings
fix-line-endings: ## fix line endings
	find cmd pkg internal -type f -name "*.go" -exec sed -i -e "s/\r//g" {} +

.PHONY: vendor
vendor: ## vendores the go modules
	GO111MODULE=on go mod vendor

.PHONY: install
install: install-notation install-docker-plugins ## install the notation cli and docker plugins

.PHONY: install-notation
install-notation: bin/notation ## installs the notation cli
	cp $< ~/bin/

.PHONY: install-docker-%
install-docker-%: bin/docker-%
	cp $< ~/.docker/cli-plugins/

.PHONY: install-docker-plugins
install-docker-plugins: $(addprefix install-,$(DOCKER_PLUGINS)) ## installs the docker plugins
	cp $(addprefix bin/,$(DOCKER_PLUGINS)) ~/.docker/cli-plugins/
