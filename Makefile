VERSION ?= 0.8.0-dev
CLI_NAME ?= kubectl-prof
CLI_DIR ?= ./cmd/cli/
AGENT_NAME ?= agent
AGENT_DIR ?= ./cmd/agent/
BUILD_DIR ?= bin
REGISTRY ?= docker.io
DOCKER_BASE_IMAGE ?= josepdcs/kubectl-prof
DOCKER_JVM_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-jvm
DOCKERFILE_JVM ?= ./docker/jvm/Dockerfile
DOCKER_JVM_ALPINE_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-jvm-alpine
DOCKERFILE_JVM_ALPINE ?= ./docker/jvm/alpine/Dockerfile
DOCKER_BPF_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-bpf
DOCKERFILE_BPF ?= ./docker/bpf/Dockerfile
DOCKER_PERF_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-perf
DOCKERFILE_PERF ?= ./docker/perf/Dockerfile
DOCKER_PYTHON_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-python
DOCKERFILE_PYTHON ?= ./docker/python/Dockerfile
DOCKER_RUBY_IMAGE ?= $(DOCKER_BASE_IMAGE):$(VERSION)-ruby
DOCKERFILE_RUBY ?= ./docker/ruby/Dockerfile

M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: build
build: build-cli

.PHONY: all
all: build-cli push-docker-jvm push-docker-jvm-alpine push-docker-bpf push-docker-perf push-docker-python push-docker-ruby

.PHONY: agents
agents: build-docker-bpf build-docker-jvm build-docker-jvm-alpine build-docker-perf build-docker-python build-docker-ruby

.PHONY: install-deps
install-deps: ## Get the dependencies
	$(info $(M) getting dependencies...)
	@go get -v -d ./...

.PHONY: build-cli
build-cli: install-deps ## Build the binary file
	$(info $(M) building kubectl plugin...)
	@go build -ldflags="-X 'github.com/josepdcs/kubectl-prof/internal/cli/version.semver=$(VERSION)'" -o $(BUILD_DIR)/$(CLI_NAME) -v $(CLI_DIR)

.PHONY: build-agent
build-agent: install-deps ## Build the binary file
	$(info $(M) building agent...)
	@go build -o $(BUILD_DIR)/$(AGENT_NAME) -v $(AGENT_DIR)

.PHONY: build-docker-jvm
build-docker-jvm:
	$(info $(M) building JVM docker image...)
	@docker build -t ${DOCKER_JVM_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_JVM) .

.PHONY: push-docker-jvm
push-docker-jvm: build-docker-jvm
	$(info $(M) pushing JVM docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_JVM_IMAGE)

.PHONY: build-docker-jvm-alpine
build-docker-jvm-alpine:
	$(info $(M) building JVM Alpine docker image...)
	@docker build -t ${DOCKER_JVM_ALPINE_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_JVM_ALPINE) .

.PHONY: push-docker-jvm-alpine
push-docker-jvm-alpine: build-docker-jvm-alpine
	$(info $(M) pushing JVM Alpine docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_JVM_ALPINE_IMAGE)

.PHONY: build-docker-bpf
build-docker-bpf:
	$(info $(M) building BPF docker image...)
	docker build -t ${DOCKER_BPF_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_BPF) .

.PHONY: push-docker-bpf
push-docker-bpf: build-docker-bpf
	$(info $(M) pushing BPF docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_BPF_IMAGE)

.PHONY: build-docker-perf
build-docker-perf:
	$(info $(M) building PERF docker image...)
	docker build --no-cache -t ${DOCKER_PERF_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_PERF) .

.PHONY: push-docker-perf
push-docker-perf: build-docker-perf
	$(info $(M) pushing PERF docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_PERF_IMAGE)

.PHONY: build-docker-python
build-docker-python:
	$(info $(M) building PYTHON docker image...)
	docker build -t ${DOCKER_PYTHON_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_PYTHON) .

.PHONY: push-docker-python
push-docker-python: build-docker-python
	$(info $(M) pushing PYTHON docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_PYTHON_IMAGE)

.PHONY: build-docker-ruby
build-docker-ruby:
	$(info $(M) building RUBY docker image...)
	docker build -t ${DOCKER_RUBY_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_RUBY) .

.PHONY: push-docker-ruby
push-docker-ruby: build-docker-ruby
	$(info $(M) pushing RUBY docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_RUBY_IMAGE)

.PHONY: test
test:
	$(info $(M) running tests...)
	GOARCH=amd64 GOOS=linux go test ./... -coverprofile=coverage.out

.PHONY: coverage
coverage: test
	$(info $(M) running tests and coverage...)
	GOARCH=amd64 GOOS=linux go tool cover -html=coverage.out && unlink coverage.out

.PHONY: debug
debug: clean
	GOARCH=amd64 GOOS=linux go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(CLI_NAME)

## lint: Run go vet
.PHONY: gocyclo
gocyclo:
	$(info $(M) running go gocyclo (complexity)...)
	@gocyclo -over 14 .

## lint: Run go vet
.PHONY: vet
vet:
	$(info $(M) running go vet...)
	@go vet ./...

## check: Check the code
.PHONY: check
check: vet;
	$(info $(M) checking code...)

.PHONY: clean
clean: ## Remove previous build
	$(info $(M) cleaning all..)
	@rm -f coverage.out
	@rm -rf $(BUILD_DIR)