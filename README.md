# grpc-rest-microservice

# Install

```sh
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
```

# grpc-gen-protoc

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

docker run --rm --name protoc-gen -v ${pwd}:/defs namely/gen-grpc-gateway -f . -s ServiceA -o ..\..\..\pkg\api\v2\gen\grpc-gateway
```

# Running

## grpc-gateway
```sh
# with automatically generated server
make grpc-gw-gen

# or manually dev
make grpc-gw-man

# simple testing locally (only unary request)
make proxy-test
# or complex testing with streaming request & response
make grpc-gw-client
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
- https://github.com/grpc/grpc-web/tree/master/net/grpc/gateway/examples/helloworld
- https://github.com/grpc/grpc-web/blob/master/net/grpc/gateway/examples/echo/tutorial.md

## External examples
- https://github.com/thinhdanggroup/benchmark-grpc-web-gateway/
- https://medium.com/zalopay-engineering/buildingdocker-grpc-gateway-e2efbdcfe5c
- https://zalopay-oss.github.io/go-advanced/ch3-rpc/ch3-05-grpc-ext.html
  
## Others
- https://webpack.js.org/configuration/dev-server/