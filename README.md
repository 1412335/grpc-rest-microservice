# grpc-rest-microservice

# protoc
go get -u github.com/golang/protobuf/protoc-gen-go

# grpc-gateway
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway

# grpc-gateway with swagger
go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger

# grpc-gen-protoc

## protoc direct
```sh
make gen
```

## Using namely/gen-grpc-gateway

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
```sh
# with automatically generated server
make gateway-gen

# or manually dev
make gateway-man

# testing locally
make proxy-test
```

# NOTE
- Copy & paste /include/google into protobuf folder (eg: ./api/proto/v2)