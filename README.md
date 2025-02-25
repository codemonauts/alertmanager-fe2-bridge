# alertmanager-fe2-bridge

A little gateway which receives alerts from the Prometheus Alertmanager via
the [webhook
receiver](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config)
and sends them to the HTTP interface of your [Alamos
FE2](https://www.alamos-gmbh.com/service/fe2/) installation.

## Installation
Install manually from source:
```
go get github.com/codemonauts/alertmanager-fe2-bridge
```
or get the latest compiled binary from the 
[Releases](https://github.com/codemonauts/alertmanager-fe2-bridge/releases) page on GitHub.

## Alertmanager configuration
```yaml
receivers:
  - name: "alamos"
    webhook_configs:
      - url: "http://<hostname>:<port>/input"
```

## Alamos FE2 configuration
