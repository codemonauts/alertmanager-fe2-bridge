# alertmanager-fe2-bridge

A little gateway which receives alerts from the Prometheus Alertmanager via the (webhook
receiver)[https://prometheus.io/docs/alerting/latest/configuration/#webhook_config] and sends them to the HTTP
interface of your (Alamos FE2)[https://www.alamos-gmbh.com/service/fe2/] installation.

## Installation

## Alertmanager configuration
```
receivers:
  - name: 'alamos'
    webhook_configs:
      - url: 'http://<hostname>/input'
```

## Alamos FE2 configuration
