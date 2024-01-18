VERSION ?= v1.3.0-dev
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

## build-cli: Build the kubectl-prof plugin and push all docker images
.PHONY: all
all: build-cli push-docker-jvm push-docker-jvm-alpine push-docker-bpf push-docker-perf push-docker-python push-docker-ruby

## build: Build the kubectl-prof plugin and the agent binary
.PHONY: build
build: build-cli build-agent

## build-docker-agents: Build the docker images
.PHONY: build-docker-agents
build-docker-agents: build-docker-bpf build-docker-jvm build-docker-jvm-alpine build-docker-perf build-docker-python build-docker-ruby

## install-deps: install dependencies if needed
.PHONY: install-deps
install-deps: ## Get the dependencies
	$(info $(M) getting dependencies...)
	@go get -v -d ./...

## build-cli: Build the kubectl-prof plugin
.PHONY: build-cli
build-cli: install-deps ## Build the binary file
	$(info $(M) building kubectl plugin...)
	@go build -ldflags="-X 'github.com/josepdcs/kubectl-prof/internal/cli/version.semver=$(VERSION)'" -o $(BUILD_DIR)/$(CLI_NAME) -v $(CLI_DIR)

## build-agent: Build the agent
.PHONY: build-agent
build-agent: install-deps ## Build the binary file
	$(info $(M) building agent...)
	@go build -o $(BUILD_DIR)/$(AGENT_NAME) -v $(AGENT_DIR)

## build-docker-jvm: Build the JVM docker image
.PHONY: build-docker-jvm
build-docker-jvm:
	$(info $(M) building JVM docker image...)
	@docker build -t ${DOCKER_JVM_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_JVM) .

## push-docker-jvm: Build and push the JVM docker image
.PHONY: push-docker-jvm
push-docker-jvm: build-docker-jvm
	$(info $(M) pushing JVM docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_JVM_IMAGE)

## build-docker-jvm-alpine: Build the JVM docker image for Alpine
.PHONY: build-docker-jvm-alpine
build-docker-jvm-alpine:
	$(info $(M) building JVM Alpine docker image...)
	@docker build -t ${DOCKER_JVM_ALPINE_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_JVM_ALPINE) .

## push-docker-jvm-alpine: Build and push the JVM docker image for Alpine
.PHONY: push-docker-jvm-alpine
push-docker-jvm-alpine: build-docker-jvm-alpine
	$(info $(M) pushing JVM Alpine docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_JVM_ALPINE_IMAGE)

## build-docker-bpf: Build the BPF docker image
.PHONY: build-docker-bpf
build-docker-bpf:
	$(info $(M) building BPF docker image...)
	docker build -t ${DOCKER_BPF_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_BPF) .

## push-docker-bpf: Build and push the BPF docker image
.PHONY: push-docker-bpf
push-docker-bpf: build-docker-bpf
	$(info $(M) pushing BPF docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_BPF_IMAGE)

## build-docker-perf: Build the PERF docker image
.PHONY: build-docker-perf
build-docker-perf:
	$(info $(M) building PERF docker image...)
	docker build --no-cache -t ${DOCKER_PERF_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_PERF) .

## push-docker-perf: Build and push the PERF docker image
.PHONY: push-docker-perf
push-docker-perf: build-docker-perf
	$(info $(M) pushing PERF docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_PERF_IMAGE)

## build-docker-python: Build the PYTHON docker image
.PHONY: build-docker-python
build-docker-python:
	$(info $(M) building PYTHON docker image...)
	docker build -t ${DOCKER_PYTHON_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_PYTHON) .

## push-docker-python: Build and push the PYTHON docker image
.PHONY: push-docker-python
push-docker-python: build-docker-python
	$(info $(M) pushing PYTHON docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_PYTHON_IMAGE)

## build-docker-ruby: Build the RUBY docker image
.PHONY: build-docker-ruby
build-docker-ruby:
	$(info $(M) building RUBY docker image...)
	docker build -t ${DOCKER_RUBY_IMAGE} --label git-commit=$(shell git rev-parse HEAD) -f $(DOCKERFILE_RUBY) .

## push-docker-ruby: Build and push the RUBY docker image
.PHONY: push-docker-ruby
push-docker-ruby: build-docker-ruby
	$(info $(M) pushing RUBY docker image to DockerHub...)
	@docker push $(REGISTRY)/$(DOCKER_RUBY_IMAGE)

## push-docker-all: Build and push all docker images
.PHONY: push-docker-all
push-docker-all: push-docker-jvm push-docker-jvm-alpine push-docker-bpf push-docker-perf push-docker-python push-docker-ruby

## test: Run unit tests
.PHONY: test
test:
	$(info $(M) running tests...)
	GOARCH=amd64 GOOS=linux go test -p 1 ./... -coverprofile=coverage.out

## coverage: Run unit tests and show coverage
.PHONY: coverage
coverage: test
	$(info $(M) running tests and coverage...)
	GOARCH=amd64 GOOS=linux go tool cover -html=coverage.out && unlink coverage.out

