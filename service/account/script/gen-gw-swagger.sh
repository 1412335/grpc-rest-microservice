#!/bin/bash

# GOOGLE_APIS=$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis

PROTO_DIR=./proto
OUT_DIR=./api
OPENAPI_DIR=$OUT_DIR/third_party/OpenAPI
NAMESPACE=account

mkdir -m 777 -p $OPENAPI_DIR

echo "===> gen grpc-gateway + swagger"
protoc -I $GOPATH/src \
    -I ../../vendor/github.com/grpc-ecosystem/grpc-gateway/ \
    -I ../../vendor/ \
    -I $PROTO_DIR/ \
    --go_out=plugins=grpc:$OUT_DIR/ \
    --grpc-gateway_out=$OUT_DIR/ \
    --swagger_out=$OPENAPI_DIR/ \
    $PROTO_DIR/*.proto

# echo "===> gen openapi ui with statik"
# # rm -rf $OUT_DIR/statik
# statik -m -f -ns $NAMESPACE -src $OPENAPI_DIR/ --dest $OUT_DIR