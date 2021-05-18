#!/bin/bash
rm *.pem

# Inspired from: https://dev.to/techschoolguru/how-to-create-sign-ssl-tls-certificates-2aai

# Output files
# ca-key: Certificate Authority private key file (this shouldn't be shared in real-life)
# ca-cert: Certificate Authority trust certificate (this should be shared with users/clients in real-life)

### for server
# server-key: Server private key, password protected (this shouldn't be shared)
# server-req: Server certificate signing request (this should be shared with the CA owner)
# server-cert: Server certificate signed by the CA (this would be sent back by the CA owner) - keep on server
# server-ext.cnf: custom certificate dns

# Summary 
# Private files: ca-key, server-key, server-cert
# "Share" files: ca-cert (needed by the client), server-req (needed by the CA)

# 1. Generate CA's private key and self-signed certificate
openssl req -x509 -newkey rsa:4096 -days 365 -nodes -keyout ca-key.pem -out ca-cert.pem -subj "/C=FR/ST=Occitanie/L=Toulouse/O=Tech School/OU=Education/CN=*.techschool.guru/emailAddress=techschool.guru@gmail.com"

echo "CA's self-signed certificate"
openssl x509 -in ca-cert.pem -noout -text

# 2. Generate weh server's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout server-key.pem -out server-req.pem -subj "/C=FR/ST=Ile de France/L=Paris/O=PC Book/OU=Computer/CN=*.pcbook.com/emailAddress=pcbook@gmail.com"

# 3. Use CA's private key to sign web server's CSR and get back the signed certificate
openssl x509 -req -in server-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile server-ext.cnf

echo "Server's signed certificate"
openssl x509 -in server-cert.pem -noout -text

## for client
# client-key: Client private key, password protected (this shouldn't be shared)
# client-req: Client certificate signing request (this should be shared with the CA owner)
# client-cert: Client certificate signed by the CA (this would be sent back by the CA owner) - keep on client
# client-ext.cnf: custom certificate dns

# 4. Generate client's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout client-key.pem -out client-req.pem -subj "/C=FR/ST=Alsace/L=Strasbourg/O=PC Client/OU=Computer/CN=*.pcclient.com/emailAddress=pcclient@gmail.com"

# 5. Use CA's private key to sign client's CSR and get back the signed certificate
openssl x509 -req -in client-req.pem -days 60 -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -out client-cert.pem -extfile client-ext.cnf

echo "Client's signed certificate"
openssl x509 -in client-cert.pem -noout -text
