---
permalink: /unstable/monitoring/v1/probe/
---

# monitoring.v1.probe

Probe defines monitoring for a set of static targets or ingresses.

## Index

* [`fn new(name)`](#fn-new)
* [`obj metadata`](#obj-metadata)
  * [`fn withAnnotations(annotations)`](#fn-metadatawithannotations)
  * [`fn withAnnotationsMixin(annotations)`](#fn-metadatawithannotationsmixin)
  * [`fn withClusterName(clusterName)`](#fn-metadatawithclustername)
  * [`fn withCreationTimestamp(creationTimestamp)`](#fn-metadatawithcreationtimestamp)
  * [`fn withDeletionGracePeriodSeconds(deletionGracePeriodSeconds)`](#fn-metadatawithdeletiongraceperiodseconds)
  * [`fn withDeletionTimestamp(deletionTimestamp)`](#fn-metadatawithdeletiontimestamp)
  * [`fn withFinalizers(finalizers)`](#fn-metadatawithfinalizers)
  * [`fn withFinalizersMixin(finalizers)`](#fn-metadatawithfinalizersmixin)
  * [`fn withGenerateName(generateName)`](#fn-metadatawithgeneratename)
  * [`fn withGeneration(generation)`](#fn-metadatawithgeneration)
  * [`fn withLabels(labels)`](#fn-metadatawithlabels)
  * [`fn withLabelsMixin(labels)`](#fn-metadatawithlabelsmixin)
  * [`fn withManagedFields(managedFields)`](#fn-metadatawithmanagedfields)
  * [`fn withManagedFieldsMixin(managedFields)`](#fn-metadatawithmanagedfieldsmixin)
  * [`fn withName(name)`](#fn-metadatawithname)
  * [`fn withNamespace(namespace)`](#fn-metadatawithnamespace)
  * [`fn withOwnerReferences(ownerReferences)`](#fn-metadatawithownerreferences)
  * [`fn withOwnerReferencesMixin(ownerReferences)`](#fn-metadatawithownerreferencesmixin)
  * [`fn withResourceVersion(resourceVersion)`](#fn-metadatawithresourceversion)
  * [`fn withSelfLink(selfLink)`](#fn-metadatawithselflink)
  * [`fn withUid(uid)`](#fn-metadatawithuid)
* [`obj spec`](#obj-spec)
  * [`fn withInterval(interval)`](#fn-specwithinterval)
  * [`fn withJobName(jobName)`](#fn-specwithjobname)
  * [`fn withModule(module)`](#fn-specwithmodule)
  * [`fn withScrapeTimeout(scrapeTimeout)`](#fn-specwithscrapetimeout)
  * [`obj spec.basicAuth`](#obj-specbasicauth)
    * [`obj spec.basicAuth.password`](#obj-specbasicauthpassword)
      * [`fn withKey(key)`](#fn-specbasicauthpasswordwithkey)
      * [`fn withName(name)`](#fn-specbasicauthpasswordwithname)
      * [`fn withOptional(optional)`](#fn-specbasicauthpasswordwithoptional)
    * [`obj spec.basicAuth.username`](#obj-specbasicauthusername)
      * [`fn withKey(key)`](#fn-specbasicauthusernamewithkey)
      * [`fn withName(name)`](#fn-specbasicauthusernamewithname)
      * [`fn withOptional(optional)`](#fn-specbasicauthusernamewithoptional)
  * [`obj spec.bearerTokenSecret`](#obj-specbearertokensecret)
    * [`fn withKey(key)`](#fn-specbearertokensecretwithkey)
    * [`fn withName(name)`](#fn-specbearertokensecretwithname)
    * [`fn withOptional(optional)`](#fn-specbearertokensecretwithoptional)
  * [`obj spec.prober`](#obj-specprober)
    * [`fn withPath(path)`](#fn-specproberwithpath)
    * [`fn withScheme(scheme)`](#fn-specproberwithscheme)
    * [`fn withUrl(url)`](#fn-specproberwithurl)
  * [`obj spec.targets`](#obj-spectargets)
    * [`obj spec.targets.ingress`](#obj-spectargetsingress)
      * [`fn withRelabelingConfigs(relabelingConfigs)`](#fn-spectargetsingresswithrelabelingconfigs)
      * [`fn withRelabelingConfigsMixin(relabelingConfigs)`](#fn-spectargetsingresswithrelabelingconfigsmixin)
      * [`obj spec.targets.ingress.namespaceSelector`](#obj-spectargetsingressnamespaceselector)
        * [`fn withAny(any)`](#fn-spectargetsingressnamespaceselectorwithany)
        * [`fn withMatchNames(matchNames)`](#fn-spectargetsingressnamespaceselectorwithmatchnames)
        * [`fn withMatchNamesMixin(matchNames)`](#fn-spectargetsingressnamespaceselectorwithmatchnamesmixin)
      * [`obj spec.targets.ingress.selector`](#obj-spectargetsingressselector)
        * [`fn withMatchExpressions(matchExpressions)`](#fn-spectargetsingressselectorwithmatchexpressions)
        * [`fn withMatchExpressionsMixin(matchExpressions)`](#fn-spectargetsingressselectorwithmatchexpressionsmixin)
        * [`fn withMatchLabels(matchLabels)`](#fn-spectargetsingressselectorwithmatchlabels)
        * [`fn withMatchLabelsMixin(matchLabels)`](#fn-spectargetsingressselectorwithmatchlabelsmixin)
    * [`obj spec.targets.staticConfig`](#obj-spectargetsstaticconfig)
      * [`fn withLabels(labels)`](#fn-spectargetsstaticconfigwithlabels)
      * [`fn withLabelsMixin(labels)`](#fn-spectargetsstaticconfigwithlabelsmixin)
      * [`fn withRelabelingConfigs(relabelingConfigs)`](#fn-spectargetsstaticconfigwithrelabelingconfigs)
      * [`fn withRelabelingConfigsMixin(relabelingConfigs)`](#fn-spectargetsstaticconfigwithrelabelingconfigsmixin)
      * [`fn withStatic(static)`](#fn-spectargetsstaticconfigwithstatic)
      * [`fn withStaticMixin(static)`](#fn-spectargetsstaticconfigwithstaticmixin)
  * [`obj spec.tlsConfig`](#obj-spectlsconfig)
    * [`fn withInsecureSkipVerify(insecureSkipVerify)`](#fn-spectlsconfigwithinsecureskipverify)
    * [`fn withServerName(serverName)`](#fn-spectlsconfigwithservername)
    * [`obj spec.tlsConfig.ca`](#obj-spectlsconfigca)
      * [`obj spec.tlsConfig.ca.configMap`](#obj-spectlsconfigcaconfigmap)
        * [`fn withKey(key)`](#fn-spectlsconfigcaconfigmapwithkey)
        * [`fn withName(name)`](#fn-spectlsconfigcaconfigmapwithname)
        * [`fn withOptional(optional)`](#fn-spectlsconfigcaconfigmapwithoptional)
      * [`obj spec.tlsConfig.ca.secret`](#obj-spectlsconfigcasecret)
        * [`fn withKey(key)`](#fn-spectlsconfigcasecretwithkey)
        * [`fn withName(name)`](#fn-spectlsconfigcasecretwithname)
        * [`fn withOptional(optional)`](#fn-spectlsconfigcasecretwithoptional)
    * [`obj spec.tlsConfig.cert`](#obj-spectlsconfigcert)
      * [`obj spec.tlsConfig.cert.configMap`](#obj-spectlsconfigcertconfigmap)
        * [`fn withKey(key)`](#fn-spectlsconfigcertconfigmapwithkey)
        * [`fn withName(name)`](#fn-spectlsconfigcertconfigmapwithname)
        * [`fn withOptional(optional)`](#fn-spectlsconfigcertconfigmapwithoptional)
      * [`obj spec.tlsConfig.cert.secret`](#obj-spectlsconfigcertsecret)
        * [`fn withKey(key)`](#fn-spectlsconfigcertsecretwithkey)
        * [`fn withName(name)`](#fn-spectlsconfigcertsecretwithname)
        * [`fn withOptional(optional)`](#fn-spectlsconfigcertsecretwithoptional)
    * [`obj spec.tlsConfig.keySecret`](#obj-spectlsconfigkeysecret)
      * [`fn withKey(key)`](#fn-spectlsconfigkeysecretwithkey)
      * [`fn withName(name)`](#fn-spectlsconfigkeysecretwithname)
      * [`fn withOptional(optional)`](#fn-spectlsconfigkeysecretwithoptional)

## Fields

### fn new

```ts
new(name)
```

new returns an instance of Probe

## obj metadata

ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create.

### fn metadata.withAnnotations

```ts
withAnnotations(annotations)
```

Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations

### fn metadata.withAnnotationsMixin

```ts
withAnnotationsMixin(annotations)
```

Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations

**Note:** This function appends passed data to existing values

### fn metadata.withClusterName

```ts
withClusterName(clusterName)
```

The name of the cluster which the object belongs to. This is used to distinguish resources with same name and namespace in different clusters. This field is not set anywhere right now and apiserver is going to ignore it if set in create or update request.

### fn metadata.withCreationTimestamp

```ts
withCreationTimestamp(creationTimestamp)
```

Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers.

### fn metadata.withDeletionGracePeriodSeconds

```ts
withDeletionGracePeriodSeconds(deletionGracePeriodSeconds)
```

Number of seconds allowed for this object to gracefully terminate before it will be removed from the system. Only set when deletionTimestamp is also set. May only be shortened. Read-only.

### fn metadata.withDeletionTimestamp

```ts
withDeletionTimestamp(deletionTimestamp)
```

Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers.

### fn metadata.withFinalizers

```ts
withFinalizers(finalizers)
```

Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. Finalizers may be processed and removed in any order.  Order is NOT enforced because it introduces significant risk of stuck finalizers. finalizers is a shared field, any actor with permission can reorder it. If the finalizer list is processed in order, then this can lead to a situation in which the component responsible for the first finalizer in the list is waiting for a signal (field value, external system, or other) produced by a component responsible for a finalizer later in the list, resulting in a deadlock. Without enforced ordering finalizers are free to order amongst themselves and are not vulnerable to ordering changes in the list.

### fn metadata.withFinalizersMixin

```ts
withFinalizersMixin(finalizers)
```

Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. Finalizers may be processed and removed in any order.  Order is NOT enforced because it introduces significant risk of stuck finalizers. finalizers is a shared field, any actor with permission can reorder it. If the finalizer list is processed in order, then this can lead to a situation in which the component responsible for the first finalizer in the list is waiting for a signal (field value, external system, or other) produced by a component responsible for a finalizer later in the list, resulting in a deadlock. Without enforced ordering finalizers are free to order amongst themselves and are not vulnerable to ordering changes in the list.

**Note:** This function appends passed data to existing values

### fn metadata.withGenerateName

```ts
withGenerateName(generateName)
```

GenerateName is an optional prefix, used by the server, to generate a unique name ONLY IF the Name field has not been provided. If this field is used, the name returned to the client will be different than the name passed. This value will also be combined with a unique suffix. The provided value has the same validation rules as the Name field, and may be truncated by the length of the suffix required to make the value unique on the server.

If this field is specified and the generated name exists, the server will NOT return a 409 - instead, it will either return 201 Created or 500 with Reason ServerTimeout indicating a unique name could not be found in the time allotted, and the client should retry (optionally after the time indicated in the Retry-After header).

Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#idempotency

### fn metadata.withGeneration

```ts
withGeneration(generation)
```

A sequence number representing a specific generation of the desired state. Populated by the system. Read-only.

### fn metadata.withLabels

```ts
withLabels(labels)
```

Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels

### fn metadata.withLabelsMixin

```ts
withLabelsMixin(labels)
```

Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels

**Note:** This function appends passed data to existing values

### fn metadata.withManagedFields

```ts
withManagedFields(managedFields)
```

ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like 'ci-cd'. The set of fields is always in the version that the workflow used when modifying the object.

### fn metadata.withManagedFieldsMixin

```ts
withManagedFieldsMixin(managedFields)
```

ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like 'ci-cd'. The set of fields is always in the version that the workflow used when modifying the object.

**Note:** This function appends passed data to existing values

### fn metadata.withName

```ts
withName(name)
```

Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names

### fn metadata.withNamespace

```ts
withNamespace(namespace)
```

Namespace defines the space within which each name must be unique. An empty namespace is equivalent to the "default" namespace, but "default" is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty.

Must be a DNS_LABEL. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/namespaces

### fn metadata.withOwnerReferences

```ts
withOwnerReferences(ownerReferences)
```

List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller.

### fn metadata.withOwnerReferencesMixin

```ts
withOwnerReferencesMixin(ownerReferences)
```

List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller.

**Note:** This function appends passed data to existing values

### fn metadata.withResourceVersion

```ts
withResourceVersion(resourceVersion)
```

An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources.

Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency

### fn metadata.withSelfLink

```ts
withSelfLink(selfLink)
```

SelfLink is a URL representing this object. Populated by the system. Read-only.

DEPRECATED Kubernetes will stop propagating this field in 1.20 release and the field is planned to be removed in 1.21 release.

### fn metadata.withUid

```ts
withUid(uid)
```

UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations.

Populated by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids

## obj spec

Specification of desired Ingress selection for target discovery by Prometheus.

### fn spec.withInterval

```ts
withInterval(interval)
```

Interval at which targets are probed using the configured prober. If not specified Prometheus' global scrape interval is used.

### fn spec.withJobName

```ts
withJobName(jobName)
```

The job name assigned to scraped metrics by default.

### fn spec.withModule

```ts
withModule(module)
```

The module to use for probing specifying how to probe the target. Example module configuring in the blackbox exporter: https://github.com/prometheus/blackbox_exporter/blob/master/example.yml

### fn spec.withScrapeTimeout

```ts
withScrapeTimeout(scrapeTimeout)
```

Timeout for scraping metrics from the Prometheus exporter.

## obj spec.basicAuth



## obj spec.basicAuth.password



### fn spec.basicAuth.password.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.basicAuth.password.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.basicAuth.password.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.basicAuth.username



### fn spec.basicAuth.username.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.basicAuth.username.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.basicAuth.username.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.bearerTokenSecret

Secret to mount to read bearer token for scraping targets. The secret needs to be in the same namespace as the probe and accessible by the Prometheus Operator.

### fn spec.bearerTokenSecret.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.bearerTokenSecret.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.bearerTokenSecret.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.prober



### fn spec.prober.withPath

```ts
withPath(path)
```

Path to collect metrics from. Defaults to `/probe`.

### fn spec.prober.withScheme

```ts
withScheme(scheme)
```

HTTP scheme to use for scraping. Defaults to `http`.

### fn spec.prober.withUrl

```ts
withUrl(url)
```

Mandatory URL of the prober.

## obj spec.targets

Targets defines a set of static and/or dynamically discovered targets to be probed using the prober.

## obj spec.targets.ingress

Ingress defines the set of dynamically discovered ingress objects which hosts are considered for probing.

### fn spec.targets.ingress.withRelabelingConfigs

```ts
withRelabelingConfigs(relabelingConfigs)
```

RelabelConfigs to apply to samples before ingestion. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config

### fn spec.targets.ingress.withRelabelingConfigsMixin

```ts
withRelabelingConfigsMixin(relabelingConfigs)
```

RelabelConfigs to apply to samples before ingestion. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config

**Note:** This function appends passed data to existing values

## obj spec.targets.ingress.namespaceSelector



### fn spec.targets.ingress.namespaceSelector.withAny

```ts
withAny(any)
```

Boolean describing whether all namespaces are selected in contrast to a list restricting them.

### fn spec.targets.ingress.namespaceSelector.withMatchNames

```ts
withMatchNames(matchNames)
```

List of namespace names.

### fn spec.targets.ingress.namespaceSelector.withMatchNamesMixin

```ts
withMatchNamesMixin(matchNames)
```

List of namespace names.

**Note:** This function appends passed data to existing values

## obj spec.targets.ingress.selector



### fn spec.targets.ingress.selector.withMatchExpressions

```ts
withMatchExpressions(matchExpressions)
```

matchExpressions is a list of label selector requirements. The requirements are ANDed.

### fn spec.targets.ingress.selector.withMatchExpressionsMixin

```ts
withMatchExpressionsMixin(matchExpressions)
```

matchExpressions is a list of label selector requirements. The requirements are ANDed.

**Note:** This function appends passed data to existing values

### fn spec.targets.ingress.selector.withMatchLabels

```ts
withMatchLabels(matchLabels)
```

matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

### fn spec.targets.ingress.selector.withMatchLabelsMixin

```ts
withMatchLabelsMixin(matchLabels)
```

matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

**Note:** This function appends passed data to existing values

## obj spec.targets.staticConfig

StaticConfig defines static targets which are considers for probing. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#static_config.

### fn spec.targets.staticConfig.withLabels

```ts
withLabels(labels)
```

Labels assigned to all metrics scraped from the targets.

### fn spec.targets.staticConfig.withLabelsMixin

```ts
withLabelsMixin(labels)
```

Labels assigned to all metrics scraped from the targets.

**Note:** This function appends passed data to existing values

### fn spec.targets.staticConfig.withRelabelingConfigs

```ts
withRelabelingConfigs(relabelingConfigs)
```

RelabelConfigs to apply to samples before ingestion. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config

### fn spec.targets.staticConfig.withRelabelingConfigsMixin

```ts
withRelabelingConfigsMixin(relabelingConfigs)
```

RelabelConfigs to apply to samples before ingestion. More info: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#relabel_config

**Note:** This function appends passed data to existing values

### fn spec.targets.staticConfig.withStatic

```ts
withStatic(static)
```

Targets is a list of URLs to probe using the configured prober.

### fn spec.targets.staticConfig.withStaticMixin

```ts
withStaticMixin(static)
```

Targets is a list of URLs to probe using the configured prober.

**Note:** This function appends passed data to existing values

## obj spec.tlsConfig



### fn spec.tlsConfig.withInsecureSkipVerify

```ts
withInsecureSkipVerify(insecureSkipVerify)
```

Disable target certificate validation.

### fn spec.tlsConfig.withServerName

```ts
withServerName(serverName)
```

Used to verify the hostname for the targets.

## obj spec.tlsConfig.ca

Struct containing the CA cert to use for the targets.

## obj spec.tlsConfig.ca.configMap



### fn spec.tlsConfig.ca.configMap.withKey

```ts
withKey(key)
```

The key to select.

### fn spec.tlsConfig.ca.configMap.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.tlsConfig.ca.configMap.withOptional

```ts
withOptional(optional)
```

Specify whether the ConfigMap or its key must be defined

## obj spec.tlsConfig.ca.secret



### fn spec.tlsConfig.ca.secret.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.tlsConfig.ca.secret.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.tlsConfig.ca.secret.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.tlsConfig.cert



## obj spec.tlsConfig.cert.configMap



### fn spec.tlsConfig.cert.configMap.withKey

```ts
withKey(key)
```

The key to select.

### fn spec.tlsConfig.cert.configMap.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.tlsConfig.cert.configMap.withOptional

```ts
withOptional(optional)
```

Specify whether the ConfigMap or its key must be defined

## obj spec.tlsConfig.cert.secret



### fn spec.tlsConfig.cert.secret.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.tlsConfig.cert.secret.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.tlsConfig.cert.secret.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.tlsConfig.keySecret



### fn spec.tlsConfig.keySecret.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.tlsConfig.keySecret.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.tlsConfig.keySecret.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined