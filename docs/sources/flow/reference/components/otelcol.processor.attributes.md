---
title: otelcol.processor.attributes
---

# otelcol.processor.attributes

`otelcol.processor.attributes` accepts telemetry data from other `otelcol`
components and modifies attributes of a span, log, or metric.
It also supports the ability to filter and match input data to determine if 
it should be included or excluded for attribute modifications.

> **NOTE**: `otelcol.processor.attributes` is a wrapper over the upstream
> OpenTelemetry Collector `attributes` processor. Bug reports or feature requests
> will be redirected to the upstream repository, if necessary.

Multiple `otelcol.processor.attributes` components can be specified by giving them
different labels.

## Usage

```river
otelcol.processor.attributes "LABEL" {
  output {
    metrics = [...]
    logs    = [...]
    traces  = [...]
  }
}
```

## Arguments

`otelcol.processor.attributes` doesn't support any arguments and is configured fully
through inner blocks.

## Blocks

The following blocks are supported inside the definition of
`otelcol.processor.attributes`:
Hierarchy | Block | Description | Required
--------- | ----- | ----------- | --------
output | [output][] | Configures where to send received telemetry data. | yes
action | [action][] | Actions to take on the attributes of incoming metrics/logs/traces. | no
include | [include/exclude] | Filter for data to be included to this processor's actions. | no
include > regexp | [regexp] | Regex cache settings. | no
include > attribute | [attribute] | A list of attributes to match against. | no
include > resource | [resource] | A list of items to match the resources against. | no
include > library | [library] | A list of items to match the implementation library against. | no
include > log_severity_number | [library] | How to match against a log record's SeverityNumber, if defined. | no
exclude | [include/exclude] | Filter for data to be excluded from this processor's actions | no
exclude > regexp | [regexp] | Regex cache settings. | no
exclude > attribute | [attribute] | A list of attributes to match against. | no
exclude > resource | [resource] | A list of items to match the resources against. | no
exclude > library | [library] | A list of items to match the implementation library against. | no
exclude > log_severity_number | [library] | How to match against a log record's SeverityNumber, if defined. | no

[output]: #output-block
[action]: #action-block
[include/exclude]: #include/exclude-blocks
[regexp]: #regexp-block
[attribute]: #attribute-block
[resource]: #resource-block
[library]: #library-block
[log_severity_number]: #log_severity_number-block

### action block

The `action` block configures how to modify the span, log or metric.

The following attributes are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`key` | `string` | The attribute that the action relates to. |  | yes
`action` | `string` | The type of action to be performed. |  | yes
`value` | `any` | The value to populate for the key. |  | no
`pattern` | `string` | A regex pattern. | `""` | no
`from_attribute` | `string` | The attribute from the input data to use to populate the attribute value. | `""` | no
`from_context` | `string` | The context value to use to populate the attribute value.  | `""` | no
`converted_type` | `string` | The type to convert the attribute value to. | `""` | no

The type of `value` could be a number, a string or a boolean.

The supported values for `action` are:

* `insert`: Inserts a new attribute in input data where the key does not already exist.

    * `key` is required. It specifies the attribute to act upon.
    * One of `value`, `from_attribute` or `from_context` is required.
    * `action = "insert"` is required.

* `update`: Updates an attribute in input data where the key does exist.

    * `key` is required. It specifies the attribute to act upon.
    * One of `value`, `from_attribute` or `from_context` is required.
    * `action = "update"` is required.

* `upsert`: Performs insert or update. Inserts a new attribute in input data where the key does not already exist and updates an attribute in input data where the key does exist.

    * `key` is required. It specifies the attribute to act upon.
    * One of `value`, `from_attribute` or `from_context` is required:
        * `value` specifies the value to populate for the key.
        * `from_attribute` specifies the attribute from the input data to use to populate
        the value. If the attribute doesn't exist, no action is performed.
        * `from_context` specifies the context value to use to populate the attribute value.
        If the key is prefixed with `metadata.`, the values are searched
        in the receiver's transport protocol additional information like gRPC Metadata or HTTP Headers. 
        If the key is prefixed with `auth.`, the values are searched
        in the authentication information set by the server authenticator. 
        Refer to the server authenticator's documentation part of your pipeline 
        for more information about which attributes are available.
        If the key doesn't exist, no action is performed.
        If the key has multiple values the values will be joined with `;` separator.
    * `action = "upsert"` is required.

* `hash`: Hashes (SHA1) an existing attribute value.

    * `key` and/or `pattern` is required.
    * `action = "hash"` is required.

* `extract`: Extracts values using a regular expression rule from the input key to target keys specified in the rule. If a target key already exists, it will be overridden. Note: It behaves similar to the Span Processor `to_attributes` setting with the existing attribute as the source.

    * `key` is required. It specifies the attribute to extract values from. The value of `key` is NOT altered.
    * `pattern` is required. It is the regex pattern used to extract attributes from the value of `key`. The submatchers must be named. If attributes already exist, they will be overwritten.
    * `action = "extract"` is required.

* `convert`: Converts an existing attribute to a specified type.

    * `key` is required. It specifies the attribute to act upon.
    * `action = "convert"` is required.
    * `converted_type` is required and must be one of int, double or string.

* `delete`: Deletes an attribute from the input data.

    * `key` and/or `pattern` is required. It specifies the attribute to act upon.
    * `action = "delete"` is required.

### include/exclude blocks

The `include` and `exclude` blocks provide an option to include or exclude data from being fed into the [action] blocks, based on the properties of a span, log, or metric records.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`match_type` | `string` | Controls how items in "services" and "span_names" arrays are interpreted. | | yes
`services` | `list(string)` | A list of items to match the service name against. | `[]` | no
`span_names` | `list(string)` | A list of items to match the span name against. | `[]` | no
`log_bodies` | `list(string)` | A list of strings that the LogRecord's body field must match against. | `[]` | no
`log_severity_texts` | `list(string)` | A list of strings that the LogRecord's severity text field must match against. | `[]` | no
`metric_names` | `list(string)` | A list of strings to match the metric name against. | `[]` | no
`span_kinds` | `list(string)` | A list of items to match the span kind against. | `[]` | no

`match_type` is required and can be set to either `"regexp"` or `"strict"`.

One of the following is also required:
* For spans, one of `services`, `span_names`, `span_kinds`, [attribute], [resource], or [library] must be specified with a non-empty value for a valid configuration. The `log_bodies`, `log_severity_texts`, `resource_attributes` and `metric_names` fields are invalid.
* For logs, one of `log_bodies`, `log_severity_texts`, [attribute], [resource], or [library] must be specified with a non-empty value for a valid configuration. The `span_names`, `span_kinds`, `metric_names`, `resource_attributes`, and `services` fields are invalid.
* For metrics, one of `metric_names`, [resource] must be specified with a valid non-empty value for a valid configuration. The `span_names`, `span_kinds`, `log_bodies`, `log_severity_texts` and `services` fields are invalid.

For `metric_names`, a match occurs if the metric name matches at least one item in the list.
For `span_kinds`, a match occurs if the span's span kind matches at least one item in this list.

Note: If both `include` and `exclude` are specified, the include properties are checked before the exclude properties.

### regexp block

This block is optional configuration for the `match_type` of `"regexp"`.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`cacheenabled` | `bool` | Determines whether match results are LRU cached. | `false` | no
`cachemaxnumentries` | `int` | The max number of entries of the LRU cache that stores match results. | `0` | no

Enabling `cacheenabled` could make subsequent matches faster.
Cache size is unlimited unless `cachemaxnumentries` is also specified.

`cachemaxnumentries` is ignored if `cacheenabled` is false.

### attribute block

This block specifies a list of attributes to match against.
Only `match_type = "strict"` is allowed if `attribute` is specified.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`key` | `string` | The attribute key. | | yes
`value` | `any` | The value to match against. | | no

If `value` is not set, any value will match.
The type of `value` could be a number, a string or a boolean.

### resource block

This block specifies items to match the resources against.
A match occurs if the input data resources matches at least one `resource` block.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`key` | `string` | The attribute key. | | yes
`value` | `any` | The value to match against. | | no

If `value` is not set, any value will match.
The type of `value` could be a number, a string or a boolean.

### library block

This block specifies items to match the implementation library against.
A match occurs if the span's implementation library matches at least one `library` block.

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`name` | `string` | The attribute key. | | yes
`version` | | The value to match against. | | yes

### output block

{{< docs/shared lookup="flow/reference/components/output-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`input` | `otelcol.Consumer` | A value that other components can use to send telemetry data to.

`input` accepts `otelcol.Consumer` data for any telemetry signal (metrics, logs, or traces).

## Component health

`otelcol.processor.attributes` is only reported as unhealthy if given an invalid
configuration.

## Debug information

`otelcol.processor.attributes` does not expose any component-specific debug
information.

## Examples

### Various uses of the "action" block

```river
otelcol.receiver.otlp "default" {
  http {}
  grpc {}

  output {
    metrics = [otelcol.processor.attributes.default.input]
    logs    = [otelcol.processor.attributes.default.input]
    traces  = [otelcol.processor.attributes.default.input]
  }
}

otelcol.processor.attributes "default" {
    // Inserts a new attribute "attribute1" to spans where
    // the key "attribute1" doesn't exist.
    // The type of `attribute1` is inferred by the configuration.
    // `123` is an integer and is stored as an integer in the attributes.
    action {
        key = "attribute1"
        value = 123
        action = "insert"
    }

    // Inserts a new attribute with a key of "string key" and
    // a string value of "anotherkey".
    action {
        key = "string key"
        value = "anotherkey"
        action = "insert"
    }

    // Setting an attribute on all spans.
    // Any spans that already had `region` now have value `planet-earth`.
    // This can be done to set properties for all traces without
    // requiring an instrumentation change.
    action {
        key = "region"
        value = "planet-earth"
        action = "upsert"
    }

    // The following demonstrates copying a value to a new key.
    // If a span doesn't contain `user_key`, no new attribute `new_user_key` is created.
    action {
        key = "new_user_key"
        from_attribute = "user_key"
        action = "upsert"
    }

    // Hashing existing attribute values.
    action {
        key = "user.email"
        action = "hash"
    }

    // Uses the value from `key:http.url` to upsert attributes
    // to the target keys specified in the `pattern`.
    // (Insert attributes for target keys that do not exist and update keys that exist.)
    // Given http.url = http://example.com/path?queryParam1=value1,queryParam2=value2
    // then the following attributes will be inserted:
    // http_protocol: http
    // http_domain: example.com
    // http_path: path
    // http_query_params=queryParam1=value1,queryParam2=value2
    // http.url value does NOT change.
    // Note: Similar to the Span Processor, if a target key already exists,
    // it will be updated.
    action {
        key = "http.url"
        pattern = "^(?P<http_protocol>.*):\\/\\/(?P<http_domain>.*)\\/(?P<http_path>.*)(\\?|\\&)(?P<http_query_params>.*)"
        action = "extract"
    }

    // Converting the type of an existing attribute value.
    action {
        key = "http.status_code"
        converted_type = "int"
        action = "convert"
    }

    // Deleting keys from an attribute.
    action {
        key = "credit_card"
        action = "delete"
    }

    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```

### Excluding spans based on attributes

For example, the following spans match the properties and won't be processed by the processor:
* Span1 Name: "svcB", Attributes: {env: "dev", test_request: 123, credit_card: 1234}
* Span2 Name: "svcA", Attributes: {env: "dev", test_request: false}

The following spans do not match the properties and the processor actions are applied to it:
* Span3 Name: "svcB", Attributes: {env: 1, test_request: "dev", credit_card: 1234}
* Span4 Name: "svcC", Attributes: {env: "dev", test_request: false}

```river
otelcol.processor.attributes "default" {
    exclude {
        match_type = "strict"
        services = ["svcA", "svcB"]
        attribute {
            key = "env"
            value = "dev"
        }
        attribute {
            key = "test_request"
        }
    }
    action {
        key = "credit_card"
        action = "delete"
    }
    action {
        key = "duplicate_key"
        action = "delete"
    }
    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}
```

### Excluding spans based on resources

A "strict" `match_type` means that we must match the `resource` key/value pairs strictly.

```river
otelcol.processor.attributes "default" {
    exclude {
        match_type = "strict"
        resource {
            key = "host.type"
            value = "n1-standard-1"
        }
    }
    action {
        key = "credit_card"
        action = "delete"
    }
    action {
        key = "duplicate_key"
        action = "delete"
    }
    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}
```

### Excluding spans based on resources

A "strict" `match_type` means that we must match the `library` key/value pairs strictly.

```river
otelcol.processor.attributes "default" {
    exclude {
        match_type = "strict"
        library {
            name = "mongo-java-driver"
            version = "3.8.0"
        }
    }
    action {
        key = "credit_card"
        action = "delete"
    }
    action {
        key = "duplicate_key"
        action = "delete"
    }
    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}
```

### Including and excluding spans based on regex and services

This processor will remove the "token" attribute and will obfuscate the "password" attribute 
in spans where the service name matches `"auth.*"` and where the span name does not match `"login.*"`.

```river
otelcol.processor.attributes "default" {
    // Specifies the span properties that must exist for the processor to be applied.
    include {
        // "match_type" defines that "services" is an array of regexp-es.
        match_type = "regexp"
        // The span service name must match "auth.*" pattern.
        services = ["auth.*"]
    }

    exclude {
        // "match_type" defines that "span_names" is an array of regexp-es.
        match_type = "regexp"
        // The span name must not match "login.*" pattern.
        span_names = ["login.*"]
    }

    action {
        key = "password"
        action = "update"
        value = "obfuscated"
    }

    action {
        key = "token"
        action = "delete"
    }

    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}
```

### Including spans based on regex and attributes

The following demonstrates how to process spans that have an attribute that matches a regexp patterns.
This processor will obfuscate the "db.statement" attribute in spans where the "db.statement" attribute
matches a regex pattern.

```river
otelcol.processor.attributes "default" {
    include {
        // "match_type" of "regexp" defines that the "value" attributes 
        // in the "attribute" blocks are regexp-es.
        match_type = "regexp"

        // This attribute ('db.statement') must exist in the span and match 
        // the regex ('SELECT \* FROM USERS.*') for a match.
        attribute {
            key = "env"
            value = "SELECT \* FROM USERS.*"
        }
    }

    action {
        key = "db.statement"
        action = "update"
        value = "SELECT * FROM USERS [obfuscated]"
    }

    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}
```

### Including spans based on regex of log body

This processor will remove "token" attribute and will obfuscate "password"
attribute in spans where the log body matches "AUTH.*".

```river
otelcol.processor.attributes "default" {
    include {
        match_type = "regexp"
        log_bodies = ["AUTH.*"]
    }
    action {
        key = "password"
        action = "update"
        value = "obfuscated"
    }
    action {
        key = "token"
        action = "delete"
    }

    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}
```

### Including spans based on regex of log severity

The following demonstrates how to process logs that have a severity text that match regexp
patterns. This processor will remove "token" attribute and will obfuscate "password"
attribute in spans where severity matches "debug".

```river
otelcol.processor.attributes "default" {
    include {
        match_type = "regexp"
        log_severity_texts = ["debug.*"]
    }
    action {
        key = "password"
        action = "update"
        value = "obfuscated"
    }
    action {
        key = "token"
        action = "delete"
    }

    output {
        metrics = [otelcol.exporter.otlp.default.input]
        logs    = [otelcol.exporter.otlp.default.input]
        traces  = [otelcol.exporter.otlp.default.input]
    }
}
```
