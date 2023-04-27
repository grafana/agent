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

Hierarchy        | Block      | Description | Required
---------------- | ---------- | ----------- | --------
stage          | [stage][]  | Processing stage to run. | no
stage > docker | [docker][] | Configures a pre-defined Docker log format pipeline. | no
stage > cri    | [cri][]    | Configures a pre-defined CRI-format pipeline. | no
stage > json   | [json][]   | Configures a JSON processing stage.  | no
stage > logfmt | [logfmt][] | Configures a logfmt processing stage. | no
stage > labels | [labels][] | Configures a labels processing stage. | no
stage > label_keep   | [label_keep][]    | Configures a `label_keep` processing stage. | no
stage > label_drop   | [label_drop][]    | Configures a `label_drop` processing stage. | no
stage > static_labels | [static_labels][] | Configures a `static_labels` processing stage. | no
stage > regex        | [regex][]         | Configures a `regex` processing stage. | no
stage > timestamp    | [timestamp][]     | Configures a `timestamp` processing stage. | no
stage > output       | [output][]        | Configures an `output` processing stage. | no
stage > replace      | [replace][]       | Configures a `replace` processing stage. | no
stage > multiline    | [multiline][]     | Configures a `multiline` processing stage. | no
stage > match        | [match][]         | Configures a `match` processing stage. | no
stage > drop         | [drop][]          | Configures a `drop` processing stage. | no

The `>` symbol indicates deeper levels of nesting. For example, `stage > json`
refers to a `json` block defined inside of a `stage` block.

[stage]: #stage-block
[docker]: #docker-block
[cri]: #cri-block
[json]: #json-block
[logfmt]: #logfmt-block
[labels]: #labels-block
[label_keep]: #label_keep-block
[label_drop]: #label_drop-block
[static_labels]: #static_labels-block
[regex]: #regex-block
[timestamp]: #timestamp-block
[output]: #output-block
[replace]: #replace-block
[multiline]: #multiline-block
[match]: #match-block
[drop]: #drop-block

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

### logfmt block

The `logfmt` inner block configures a processing stage that reads incoming log
lines as logfmt and extracts values from them.

The following arguments are supported:

Name       | Type          | Description | Default | Required
---------- | ------------- | ----------- | ------- | --------
`mapping`  | `map(string)` | Key-value pairs of logmft fields to extract. | | yes
`source`   | `string`      | Source of the data to parse as logfmt. | `""` | no


The `source` field defines the source of data to parse as logfmt. When `source`
is missing or empty, the stage parses the log line itself, but it can also be
used to parse a previously extracted value.

