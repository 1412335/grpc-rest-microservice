#!/bin/bash

GOOGLE_APIS=$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis

PROTO_DIR=./api/proto
PROTO_V1=./api/proto/v1
PROTO_V2=./api/proto/v2

OUT_DIR=./pkg/api
API_V1=./pkg/api/v1
API_V2=./pkg/api/v2

# ls $API_V2/grpc-web/gen/*_grpc_web_pb.js

echo "GEN STUBS"
echo "===> gen v1"
protoc -I $PROTO_V1/ \
    --go_out=plugins=grpc:$API_V1 \
    $PROTO_V1/*.proto 

echo "==================================="

echo "GEN GRPC-GATEWAY"

echo "===> gen v2 grpc-gateway => grpc-gateway_out"
protoc -I $GOPATH/src \
    -I $GOOGLE_APIS \
    -I $PROTO_V2/ \
    --grpc-gateway_out=$API_V2/grpc-gateway/gen \
    --go_out=plugins=grpc:$API_V2/grpc-gateway/gen \
    $PROTO_V2/*.proto


echo "===> gen v2 grpc-gateway + swagger => swagger_out"
protoc -I $GOPATH/src \
    -I $GOOGLE_APIS \
    -I $PROTO_V2/ \
    --grpc-gateway_out=$API_V2/grpc-gateway/gen \
    --swagger_out=$API_V2/grpc-gateway/third_party/OpenAPI/ \
    --go_out=plugins=grpc:$API_V2/grpc-gateway/gen \
    $PROTO_V2/*.proto

echo "==================================="

echo "GEN GRPC-WEB"

echo "remove dir '$API_V2/grpc-web/gen'"
rm -rf $API_V2/grpc-web/gen/ && mkdir -p $API_V2/grpc-web/gen/

echo "==> gen v2 grpc-web (grpcweb with binary protobuf supported)"
echo "NOTE: add prefix _binary"
protoc -I $GOPATH/src \
    -I $GOOGLE_APIS \
    -I $PROTO_V2/ \
    --js_out=import_style=commonjs,binary:$API_V2/grpc-web/gen \
    --grpc-web_out=import_style=commonjs,mode=grpcweb:$API_V2/grpc-web/gen \
    $PROTO_V2/*.proto && \
for filename in $(ls $API_V2/grpc-web/gen/*_grpc_web_pb.js); do mv $filename ${filename%.*}_binary.js; done;


echo "===> gen v2 grpc-web (grpcwebtext)"
protoc -I $GOPATH/src \
    -I $GOOGLE_APIS \
    -I $PROTO_V2/ \
    --js_out=import_style=commonjs:$API_V2/grpc-web/gen \
    --grpc-web_out=import_style=commonjs,mode=grpcwebtext:$API_V2/grpc-web/gen \
    $PROTO_V2/*.proto