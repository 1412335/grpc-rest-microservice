#!/bin/bash

PROTO_V3=./api/proto/v3
API_V3=./pkg/api/v3/

echo "===> gen v3 => gogo_out + grpc-gateway_out + swagger_out + govalidators_out"
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

# Workaround for https://github.com/grpc-ecosystem/grpc-gateway/issues/229.
# sed -i.bak "s/empty.Empty/types.Empty/g" proto/example.pb.gw.go && rm proto/example.pb.gw.go.bak

# Generate static assets for OpenAPI UI
# rm -rf $API_V3
echo "===> gen v3 => openapi ui with statik"
statik -m -f -src $API_V3/third_party/OpenAPI/ --dest $API_V3