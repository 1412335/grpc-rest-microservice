#!/bin/bash

PROTO_DIR=./api/proto
PROTO_V3=./api/proto/v3

OUT_DIR=./pkg/api
API_V3=./pkg/api/v3/

# ls $API_V2/grpc-web/gen/*_grpc_web_pb.js

echo "GEN GRPC-GATEWAY"

echo "===> gen v3 grpc-gateway => grpc-gateway_out"
protoc -I $GOPATH/src \
    -I ./vendor/github.com/grpc-ecosystem/grpc-gateway/ \
    -I ./vendor/github.com/gogo/googleapis/ \
    -I ./vendor/ \
    -I $PROTO_V3/ \
    --gogo_out=plugins=grpc,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,\
Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types:\
$API_V3 \
    --grpc-gateway_out=allow_patch_feature=false,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,\
Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types:\
$API_V3 \
    --swagger_out=$API_V3/third_party/OpenAPI/ \
    --govalidators_out=gogoimport=true,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types,\
Mgoogle/api/annotations.proto=github.com/gogo/googleapis/google/api,\
Mgoogle/protobuf/field_mask.proto=github.com/gogo/protobuf/types:\
$API_V3 \
    $PROTO_V3/*.proto

# Generate static assets for OpenAPI UI
# rm -rf statik
statik -m -f -src $API_V3/third_party/OpenAPI/ --dest $API_V3