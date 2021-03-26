#!/bin/bash

# 1. Generate server's private key
openssl genrsa -out server.key 2048

# 2. Generate public key
openssl req -new -x509 -key server.key -out server.pem -days 3650