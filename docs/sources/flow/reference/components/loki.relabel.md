---
aliases:
- /docs/agent/latest/flow/reference/components/loki.relabel
title: loki.relabel
---

# loki.relabel

The `loki.relabel` component rewrites the label set of each log entry passed to
its receiver by applying one or more relabeling `rule`s and forwards the
results to the list of receivers in the component's arguments.

<!-- TODO(@tpaschalis) Add note about loki.transform
  To manipulate the log entry itself, you can look at the loki.transform
  component which allows you to run one or more 'stages' on the log entries.
 -->

If no labels remain after the relabeling rules are applied, then the log entry
are dropped.

The most common use of `loki.relabel` is to filter log entries or standardize
the label set that is passed to one or more downstream receivers. The `rule`
blocks are applied to the label set of each log entry in order of their
appearance in the configuration file.

Multiple `loki.relabel` components can be specified by giving them
different labels.

## Usage

```river
loki.relabel "LABEL" {
  forward_to = RECEIVER_LIST

  rule {
    ...
  }

  ...
}
```

## Arguments

The following arguments are supported:

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`forward_to` | `list(receiver)` | Where the log entries should be forwarded to, after relabeling takes place. | | yes

## Blocks

The following blocks are supported inside the definition of `loki.relabel`:

Hierarchy | Name | Description | Required
--------- | ---- | ----------- | --------
rule | [rule][] | Relabeling rules to apply to received log entries. | no

[rule]: #rule-block

### rule block

{{< docs/shared lookup="flow/reference/components/rule-block.md" source="agent" >}}

## Exported fields

The following fields are exported and can be referenced by other components:

Name | Type | Description
---- | ---- | -----------
`receiver` | `receiver` | The input receiver where log lines are sent to be relabeled.

## Component health

`loki.relabel` is only reported as unhealthy if given an invalid configuration.
In those cases, exported fields are kept at their last healthy values.

## Debug information

`loki.relabel` does not expose any component-specific debug information.

## Debug metrics


* `loki_relabel_entries_processed` (counter): Total number of log entries processed.
* `loki_relabel_entries_written` (counter): Total number of log entries forwarded.
* `loki_relabel_cache_misses` (counter): Total number of cache misses.
* `loki_relabel_cache_hits` (counter): Total number of cache hits.
* `loki_relabel_cache_size` (gauge): Total size of relabel cache.

## Example

One of Loki's core features is that, unlike other logging systems, it does not
index the log message itself, but works off indexing metadata about the
received log streams. These metadata are stored as key-value pairs called
_labels_, much like Prometheus metric labels. Having a small and efficient
index, along being able to store the log data itself in highly compressed
chunks, simplifies the operation and lowers the cost of running Loki.

All Loki log entries that pass through a Flow pipeline carry two pieces of
information: the log itself (the logged line and a timestamp), as well as the
labels associated with it.

Let's create an instance of a `loki.relabel` component and see how it acts on
some incoming log entries. We can see the component definition, as well as some
log entries coming from different files, environments, jobs, and on varying log
levels.

```river
loki.relabel "keep_error_only" {
  forward_to = [loki.write.onprem.receiver]

  rule {
    action        = "keep"
    source_labels = ["level"]
    regex         = "error"
  }
  rule {
    action        = "replace"
    source_labels = ["env", "job"]
    separator     = "/"
    target_label  = "instance"
  }
  rule {
    action = "labeldrop"
    regex  = "env|job"
  }
}
```

```
Log line     |    Labels
------------------------------
"entry1"     | "filename" = "/tmp/foo.txt", "env" = "dev",  "job" = "web", "level" = "info"
"error foo!" | "filename" = "/tmp/foo.txt", "env" = "dev",  "job" = "web", "level" = "error"
"entry2"     | "filename" = "/tmp/bar.txt", "env" = "dev",  "job" = "web", "level" = "info"
"entry3"     | "filename" = "/tmp/bar.txt", "env" = "dev",  "job" = "web", "level" = "debug"
"error bar!  | "filename" = "/tmp/bar.txt", "env" = "dev",  "job" = "sql", "level" = "error"
"error baz!  | "filename" = "/tmp/baz.txt", "env" = "prod", "job" = "sql", "level" = "error"
```

After applying the first relabeling rule, the component would only keep the entries whose level is 'error'.

```
Log line     |    Labels
------------------------------
"error foo!" | "filename" = "/tmp/foo.txt", "env" = "dev",  "job" = "web", "level" = "error"
"error bar!  | "filename" = "/tmp/bar.txt", "env" = "dev",  "job" = "sql", "level" = "error"
"error baz!  | "filename" = "/tmp/baz.txt", "env" = "prod", "job" = "sql", "level" = "error"
```

The second relabeling rule concatenates the values in the 'env' and 'jobs'
labels using a slash, and populates the instance label.

```
Log line     |    Labels
------------------------------
"error foo!" | "filename" = "/tmp/foo.txt", "env" = "dev",  "job" = "web", "level" = "error", "instance" = "dev/web"
"error bar!  | "filename" = "/tmp/bar.txt", "env" = "dev",  "job" = "sql", "level" = "error", "instance" = "dev/sql"
"error baz!  | "filename" = "/tmp/baz.txt", "env" = "prod", "job" = "sql", "level" = "error", "instance" = "prod/sql"
```

The third relabeling rule would drop the 'env' and 'job' labels, as they're no
longer needed.

```
Log line     |    Labels
------------------------------
"error foo!" | "filename" = "/tmp/foo.txt", "level" = "error", "instance" = "dev/web"
"error bar!  | "filename" = "/tmp/bar.txt", "level" = "error", "instance" = "dev/sql"
"error baz!  | "filename" = "/tmp/baz.txt", "level" = "error", "instance" = "prod/sql"
```

The log entries are processed and sent to the list of receivers in `forward_to`
in succession. This means that if a receiver is bottlenecked, it may impact the
rest.

Since relabeling can be a resource-intensive process, the component utilizes
an LRU cache to store the results of the relabling process on previously-seen
log label streams. The cache is purged whenever the component is updated with
a new relabeling configuration. You can use the component's debug metrics to
see how the cache is performing in terms of hits/misses and its current size.
