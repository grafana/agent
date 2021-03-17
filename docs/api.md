# API

The API is divided into several parts:

- [Config Management API](#config-management-api)
- [Agent API](#agent-api)
- [Ready/Healthy API](#ready--health-api)

## Config Management API

Grafana Agent exposes a REST API for managing instance configurations when
it is running in [scraping service mode](./scraping-service.md). The following
endpoints are exposed:

- List configs: [`GET /agent/api/v1/configs`](#list-configs)
- Get config: [`GET /agent/api/v1/configs/{name}`](#get-config)
- Update config: [`PUT /agent/api/v1/config/{name}`](#update-config)
- Delete config: [`DELETE /agent/api/v1/config/{name}`](#delete-config)

### API Response

All Config Management API endpoints will return responses in the following
form, unless an internal service error prevents the server from responding
properly:

```json
{
  "status": "success" | "error",
  "data": {}
}
```

Status will be either `success` or `error`. All 2xx responses will be
accompanied with a `success` value for the status field. 4xx and 5xx
responses will provide a value of `error`. All requests may potentially
return 500 on an internal error. Other non-500 responses will be documented
per API.

The data field may or may not be present, depending on the endpoint. It
provides extra information for the query. The documentation for each endpoint
will describe the full response provided.

### List Configs

```
GET /agent/api/v1/configs
```

List Configs returns a list of the named configurations currently known by the
underlying KV store.

Status code: 200 on success.
Response:

```
{
  "status": "success",
  "data": {
    "configs": [
      // list of config names:
      "a",
      "b",
      "c",
      // ...
    ]
  }
}
```

### Get Config

```
GET /agent/api/v1/configs/{name}
```

Get Config will return a single configuration by name. The configuration must
exist or an error will be returned. URL-encoded names will be retrieved in decoded
form. e.g., `hello%2Fworld` will represent the config named `hello/world`.

Status code: 200 on success, 400 on invalid config name.
Response on success:

```
{
  "status": "success",
  "data": {
    "value": "/* YAML configuration */"
  }
}
```

### Update Config

```
PUT /agent/api/v1/config/{name}
POST /agent/api/v1/config/{name}
```

Update Config will update or add a new configuration by name. If a configuration
with the same name already exists, it will be completely overwritten.

URL-encoded names will be stored in decoded form. e.g., `hello%2Fworld` will
represent the config named `hello/world`.

The request body passed to this endpoint must match the format of
[prometheus_instance_config](./configuration-reference.md#prometheus_instance_config)
defined in the Configuration Reference. The name field of the configuration is
ignored and the name in the URL takes precedence. The request body must be
formatted as YAML.

Status code: 201 with a new config, 200 on updated config.
Response on success:

```
{
  "status": "success"
}
```

### Delete Config

```
DELETE /agent/api/v1/config/{name}
```

Delete Config will attempt to delete a configuration by name. The named
configuration must exist; deleting a nonexistent config will result in an
error.

URL-encoded names will be interpreted in decoded form. e.g., `hello%2Fworld`
will represent the config named `hello/world`.

Status code: 200 on success, 400 with invalid config name.
Response on success:

```
{
  "status": "success"
}
```

## Agent API

### List current running instances

```
GET /agent/api/v1/instances
```

Status code: 200 on success.
Response on success:

```
{
  "status": "success",
  "data": [
    <strings of instance names that are currently running>
  ]
}
```

### List current scrape targets

```
GET /agent/api/v1/targets
```

This endpoint collects all targets known to the Agent across all running
instances. Only targets being scraped from the local Agent will be returned. If
running in scraping service mode, this endpoint must be invoked in all Agents
separately to get the combined set of targets across the whole Agent cluster.

The `labels` fields shows the labels that will be added to metrics from the
target, while the `discovered_labels` field shows all labels found during
service discovery.

Status code: 200 on success.
Response on success:

```
{
  "status": "success",
  "data": [
    {
      "instance": <string, instance config name>,
      "target_group": <string, scrape config group name>,
      "endpoint": <string, URL being scraped>
      "state": <string, one of up, down, unknown>,
      "discovered_labels": {
        "__address__": "<address>",
        ...
      },
      "labels": {
        "label_a": "value_a",
        ...
      },
      "last_scrape": <string, RFC 3339 timestamp of last scrape>,
      "scrape_duration_ms": <number, last scrape duration in milliseconds>,
      "scrape_error": <string, last error. empty if scrape succeeded>
    },
    ...
  ]
}
```

## Ready / Health API

### Readiness Check

```
GET /-/ready
```

Status code: 200 if ready.

Response:
```
Agent is Ready.
```

### Healthiness Check

```
GET /-/healthy
```

Status code: 200 if healthy.

Response:
```
Agent is Healthy.
```
