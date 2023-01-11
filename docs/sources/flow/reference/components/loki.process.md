---
aliases:
- /docs/agent/latest/flow/reference/components/loki.process
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

Hierarchy        | Block      | Description | Required
---------------- | ---------- | ----------- | --------
stage          | [stage][]  | Processing stage to run. | no
stage > docker | [docker][] | Configures a pre-defined Docker log format pipeline. | no
stage > cri    | [cri][]    | Configures a pre-defined CRI-format pipeline. | no
stage > json   | [json][]   | Configures a JSON processing stage.  | no
stage > labels | [labels][] | Configures a labels processing stage. | no
stage > label_keep   | [label_keep][]    | Configures a `label_keep` processing stage. | no
stage > label_drop   | [label_drop][]    | Configures a `label_drop` processing stage. | no
stage > static_labels | [static_labels][] | Configures a `static_labels` processing stage. | no
stage > regex        | [regex][]         | Configures a `regex` processing stage. | no
stage > timestamp    | [timestamp][]     | Configures a `timestamp` processing stage. | no
stage > output       | [output][]        | Configures an `output` processing stage. | no

The `>` symbol indicates deeper levels of nesting. For example, `stage > json`
refers to a `json` block defined inside of a `stage` block.

[stage]: #stage-block
[docker]: #docker-block
[cri]: #cri-block
[json]: #json-block
[labels]: #labels-block
[label_keep]: #label_keep-block
[label_drop]: #label_drop-block
[static_labels]: #static_labels-block
[regex]: #regex-block
[timestamp]: #timestamp-block
[output]: #output-block

### stage block

The `stage` block describes a single processing step to run log entries
through. As such, each block must have exactly _one_ inner block to match the
type of stage to configure. Multiple processing stages must be defined in
different blocks and are applied on the incoming log entries in top-down order.

The `stage` block does not support any arguments and is configured only via
inner blocks.

### docker block

The `docker` inner block enables a predefined pipeline which reads log lines in
the standard format of Docker log files.

The `docker` block does not support any arguments or inner blocks, so it is
always empty.

```river
stage {
	docker {}
}
```

Docker log entries are formatted as JSON with the following keys:

* `log`: The content of log line
* `stream`: Either `stdout` or `stderr`
* `time`: The timestamp string of the log line

Given the following log line, the subsequent key-value pairs are created in the
shared map of extracted data:

```
{"log":"log message\n","stream":"stderr","time":"2019-04-30T02:12:41.8443515Z"}

output: log message\n
stream: stderr
timestamp: 2019-04-30T02:12:41.8443515
```

### cri block

The `cri` inner block enables a predefined pipeline which reads log lines using
the CRI logging format.

The `cri` block does not support any arguments or inner blocks, so it is always
empty.

```river
stage {
	cri {}
}
```

CRI specifies log lines as single space-delimited values with the following
components:

* `time`: The timestamp string of the log
* `stream`: Either `stdout` or `stderr`
* `flags`: CRI flags including `F` or `P`
* `log`: The contents of the log line

Given the following log line, the subsequent key-value pairs are created in the
shared map of extracted data:
```
"2019-04-30T02:12:41.8443515Z stdout F message"

content: message
stream: stdout
timestamp: 2019-04-30T02:12:41.8443515
```

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

### label_keep block

The `label_keep` inner block configures a processing stage that filters the
label set of an incoming log entry down to a subset.

The following arguments are supported:

Name        | Type           | Description                                   | Default        | Required
----------- | -------------- | --------------------------------------------- | -------------- | --------
`values`    | `list(string)` | Configures a `label_keep` processing stage.   | `{}`           | no


```river
stage {
	label_keep {
		values = [ "kubernetes_pod_name", "kubernetes_container_name" ]
	}
}
```

### label_drop block

The `label_drop` inner block configures a processing stage that drops labels
from incoming log entries.

The following arguments are supported:

Name         | Type           | Description                                  | Default        | Required
------------ | -------------- | -------------------------------------------- | -------------- | --------
`values`     | `list(string)` | Configures a `label_drop` processing stage.   | `{}`           | no

