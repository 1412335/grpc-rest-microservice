# gen stubs
.PHONY: gen
gen:
	@echo "====gen stubs===="
	sh ./script/gen-gw-swagger.sh

# gofmt
.PHONY: fmt
fmt:
	go fmt -mod=mod $(go list ./... | grep -v /pkg/api/)

# go-lint
.PHONY: lint
lint: fmt
	golangci-lint run $(go list ./... | grep -v /vendor/)
