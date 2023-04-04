# Create Prometheus Exporter Flow Components

This guide will walk you through the process of creating a new Prometheus exporter Flow component and best practices for implementing it. It is required that the exporter has an existing integration in order to wrap it as a Flow component.

## Arguments (Configuration)

`Arguments` struct defines the arguments that can be passed to the component. In most cases, this would be exactly the same as the arguments that the integration for this exporter the uses. Some recommendations:

- Use `attr` tag to define the name of the argument in the Flow component.
- Use `optional` tag for optional arguments.
- Use `rivertypes.Secret` type for sensitive arguments (e.g. API keys, passwords, etc). The original integration should have a similar field type called `Secret` from Prometheus.
- If one of the arguments is a slice, use `block`. For example, the [process_exporter](../../component/prometheus/exporter/process/process.go) `Arguments` struct has `ProcessExporter` param which is a `[]MatcherGroup`. The name of the parameter should be in singular. This will allow the user to define multiple blocks of the same type.

The river config would look like this using `matcher` block multiple times:

```river
prometheus.exporter.process "example" {
  track_children = false
  matcher {
    comm = ["grafana-agent"]
  }
  matcher {
    comm = ["firefox"]
  }  
}
```

- If one of the argument is a map, use `block` and define the key as `river:",label"` in the struct. For example, the [blackbox_exporter](../../component/prometheus/exporter/blackbox/blackbox.go) `BlackboxTarget` struct has a `Name` param which represents the name of the block. 

The river config would look like this:

```river
prometheus.exporter.blackbox "example" { 
	config_file = "blackbox_modules.yml"
	
	target "example" {
		address = "http://example.com"
		module  = "http_2xx"
	}
}
```

- Define `DefaultArguments` as a global variable to define the default arguments for the component. 

## Functions

- Define `UnmarshalRiver` function to unmarshal the arguments from the river config into the `Arguments` struct. Please, add a test to validate the unmarshalling covering as many cases as possible.

- Define a `Convert` function to convert nested structs to the ones that the integration uses. Please, also add a test to validate the conversion covering as many cases as possible.

- If the exporter follows the multi-target pattern, add a function to define Prometheus discovery targets and use `exporter.NewMultiTarget` for the `Build` param of the `component.Register` function. 

## Documentation

Writing the documentation for the component is very important. Please, follow the [Writing documentation for Flow components](./writing-flow-component-documentation.md) and take a look at the existing documentation for other exporters.