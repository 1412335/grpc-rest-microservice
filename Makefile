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
.PHONY: cert
cert:
	cd ./cert; ./gen.sh; cd ../


# gen stubs
.PHONY: gen
gen:
	@echo "====gen stubs===="
	sh gen-proto.sh

.PHONY: genv3
genv3:
	@echo "====gen stubs v3===="
	sh gen-proto-v3.sh

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

# run cli
.PHONY: run
run:
	@echo "====Run grpc server v1===="
	go run main.go v1

.PHONY: grpc
grpc:
	@echo "====Run grpc server with docker===="
	# docker-compose up -d mysql
	# sleep 20s
	docker-compose up -d

# Evans cli: calling grpc service (reflection.Register(server))
.PHONY: cli
cli:
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

.PHONY: grpc-server
# run locally grpc server & client
grpc-server:
	@echo "====Run grpc server service===="
	docker-compose up --build client-service
	# go run ./cmd/server/main.go -grpc-port=9090 -db-host=:3306 -db-user=user -db-password=pwd -db-scheme=

.PHONY: grpc-client
grpc-client:
	@echo "====Run grpc client===="
	docker-compose -f docker-compose.client.yml up --build client
	# go run ./cmd/client-grpc/main.go -server=localhost:9090

# run grpc using docker
.PHONY: service-build-run
service-build-run:
	@echo "===build grpc service image==="
	docker build --build-arg GRPC_PORT=9090 -t grpc:service -f Dockerfile.service .
	@echo ""
	@echo "===run grpc service container==="
	docker run --rm --name grpc-service -p 9090:9090 grpc:service

.PHONY: gateway-build-run
gateway-build-run:
	@echo "===build grpc gateway image==="
	docker build --build-arg GRPC_HOST=localhost:9090 --build-arg PROXY_PORT=8080 -t grpc-gateway:gw -f Dockerfile .
	@echo ""
	@echo "===run grpc gateway container==="
	docker run --rm --name grpc-gw -p 8080:8080 grpc-gateway:gw


# run grpc gateway using docker-compose with server initialized manually
.PHONY: grpc-gw-man
grpc-gw-man:
	@echo "===run grpc gateway with manually writted server===="
	# docker-compose down
	docker-compose up --build client-service grpc-gateway


# run grpc gateway using docker-compose with server auto generated
.PHONY: grpc-gw-gen
grpc-gw-gen:
	@echo "===run grpc gateway with generated server===="
	# docker-compose down
	docker-compose up --build client-service grpc-gateway-gen
	# docker-compose -f docker-compose.gen.yml up --build

.PHONY: grpc-gw-client
grpc-gw-client:
	docker-compose -f docker-compose.client.yml up --build client

# grpc-web with envoy & node client
.PHONY: grpc-web
grpc-web:
	@echo "===grpc-web with envoy & node client===="
	# docker-compose down
	docker-compose up --build client-service envoy

.PHONY: grpc-web-client
grpc-web-client:
	docker-compose -f docker-compose.client.yml up --build client-web

# gofmt
.PHONY: fmt
fmt:
	go fmt -mod=mod ./... 

# go-lint
.PHONY: lint
lint: fmt
	golangci-lint run ./...

# cleaning
.PHONY: clean
clean:
	@echo "====cleaning env==="
	docker-compose down -v --remove-orphans
	rm -rf ./docker/mysql/data
	# docker system prune -af --volumes
	# docker rm $(docker ps -aq -f "status=exited")