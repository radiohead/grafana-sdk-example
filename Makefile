BIN_DIR              := bin
SOURCES              := $(shell find . -type f -name "*.go")
MOD_FILES            := go.mod go.sum
COVOUT               := coverage.out

OPERATOR_DOCKERIMAGE := "grafana-sdk-example"

SDK_VER              := 0.14.7
LINTER_VERSION       := 1.55.2

.PHONY: all
all: deps generate lint test build

.PHONY: deps
deps:
	@go mod tidy

LINTER_BINARY  := $(BIN_DIR)/golangci-lint-$(LINTER_VERSION) $@
LINT_ARGS := --max-same-issues=0 --max-issues-per-linter=0 --exclude-use-default=false

.PHONY: lint
lint: $(LINTER_BINARY)
	$(LINTER_BINARY) run $(LINT_ARGS)

$(LINTER_BINARY):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(BIN_DIR) v$(LINTER_VERSION)
	@mv $(BIN_DIR)/golangci-lint $@

.PHONY: test
test:
	go test -count=1 -cover -covermode=atomic -coverprofile=$(COVOUT) ./...

.PHONY: coverage
coverage: test
	go tool cover -html=$(COVOUT)

.PHONY: build
build: build/plugin build/operator

.PHONY: build/plugin
build/plugin: build/plugin-backend build/plugin-frontend

.PHONY: build/plugin-frontend
build/plugin-frontend:
ifeq ("$(wildcard plugin/src/plugin.json)","plugin/src/plugin.json")
	@cd plugin && yarn install && yarn build
else
	@echo "No plugin.json found, skipping frontend build"
endif

.PHONY: build/plugin-backend
build/plugin-backend:
ifeq ("$(wildcard plugin/Magefile.go)","plugin/Magefile.go")
	@cd plugin && mage -v
else
	@echo "No Magefile.go found, skipping backend build"
endif

.PHONY: build/operator
build/operator:
	docker build -t $(OPERATOR_DOCKERIMAGE) -f cmd/operator/Dockerfile .

.PHONY: compile/operator
compile/operator:
	@go build cmd/operator -o target/operator

SDK_CLI := $(BIN_DIR)/grafana-app-sdk-$(SDK_VER)
$(SDK_CLI):
	mkdir -p $(BIN_DIR)
	@./scripts/install-sdk.sh $(SDK_VER) $(BIN_DIR)

.PHONY: generate
generate: $(SDK_CLI)
	@$(SDK_CLI) generate -c kinds
	@./scripts/prettify-crds.sh

.PHONY: local/up
local/up: local/generate
	@sh local/scripts/cluster.sh create "local/generated/k3d-config.json"
	@cd local && tilt up

.PHONY: local/generate
local/generate: $(SDK_CLI)
	@$(SDK_CLI) project local generate

.PHONY: local/down
local/down:
	@cd local && tilt down

.PHONY: local/deploy_plugin
local/deploy_plugin:
	-tilt disable grafana
	cp -R plugin/dist local/mounted-files/plugin/dist
	-tilt enable grafana

.PHONY: local/push_operator
local/push_operator:
	# Tag the docker image as part of localhost, which is what the generated k8s uses to avoid confusion with the real operator image
	@docker tag "$(OPERATOR_DOCKERIMAGE):latest" "localhost/$(OPERATOR_DOCKERIMAGE):latest"
	@sh local/scripts/push_image.sh "localhost/$(OPERATOR_DOCKERIMAGE):latest"

.PHONY: local/clean
local/clean: local/down
	@sh local/scripts/cluster.sh delete

.PHONY: clean
clean:
	@rm -f $(COVOUT)
