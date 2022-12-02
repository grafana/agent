# Multi-tenancy support in Flow

* Date: 2022-12-01
* Author: Paulin Todev (@todevpau)
* PR: [grafana/agent#2583](https://github.com/grafana/agent/pull/2583)
* Status: Draft

## Background

Grafana databases (e.g. Mimir/Loki/Tempo) support the notion of "tenancy". In GEx, a "tenant" is both a piece of metadata and an authentication/trust boundary. It can be used to manage rate limits, distribution of data across the cluster, limiting the impact of expensive queries, etc. 

In order for the Grafana databases to know the tenant of a series, they expect to see the tenant ID in the "X-Scope-OrgID" HTTP header of the remote write request. In GEx this is done via a [Gateway](https://grafana.com/docs/enterprise-traces/latest/gateway/) component which sits between the Agent and the database. The gateway assigns tenancy information based on the security credentials which the Agent used to connect to it.

The above works fine for GEx. However, there are also Grafana customers who use on-prem hosted databases. They are not able to do multi-tenancy for two reasons:
* The aforementioned "gateway" component is not open source
* The agent and the database could be communicating via a non-secure connection. Hence, even if a Gateway is used, deriving tenancy information from authentication details would not be an option.

## Problem statement

In order to support the needs of such on-prem customers we could build support for multi-tenancy within the Grafana Agent. It could retrieve the tenant ID from one of several different sources:
* The name of a Kubernetes pod.
* An environment variable.
* Directly specified in the Agent config (e.g. always using a certain tenant ID for a specific remote write component).

For this to work, somehow the Agent has to keep track of which metrics belong to which tenant, then attach the tenant information in the "X-Scope-OrgID" HTTP header when remote writing them to the DB.

## Potential solutions

### Solution 1 - Support multi-tenancy only in the Gateway instead of the Agent

The simplest solution would be to not support multi-tenancy in the Agent at all. Instead, the Enterprise Gateway component (or parts of it) could be open sourced, which could allow on-prem users to run it on their premises. The Gateway could also be expanded to assign tenancy information based on something other than security credentials.

#### Pros

* No changes to the Agent required.
* Clear separation of responsibility - only the Gateway manages tenant information.
* The system architecture of on-prem users would be the same as the enterprise one. Hence, it would be easier to understand and maintain.
* Good security. Allowing the Agent to set the tenant could mean that an Agent is configured to send data as the wrong tenant (either due to malicious activity or by mistake). Setting tenant information within the Gateway, a server-side component, would be more restricted and secure.

#### Cons

* It is extra work for customers to introduce and operate an extra component.
* Development work on the Gateway is required in order to open source it and make it more configurable.

### Solution 2 - A generic "sharding" configuration for the Agent remote write

The [```prometheus.remote_write```](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.remote_write/) Flow component could be expanded to be able to shard the writes into different remote writes under the hood. Each of these shards would have its own write-ahead log (WAL). Also, each shard would be able to remote write to a different endpoint address and with different HTTP header information.

#### Step 1 - Add a tenant label prior to the remote write

The “replace” action of the [```prometheus.relabel```](https://grafana.com/docs/agent/latest/flow/reference/components/prometheus.relabel/) or [```discovery.relabel```](https://grafana.com/docs/agent/latest/flow/reference/components/discovery.relabel/) components could be used. It would work by attempting to replace a label which doesn’t exist - then, the “relabel” component would create a new label. However, the disadvantage to this is that we could end up replacing a label which we assumed did not exist.

The label is internal, so could start with two underscores, e.g. ```__tenant__```.

What if we try to add a tenant label, but such a label already exists? How do we detect whether the label already exists? If it already exists, then replacing it might be the wrong thing to do. We should avoid sending bad data downstream. There are two options to deal with this:
* Drop these metrics and refuse to send them. However, in this case customers would lose data and be unhappy. We cannot go with this approach.
* Set an empty tenant ID in the HTTP header.
* Set the tenant ID in the HTTP header to a special reserved value, e.g. “conflict”.

There are various places where the label can be set:
* In discovery.relabel. Allows us to set the tenant ID to a kubernetes property, e.g. the name of a pod.
* In prometheus.relabel. That way we can set the tenant ID to either one of:
  * an environment variable
  * a string specified in the Agent config
  * the contents of another label
* The label might already exist on the scraped metric itself.
* The label might already be set when converting from an OpenTelemetry metric.
* In prometheus.remotewrite through an external_label field.

#### Step 2 - Remote write using the tenant HTTP header

1. Remove the tenant label (this is optional).
2. Set the tenant ID as an HTTP header and remote write the data to the database. Hence, data for different tenants would have to be sent on different HTTP requests.
3. We could achieve this by starting multiple remote write sub-components inside the main remote write component. Each sub-component would have its own WAL. 
4. When a sub-component queue is empty and it has not been given metrics to send for a certain amount of time, we could remove it from the pool of remote write sub-components.

#### New configuration for the remote-write Flow component

We would need to add a new “sharding” config block to the remote_write Flow component, with the following config options:
* policy: this could be either "none" or "one_shard_per_label_value" for now. 
In the future, we could add more policies such as grouping certain labels into one shard.
If the policy is “none”, then the behavior and performance would be identical to the pre-tenant implementation of the Agent. 
* label: This would be an option used by the "one_shard_per_label_value" sharding policy. However, it might not be used by other sharding policies.
* config_overrides: This is where we can set the http header to the value of this label. Ideally this should be generic enough so that we can override any part of the remote write configuration.

Example configuration:
```
prometheus.remote_write "staging" {
  // Send metrics to a locally running Mimir.
  endpoint {
    url = "http://mimir:9009/api/v1/push"

    http_client_config {
      basic_auth {
        username = "example-user"
        password = "example-password"
      }
    }
  }
  // Configure multi-tenancy
  sharding {
    policy = "one_shard_per_label_value"

    per_label_value_config_override {
      label = "__tenant__"
      config_overrides {
        endpoint {
          headers = {"X-Scope-OrgID" = "__shard_label_value__"}
        }
      }
    }
  }
}
```

Above, ```__tenant__``` is a label which could be created by relabeling another label like a kubernetes service discovery property.

#### Pros

* It is possible to configure “sharding” for use cases other than multi-tenancy.
* We don’t have to touch Prometheus code:
  * It will be fast to get the change out
  * Flow acts as a layer of abstraction on top of Prometheus and extends its features rather than competes with it

#### Cons

* Harder for users to understand and configure than a tenancy-specific configuration.
* It is not clear what is the best way to represent some things in configuration:
  * The HTTP header overrides.
  * The value for the HTTP header could be a label value (e.g. a ```__tenant__``` label used for sharding), but how to we represent this in River config? We do not know what the label value is at the time of creating the config. Only the label name.
* It’s a very generic feature and might complicate the code as the feature gets extended. We could try to manage this risk by having few and simple sharding policies.
* We need to come up with a cleanup policy for unnecessary shards. Potentially this should be configurable.

### Solution 3 - A non-generic, tenant-specific configuration for the Agent

Instead of a generic "sharding" configuration, we could just have a tenancy-specific configuration. The sharding will be done under the hood and the users of the Agent would not be aware of it. This would work in the same say as solution 2 with a "one_shard_per_label_value" sharding policy. From a user's point of view, this configuration could be similar to [Promtail's tenant stage](https://grafana.com/docs/loki/latest/clients/promtail/stages/tenant/).

Example configuration:
```
prometheus.remote_write "staging" {
  // Send metrics to a locally running Mimir.
  endpoint {
    url = "http://mimir:9009/api/v1/push"

    http_client_config {
      basic_auth {
        username = "example-user"
        password = "example-password"
      }
    }
  }
  // Configure multi-tenancy
  tenancy {
    # Name from labels to whose value should be set as tenant ID.
    label = "__tenant__"
  }
}
```

#### Pros

* Easier to create than a generic component.
* Easier to use for customers who just want to do multi-tenancy.
* It could be a good starting point for experimentation - even if solution 2 (a generic sharding feature) is our preferred one, we could start with solution 3 and then over time implement its internals as a special case of solution 2. This could give us time to see if users actually pick up the feature.

#### Cons

* It's a lot of work for a feature which would not be used in GEx.

### Solution 4 - Extend the Prometheus remote_write code to support this use case

Prometheus already supports remote writing to different hosts. However, it sends the same data to all hosts. For multi-tenancy we would need it to send different data on each host based on some property like a label value in the series.

Also, Prometheus already supports sharding the series into different remote writes. However, all of the shards go to the same remote endpoint. Also, at the time of this writing it is not possible to configure each shard to have its own http headers.

Refer to the Prometheus documentation [here](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write) and [here](https://prometheus.io/docs/practices/remote_write/) for more details.

#### Pros

* Grafana customers who use Prometheus instead of the Agent would be able to fulfill their multi-tenancy needs too.
* Little work required on the Agent side

#### Cons

* It would be slow to merge such a major feature into Prometheus

## Edge cases to keep in mind

### What if tenants get renamed?

* Mimir can have some limits, e.g. on the max number of tenants. It also supports a list of enabled/disabled tenants, so a renaming might result in data being dropped or handled differently.
* On the Agent side, it would be nice if we could handle a change of the tenant ID at runtime, without restarting?

These issues are generally not of concern to the Agent. However, it is good to mention them here just to remind ourselves of potential problems down the road.

### Security

* Is it safe to let clients configure their tenant ID in their Agent configuration?
* Could they not set it to the ID of another customer, and then end up pushing data which will be associated with this other customer? 
* If we don't use a Gateway, how would the backend DB know that a client is allowed to use a certain tenant ID?

## Author's recommendation

Firstly, it is important to understand what problem we are trying to solve:
1. Is this for customers who use unsecured connections and who think it's too much hassle to run a gateway? One of the main advantages of multi-tenancy is extra security due to the segregation of tenants. Is using unsecured connection even a good practice, and can we discourage it by not providing extensive support for it? Instead, a workaround such as a Gateway + dummy credentials or non-auth tenant IDs could help.
2. Or is this for customers who already use secure authentication? In this case, would they be willing to run a Gateway component similarly to GEx?

The author of the RFC has been advised by other Grafana developers that:
* Multi-tenancy in the Agent could be useful for all kinds of customers, including ones who use unsecured connections.
* Expanding the open source offerings are good for Grafana, so even if a feature is not used in GEx there is value in implementing it anyway.

Should the multi-tenancy feature be implemented, the author of this RFC would ideally prefer solution 1 - supporting multi-tenancy in the Gateway instead of the Agent. However, if this is not possible, solution 2 would be the next best option (albeit also a lot more work) - a generic "sharding" configuration for the Agent. Solution 3 could be a good one if we are not sure what a generic sharding configuration would look like. Solution 4 is not considered realistic by the author because it would take a lot of time and the benefits are not clear.