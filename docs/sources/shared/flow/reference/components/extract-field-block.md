---
aliases:
- /docs/agent/shared/flow/reference/components/extract-field-block/
- /docs/grafana-cloud/agent/shared/flow/reference/components/extract-field-block/
- /docs/grafana-cloud/monitor-infrastructure/agent/shared/flow/reference/components/extract-field-block/
- /docs/grafana-cloud/monitor-infrastructure/integrations/agent/shared/flow/reference/components/extract-field-block/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/extract-field-block/
description: Shared content, extract field block
headless: true
---

The following attributes are supported:

Name | Type           | Description                                                                                              | Default | Required
---- |----------------|----------------------------------------------------------------------------------------------------------|---------| --------
`tag_name` | `string` | The name of the resource attribute that will be added to logs, metrics, or spans.      | `""` | no
`key` | `string` | The annotation (or label) name. This must exactly match an annotation (or label) name.    |  `""` | no
`key_regex` | `string` | A regular expression used to extract a key that matches the regex.                           | `""` | no
`regex` | `string` | An optional field used to extract a sub-string from a complex field value.                      | `""` | no
`from` | `string` | The source of the labels or annotations. Allowed values are `pod` and `namespace`.          | `pod`    | no

When `tag_name` is not specified, a default tag name will be used with the format:
* `k8s.pod.annotations.<annotation key>`
* `k8s.pod.labels.<label key>`

For example, if `tag_name` is not specified and the key is `git_sha`, then the attribute name will be
`k8s.pod.annotations.git_sha`.

Either the `key` attribute or the `key_regex` attribute should be set, not both.
When `key_regex` is present, `tag_name` supports back reference to both
named capturing and positioned capturing.

For example, assume your pod spec contains the following labels:
* `app.kubernetes.io/component: mysql`
* `app.kubernetes.io/version: 5.7.21`

If you'd like to add tags for all labels with prefix `app.kubernetes.io/` and trim the prefix, 
then you can specify the following extraction rules:

```river
extract {
	label {
	    from = "pod"
		key_regex = "kubernetes.io/(.*)"
		tag_name  = "$1"
	}
}
```

These rules will add the `component` and `version` tags to the spans or metrics.

The `from` attribute can be set to either `"pod"` or `"namespace"`.
