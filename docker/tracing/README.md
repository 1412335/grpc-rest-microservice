# Grafana Integration + Loki + Prometheus

## Loki
- Collect docker logs
- [Logql](https://grafana.com/docs/loki/latest/logql/)

## Grafana
- Dashboard: [example](./grafana/dashboards/hotrod_metrics_logs.json)
- Setup datasources.yaml: connect to container

## Prometheus
- Metrics

## Install

1. Adding Loki as a Logging Driver
```sh
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions

# restart the docker daemon
sudo systemctl restart docker

docker plugin ls
```

2. Using Loki
```sh
# add to docker-compose
logging:
  driver: loki
  options:
    loki-url: "http://host.docker.internal:3100/loki/api/v1/push"


# docker-compose v3.4
# logger driver - change this driver to ship all container logs to a different location
x-logging: &logging
  logging:
    driver: loki
    options:
      loki-url: "http://host.docker.internal:3100/loki/api/v1/push"
services:
  my_service:
    *logging
    container_name: xxx
    image: xxx/xxx
  another_cool_service:
    *logging
    container_name: xxxx
    image: xxxx/xxxx
```

## Ref
- [Collecting-Docker-Logs-With-Loki](https://yuriktech.com/2020/03/21/Collecting-Docker-Logs-With-Loki/)
- [Hot R.O.D. - Rides on Demand - Grafana integration](https://github.com/jaegertracing/jaeger/tree/50fb11f2a90553ef10b9c3a9049270de385a8c17/examples/grafana-integration)