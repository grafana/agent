With the change to modules we need the ability to load subgraphs. In essence every layer of flow becomes a subgraph.

## Concepts

### Subgraph

A subgraph consists of a full running flow system. The primary parts of a subgraph are:
* Delegate Component
* Metrics Handler
* Scheduler
* Loader
* Update Queue
* Child Subgraphs

####  Namespace ID

Namespace ID is the fully qualified ID, including the parent namespace ID and the current ID. For instance if I have

```river
module.string "mod1" {
   content = `
   prometheus.scrape "default" {
   }
   `
}

```

The ID of the inner scrape is `prometheus.scrape.default` but the namespace id is `module.string.mod1.prometheus.scrape.default`. This allows a fully unique name for every component. This impacts several items:

* The data path
* The metrics component_id
* The tracing trace_id
* The UI

When something is needed to be garaunteed to be unique then the namespace ID is the value to use.

### Delegate Component

This is the creator of the subgraph. Currently the supported parent delegates are:
* Flow
* module.string

Delegate Components own the subgraph lifecycle and can reload, stop and start the subgraph. Along with extracting information about the individual component nodes.

Delegate Component is an interface that exposes the full `ID` and the delimted `IDs`. These are used to build the name space.

### Metrics Handler

The metrics handler is an interface supported by the subgraph itself that allows a delegate component to instantiate a new subgraph. Whenever a component is created the current subgraph is passed via `component.options` that allows the delegate to call `LoadSubgraph`, when called this creates a new subgraph, adds that subgraph to the children of the current and then starts the subgraph.

### Scheduler

The scheduler is an existing flow component that handles the update cycle, this is scheduler is responsible for handling the update queue and then calling for a reevaluation of the dependencies. Each component is then ran from the scheduler via the `Synchronize` function.  The `Synchronize` function removes any components no longer needed, and creates any new components.

### Loader

The loader loads a given config parsing the AST and generating the components. The loader is also responsibile for building the dependency graph and applying the changes from an update via `EvaluateDependencies`, which then walks the tree and chaining more updates as needed.

### Update Queue

The update queue is a small object that enqueus and dequeues `OnStateChange` messages. These messages are passed to the loader

### Child Subgraphs

Each time the `LoadSubgraph` is successfully called a new subgraph is created with an entirely new queue, loader, vmscope, and scheduler. The components are loaded with the newly created subgraphas the `DelegateHandler` in this way a chain can be created. If a subgraph is changed or stopped then this triggers a propagation of those actions to all children.

The separated `vm.scope` ensures that local variables stay localized to the namespace they are in.

## Modules

Given the implementation of subgraphs the topic of modules can now be discussed.

Modules are defined as a subgraph loaded from a river configuration. Modules are unable to contain references to singleton components or configuration blocks. Modules instantiate a new subgraph and are marked as the owner of that subgraph. Modules can then control the lifecycle of the subgraph.

Modules communicate to their subgraph and the parent subgraph via `module.export` and `module.argument`.

### module.argument

The `module.argument` allows a value to be passed into a subgraph defined by a module. On loading the subgraph a list of all components is passed to the module that then iterates over the component list finding any arguments. If any are found they are registered and the parent value pushed down via the argument's `Update`. This then updates via `OnStateChange` This means that the only way a `module.argument Update`  can be called is via the parent module. This value then propagates as normal to the subgraph the `module.argument` lives in.

#### module.export

The `module.export` allows a value to be passed from the child subgraph to used in the parent subgraph. When a module loads the subgraph the module receives a list of all the components and iterates through these components. Any that are `module.export` are injected with a callback via `UpdateInform`. When the `Update` is called on `module.export` the export calls `inform` that pushes the value to the parent argument which then calls `OnStateChange`. This then propagates the value in the subgraph of the module.

## Diagrams

### Basic Layout

In this layout creating a simple scraper inside the module `module.string.mod1`, this entire config has two subgraphs. The `root` module created by flow, the ID of this root module is `""` , a blank string. And the `module.string.mod1` subgraph created by `module.string.mod1` whose parent subgraph is `root`. Both this modules have separate loaders, schedulers and queues. This allows them to operate indepedently.

```river

module.string.mod1 {
	content = `
	   module.argument "targets" {
	   }
	   prometheus.scraper.scrape1 {
	      targets = module.arguments.targets
	   } 
	`
}

```

```
┌────────────────────────────────────────────────────────────────────────────┐
│Flow System                                                                 │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │root subgraph            ┌──────────────────┐                        │  │
│   │- parent `flow`          │children          │                        │  │
│   │ ┌──────────────────┐    ├──────────────────┴─────────────────────┐  │  │
│   │ │controller.Queue  │    │module.string.mod1                      │  │  │
│   │ └──────────────────┘    │- parent `root subgraph`                │  │  │
│   │ ┌──────────────────┐    │ ┌──────────────────┐                   │  │  │
│   │ │controller.Loader │    │ │controller.Queue  │                   │  │  │
│   │ └──────────────────┘    │ └──────────────────┘                   │  │  │
│   │ ┌──────────────────┐    │ ┌──────────────────┐                   │  │  │
│   │ │controller.queue  │    │ │controller.Loader │                   │  │  │
│   │ └──────────────────┘    │ └──────────────────┘                   │  │  │
│   │ ┌─────────────────────┐ │ ┌──────────────────┐                   │  │  │
│   │ │Components           │ │ │controller.queue  │                   │  │  │
│   │ │- module.string.mod1 │ │ └──────────────────┘                   │  │  │
│   │ │                     │ │ ┌────────────────────────────────┐     │  │  │
│   │ │                     │ │ │components                      │     │  │  │
│   │ │                     │ │ │- prometheus.scrape.scrape1     │     │  │  │
│   │ │                     │ │ │- module.argument.targets       │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ │                     │ │ │                                │     │  │  │
│   │ └─────────────────────┘ │ │                                │     │  │  │
│   │                         │ └────────────────────────────────┘     │  │  │
│   │                         │                                        │  │  │
│   │                         │                                        │  │  │
│   │                         └────────────────────────────────────────┘  │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────────────────────────────┘
```

## Todo

* Unit Tests
* Documentation
* Reload
* Metrics
* Closing cleanly
* Stack limit
* Disabling singleton components