This stage uses the [go-logfmt](https://github.com/go-logfmt/logfmt)
unmarshaler, so that numeric or boolean types are unmarshalled into their
correct form. The stage does not perform any other type conversions. If the
extracted value is a complex type, it is treated as a string.

Let's see how this works on the following log line and stages.

```
time=2012-11-01T22:08:41+00:00 app=loki level=WARN duration=125 message="this is a log line" extra="user=foo"

stage {
	logfmt {
		mapping = { "extra" = "" }
	}
}

stage {
	logfmt {
		mapping = { "username" = "user" }
		source  = "extra"
	}
}
```

The first stage parses the log line itself and inserts the `extra` key in the
set of extracted data, with the value of `user=foo`.

The second stage parses the contents of `extra` and appends the `username: foo`
key-value pair to the set of extracted data.


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


The `expression` field needs to be a RE2 regex string. Every matched capture
group is added to the extracted map, so it must be named like: `(?P<name>re)`.
The name of the capture group is then used as the key in the extracted map for
the matched value.

Because of how River strings work, any backslashes in `expression` must be
escaped with a double backslash; for example `"\\w"` or `"\\S+"`.

If the `source` is empty or missing, then the stage parses the log line itself.
If it's set, the stage parses a previously extracted value with the same name.

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
arbitrary reference point in history should
be interpreted by the stage. The arbitrary reference point is Mon Jan 2 15:04:05 -0700 MST 2006.

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

### replace block

The `replace` inner block configures a stage that parses a log line using a
regular expression and replaces the log line contents. Named capture groups in
the regex also support adding data into the shared extracted map.

The following arguments are supported:

Name           | Type      | Description                                        | Default   | Required
-------------- | --------- | -------------------------------------------------- | --------- | --------
`expression`   | `string`  | Name from extracted data to use for the log entry. |           | yes
`source`       | `string`  | Source of the data to parse. If empty, it uses the log message. | | no
`replace`      | `string`  | Value replaced by the capture group.               |           | no


The `source` field defines the source of data to parse using `expression`. When
`source` is missing or empty, the stage parses the log line itself, but it can
also be used to parse a previously extracted value. The replaced value is
assigned back to the `source` key.

The `expression` must be a valid RE2 regex. Every named capture group
`(?P<name>re)` is set into the extracted map with its name.

Because of how River treats backslashes in double-quoted strings, note that all
backslashes in a regex expression must be escaped like `"\\w*"`.

Let's see how this works with the following log line and stage. Since `source`
is omitted, the replacement occurs  on the log line itself.

```
2023-01-01T01:00:00.000000001Z stderr P i'm a log message who has sensitive information with password xyz!

stage {
	replace {
		expression = "password (\\S+)"
		replace    = "*****"
	}
}
```

The log line is transformed to 
```
2023-01-01T01:00:00.000000001Z stderr P i'm a log message who has sensitive information with password *****!
```

If `replace` is empty, then the captured value is omitted instead.

In the following example, `source` is defined.
```
{"time":"2023-01-01T01:00:00.000000001Z", "level": "info", "msg":"11.11.11.11 - \"POST /loki/api/push/ HTTP/1.1\" 200 932 \"-\" \"Mozilla/5.0\"}

stage {
	json {
		expressions = { "level" = "", "msg" = "" }
	}
}

stage {
	replace {
		expression = "\\S+ - \"POST (\\S+) .*"
		source     = "msg"
		replace    = "redacted_url"
	}
}
```

The JSON stage adds the following key-value pairs into the extracted map:
```
time: 2023-01-01T01:00:00.000000001Z
level: info
msg: "11.11.11.11 - "POST /loki/api/push/ HTTP/1.1" 200 932 "-" "Mozilla/5.0"
```

The `replace` stage acts on the `msg` value. The capture group matches against
`/loki/api/push` and is replaced by `redacted_url`.

The `msg` value is finally transformed into:
```
msg: "11.11.11.11 - "POST redacted_url HTTP/1.1" 200 932 "-" "Mozilla/5.0"
```

The `replace` field can use a set of templating functions, by utilizing Go's
[text/template](https://pkg.go.dev/text/template) package.

Let's see how this works with named capture groups with a sample log line
and stage.
```
11.11.11.11 - agent [01/Jan/2023:00:00:01 +0200]

stage {
	replace {
		expression = "^(?P<ip>\\S+) (?P<identd>\\S+) (?P<user>\\S+) \\[(?P<timestamp>[\\w:/]+\\s[+\\-]\\d{4})\\]"
		replace    = "{{ .Value | ToUpper }}"
	}
}
```

Since `source` is empty, the regex parses the log line itself and extracts the
named capture groups to the shared map of values. The `replace` field acts on
these extracted values and converts them to uppercase:
```
ip: 11.11.11.11
identd: -
user: FRANK
timestamp: 01/JAN/2023:00:00:01 +0200
```

and the log line becomes:
```
11.11.11.11 - FRANK [01/JAN/2023:00:00:01 +0200] 
```

The following list contains available functions with examples of
more complex `replace` fields.
```
ToLower, ToUpper, Replace, Trim, TrimLeftTrimRight, TrimPrefix, TrimSuffix, TrimSpace, Hash, Sha2Hash, regexReplaceAll, regexReplaceAllLiteral

"{{ if eq .Value \"200\" }}{{ Replace .Value \"200\" \"HttpStatusOk\" -1 }}{{ else }}{{ .Value | ToUpper }}{{ end }}"
"*IP4*{{ .Value | Hash "salt" }}*"
```


### multiline block

The `multiline` inner block merges multiple lines into a single block before
passing it on to the next stage in the pipeline.

The following arguments are supported:

Name                | Type           | Description                                           | Default  | Required
------------------- | -------------- | ----------------------------------------------------- | -------- | --------
`firstline`         | `string`       | Name from extracted data to use for the log entry.    |          | yes
`max_wait_time`     | `duration`     | The maximum time to wait for a multiline block.       |  `"3s"`  | no 
`max_lines`         | `int`          | The maximum number of lines a block can have.         |  `128`   | no


A new block is identified by the RE2 regular expression passed in `firstline`.


Any line that does _not_ match the expression is considered to be part of the
block of the previous match. If no new logs arrive with `max_wait_time`, the
block is sent on. The `max_lines` field defines the maximum number of lines a
block can have. If this is exceeded, a new block is started.

Let's see how this works in practice with an example stage and a stream of log
entries from a Flask web service.

```
stage {
	multiline {
		firstline     = "^\[\d{4}-\d{2}-\d{2} \d{1,2}:\d{2}:\d{2}\]"
		max_wait_time = "10s"
	}
}

[2023-01-18 17:41:21] "GET /hello HTTP/1.1" 200 -
[2023-01-18 17:41:25] ERROR in app: Exception on /error [GET]
Traceback (most recent call last):
  File "/home/pallets/.pyenv/versions/3.8.5/lib/python3.8/site-packages/flask/app.py", line 2447, in wsgi_app
    response = self.full_dispatch_request()
  File "/home/pallets/.pyenv/versions/3.8.5/lib/python3.8/site-packages/flask/app.py", line 1952, in full_dispatch_request
    rv = self.handle_user_exception(e)
  File "/home/pallets/.pyenv/versions/3.8.5/lib/python3.8/site-packages/flask/app.py", line 1821, in handle_user_exception
    reraise(exc_type, exc_value, tb)
  File "/home/pallets/.pyenv/versions/3.8.5/lib/python3.8/site-packages/flask/_compat.py", line 39, in reraise
    raise value
  File "/home/pallets/.pyenv/versions/3.8.5/lib/python3.8/site-packages/flask/app.py", line 1950, in full_dispatch_request
    rv = self.dispatch_request()
  File "/home/pallets/.pyenv/versions/3.8.5/lib/python3.8/site-packages/flask/app.py", line 1936, in dispatch_request
    return self.view_functions[rule.endpoint](**req.view_args)
  File "/home/pallets/src/deployment_tools/hello.py", line 10, in error
    raise Exception("Sorry, this route always breaks")
Exception: Sorry, this route always breaks
[2023-01-18 17:42:24] "GET /error HTTP/1.1" 500 -
[2023-01-18 17:42:29] "GET /hello HTTP/1.1" 200 -
```

All 'blocks' that form log entries of separate web requests start with a
timestamp in square brackets. The stage detects this with the regular
expression in `firstline` to collapse all lines of the traceback into a single
block and thus a single Loki log entry.


### match block

The `match` inner block configures a filtering stage that can conditionally
either apply a nested set of processing stages or drop an entry when a log
entry matches a configurable LogQL stream selector and filter expressions.

The following arguments are supported:

Name            | Type      | Description                                                         | Default | Required
--------------- | --------- | ------------------------------------------------------------------- | ------- | --------
`selector`      | `string`  | The LogQL stream selector and filter expressions to use.            |         | yes
`pipeline_name` | `string`  | A custom name to use for the nested pipeline.                       | `""`    | no
`action`        | `string`  | The action to take when the selector matches the log line. Supported values are `"keep"` and `"drop"` | `"keep"` | no
`drop_counter_reason` | `string` | A custom reason to report for dropped lines.                   | `"match_stage"` | no 

The `match` block supports a number of `stage` inner blocks, like the top-level
block. These are used to construct the nested set of stages to run if the
selector matches the labels and content of the log entries.

The following blocks are supported inside the definition of `stage > match`:

Hierarchy      | Block      | Description | Required
-------------- | ---------- | ----------- | --------
stage          | [stage][]  | Processing stage to run. | no


If the specified action is `"drop"`, the metric
`loki_process_dropped_lines_total` is incremented with every line dropped.
By default, the reason label is `"match_stage"`, but a custom reason can be
provided by using the `drop_counter_reason` argument.

Let's see this in action, with the following log lines and stages
```
{ "time":"2023-01-18T17:08:41+00:00", "app":"foo", "component": ["parser","type"], "level" : "WARN", "message" : "app1 log line" }
{ "time":"2023-01-18T17:08:42+00:00", "app":"bar", "component": ["parser","type"], "level" : "ERROR", "message" : "foo noisy error" }

stage {
	json {
		expressions = { "appname" = "app" }
	}
}

stage {
	labels {
		values = { "applbl" = "appname" }
	}
}

stage {
	match {
		selector = '{applbl="foo"}'
		stage {
			json {
				expressions = { "msg" = "message" }
			}
		}
	}
}

stage {
	match {
		selector = '{applbl="qux"}'
		stage {
			json {
				expressions = { "msg" = "msg" }
			}
		}
	}
}

stage {
	match {
		selector = '{applbl="bar"} |~ ".*noisy error.*"'
		action   = "drop"

		drop_counter_reason = "discard_noisy_errors"
	}
}

stage {
	output {
		source = "msg"
	}
}
```

The first two stages parse the log lines as JSON, decode the `app` value into
the shared extracted map as `appname`, and use its value as the `applbl` label.

The third stage uses the LogQL selector to only execute the nested stages on
lines where the `applbl="foo"`. So, for the first line, the nested JSON stage
adds `msg="app1 log line"` into the extracted map.

The fourth stage uses the LogQL selector to only execute on lines where
`applbl="qux"`; that means it won't match any of the input, and the nested
JSON stage does not run.

The fifth stage drops entries from lines where `applbl` is set to 'bar' and the
line contents matches the regex `.*noisy error.*`. It also increments the
`loki_process_dropped_lines_total` metric with a label
`drop_counter_reason="discard_noisy_errors"`.

The final output stage changes the contents of the log line to be the value of
`msg` from the extracted map. In this case, the first log entry's content is
changed to `app1 log line`.


### drop block

The `drop` inner block configures a filtering stage that drops log entries
based on several options. If multiple options are provided, they're treated
as AND clauses and must _all_ be true for the log entry to be dropped.
To drop entries with an OR clause, specify multiple `drop` blocks in sequence.

The following arguments are supported:

Name                  | Type       | Description                                           | Default   | Required
--------------------- | ---------- | ----------------------------------------------------- | --------- | --------
`source`              | `string`   | Name from extracted data to parse. If empty or not defined, it uses the log message.    | `""` | no
`expression`          | `string`   | A valid RE2 regular expression. | `""` | no
`value`               | `string`   | If both `source` and `value` are specified, the stage drops lines where `value` exactly matches the source content. | `""` | no
`older_than`          | `duration` | If specified, the stage drops lines whose timestamp is older than the current time minus this duration. | `""` | no
`longer_than`         | `string`   | If specified, the stage drops lines whose size exceeds the configured value. | `""` | no
`drop_counter_reason` | `string`   | A custom reason to report for dropped lines. | `"drop_stage"` | no

The `expression` field needs to be a RE2 regex string. If `source` is empty or
not provided, the regex attempts to match the log line itself. If source is
provided, the regex attempts to match the corresponding value from the
extracted map.

The `value` field can only work with values from the extracted map, and must be
specified together with `source`. Entries are dropped when there is an exact
match between the two.

Whenever an entry is dropped, the metric `loki_process_dropped_lines_total`
is incremented. By default, the reason label is `"drop_stage"`, but you can
provide a custom label using the `drop_counter_reason` argument.

The following stage drops log entries that contain the word `debug` _and_ are
longer than 1KB.

```
stage {
	drop {
		expression  = ".*debug.*"
		longer_than = "1KB"
	}
}
```

On the following example, we define multiple `drop` blocks so `loki.process`
will drop entries that are either 24h or older, are longer than 8KB, _or_ the
extracted value of 'app' is equal to foo.

```
stage {
	drop {
		older_than  = "24h"
		drop_reason = "too old"
	}
}

stage {
	drop {
		older_than  = "8KB"
		drop_reason = "too long"
	}
}

stage {
	drop {
		source = "app"
		value  = "foo"
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

