---
aliases:
- /docs/agent/shared/flow/reference/components/rule-block/
canonical: https://grafana.com/docs/agent/latest/shared/flow/reference/components/rule-block-logs/
headless: true
---

The `rule` block contains the definition of any relabeling rules that can be
applied to an input log. If more than one `rule` block is defined, the
transformations are applied in top-down order.

The following arguments can be used to configure a `rule`. All arguments are
optional. Omitted fields take their default values.

Name | Type | Description | Default | Required
---- | ---- | ----------- | ------- | --------
`source_labels` | `list(string)` | The list of labels whose values are to be selected. Their content is concatenated using the `separator` and matched against `regex`. | | no
`separator`     | `string`       | The separator used to concatenate the values present in `source_labels`. | ; | no
`regex`         | `string`       | A valid RE2 expression with support for parenthesized capture groups. Used to match the extracted value from the combination of the `source_label` and `separator` fields or filter labels during the `labelkeep/labeldrop/labelmap` actions. | `(.*)` | no
`modulus`       | `uint`         | A positive integer used to calculate the modulus of the hashed source label values. | | no
`target_label`  | `string`       | Label to which the resulting value will be written to. | | no
`replacement`   | `string`       | The value against which a regex replace is performed, if the regex matches the extracted value. Supports previously captured groups. | $1 | no
`action`        | `string`       | The relabeling action to perform. | replace | no

Here's a list of the available actions, along with a brief description of their usage.

* `replace`   - Matches `regex` to the concatenated labels. If there's a match, it replaces the content of the `target_label` using the contents of the `replacement` field.
* `keep`      - Keeps logs where `regex` matches the string extracted using the `source_labels` and `separator`.
* `drop`      - Drops logs where `regex` matches the string extracted using the `source_labels` and `separator`.
* `hashmod`   - Hashes the concatenated labels, calculates its modulo `modulus` and writes the result to the `target_label`.
* `labelmap`  - Matches `regex` against all label names. Any labels that match are renamed according to the contents of the `replacement` field.
* `labeldrop` - Matches `regex` against all label names. Any labels that match are removed from the logs's label set.
* `labelkeep` - Matches `regex` against all label names. Any labels that don't match are removed from the logs's label set.
* `keepequal` - Drop targets for which the concatenated `source_labels` do not match `target_label`.
* `dropequal` - Drop targets for which the concatenated `source_labels` do match `target_label`.
* `lowercase` - Sets `target_label` to the lowercase form of the concatenated `source_labels`.
* `uppercase` - Sets `target_label` to the uppercase form of the concatenated `source_labels`.

Finally, note that the regex capture groups can be referred to using either the
`$CAPTURE_GROUP_NUMBER` or `${CAPTURE_GROUP_NUMBER}` notation.
