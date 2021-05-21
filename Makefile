# export GO111MODULE=on

# install
.PHONY: install
install:
	# go get -u \
	# 	github.com/golang/protobuf/protoc-gen-go \
	# 	github.com/gogo/protobuf/protoc-gen-gogo \
	# 	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
	# 	github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
	# 	github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
	# 	github.com/mwitkow/go-proto-validators/protoc-gen-govalidators \
	# 	github.com/rakyll/statik
	go get \
		github.com/gogo/protobuf/protoc-gen-gogo \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
		github.com/mwitkow/go-proto-validators/protoc-gen-govalidators \
		github.com/rakyll/statik

# gen cert
.PHONY: gen-cert
gen-cert:
	cd ./cert; sh gen.sh; cd ../

.PHONY: gen-rsa
gen-rsa:
	cd ./cert; sh gen-rsa.sh; cd ../

# gen stubs
.PHONY: gen
gen:
	@echo "====gen stubs===="
	sh ./script/gen-proto.sh

.PHONY: genv3
genv3:
	@echo "====gen stubs v3===="
	sh ./script/gen-proto-v3.sh

.PHONY: gen-demo
gen-demo:
	@echo "====gen demo using namely/protoc-all===="
	cd ./api/proto/v2/ && \
	docker run --rm --name protoc-gen -v `pwd`:/defs namely/protoc-all -f common.proto -l go

.PHONY: gen-gateway-unix
gen-gateway-unix:
	@echo "====gen gateway using namely/gen-grpc-gateway===="
	cd ./api/proto/v2/ && \
	docker run --rm --name protoc-gen -v `pwd`:/defs namely/gen-grpc-gateway -f . -s ServiceA -o ..\..\..\pkg\api\v2\gen\grpc-gateway
# docker run --rm --name protoc-gen -v `pwd`:/defs namely/protoc-all -d . -l go --with-gateway

.PHONY: gen-openapi
gen-openapi: ## gen statik openapi
	@echo "====gen openapi===="
	sh ./script/gen-openapi.sh

.PHONY: run
run: clean run-tracing ## Setup env && Run service v3
	@echo "====Running postgres===="
	docker-compose up -d postgres
	sleep 10s
	@echo "====Running v3===="
	docker-compose up -d --build v3
	docker-compose logs -f v3

.PHONY: build
build: ## Build service v3
	@echo "====Build v3===="
	docker-compose up --build v3

.PHONY: run-tracing
run-tracing: ## Running tracing containers
	# docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
	# docker plugin ls
	docker-compose up -d grafana prometheus loki jaeger

.PHONY: cli
cli: ## Evans cli: calling grpc service (reflection.Register(server)) https://github.com/ktr0731/evans
	evans -r repl -p 8080

v2cli:
	evans --header x-request-id=1 -r repl --host localhost -p 8081

v2curl:
	@echo "====Testing proxy====="
	curl -H "x-request-id:1" -X GET localhost:8001/v2/ping/1
	# curl -H "Grpc-Metadata-request-id:1" -X GET localhost:8001/v2/ping/1 # with DefaultHeaderMatcher
	@echo "--- GET ---"
	curl -H "x-request-id:1" localhost:8001/v2/ping/70000
	@echo ""
	curl -H "x-request-id:1" localhost:8001/v2/extra/ping/70000
	@echo ""; echo "--- POST ---"
	curl -H "x-request-id:1" -X POST localhost:8001/v2/post -d '{"timestamp": 7000}'
	@echo ""
	curl -H "x-request-id:1" -X POST localhost:8001/v2/extra/post -d '{"timestamp": 7000}'

# grpc-web with envoy & node client
.PHONY: grpc-web
grpc-web:
	@echo "===grpc-web with envoy & node client===="
	# docker-compose down
	docker-compose up -d --build v2 envoy

.PHONY: grpc-web-client
grpc-web-client:
	docker-compose -f docker-compose.client.yml up --build client-web

.PHONY: fmt
fmt: ## gofmt
	go fmt -mod=mod $(go list ./... | grep -v /pkg/api/)

.PHONY: lint
lint: fmt ## gofmt & golangci-lint
	golangci-lint run $(go list ./... | grep -v /vendor/)

.PHONY: test
test: lint ## gofmt & golangci-lint & go test
	go test -v -short -race -coverprofile=coverage.out -covermode=atomic $(go list ./... | grep -v /vendor/)

.PHONY: cover
cover: test  ## Run unit tests and open the coverage report
	go tool cover -html=coverage.out

.PHONY: clean
clean: ## stop containers & clean go test result
	@echo "====cleaning env==="
	docker-compose down -v --remove-orphans
	rm -rf ./docker/mysql/data
	# docker system prune -af --volumes
	# docker rm $(docker ps -aq -f "status=exited")
	go clean -testcache
	rm -f coverage.out

.PHONY: help
help:  ## Print usage information
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.DEFAULT_GOAL := help