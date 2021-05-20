#!/bin/bash

API_V2=./pkg/api/v2
API_V3=./pkg/api/v3

# Generate static assets for OpenAPI UI

# rm -rf statik
statik -m -f -ns v2 -src $API_V2/grpc-gateway/third_party/OpenAPI/ --dest $API_V2/grpc-gateway

statik -m -f -ns v3 -src $API_V3/third_party/OpenAPI/ --dest $API_V3