# FAQ

## Where do I find information on the supported values for the CustomResourceDefinitions?

Once you've [deployed the CustomResourceDefinitions](./getting-started.md#deploying-customresourcedefinitions)
to your Kubernetes cluster, use `kubectl explain <resource>` to get access to
the documentation for each resource. For example, `kubectl explain GrafanaAgent`
will describe the GrafanaAgent CRD, and `kubectl explain GrafanaAgent.spec` will
give you information on its spec field.
