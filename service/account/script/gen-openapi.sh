#!/bin/bash

OUT_DIR=./api
OPENAPI_DIR=$OUT_DIR/third_party/OpenAPI
NAMESPACE=account

mkdir -m 777 -p $OPENAPI_DIR

echo "===> gen openapi ui with statik"
# rm -rf $OUT_DIR/statik
statik -m -f -ns $NAMESPACE -src $OPENAPI_DIR/ --dest $OUT_DIR