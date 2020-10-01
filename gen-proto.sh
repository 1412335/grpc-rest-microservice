#!/bin/bash

# echo "gen v1"
# protoc -I ./api/proto/v1/ \
#     --go_out=plugins=grpc:./pkg/api/v1 \
#     ./api/proto/v1/*.proto 


# echo "gen v2 grpc-gateway => grpc-gateway_out"
# protoc -I $GOPATH/src \
#     -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#     -I ./api/proto/v2/ \
#     --grpc-gateway_out=./pkg/api/v2 \
#     --go_out=plugins=grpc:./pkg/api/v2 \
#     ./api/proto/v2/*.proto


# echo "gen v2 grpc-gateway + swagger => swagger_out"
# protoc -I $GOPATH/src \
#     -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#     -I ./api/proto/v2/ \
#     --grpc-gateway_out=./pkg/api/v2 \
#     --swagger_out=./pkg/api/v2 \
#     --go_out=plugins=grpc:./pkg/api/v2 \
#     ./api/proto/v2/*.proto


echo "gen v2 grpc-web (grpcwebtext)"
protoc -I $GOPATH/src \
    -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -I ./api/proto/v2/ \
    --js_out=import_style=commonjs:./pkg/api/v2/gen/grpc-web/gen \
    --grpc-web_out=import_style=commonjs,mode=grpcwebtext:./pkg/api/v2/gen/grpc-web/gen \
    ./api/proto/v2/*.proto

echo "gen v2 grpc-web (grpcweb with binary protobuf supported)"
protoc -I $GOPATH/src \
    -I $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -I ./api/proto/v2/ \
    --js_out=import_style=commonjs:./pkg/api/v2/gen/grpc-web/gen \
    --grpc-web_out=import_style=commonjs,mode=grpcweb:./pkg/api/v2/gen/grpc-web/gen \
    ./api/proto/v2/*.proto