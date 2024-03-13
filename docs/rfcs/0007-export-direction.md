# Direction of Flow component exports

* Date: 2023-03-22
* Author: Paulin Tpdev (@ptodev)
* PR:
* Status: Draft

## Summary
It may be confusing that component exports can be either the component's input or its output. At the moment Flow doesn't enforce either model. People are free to fix and match them using their best judgement. This RFC evaluates the pros and cons of each approach, and also states if it'd be worth it to change our current approach.

## Outcome
The users would like to see the following:

- When reading a component's documentation, know what is an input and what is an output at a glance.
- Is should be easy to tell which component's output can be piped to which component's input (composition).
- Too much reliance on examples is a sign that the system is too hard to use. Browsing the documentation for a component's properties should be enough most of the time. 
- Learning new components should be easy and feel familiar. It should not feel like a new experience.

## Examples

### Example 1: Prometheus scrape and remote write
#### Example 1.1: How Flow works at the moment
```
prometheus.remote_write "staging" {
  endpoint {
    url = "http://mimir:9009/api/v1/push"
  }
}

prometheus.scrape "demo1" {
  targets = [{"__address__" = "127.0.0.1:12345"}]
  forward_to = [prometheus.remote_write.staging.receiver]
}

prometheus.scrape "demo2" {
  targets = [{"__address__" = "127.0.0.2:12345"}]
  forward_to = [prometheus.remote_write.staging.receiver]
}
```

#### Example 1.2: How Flow could work if we do not use exports as inputs
```
prometheus.remote_write "staging" {
  receiver = [prometheus.scrape.demo1, prometheus.scrape.demo2] 
  endpoint {
    url = "http://mimir:9009/api/v1/push"
  }
}

prometheus.scrape "demo1" {
  targets = [{"__address__" = "127.0.0.1:12345"}]
}

prometheus.scrape "demo2" {
  targets = [{"__address__" = "127.0.0.2:12345"}]
}
```

### Example 2: Prometheus scrape converted to Otel format and sent to an Otel endpoint
#### Example 2.1: How Flow works at the moment
```
prometheus.scrape "default" {
  targets = [{"__address__"   = "127.0.0.1:12345"}]
  forward_to = [otelcol.receiver.prometheus.default.receiver]
}

otelcol.receiver.prometheus "default" {
  output {
    metrics = [otelcol.exporter.otlp.default.input]
  }
}

otelcol.exporter.otlp "default" {
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```

#### Example 2.2: How Flow could work if we do not use exports as inputs
```
prometheus.scrape "default" {
    targets = [{"__address__"   = "127.0.0.1:12345"}]
}

otelcol.receiver.prometheus "default" {
  receiver = prometheus.scrape.output
}

otelcol.exporter.otlp "default" {
  // otelcol.receiver.prometheus.default.output will be of type 
  // "otelcol.Consumer" instead of an "output" block
  input = otelcol.receiver.prometheus.default.output
  client {
    endpoint = env("OTLP_ENDPOINT")
  }
}
```

## Potential approaches

### Approach 1 - Use exports as both inputs and outputs (current status quo)
With this approach the component's exports could be either its inputs or its outputs. The writer of each component uses their best judgement on whether having the exports as inputs or outputs would be most appropriate.

#### Pros
- Maximum flexibility.
- Prometheus uses a pull model. Otel uses a push model. Supporting both models in Flow could make it easier to express Prometheus and Otel logic. 
#### Cons
- Lack of consistency between components can be confusing.
- Having components in the same config file using different paradigms looks confusing.

### Approach 2 - Use exports as inputs only
With this approach the component's exports are its inputs. Each component will have to specify to which other components its output data is going to.

#### Pros

#### Cons
- It may be awkward to modify components which are already configured, so that they send data to more downstream components.

### Approach 3 - Use exports as outputs only
With this approach the component's exports are its outputs. Other components will feed those outputs to their inputs. Each component will have to specify from which other components its input data is coming from.

