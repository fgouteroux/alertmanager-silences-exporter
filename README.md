# alertmanager-silences-exporter
Prometheus exporter exposing [AlertManager Silences](https://prometheus.io/docs/alerting/alertmanager/#silences) as metrics.

## Getting Started

### Prerequisites

To run this project, you will need a [working Go environment](https://golang.org/doc/install).

### Installing

```bash
go get -u github.com/fgouteroux/alertmanager-silences-exporter
```

## Building

Build the sources with

```bash
make build
```

## Run the binary

The exporter expects a config file as one of its arguments:

```bash
./alertmanager-silences-exporter --config.file=sample-config.yml
```

The exporter's Alertmanager API connection can also be configured by defining the following environment variable(s). If they are present, they will take precedence over the corresponding variables in the config file.

Environment Variable | Description
---------------------| -----------
ALERTMANAGER_URL | URL of the exported alertmanager api (eg: "http://localhost:9093/")
ALERTMANAGER_USERNAME | alertmanager api username
ALERTMANAGER_PASSWORD | alertmanager api password


Use -h flag to list available options.

## Testing

### Running unit tests

```bash
make test
```

## Configuration

An example can be found in
[sample-config.yml](https://github.com/fgouteroux/alertmanager-silences-exporter/blob/master/sample-config.yml).

Configuration element | Description
--------------------- | -----------
alertmanager_url | (Mandatory) URL of the exported alertmanager api (eg: "http://localhost:9093/") (string)
alertmanager_username | (Optional) alertmanager api username (string)
alertmanager_password | (Optional) alertmanager api password (string)
expired_silences | (Optional) Collect expired silences metrics (boolean)
tenants          | (Optional) Collect multi-tenant silences metrics (list of string)

## Exposed metrics

Metric | Description
------ | -----------
alertmanager_silence_info | [AlertManager Silences](https://prometheus.io/docs/alerting/alertmanager/#silences) exposed as gauge values. Returns 1 if a silence is active, and 0 for all other statuses.
alertmanager_silence_start_seconds | The start time of an Alertmanager Silence, exposed in unix/epoch time.
alertmanager_silence_end_seconds | The end time of an Alertmanager Silence, exposed in unix/epoch time.

Example:

```
# HELP alertmanager_silence_end_seconds Alertmanager silence end time, elapsed seconds since epoch
# TYPE alertmanager_silence_end_seconds gauge
alertmanager_silence_end_seconds{id="7aa4fb96-9aac-4a3a-899c-ee5f20afd730"} 1.5829092e+09
# HELP alertmanager_silence_info Alertmanager silence info metric
# TYPE alertmanager_silence_info gauge
alertmanager_silence_info{comment="me",createdBy="me",id="7aa4fb96-9aac-4a3a-899c-ee5f20afd730",matcher_customer="foo",status="active"} 1
# HELP alertmanager_silence_start_seconds Alertmanager silence start time, elapsed seconds since epoch
# TYPE alertmanager_silence_start_seconds gauge
alertmanager_silence_start_seconds{id="7aa4fb96-9aac-4a3a-899c-ee5f20afd730"} 1.582571132e+09
```

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/fgouteroux/alertmanager-silences-exporter/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/fgouteroux/alertmanager-silences-exporter/blob/master/LICENSE).
