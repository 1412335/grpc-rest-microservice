# Acccount Service

## Running

```sh
make run
```

## Note

### Update private package

```sh
export GOPRIVATE=github.com/1412335/grpc-rest-microservice
git config --global url.git@github.com:.insteadOf https://github.com/
go get -u github.com/1412335/grpc-rest-microservice
```

### Config variables priority

flags > environment variables > configuration files > flag defaults

## Ref

- <https://www.smartystreets.com/blog/2018/09/private-dependencies-in-docker-and-go/>