```river
stage {
	label_drop {
		values = [ "kubernetes_node_name", "kubernetes_namespace" ]
	}
}
```

### static_labels block

The `static_labels` inner block configures a static_labels processing stage
that adds a static set of labels to incoming log entries.

The following arguments are supported:

Name         | Type          | Description                                      | Default        | Required
------------ | --------------| ------------------------------------------------ | -------------- | --------
`values`     | `map(string)` | Configures a `static_labels` processing stage.   | `{}`           | no


```river
stage {
	static_labels {
		values = {
			foo = "fooval",
			bar = "barval",
		}
	}
}
```

### regex block

The `regex` inner block configures a processing stage that parses log lines
using regular expressions and uses named capture groups for adding data into
the shared extracted map of values.

The following arguments are supported:

Name          | Type      | Description                                                         | Default | Required
------------- | --------- | ------------------------------------------------------------------- | ------- | --------
`expression`  | `string`  | A valid RE2 regular expression. Each capture group must be named.   |         | yes
`source`      | `string`  | Name from extracted data to parse. If empty, uses the log message.  | `""`    | no


The `expression` field needs to be a Go RE2 regex string. Every capture group (re) is set into the extracted map, so it must be named like: `(?P<name>re)`. The name of the capture group is then used as the key in the extracted map.

<!--
We don't care about YAML, what does River do instead???

Because of how YAML treats backslashes in double-quoted strings, note that all backslashes in a regex expression must be escaped when using double quotes. For example, all of these are valid:

expression: \w*
expression: '\w*'
expression: "\\w*"
But these are not:

expression: \\w* (only escape backslashes when using double quotes)
expression: '\\w*' (only escape backslashes when using double quotes)
expression: "\w*" (backslash must be escaped)
-->

If the `source` is empty, then the stage uses attempts to parse the log line
itself.

Given the following log line and regex stage, the extracted values are shown
below:

```
2019-01-01T01:00:00.000000001Z stderr P i'm a log message!

stage {
  regex {
    expression = "^(?s)(?P<time>\\S+?) (?P<stream>stdout|stderr) (?P<flags>\\S+?) (?P<content>.*)$"
  }
}


time: 2019-01-01T01:00:00.000000001Z,
stream: stderr,
flags: P,
content: i'm a log message
```

On the other hand, if the `source` value is set, then the regex is applied to
the value stored in the shared map under that name.

Let's see what happens when the following log line is put through this
two-stage pipeline:
```
{"timestamp":"2022-01-01T01:00:00.000000001Z"}

stage {
  json {
    expressions = { time = "timestamp" }
  }
}
stage {
  regex {
    expression = "^(?P<year>\\d+)"
    source     = "time"
  }
}
```

The first stage adds the following key-value pair into the extracted map:
```
time: 2022-01-01T01:00:00.000000001Z
```

Then, the regex stage parses the value for time from the shared values and
appends the subsequent key-value pair back into the extracted values map:
```
year: 2022
```


### timestamp block

The `timestamp` inner block configures a processing stage that sets the
timestamp of log entries before they're forwarded to the next component. When
no timestamp stage is set, the log entry timestamp defaults to the time when
the log entry was scraped.

The following arguments are supported:

Name                | Type           | Description                                                 | Default   | Required
------------------- | -------------- | ----------------------------------------------------------- | --------- | --------
`source`            | `string`       | Name from extracted values map to use for the timestamp.    |           | yes
`format`            | `string`       | Determines how to parse the source string.                  |           | yes
`fallback_formats`  | `list(string)` | Fallback formats to try if the `format` field fails.        | `[]`      | no
`location`          | `string`       | IANA Timezone Database location to use when parsing.        | `""`      | no
`action_on_failure` | `string`       | What to do when the timestamp can't be extracted or parsed. | `"fudge"` | no

The `source` field defines which value from the shared map of extracted values
the stage should attempt to parse as a timestamp.

The `format` field defines _how_ that source should be parsed.

