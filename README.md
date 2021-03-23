# grpc-rest-microservice

# Install

```sh
make install

# protoc
https://github.com/protocolbuffers/protobuf/releases

# protoc go
go get -u github.com/golang/protobuf/protoc-gen-go

# grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

# grpc-gateway with swagger
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

# grpc-web
https://github.com/grpc/grpc-web/releases

# install deps into vendor dir
go mod vendor
```

# Gen proto

## Using `protoc`
```sh
make gen
```

## Using `namely/gen-grpc-gateway`

### On Linux/Mac
```sh
make gen-gateway-unix
```

### On Windows (with powershell)
```sh
cd ./api/proto/v2

docker run --rm --name protoc-gen -v ${pwd}:/defs namely/gen-grpc-gateway \
    -f . -s ServiceA \
    -o ..\..\..\pkg\api\v2\gen\grpc-gateway
```

# Gen OpenAPI with statik
```
make gen-openapi
```

# Running

## All
```
make grpc
```

## Internal grpc
```sh
# start grpc service
docker-compose up -d --build v2

# run grpc client
docker-compose up -d --build v2-client
```

## grpc-gateway
```sh
# simple testing locally (only unary request)
make v2curl
```

## grpc-web
```sh
# grpc service + envoy proxy
make grpc-web

# grpc js client
make grpc-web-client

# testing
curl -X GET localhost:8081
```

## Cli with evans
```sh
# v1
make cli
# v2
make v2cli
```

## Jaeger UI
- Jaeger UI [http://127.0.0.1:16686](http://127.0.0.1:16686)
- Grafana [http://127.0.0.1:3000](http://127.0.0.1:3000)

## OpenAPI SwaggerUI
- v2 [http://127.0.0.1:8001/openapi-ui](http://127.0.0.1:8001/openapi-ui)

# Note
- Copy & paste /include/google into protobuf folder (eg: ./api/proto/v2)

# Docs

## Overall
- https://developers.google.com/protocol-buffers/docs/gotutorial
- https://github.com/grpc/grpc-go
- https://github.com/grpc/grpc-web
- https://github.com/namely/docker-protoc

## Official examples
- https://github.com/grpc/grpc-go/blob/master/examples/route_guide
- https://github.com/grpc-ecosystem/grpc-gateway/tree/master/examples/internal/proto/examplepb
- https://github.com/grpc/grpc-web/tree/master/net/grpc/gateway/examples/helloworld
- https://github.com/grpc/grpc-web/blob/master/net/grpc/gateway/examples/echo/tutorial.md

## External examples
- https://github.com/thinhdanggroup/benchmark-grpc-web-gateway/
- https://medium.com/zalopay-engineering/buildingdocker-grpc-gateway-e2efbdcfe5c
- https://zalopay-oss.github.io/go-advanced/ch3-rpc/ch3-05-grpc-ext.html
- https://mycodesmells.com/post/pooling-grpc-connections
- https://blog.gopheracademy.com/advent-2017/go-grpc-beyond-basics
- https://medium.com/@shijuvar/writing-grpc-interceptors-in-go-bf3e7671fe48
- https://dev.to/techschoolguru/use-grpc-interceptor-for-authorization-with-jwt-1c5h
- https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md#enable-server-reflection
- https://dev.to/techschoolguru/grpc-reflection-and-evans-cli-3oia
- https://github.com/golangci/golangci-lint-action

## Others
- https://webpack.js.org/configuration/dev-server/
- https://github.com/Colt/webpack-demo-app/

- [grpc-gateway](https://grpc-ecosystem.github.io/)

- https://docs.github.com/en/github/managing-large-files/removing-files-from-a-repositorys-history