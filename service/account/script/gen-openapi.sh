#!/bin/bash

OUT_DIR=./api
OPENAPI_DIR=$OUT_DIR/third_party/OpenAPI
SERVICE_NAME=account

export $(grep -v '^#' .env | xargs)

echo $SERVICE_NAME

mkdir -m 777 -p $OPENAPI_DIR

echo "===> gen openapi ui with statik"
# rm -rf $OUT_DIR/statik
statik -m -f -ns $SERVICE_NAME -src $OPENAPI_DIR/ --dest $OUT_DIR