First off, the `format` can be set to one of the following shorthand values for
commonly-used forms:
```
ANSIC: Mon Jan _2 15:04:05 2006
UnixDate: Mon Jan _2 15:04:05 MST 2006
RubyDate: Mon Jan 02 15:04:05 -0700 2006
RFC822: 02 Jan 06 15:04 MST
RFC822Z: 02 Jan 06 15:04 -0700
RFC850: Monday, 02-Jan-06 15:04:05 MST
RFC1123: Mon, 02 Jan 2006 15:04:05 MST
RFC1123Z: Mon, 02 Jan 2006 15:04:05 -0700
RFC3339: 2006-01-02T15:04:05-07:00
RFC3339Nano: 2006-01-02T15:04:05.999999999-07:00
```

Additionally, support for common Unix timestamps is supported with the
following format values:
```
Unix: 1562708916 or with fractions 1562708916.000000123
UnixMs: 1562708916414
UnixUs: 1562708916414123
UnixNs: 1562708916000000123
```

Otherwise, the field accepts a custom format string that defines how an
arbitrary reference point in history (Mon Jan 2 15:04:05 -0700 MST 2006) should
be interpreted by the stage.

The string value of the field is passed directly to the layout parameter in
Go's [`time.Parse`](https://pkg.go.dev/time#Parse) function.

If the custom format has no year component, the stage uses the current year,
according to the system's clock.

The following table shows the supported reference values to use when defining a
custom format.

Timestamp Component | Format value 
------------------- | --------------
Year                | 06, 2006
Month               | 1, 01, Jan, January
Day                 | 2, 02, _2 (two digits right justified)
Day of the week     | Mon, Monday
Hour                | 3 (12-hour), 03 (12-hour zero prefixed), 15 (24-hour)
Minute              | 4, 04
Second              | 5, 05
Fraction of second  | .000 (ms zero prefixed), .000000 (μs), .000000000 (ns), .999 (ms without trailing zeroes), .999999 (μs), .999999999 (ns)
12-hour period      | pm, PM
Timezone name       | MST
Timezone offset     | -0700, -070000 (with seconds), -07, 07:00, -07:00:00 (with seconds)
Timezone ISO-8601   | Z0700 (Z for UTC or time offset), Z070000, Z07, Z07:00, Z07:00:00

The `fallback_formats` field defines one or more format fields to try and parse
the timestamp with, if parsing with `format` fails.

The `location` field must be a valid IANA Timezone Database location and
determines in which timezone the timestamp value is interpreted to be in.

The `action_on_failure` field defines what should happen when the source field
doesn't exist in the shared extracted map, or if the timestamp parsing fails.

The supported actions are:

* fudge (default): Change the timestamp to the last known timestamp, summing up
  1 nanosecond (to guarantee log entries ordering).
* skip: Do not change the timestamp and keep the time when the log entry was
  scraped.

The following stage fetches the `time` value from the shared values map, parses
it as a RFC3339 format, and sets it as the log entry's timestamp.

```
stage {
	timestamp {
		source = "time"
		format = "RFC3339"
	}
}
```


### output block

The `output` inner block configures a processing stage that reads from the
extracted map and changes the content of the log entry that is forwarded
to the next component.

The following arguments are supported:

Name                | Type           | Description                                           | Default   | Required
------------------- | -------------- | ----------------------------------------------------- | --------- | --------
`source`            | `string`       | Name from extracted data to use for the log entry.    |           | yes


Let's see how this works for the following log line and three-stage pipeline:

```
{"user": "John Doe", "message": "hello, world!"}

stage {
	json {
		expressions = { "user" = "user", "message" = "message" }
	}
}

stage {
	labels {
		values = { "user" = "user" }
	}
}

stage {
	output {
		source = "message"
	}
}
```

The first stage extracts the following key-value pairs into the shared map:
```
user: John Doe
message: hello, world!
```

Then, the second stage adds `user="John Doe"` to the label set of the log
entry, and the final output stage changes the log line from the original
JSON to `hello, world!`.


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