## build-debug: Build the kubectl prof with debug info
.PHONY: build-debug
debug: clean
	GOARCH=amd64 GOOS=linux go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(CLI_NAME)

## gocyclo: Run gocyclo
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

## clean: Clean build files. Runs `go clean` internally.
.PHONY: clean
clean:
	$(info $(M) cleaning all..)
	@rm -f coverage.out
	@rm -rf $(BUILD_DIR)
	@$(GO) clean

## version: Show the project version
.PHONY: version
version:
	@echo $(VERSION)

## minikube-all: Start and build all minikube environment
.PHONY: minikube-all
minikube-all:
	@test/minikube-lab/run_orchestration.sh

## minikube-start-clusters: Start minikube clusters
.PHONY: minikube-start-clusters
minikube-start-clusters:
	@test/minikube-lab/start_clusters.sh

## minikube-stop-clusters: Stop minikube clusters
.PHONY: minikube-stop-clusters
minikube-stop-clusters:
	@test/minikube-lab/stop_clusters.sh

## minikube-build-and-push-all: Build all images (stupid-apps and agents) and load them into minikube
.PHONY: minikube-build-and-push-all
minikube-build-and-push-all: minikube-build-and-push-stupid-apps minikube-build-and-push-agents

## minikube-build-and-push-stupid-apps: Build images of stupid stupid-apps and load them into minikube
.PHONY: minikube-build-and-push-stupid-apps
minikube-build-and-push-stupid-apps:
	@test/minikube-lab/build_and_push_stupid_apps.sh

## minikube-build-and-push-agents: Build images of agents and load them into minikube
.PHONY: minikube-build-and-push-agents
minikube-build-and-push-agents:
	@test/minikube-lab/build_and_push_agents.sh

## minikube-build-and-push-jvm-agent: Build image of jvm agent and load it into minikube
.PHONY: minikube-build-and-push-jvm-agent
minikube-build-and-push-jvm-agent:
	@test/minikube-lab/build_and_push_image.sh "docker/jvm" "docker" "jvm"

## minikube-build-and-push-jvm-alpine-agent: Build image of jvm alpine agent and load it into minikube
.PHONY: minikube-build-and-push-jvm-alpine-agent
minikube-build-and-push-jvm-alpine-agent:
	@test/minikube-lab/build_and_push_image.sh "docker/jvm/alpine" "docker" "jvm-alpine"

## minikube-build-and-push-bpf-agent: Build image of bpf agent and load it into minikube
.PHONY: minikube-build-and-push-bpf-agent
minikube-build-and-push-bpf-agent:
	@test/minikube-lab/build_and_push_image.sh "docker/bpf" "docker" "bpf"

## minikube-build-and-push-python-agent: Build image of python agent and load it into minikube
.PHONY: minikube-build-and-push-python-agent
minikube-build-and-push-python-agent:
	@test/minikube-lab/build_and_push_image.sh "docker/python" "docker" "python"

## minikube-build-and-push-ruby-agent: Build image of ruby agent and load it into minikube
.PHONY: minikube-build-and-push-ruby-agent
minikube-build-and-push-ruby-agent:
	@test/minikube-lab/build_and_push_image.sh "docker/ruby" "docker" "ruby"

## minikube-deploy-stupid-apps: Deploy stupid apps into minikube
.PHONY: minikube-deploy-stupid-apps
minikube-deploy-stupid-apps:
	@test/minikube-lab/deploy_stupid_apps.sh

## minikube-build-and-push-ruby-stupid-app: Build image of ruby stupid app and load it into minikube
.PHONY: minikube-build-and-push-ruby-stupid-app
minikube-build-and-push-ruby-stupid-app:
	@test/minikube-lab/build_and_push_image.sh "test/stupid-apps/ruby" "stupid-apps" "ruby"

## minikube-build-and-push-node-stupid-app: Build image of node stupid app and load it into minikube
.PHONY: minikube-build-and-push-node-stupid-app
minikube-build-and-push-node-stupid-app:
	@test/minikube-lab/build_and_push_image.sh "test/stupid-apps/node" "stupid-apps" "node"

## minikube-build-and-push-python-stupid-app: Build image of python stupid app and load it into minikube
.PHONY: minikube-build-and-push-python-stupid-app
minikube-build-and-push-python-stupid-app:
	@test/minikube-lab/build_and_push_image.sh "test/stupid-apps/python" "stupid-apps" "python"

## minikube-build-and-push-jvm-stupid-app: Build image of jvm stupid app and load it into minikube
.PHONY: minikube-build-and-push-jvm-stupid-app
minikube-build-and-push-jvm-stupid-app:
	@test/minikube-lab/build_and_push_image.sh "test/stupid-apps/jvm" "stupid-apps" "jvm"

## minikube-configure-profiling: Configure all needed for profiling (service account, namespace, etc.)
.PHONY: minikube-configure-profiling
minikube-configure-profiling:
	@test/minikube-lab/conf_profiling.sh

## help: This message
.PHONY: help
help: Makefile
	@echo
	@echo " Choose a command:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo