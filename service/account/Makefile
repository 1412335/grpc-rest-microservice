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

update-pkg-cache:
    GOPROXY=https://proxy.golang.org GO111MODULE=on \
    go get github.com/$(USER)/$(PACKAGE)@v$(VERSION)