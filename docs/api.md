# API

## Config Management API

Grafana Cloud Agent exposes a REST API for managing instance configurations when
it is running in scraping service mode. The following endpoints are exposed:

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
responses will provide a value of `error`.

The data field may or may not be present, depending on the endpoint. It
provides extra information for the query. The documentation for each endpoint
will describe the full response provided.

### List Configs

```
GET /agent/api/v1/configs
```

List Configs returns a list of the named configurations currently known by the
underlying KV store.

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

### Get Config:

```
GET /agent/api/v1/configs/{name}
```

Get Config will return a single configuration by name. The configuration must
exist or an error will be returned.

Response on success:

```
{
  "status": "success",
  "data": {
    "value": "/* configuration in the string */"
  }
}
```

By default, the configuration contained in "value" will be YAML, even if it was
initially stored as JSON. To get back a JSON config string, add a
`Content-Type: application/json` header to the request.

Get config

### Update Config

```
PUT /agent/api/v1/config/{name}
POST /agent/api/v1/config/{name}
```

Update Config will update or add a new configuration by name. If a configuration
with the same name already exists, it will be completely overwritten.

The request body passed to this endpoint must match the format of
[prometheus_instance_config](./configuration-reference.md#prometheus_instance_config)
defined in the Configuration Reference. The name field of the configuration is
ignored and the name in the URL takes precedence.

By default, the request body is expected to be JSON. To pass a YAML config, add
a `Content-Type: text/yaml` header to the request.

Response on success:

```
{
  "status": "success"
}
```

### Delete Config:

```
DELETE /agent/api/v1/config/{name}
```

Delete Config will attempt to delete a configuration by name. The named
configuration must exist; deleting a nonexistent config will result in an
error.

Response on success:

```
{
  "status": "success"
}
```
