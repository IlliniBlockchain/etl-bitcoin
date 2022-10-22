PROJECT_NAME := "etl-bitcoin"
PKG := "github.com/IlliniBlockchain/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/ | grep -v mocks)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all dep build clean test coverage coverhtml lint docker.start docker.stop docker.restart test.unit test.integration test.integration.rpcclient

all: build

docker.start: docker.start.bitcoin-core ## Start docker containers

docker.start.bitcoin-core: ## Start bitcoin-core docker container
	@docker run --rm -d \
		-p 18443:18443 \
		-p 18444:18444 \
		ruimarinho/bitcoin-core \
		-printtoconsole \
		-regtest=1 \
		-rpcallowip=172.17.0.0/16 \
		-rpcbind=0.0.0.0 \
		-rpcuser=test \
		-rpcpassword=test

docker.stop: ## Stop docker containers
	@docker stop $$(docker ps -q)

lint: ## Lint the files
	@golangci-lint run

report: ## Run goreportcard
	@goreportcard-cli -v

test: test.unit test.integration ## Run all tests

test.unit: ## Run unit tests
	@go test -timeout 120s -short ${PKG_LIST}

test.integration: test.integration.rpcclient ## Run integration tests

test.integration.rpcclient: docker.start.bitcoin-core ## Run rpcclient integration tests
	@go test -run Integration -run RPCClient ${PKG_LIST}
	@make docker.stop

race: dep ## Run data race detector
	@go test -race -short ${PKG_LIST}

msan: dep ## Run memory sanitizer
	CC=clang CXX=clang++ CGO_ENABLED=1 go test -msan -short ${PKG_LIST}

coverage: ## Generate global code coverage report
	@go test -coverprofile=coverage.cov ${PKG_LIST}
	@go tool cover -func coverage.cov
	@rm "coverage.cov";

coverhtml: ## Generate global code coverage report in HTML
	@go test -coverprofile=coverage.cov ${PKG_LIST}
	@go tool cover -func coverage.cov
	@go tool cover -html=coverage.cov -o coverage.html
	@rm "coverage.cov";

dep: ## Get the dependencies
	@go get -v -d ./...
	@go get -u golang.org/x/tools/cmd/cover

build: dep ## Build the binary file
	@go build -o etl_bitcoin.exe -v $(PKG)/cmd

clean: ## Remove previous build
	@go clean -cache
	@rm -f $(PROJECT_NAME)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
