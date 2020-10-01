############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add git
# Create appuser.
RUN adduser -D -g '' appuser
# WORKDIR $GOPATH/src/mypackage/myapp/
WORKDIR /myapp
ENV GO111MODULE=on
COPY go.mod go.sum ./

RUN go mod download

COPY . /myapp

RUN ls $GOPATH/src

# Build the binary.
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-w -s" -o /go/bin/myapp cmd/proxy/main.go 
############################
# STEP 2 build a small image
############################
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

# Import the user and group files from the builder.
# COPY --from=builder /etc/passwd /etc/passwd

WORKDIR /root/

# Copy our static executable.
COPY --from=builder /go/bin/myapp ./

# Use an unprivileged user.
# USER appuser

ARG GRPC_HOST=":9090"
ENV GRPC_HOST=${GRPC_HOST}

ARG PROXY_PORT=8080
ENV PROXY_PORT=${PROXY_PORT}

# Run the hello binary. 
EXPOSE $PROXY_PORT

# ENTRYPOINT [ "./myapp" ] 
# CMD -proxy-port=$PROXY_PORT -grpc-host=$GRPC_HOST
CMD ["sh", "-c", "/root/myapp -proxy-port=$PROXY_PORT -grpc-host=$GRPC_HOST"]