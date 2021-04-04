GO_BUILD_FLAGS = 
DOCKER_PLUGINS = docker-generate docker-nv2
COMMANDS       = nv2 $(DOCKER_PLUGINS)

define BUILD_BINARY =
	go build $(GO_BUILD_FLAGS) -o $@ ./$<
endef

.PHONY: all
all: build

bin/%: cmd/%
	$(BUILD_BINARY)

.PHONY: build
build: $(addprefix bin/,$(COMMANDS))

.PHONY: clean
clean:
	git status --ignored --short | grep '^!! ' | sed 's/!! //' | xargs rm -rf

.PHONY: check-encoding
check-encoding:
	! find cmd pkg internal -name "*.go" -type f -exec file "{}" ";" | grep CRLF

.PHONY: fix-encoding
fix-encoding:
	find cmd pkg internal -type f -name "*.go" -exec sed -i -e "s/\r//g" {} +

.PHONY: vendor
vendor:
	GO111MODULE=on go mod vendor

.PHONY: install-docker-plugins
install-docker-plugins: $(addprefix bin/,$(DOCKER_PLUGINS))
	cp $(addprefix bin/,$(DOCKER_PLUGINS)) ~/.docker/cli-plugins/
