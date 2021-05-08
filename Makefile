GO_BUILD_FLAGS =
DOCKER_PLUGINS = docker-generate docker-nv2
COMMANDS       = nv2 $(DOCKER_PLUGINS)

define BUILD_BINARY =
	go build $(GO_BUILD_FLAGS) -o $@ ./$<
endef

.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: build

.PHONY: FORCE
FORCE:

bin/%: cmd/% FORCE
	$(BUILD_BINARY)

.PHONY: download
download: ## download dependencies via go mod
	go mod download

.PHONY: build
build: $(addprefix bin/,$(COMMANDS)) ## builds binaries

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
install: install-nv2 install-docker-plugins

.PHONY: install-nv2
install-nv2: bin/nv2
	cp $< ~/bin/

.PHONY: install-docker-%
install-docker-%: bin/docker-%
	cp $< ~/.docker/cli-plugins/

.PHONY: install-docker-plugins
install-docker-plugins: $(addprefix install-,$(DOCKER_PLUGINS)) ## installs the docker plugins
	cp $(addprefix bin/,$(DOCKER_PLUGINS)) ~/.docker/cli-plugins/
