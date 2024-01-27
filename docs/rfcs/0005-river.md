# River: A Flow-optimized config language

* Date: 2022-06-27
* Author: Robert Fratto (@rfratto), Matt Durham (@mattdurham)
* PR: [grafana/agent#1839](https://github.com/grafana/agent/pull/1839)

## Summary

Grafana Agent developers have been working towards a feature called Grafana
Agent Flow ([RFC-0004][]), a component-based re-imagining of Grafana Agent
which compartmentalize the different configurable pieces of the agent, allowing
users to more easily understand and debug configuration issues. Grafana Agent
Flow was purposefully scoped broadly to allow for exploring many different
component-based approaches for prototyping the experimental feature.

The current implementation strategy focuses around expressions: settings for
components can be derived from expressions which can reference and mutate the
outputs of other components. Values can refer to arbitrary Go values like
interfaces or channels, enabling component developers to easily allow users to
construct data pipelines using Go APIs without requiring knowledge of the
underlying implementation.

The initial expressions prototype used [HCL][], which initially fit Flow's
needs during early prototyping. However, the growing dependency on passing
around arbitrary Go values to build pipelines started to conflict with the
limitations of HCL, making HCL increasingly insufficient for Flow's specific
use case.

We examined alternatives to HCL such as YAML, CUE, Jsonnet, Lua, and Go itself.
Eventually, we determined that the way we use arbitrary Go values in
expressions for constructing pipelines is a new use case warranting a
custom-built language.

This document proposes River, an HCL-inspired declarative expressions-based
language for continuous runtime evaluation. The decision to propose a new
language is not taken lightly, and is seen as the last resort. As such, much of
this proposal will focus on the rationale leading to this choice.

## Goals

* Minimize learning curve as much as possible to reduce friction
* Make it easy for developers to create Flow components which operate with
  arbitrary Go values (interfaces, channels, etc.)
* Expose error messages in an easily understandable and actionable way
* Natively support using Go values of any type in config expressions
* Natively support passing around and invoking real Go functions in config expressions

The language design will be scoped as small as possible, and new features will
only be added over time as they are determined to be strictly necessary for
Flow.

## Non-Goals

We are not aiming to create a general purpose configuration language. While it
would be possible for River to eventually be used in different contexts by
different projects, the primary goal today is specifically targeting Grafana
Agent Flow.

We will not provide a full specification for River here, only lightly
describing it to allow implementation details to change over time.

## Rationale

### Why an expression language? Why not YAML?

The entire rationale for creating a new language depends on the rationale that
expressions provide a useful amount of capabilities to users. Expressions
enable users to manipulate values to meet their own use cases in ways that
otherwise would require dedicated feature work, such as:

* Allowing users to merge metadata together from distinct sources when adding
  labels to metrics, such as merging labels from discovered Kubernetes
  namespaces with discovered Kubernetes pods.

* Allowing users to chain Prometheus service discoveries (e.g., feed the output
  of Kubernetes Service Discovery into HTTP Service Discovery)

* Allowing users to perform custom conditional logic, such as increasing rate
  limits during busier business months.

Without expressions, we would need more components for common tasks. A
`concat()` function call can be used to combine lists of discovered Prometheus
targets, but without expressions, there would likely need to be a dedicated
component for aggregating sets of targets together.

The belief is that the work required to use and maintain an expression language
is far less than the combined work to implement features that would be handled
by expressions out of the box.

YAML by itself does not support expressions. While expressions could be added
to YAML through the use of templates (e.g., `field_a: {{ some_variable + 5
}}`), it is beyond the scope of what YAML was intended for and would be more
cumbersome to use compared to a language where expressions are a first-class
concept.

### Why an embedded language?

We are using the term "embedded languages" to refer to languages typically
known for the ability for maintainers of the project to expose APIs to users of
the embedded language, such as the Lua API used by Neovim. Embedded languages
typically imply tight integration with the application embedding them, as
opposed to something like YAML which is a language consumed once at load time.

An embedded language is a good fit for Flow:

* It makes it easy for developers to expose APIs which users can interact with
  or pass around. These APIs can be opaque arbitrary Go types which the user
  doesn't need to know the detail of, only that it refers to something like a
  stream of metric samples.

* It is well-suited for continuous evaluation (i.e., the core feature of Flow)
  so configuration can adapt to a changing environment.

### Why a declarative language? Why not Lua?

The language Flow relies on should have a minimal learning curve. While a
language like Lua could likely be a decent fit for Flow, imperative languages
have steeper learning curves compared to declarative languages.

Declarative languages natively map to configuration files, since configuration
files are used to tell the application the desired state, reducing the learning
curve for the language and making it easier for users to reason about what the
final config state should be.

### Why not HCL?

> For some background, it's important to note that HCL can be considered two
> separate projects: `hashicorp/hcl` (the language and expression evaluator)
> and `zclconf/go-cty` (the value and type system used by HCL).

HCL was the obvious first choice for the Flow prototype: it supports
expressions, you can expose functions for users to call, and its syntax has a
small learning curve.

However, I found the [schema-driven processing][] API exposed by HCL to be
difficult to work with for Flow, requiring a lot of boilerplate. While there is
a library to interoperate with tagged Go structs, it was insufficient for
passing around arbitrary Go values, requiring me to [fork][gohcl] both
github.com/hashicorp/hcl/v2/gohcl and github.com/zclconf/go-cty/cty/gocty to
reduce boilerplate. While the fork lets us avoid the boilerplate of hand-crafting
schema definitions for components, it contains a non-trivial amount of changes
that would need to be contributed upstream to be tenable long-term.

Additionally, there is desired functionality that is not supported today in
HCL/go-cty:

1. A stronger focus on performance and memory usage, changing go-cty to operate
   around Go values instead of converting Go values to a custom representation.
   The performance gain will suit our needs for doing continuous evaluation of
   expressions.
2. Ability to disable go-cty's requirement that strings are UTF-8 encoded
3. Pass around functions as go-cty values (e.g., to allow a clustering
   component to expose a function to check for ownership of key against a hash
   ring)
4. Ability to declare local variables in a scope without needing a `locals`
   block like as seen in Terraform.

The combination of desired changes across gohcl and go-cty, the fork that was
already necessary to make it easier to adopt HCL for Flow, and the desire to
have a stronger interaction with arbitrary Go values led to the decision that a
new Flow-specific language was warranted.

### Why now?

Grafana Agent Flow is already a dramatic change to the Agent. To avoid users
being exhausted from the frequency of dramatic changes, it would be ideal for
Grafana Agent Flow to ship with River instead of eventually migrating to River.

## Minimizing impact

New languages always have some amount of learning curve, and if the learning
curve is too steep, the language will fail to be adopted.

We will minimize this impact of a new language by:

* Minimizing the learning curve as much as possible by not creating
  too many novel ideas at the language level.

* Tend the syntax towards allowing users to copy-and-paste examples to learn as
  they go.

* Heavily document the language so that all questions a user may have is
  answered.

* Ensuring that error messages explain the problem and the resolution is
  obvious.

## Proposal

River's syntax is inspired by HCL. However, some of the syntax will be changed
from HCL to make River more easily identifiable as a different language and
avoid a situations where users confuse the two.

River focuses on expressions, attributes, and blocks.

### Expressions

Expressions resolve to values used by River. The type of expressions are:

* Literal expressions:
  * Booleans: `true`, `false`
  * Numbers: `3`, `3.5`, `3e+10`, etc.
  * Strings: `"Hello, world!"`
* Unary operations:
  * Logical NOT: `!true`
  * Negative: `-5`
* Binary operations:
  * Math operators: `+`, `-`, `*`, `/`, `^` (pow)
  * Equality operators: `==`, `!=`, `<`, `<=`, `>`, `>=`
  * Logical operators: `||`, `&&`
* Lists: `[1, 2, 3]`
* Objects: `{ a = 5, b = 6 }`
* Variable reference: `foobar`
* Indexing: `some_list[0]`
* Field access: `some_object.field_a`
* Function calls: `concat([0, 1], [2, 3])`
* Parenthesized expression: `(3 + 5)`

### Attributes

Attributes are key-value pairs which set individual settings, formatted as
`<identifier> = <expression>`:

```
log_level  = "debug"
log_format = "logfmt"
```

### Blocks

Blocks are named groupings of attributes, wrapping in curly braces. Blocks can
also contain other blocks.

```
server {
  http_address = "127.0.0.1:12345"
}

prometheus.storage {
  remote_write {
    url = "http://localhost:9090/api/v1/write"
  }

  remote_write {
    url = "http://localhost:9091/api/v1/write"
  }
}
```

Block names must consist of one or more identifiers separated by `.`. Blocks
can also be given user-specified labels, denoted as a string wrapped in quotes:

```
prometheus.storage "primary" {
  // ...
}

prometheus.storage "secondary" {
  // ...
}
```

### Type system

Values are categorized as being one of the following:

* `number`
* `bool`
* `string`
* `list`
  * Elements within the list do not have to be the same type.
* `object`
* `function`
  * Function values differentiate River from HCL/go-cty, which does not support
    passing around or invoking function values.
* `capsule`
  * Capsule is a catch-all type which refers to some arbitrary Go value which
    is not one of the other types. For example, `<-chan int` would be
    represented as a capsule in River.

River types map to Go types as follows:

* `number`: Go `int*`, `uint*`, `float*`
* `bool`: Go `bool`
* `string`: Go `string`, `[]byte`
* `list`: Go `[]T`, `[...]T`.
* `object`: Go `map[string]T`, and structs
* `function`: Any Go function.
  * If the final return value of the Go function is an error, it will be
    checked on calling; a non-nil error will cause the evaluation of the
    function to fail.
* `capsule`: All other Go values.
  * Additionally, type which implements `interface { RiverCapsuleMarker() }`
    will also be treated as a capsule.

River acts like a combination of a configuration language like HCL and an
embedded language like Lua due to its focus on supporting all Go values,
including values which cannot be directly represented by the user (such as Go
interfaces). This enables developers to use native Go types for easily passing
around business logic which users wire together through their configuration.

### River struct tags

River struct tags are used to converting between River values and Go structs.
Tags take one of the following forms:

* `river:"example,attr"`: required attribute named `example`
* `river:"example,attr,optional"`: optional attribute named `example`
* `river:"example,block"`: required block named `example`
* `river:"example,block,optional"`: optional block named `example`
* `river:",label"`: Used for decoding block labels into a `string`.

Attribute and block names must be unique across the whole type. When encoding a
Go struct, inner blocks are converted into objects. Attributes are converted
into River values of the appropriate type.

Fields without struct tags are ignored.

### Errors

There are multiple types of errors which may occur:

* Lexing / parsing errors
* Evaluation errors (when evaluating an expression into a River value)
* Decoding errors (when converting a River value into a Go value)
* Validation errors (when Go code validates a value)

Errors should be displayed to the user in a way that gives as much information
as possible. Errors which involve unexpected values should print the value to
ease debugging.

For this `example.river` config file which expects the `targets` field to be a
list of objects:

```
prometheus.scrape "example1" {
  targets = 5
}

prometheus.scrape "example2" {
  targets = [5]
}

prometheus.scrape "example3" {
  targets = some_list_of_objects + 5
}
```

Errors could be shown to the user like:

```
example.river:2:3: targets expects list value, got number

  | targets = 5

  Value:
    5

example.river:6:3: list element 0 must be object, got number

  | targets = [5]

  Value:
    5

example.river:10:13: cannot perform `+` on types list and number

  | some_list_of_objects + 5

  Expression:
    [{}] + 5
```

The errors print out the offending portion of the config file alongside the
offending values. Printing out the offending values is useful when the values
come from the result of referring to a variable or calling a function.

### Concerns

No existing tooling for River will exist from day one. While the initial
implementation should include a formatter, tools like syntax highlighting or
LSPs won't exist and will need to be created over time.

## Alternatives considered

### Handles

Instead of passing around literal arbitrary Go values, handles could be used to
_refer_ to arbitrary Go values. For example, a number could refer to some entry
in an in-memory store which holds a Go channel or interface.

Pros:
* Works better with HCL in its current state without needing the gohcl fork
* Would enable YAML, CUE, and Jsonnet to pass around arbitrary values

Cons:
* Still wouldn't allow HCL to pass around functions as values
* More tedious for developers to work with (they now have to exchange handles
  for values).
* Developers will have to deal with extra logic for handling stale handles,
  whereas arbitrary Go values would continue to exist until they've been
  garbage collected.

[RFC-0004]: ./0004-agent-flow.md
[HCL]: https://github.com/hashicorp/hcl
[go-cty]: github.com/zclconf/go-cty
[gohcl]: https://github.com/rfratto/gohcl
[schema-driven processing]: https://github.com/hashicorp/hcl/blob/main/spec.md#schema-driven-processing
