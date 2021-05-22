.PHONY: gen-openapi
gen-openapi: ## statik gen openapi
	@echo "====gen openapi===="
	sh ./script/gen-openapi.sh

.PHONY: gen-api
gen-api: ## gen grpc + gateway
	@echo "====gen stubs===="
	sh ./script/gen-gw-swagger.sh

.PHONY: gen
gen: gen-api gen-openapi ## gen grpc + gateway + openui w swagger

.PHONY: fmt
fmt: ## go fmt
	go fmt -mod=mod $(go list ./... | grep -v /api)

.PHONY: lint
lint: fmt ## gofmt && golangci-lint
	golangci-lint run $(go list ./... | grep -v /api)

.PHONY: test
test: lint ## gofmt & golangci-lint & go test
	go test -v -short -race -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: cover
cover: test  ## Run unit tests and open the coverage report
	go tool cover -html=coverage.out -o coverage.html

# export SSH_PRIVATE_KEY=$(cat ~/.ssh/id_ed25519)
run: clean cover ## clean && lint && test && build & run service
	docker-compose up -d --build account
	docker-compose logs -f account

.PHONY: cli
cli: ## Evans cli: calling grpc service (reflection.Register(server)) https://github.com/ktr0731/evans
	evans -r repl -p 8083

.PHONY: clean
clean: ## cleaning
	@echo "====cleaning env==="
	docker-compose down -v --remove-orphans
	go clean -testcache
	rm -f coverage.*

.PHONY: help
help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort