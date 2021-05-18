.PHONY: gen
gen: # gen grpc + gateway + openui w swagger
	@echo "====gen stubs===="
	sh ./script/gen-gw-swagger.sh

.PHONY: fmt
fmt: ## go fmt
	go fmt -mod=mod $(go list ./... | grep -v /api/)

.PHONY: lint
lint: fmt ## gofmt && golangci-lint
	golangci-lint run $(go list ./... | grep -v /vendor/)

update-pkg-cache:
    GOPROXY=https://proxy.golang.org GO111MODULE=on \
    go get github.com/$(USER)/$(PACKAGE)@v$(VERSION)

# export SSH_PRIVATE_KEY=$(cat ~/.ssh/id_ed25519)
run: clean lint ## clean && lint && build & run service
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
	rm -f coverage.out

.PHONY: help
help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort