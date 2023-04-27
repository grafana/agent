---
title: loki.process
---

# loki.process

`loki.process` receives log entries from other loki components, applies one or
more processing _stages_, and forwards the results to the list of receivers
in the component's arguments.

A stage is a multi-purpose block that can parse, transform, and filter log
entries before they're passed to a downstream component. These stages are
applied to each log entry in order of their appearance in the configuration
file. All stages within a `loki.process` block have access to the log entry's
label set, the log line, the log timestamp, as well as a shared map of
'extracted' values so that the results of one stage can be used in a subsequent
one.

Multiple `loki.process` components can be specified by giving them
different labels.

## Usage

```river
loki.process "LABEL" {
  forward_to = RECEIVER_LIST

  stage {
    ...
  }
  ...
}
```

## Arguments

`loki.process` supports the following arguments:

Name              | Type                 | Description                                      | Default | Required
----------------- | -------------------- | ------------------------------------------------ | ------- | --------
`forward_to`      | `list(LogsReceiver)` | Where to forward log entries after processing. | | yes

## Blocks

The following blocks are supported inside the definition of `loki.process`:

Hierarchy      | Block      | Description | Required
-------------- | ---------- | ----------- | --------
stage          | [stage][]  | Processing stage to run. | no
stage > json   | [json][]   | Configures a JSON processing stage.  | no
stage > labels | [labels][] | Configures a labels processing stage. | no

The `>` symbol indicates deeper levels of nesting. For example, `stage > json`
refers to a `json` block defined inside of a `stage` block.

[stage]: #stage-block
[json]: #json-block
[labels]: #labels-block

### stage block

The `stage` block describes a single processing step to run log entries
through. As such, each block must have exactly _one_ inner block to match the
type of stage to configure. Multiple processing stages must be defined in
different blocks and are applied on the incoming log entries in top-down order.

The `stage` block does not support any arguments and is configured only via
inner blocks.

### json block

The `json` inner block configures a JSON processing stage that parses incoming
log lines or previously extracted values as JSON and uses
[JMESPath expressions](https://jmespath.org/tutorial.html) to extract new
values from them.

The following arguments are supported:

Name             | Type          | Description | Default | Required
---------------- | ------------- | ----------- | ------- | --------
`expressions`    | `map(string)` | Key-value pairs of JMESPath expressions. | | yes
`source`         | `string`      | Source of the data to parse as JSON. | `""` | no
`drop_malformed` | `bool`        | Drop lines whose input cannot be parsed as valid JSON.| `false` | no

When configuring a JSON stage, the `source` field defines the source of data to
parse as JSON. By default, this is the log line itself, but it can also be a
previously extracted value.

The `expressions` field is the set of key-value pairs of MESPath expressions to
run. The map key defines the name with which the data is extracted, while the
map value is the expression used to populate the value.

Here's a given log line and two JSON stages to run.

```river
{"log":"log message\n","extra":"{\"user\":\"agent\"}"}

loki.process "username" {
	stage {
		json {
			expressions = {output = log, extra = ""}
		}
	}

	stage {
		json {
			source      = "extra"
			expressions = {username = "user"}
		}
	}
}
```

In this example, the first stage uses the log line as the source and populates
these values in the shared map. An empty expression means using the same value
as the key (so `extra="extra"`).
```
output: log message\n
extra: {"user": "agent"}
```

The second stage uses the value in `extra` as the input and appends the
following key-value pair to the set of extracted data.
```
username: agent
```

### labels block

The `labels` inner block configures a labels processing stage that can read
data from the extracted values map and set new labels on incoming log entries.

The following arguments are supported:

Name                  | Type          | Description                               | Default        | Required
--------------------- | --------------| ----------------------------------------- | -------------- | --------
`values`              | `map(string)` | Configures a `labels` processing stage.   | `{}`           | no

In a labels stage, the map's keys define the label to set and the values are
how to look them up.  If the value is empty, it is inferred to be the same as
the key.

```river
stage {
	labels {
		values = {
			env  = "",         // Sets up an 'env' label, based on the 'env' extracted value.
			user = "username", // Sets up a 'user' label, based on the 'username' extracted value.
		}
	}
}
```

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `LogsReceiver` | A value that other components can use to send log entries to.

## Component health

`loki.process` is only reported as unhealthy if given an invalid configuration.

## Debug information

`loki.process` does not expose any component-specific debug information.

## Debug metrics
* `loki_process_dropped_lines_total` (counter): Number of lines dropped as part of a processing stage.

## Example

This example creates a `loki.process` component that extracts the `environment`
value from a JSON log line and sets it as a label named 'env'.

```river
loki.process "local" {
  forward_to = [loki.write.onprem.receiver]

  stage {
    json {
      expressions = { "extracted_env" = "environment" }
    }
  }

  stage {
    labels = { "env" = "extracted_env" }
  }
}
```

