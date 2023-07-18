# Copyright The Notary Project Authors.
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

MODULE         = github.com/notaryproject/notation
COMMANDS       = notation
GIT_TAG        = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_COMMIT     = $(shell git rev-parse HEAD)

# if the commit was tagged, BuildMetadata is empty.
ifndef BUILD_METADATA
	BUILD_METADATA := unreleased
endif

ifneq ($(GIT_TAG),)
	BUILD_METADATA := 
endif

# set flags
LDFLAGS := -w \
 -X $(MODULE)/internal/version.GitCommit=$(GIT_COMMIT) \
 -X $(MODULE)/internal/version.BuildMetadata=$(BUILD_METADATA)

GO_BUILD_FLAGS = --ldflags="$(LDFLAGS)"

.PHONY: help
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: build

.PHONY: FORCE
FORCE:

bin/%: cmd/% FORCE
	go build $(GO_INSTRUMENT_FLAGS) $(GO_BUILD_FLAGS) -o $@ ./$<

.PHONY: download
download: ## download dependencies via go mod
	go mod download

.PHONY: build
build: $(addprefix bin/,$(COMMANDS)) ## builds binaries

.PHONY: test
test: vendor check-line-endings ## run unit tests
	go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...


.PHONY: e2e
e2e: build ## build notation cli and run e2e test
	NOTATION_BIN_PATH=`pwd`/bin/$(COMMANDS); \
	cd ./test/e2e; \
	./run.sh zot $$NOTATION_BIN_PATH; \

.PHONY: e2e-covdata
e2e-covdata:
	export GOCOVERDIR=$(CURDIR)/test/e2e/.cover; \
	rm -rf $$GOCOVERDIR; \
	mkdir -p $$GOCOVERDIR; \
	export GO_INSTRUMENT_FLAGS='-coverpkg "github.com/notaryproject/notation/internal/...,github.com/notaryproject/notation/pkg/...,github.com/notaryproject/notation/cmd/..."'; \
	$(MAKE) e2e && go tool covdata textfmt -i=$$GOCOVERDIR -o "$(CURDIR)/test/e2e/coverage.txt"

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
install: install-notation ## install the notation cli

.PHONY: install-notation
install-notation: bin/notation ## installs the notation cli
	mkdir -p ~/bin
	cp $< ~/bin/

.PHONY: install-docker-%
install-docker-%: bin/docker-%
	cp $< ~/.docker/cli-plugins/