#### Pros
- More intuitive. Newcomers to Flow will browse the documentation and look for inputs and outputs of each component. They will likely assume that the non-exported arguments are the component's inputs, whereas the exported ones are the component's outputs.

#### Cons

## Other considerations

### Which approach would be less verbose to write?

Depending on the situation, either approach could be more verbose. For example:

- A single Prometheus remote write being referenced by multiple scrapers is less verbose when the remote write specifies its inputs in the remote write itself (as an input). This is example 1.2 above.
- A single scraper referencing multiple remote writes would be less verbose if the remote writes are specified in the scraper itself (as an output).

Hence, nether option can be considered less verbose in the general case.

### Ease of reading

Ideally, one should be able to read a River config file from top to bottom easily. In addition:
* It should be easy to see where a component's inputs are coming from.
* It should be easy to see where a component's outputs are going to.

It is not possible to achieve both in practice, just by reading a config file. That's because we only specify either where data came from, or where it's going to. It would be redundant to specify this in both the source component and the destination component. The only way to achieve both of those things is to use a UI tool for visualizing the graph.

Approach 1 makes tracing dependencies easy for some components like this one...
```
prometheus.scrape "default" {
  targets    = discovery.kubernetes.pods.targets
  forward_to = [prometheus.remote_write.default.receiver]
}
```

... but it makes it extra hard for other components such as this:
```
discovery.kubernetes "pods" {
  role = "pod"
}
```

It may have been easier to read the file if `discovery.kubernetes` listed it outputs like so:
```
discovery.kubernetes "pods" {
  role = "pod"
  targets = [prometheus.scrape.default]
}
```

It could also help readability if we advise users to put downstream components at the top, or at the bottom, depending on whether we go with approach 2 or 3.

### Could any approach lead to composition issues?

In River, it is possible to set any of a component's attributes via exports from another component. If the exports of component A are actually used to feed inputs to component A, then component B would not be able to use all exports from component A to feed a component B.

It would be much simpler if we could just say to the users that "every argument can be set via another component's exports - provided that the types match".

Approach 1 can make it harder for users to figure out how to compose components, which then means we have to provide examples for lots of potential use cases. It would be better if there is no need for excessive use of examples.

### Learning curve

The more consistent our approach, the easiest it is to explain in documentation. In order to link two components, at the moment sometimes users specify the next component in the "outputs" section, and sometimes they specify the previous component in the "inputs" section. This can be confusing. 

### Using components from a River file which cannot or should not be modified (River libraries?)

In this case, it is best for the upstream component not to specify all the downstream components which use its output. That way components which need data that the upstream exports could simply use it in any of the downstream's attributes.

### Preventing dangling references to shut down components

If a component shuts down, anything which references it should work properly afterwards.
For example, if a `prometheus.remote_write` shuts down, there should be no stale references to it.
A bugfix was already done for such a problem - see issue [#2216](https://github.com/grafana/agent/issues/2216)

### What works best with modules?

In the example below, the module input ("logs_output") is where the logs will be output to, whereas the module export ("logs_input") is where other modules/components would be inputting logs to this module.

```
argument "logs_output" { }

loki.process "filter" {
  forward_to = argument.logs_output.value
  
  stage.match {
    selector = "{job!=\"\"} |~ \"level=(debug|info)\""
    action   = "drop"
}

export "logs_input" {
  value = loki.process.filter.receiver
}
```

It would have been more intuitive to have the module inputs be the data flowing into the module, and the module exports be the data flowing out of the module.

### Would it be worth the effort to change what we do at the moment?

What would be the way to solve this problem with the minimum required effort? How to minimize changes to code and documentation?

## Conclusion
It may be easiest to go with approach 3. To help with config readability, we could recommend writing configs in which downstream components at the top of the config. For example, discovery components would be at the bottom and remote write components could be at the top. This is similar to example 2.2 That way it would be easier ot read a River file from top to bottom. Or, we could just recommend that people read the configs from the bottom if they want to trace the dependencies.