# Acccount Service

## Running

```sh
make run
```

## Note

```sh
# update private package
export GOPRIVATE=github.com/1412335/grpc-rest-microservice
git config --global url.git@github.com:.insteadOf https://github.com/
go get -u github.com/1412335/grpc-rest-microservice
```

## Ref

- <https://www.smartystreets.com/blog/2018/09/private-dependencies-in-docker-and-go/>