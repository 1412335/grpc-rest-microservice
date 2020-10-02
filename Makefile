# export GO111MODULE=on

# gen stubs
gen:
	@echo "====gen stubs===="
	sh gen-proto.sh

gen-demo:
	@echo "====gen demo using namely/protoc-all===="
	cd ./api/proto/v2/ && \
	docker run --rm --name protoc-gen -v `pwd`:/defs namely/protoc-all -f common.proto -l go

gen-gateway-unix:
	@echo "====gen gateway using namely/gen-grpc-gateway===="
	cd ./api/proto/v2/ && \
	docker run --rm --name protoc-gen -v `pwd`:/defs namely/gen-grpc-gateway -f . -s ServiceA -o ..\..\..\pkg\api\v2\gen\grpc-gateway
# docker run --rm --name protoc-gen -v `pwd`:/defs namely/protoc-all -d . -l go --with-gateway


# run locally grpc server & client
grpc-server:
	@echo "====Run grpc server service===="
	go run ./cmd/server/main.go -grpc-port=9090 -db-host=:3306 -db-user=user -db-password=pwd -db-scheme=

grpc-client:
	@echo "====Run grpc client===="
	go run ./cmd/client-grpc/main.go -server=localhost:9090


# run grpc gateway locally
proxy-server:
	@echo "====Run server===="
	go run ./cmd/proxy/grpc.go -grpc-port=9090

proxy-client:
	@echo "====Run rest client===="
	go run ./cmd/proxy/main.go -grpc-port=9090 -proxy-port=8000

# testing
proxy-test:
	@echo "====Testing proxy====="
	@echo "--- GET ---"
	curl localhost:8000/v2/ping/70000
	@echo ""
	curl localhost:8000/v2/extra/ping/70000
	@echo ""; echo "--- POST ---"
	curl localhost:8000/v2/post -X POST -d '{"timestamp": 7000}'
	@echo ""
	curl localhost:8000/v2/extra/post -X POST -d '{"timestamp": 7000}'


# run grpc gateway using docker
service-build-run:
	@echo "===build grpc service image==="
	docker build --build-arg GRPC_PORT=9090 -t grpc-gateway:client -f Dockerfile.service .
	@echo ""
	@echo "===run grpc service container==="
	docker run --rm --name grpc-gw-service -p 9090:9090 grpc-gateway:client

gateway-build-run:
	@echo "===build grpc gateway image==="
	docker build --build-arg GRPC_HOST=localhost:9090 --build-arg PROXY_PORT=8080 -t grpc-gateway:gw -f Dockerfile .
	@echo ""
	@echo "===run grpc gateway container==="
	docker run --rm --name grpc-gw -p 8080:8080 grpc-gateway:gw


# run grpc gateway using docker-compose with server initialized manually
grpc-gw-man:
	@echo "===run grpc gateway with manually writted server===="
	# docker-compose down
	docker-compose up --build client-service grpc-gateway


# run grpc gateway using docker-compose with server auto generated
grpc-gw-gen:
	@echo "===run grpc gateway with generated server===="
	# docker-compose down
	docker-compose up --build client-service grpc-gateway-gen
	# docker-compose -f docker-compose.gen.yml up --build

grpc-gw-client:
	docker-compose -f docker-compose.client.yml up --build client-gateway

# grpc-web with envoy & node client
grpc-web:
	@echo "===grpc-web with envoy & node client===="
	# docker-compose down
	docker-compose up --build client-service envoy

grpc-web-client:
	docker-compose -f docker-compose.client.yml up --build client-web

# cleaning
clean:
	@echo "====cleaning env==="
	docker-compose down -v
	# docker system prune -af --volumes
	# docker rm $(docker ps -aq -f "status=exited")