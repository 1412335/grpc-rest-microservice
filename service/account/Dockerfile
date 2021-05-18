############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add git \
    openssh

# Create appuser.
RUN adduser -D -g '' appuser
# WORKDIR $GOPATH/src/mypackage/myapp/
WORKDIR /myapp

# add credentials on build
ARG SSH_PRIVATE_KEY
RUN mkdir /root/.ssh/ && umask 0077 && echo "${SSH_PRIVATE_KEY}" > /root/.ssh/id_rsa && \
    git config --global url.git@github.com:.insteadOf https://github.com/ && \
    ssh-keyscan -t rsa github.com > ~/.ssh/known_hosts
    # echo "StrictHostKeyChecking no " > /root/.ssh/config

ENV GO111MODULE=on
COPY go.mod go.sum ./

RUN go mod download

COPY . /myapp

# Build the binary.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-w -s" -o /go/bin/myapp main.go

############################
# STEP 2 build a small image
############################
# FROM scratch
FROM alpine:latest  
RUN apk --no-cache add ca-certificates
# Import the user and group files from the builder.
COPY --from=builder /etc/passwd /etc/passwd

WORKDIR /root/
# Copy our static executable.
COPY --from=builder /go/bin/myapp .
COPY --from=builder /myapp/config.yml .

# Use an unprivileged user.
# USER appuser

# gRPC server port
EXPOSE 9090
# proxy gateway
EXPOSE 8000

# CMD exec /bin/sh -c "trap : TERM INT; (while true; do sleep 1000; done) & wait"
ENTRYPOINT ["./myapp"]
CMD [ "grpc-gateway" ]

# ENTRYPOINT: append
# CMD: replace/default command