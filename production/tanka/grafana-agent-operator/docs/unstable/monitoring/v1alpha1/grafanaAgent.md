---
permalink: /unstable/monitoring/v1alpha1/grafanaAgent/
---

# monitoring.v1alpha1.grafanaAgent

GrafanaAgent defines a Grafana Agent deployment.

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
  * [`fn withConfigMaps(configMaps)`](#fn-specwithconfigmaps)
  * [`fn withConfigMapsMixin(configMaps)`](#fn-specwithconfigmapsmixin)
  * [`fn withContainers(containers)`](#fn-specwithcontainers)
  * [`fn withContainersMixin(containers)`](#fn-specwithcontainersmixin)
  * [`fn withImage(image)`](#fn-specwithimage)
  * [`fn withImagePullSecrets(imagePullSecrets)`](#fn-specwithimagepullsecrets)
  * [`fn withImagePullSecretsMixin(imagePullSecrets)`](#fn-specwithimagepullsecretsmixin)
  * [`fn withInitContainers(initContainers)`](#fn-specwithinitcontainers)
  * [`fn withInitContainersMixin(initContainers)`](#fn-specwithinitcontainersmixin)
  * [`fn withLogFormat(logFormat)`](#fn-specwithlogformat)
  * [`fn withLogLevel(logLevel)`](#fn-specwithloglevel)
  * [`fn withNodeSelector(nodeSelector)`](#fn-specwithnodeselector)
  * [`fn withNodeSelectorMixin(nodeSelector)`](#fn-specwithnodeselectormixin)
  * [`fn withPaused(paused)`](#fn-specwithpaused)
  * [`fn withPortName(portName)`](#fn-specwithportname)
  * [`fn withPriorityClassName(priorityClassName)`](#fn-specwithpriorityclassname)
  * [`fn withSecrets(secrets)`](#fn-specwithsecrets)
  * [`fn withSecretsMixin(secrets)`](#fn-specwithsecretsmixin)
  * [`fn withServiceAccountName(serviceAccountName)`](#fn-specwithserviceaccountname)
  * [`fn withTolerations(tolerations)`](#fn-specwithtolerations)
  * [`fn withTolerationsMixin(tolerations)`](#fn-specwithtolerationsmixin)
  * [`fn withTopologySpreadConstraints(topologySpreadConstraints)`](#fn-specwithtopologyspreadconstraints)
  * [`fn withTopologySpreadConstraintsMixin(topologySpreadConstraints)`](#fn-specwithtopologyspreadconstraintsmixin)
  * [`fn withVersion(version)`](#fn-specwithversion)
  * [`fn withVolumeMounts(volumeMounts)`](#fn-specwithvolumemounts)
  * [`fn withVolumeMountsMixin(volumeMounts)`](#fn-specwithvolumemountsmixin)
  * [`fn withVolumes(volumes)`](#fn-specwithvolumes)
  * [`fn withVolumesMixin(volumes)`](#fn-specwithvolumesmixin)
  * [`obj spec.affinity`](#obj-specaffinity)
    * [`obj spec.affinity.nodeAffinity`](#obj-specaffinitynodeaffinity)
      * [`fn withPreferredDuringSchedulingIgnoredDuringExecution(preferredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitynodeaffinitywithpreferredduringschedulingignoredduringexecution)
      * [`fn withPreferredDuringSchedulingIgnoredDuringExecutionMixin(preferredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitynodeaffinitywithpreferredduringschedulingignoredduringexecutionmixin)
      * [`obj spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution`](#obj-specaffinitynodeaffinityrequiredduringschedulingignoredduringexecution)
        * [`fn withNodeSelectorTerms(nodeSelectorTerms)`](#fn-specaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionwithnodeselectorterms)
        * [`fn withNodeSelectorTermsMixin(nodeSelectorTerms)`](#fn-specaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionwithnodeselectortermsmixin)
    * [`obj spec.affinity.podAffinity`](#obj-specaffinitypodaffinity)
      * [`fn withPreferredDuringSchedulingIgnoredDuringExecution(preferredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitypodaffinitywithpreferredduringschedulingignoredduringexecution)
      * [`fn withPreferredDuringSchedulingIgnoredDuringExecutionMixin(preferredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitypodaffinitywithpreferredduringschedulingignoredduringexecutionmixin)
      * [`fn withRequiredDuringSchedulingIgnoredDuringExecution(requiredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitypodaffinitywithrequiredduringschedulingignoredduringexecution)
      * [`fn withRequiredDuringSchedulingIgnoredDuringExecutionMixin(requiredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitypodaffinitywithrequiredduringschedulingignoredduringexecutionmixin)
    * [`obj spec.affinity.podAntiAffinity`](#obj-specaffinitypodantiaffinity)
      * [`fn withPreferredDuringSchedulingIgnoredDuringExecution(preferredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitypodantiaffinitywithpreferredduringschedulingignoredduringexecution)
      * [`fn withPreferredDuringSchedulingIgnoredDuringExecutionMixin(preferredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitypodantiaffinitywithpreferredduringschedulingignoredduringexecutionmixin)
      * [`fn withRequiredDuringSchedulingIgnoredDuringExecution(requiredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitypodantiaffinitywithrequiredduringschedulingignoredduringexecution)
      * [`fn withRequiredDuringSchedulingIgnoredDuringExecutionMixin(requiredDuringSchedulingIgnoredDuringExecution)`](#fn-specaffinitypodantiaffinitywithrequiredduringschedulingignoredduringexecutionmixin)
  * [`obj spec.apiServer`](#obj-specapiserver)
    * [`fn withBearerToken(bearerToken)`](#fn-specapiserverwithbearertoken)
    * [`fn withBearerTokenFile(bearerTokenFile)`](#fn-specapiserverwithbearertokenfile)
    * [`fn withHost(host)`](#fn-specapiserverwithhost)
    * [`obj spec.apiServer.basicAuth`](#obj-specapiserverbasicauth)
      * [`obj spec.apiServer.basicAuth.password`](#obj-specapiserverbasicauthpassword)
        * [`fn withKey(key)`](#fn-specapiserverbasicauthpasswordwithkey)
        * [`fn withName(name)`](#fn-specapiserverbasicauthpasswordwithname)
        * [`fn withOptional(optional)`](#fn-specapiserverbasicauthpasswordwithoptional)
      * [`obj spec.apiServer.basicAuth.username`](#obj-specapiserverbasicauthusername)
        * [`fn withKey(key)`](#fn-specapiserverbasicauthusernamewithkey)
        * [`fn withName(name)`](#fn-specapiserverbasicauthusernamewithname)
        * [`fn withOptional(optional)`](#fn-specapiserverbasicauthusernamewithoptional)
    * [`obj spec.apiServer.tlsConfig`](#obj-specapiservertlsconfig)
      * [`fn withCaFile(caFile)`](#fn-specapiservertlsconfigwithcafile)
      * [`fn withCertFile(certFile)`](#fn-specapiservertlsconfigwithcertfile)
      * [`fn withInsecureSkipVerify(insecureSkipVerify)`](#fn-specapiservertlsconfigwithinsecureskipverify)
      * [`fn withKeyFile(keyFile)`](#fn-specapiservertlsconfigwithkeyfile)
      * [`fn withServerName(serverName)`](#fn-specapiservertlsconfigwithservername)
      * [`obj spec.apiServer.tlsConfig.ca`](#obj-specapiservertlsconfigca)
        * [`obj spec.apiServer.tlsConfig.ca.configMap`](#obj-specapiservertlsconfigcaconfigmap)
          * [`fn withKey(key)`](#fn-specapiservertlsconfigcaconfigmapwithkey)
          * [`fn withName(name)`](#fn-specapiservertlsconfigcaconfigmapwithname)
          * [`fn withOptional(optional)`](#fn-specapiservertlsconfigcaconfigmapwithoptional)
        * [`obj spec.apiServer.tlsConfig.ca.secret`](#obj-specapiservertlsconfigcasecret)
          * [`fn withKey(key)`](#fn-specapiservertlsconfigcasecretwithkey)
          * [`fn withName(name)`](#fn-specapiservertlsconfigcasecretwithname)
          * [`fn withOptional(optional)`](#fn-specapiservertlsconfigcasecretwithoptional)
      * [`obj spec.apiServer.tlsConfig.cert`](#obj-specapiservertlsconfigcert)
        * [`obj spec.apiServer.tlsConfig.cert.configMap`](#obj-specapiservertlsconfigcertconfigmap)
          * [`fn withKey(key)`](#fn-specapiservertlsconfigcertconfigmapwithkey)
          * [`fn withName(name)`](#fn-specapiservertlsconfigcertconfigmapwithname)
          * [`fn withOptional(optional)`](#fn-specapiservertlsconfigcertconfigmapwithoptional)
        * [`obj spec.apiServer.tlsConfig.cert.secret`](#obj-specapiservertlsconfigcertsecret)
          * [`fn withKey(key)`](#fn-specapiservertlsconfigcertsecretwithkey)
          * [`fn withName(name)`](#fn-specapiservertlsconfigcertsecretwithname)
          * [`fn withOptional(optional)`](#fn-specapiservertlsconfigcertsecretwithoptional)
      * [`obj spec.apiServer.tlsConfig.keySecret`](#obj-specapiservertlsconfigkeysecret)
        * [`fn withKey(key)`](#fn-specapiservertlsconfigkeysecretwithkey)
        * [`fn withName(name)`](#fn-specapiservertlsconfigkeysecretwithname)
        * [`fn withOptional(optional)`](#fn-specapiservertlsconfigkeysecretwithoptional)
  * [`obj spec.podMetadata`](#obj-specpodmetadata)
    * [`fn withAnnotations(annotations)`](#fn-specpodmetadatawithannotations)
    * [`fn withAnnotationsMixin(annotations)`](#fn-specpodmetadatawithannotationsmixin)
    * [`fn withLabels(labels)`](#fn-specpodmetadatawithlabels)
    * [`fn withLabelsMixin(labels)`](#fn-specpodmetadatawithlabelsmixin)
    * [`fn withName(name)`](#fn-specpodmetadatawithname)
  * [`obj spec.prometheus`](#obj-specprometheus)
    * [`fn withEnforcedNamespaceLabel(enforcedNamespaceLabel)`](#fn-specprometheuswithenforcednamespacelabel)
    * [`fn withEnforcedSampleLimit(enforcedSampleLimit)`](#fn-specprometheuswithenforcedsamplelimit)
    * [`fn withEnforcedTargetLimit(enforcedTargetLimit)`](#fn-specprometheuswithenforcedtargetlimit)
    * [`fn withExternalLabels(externalLabels)`](#fn-specprometheuswithexternallabels)
    * [`fn withExternalLabelsMixin(externalLabels)`](#fn-specprometheuswithexternallabelsmixin)
    * [`fn withIgnoreNamespaceSelectors(ignoreNamespaceSelectors)`](#fn-specprometheuswithignorenamespaceselectors)
    * [`fn withOverrideHonorLabels(overrideHonorLabels)`](#fn-specprometheuswithoverridehonorlabels)
    * [`fn withOverrideHonorTimestamps(overrideHonorTimestamps)`](#fn-specprometheuswithoverridehonortimestamps)
    * [`fn withPrometheusExternalLabelName(prometheusExternalLabelName)`](#fn-specprometheuswithprometheusexternallabelname)
    * [`fn withRemoteWrite(remoteWrite)`](#fn-specprometheuswithremotewrite)
    * [`fn withRemoteWriteMixin(remoteWrite)`](#fn-specprometheuswithremotewritemixin)
    * [`fn withReplicaExternalLabelName(replicaExternalLabelName)`](#fn-specprometheuswithreplicaexternallabelname)
    * [`fn withReplicas(replicas)`](#fn-specprometheuswithreplicas)
    * [`fn withScrapeInterval(scrapeInterval)`](#fn-specprometheuswithscrapeinterval)
    * [`fn withScrapeTimeout(scrapeTimeout)`](#fn-specprometheuswithscrapetimeout)
    * [`fn withShards(shards)`](#fn-specprometheuswithshards)
    * [`obj spec.prometheus.arbitraryFSAccessThroughSMs`](#obj-specprometheusarbitraryfsaccessthroughsms)
      * [`fn withDeny(deny)`](#fn-specprometheusarbitraryfsaccessthroughsmswithdeny)
    * [`obj spec.prometheus.instanceNamespaceSelector`](#obj-specprometheusinstancenamespaceselector)
      * [`fn withMatchExpressions(matchExpressions)`](#fn-specprometheusinstancenamespaceselectorwithmatchexpressions)
      * [`fn withMatchExpressionsMixin(matchExpressions)`](#fn-specprometheusinstancenamespaceselectorwithmatchexpressionsmixin)
      * [`fn withMatchLabels(matchLabels)`](#fn-specprometheusinstancenamespaceselectorwithmatchlabels)
      * [`fn withMatchLabelsMixin(matchLabels)`](#fn-specprometheusinstancenamespaceselectorwithmatchlabelsmixin)
    * [`obj spec.prometheus.instanceSelector`](#obj-specprometheusinstanceselector)
      * [`fn withMatchExpressions(matchExpressions)`](#fn-specprometheusinstanceselectorwithmatchexpressions)
      * [`fn withMatchExpressionsMixin(matchExpressions)`](#fn-specprometheusinstanceselectorwithmatchexpressionsmixin)
      * [`fn withMatchLabels(matchLabels)`](#fn-specprometheusinstanceselectorwithmatchlabels)
      * [`fn withMatchLabelsMixin(matchLabels)`](#fn-specprometheusinstanceselectorwithmatchlabelsmixin)
  * [`obj spec.resources`](#obj-specresources)
    * [`fn withLimits(limits)`](#fn-specresourceswithlimits)
    * [`fn withLimitsMixin(limits)`](#fn-specresourceswithlimitsmixin)
    * [`fn withRequests(requests)`](#fn-specresourceswithrequests)
    * [`fn withRequestsMixin(requests)`](#fn-specresourceswithrequestsmixin)
  * [`obj spec.securityContext`](#obj-specsecuritycontext)
    * [`fn withFsGroup(fsGroup)`](#fn-specsecuritycontextwithfsgroup)
    * [`fn withFsGroupChangePolicy(fsGroupChangePolicy)`](#fn-specsecuritycontextwithfsgroupchangepolicy)
    * [`fn withRunAsGroup(runAsGroup)`](#fn-specsecuritycontextwithrunasgroup)
    * [`fn withRunAsNonRoot(runAsNonRoot)`](#fn-specsecuritycontextwithrunasnonroot)
    * [`fn withRunAsUser(runAsUser)`](#fn-specsecuritycontextwithrunasuser)
    * [`fn withSupplementalGroups(supplementalGroups)`](#fn-specsecuritycontextwithsupplementalgroups)
    * [`fn withSupplementalGroupsMixin(supplementalGroups)`](#fn-specsecuritycontextwithsupplementalgroupsmixin)
    * [`fn withSysctls(sysctls)`](#fn-specsecuritycontextwithsysctls)
    * [`fn withSysctlsMixin(sysctls)`](#fn-specsecuritycontextwithsysctlsmixin)
    * [`obj spec.securityContext.seLinuxOptions`](#obj-specsecuritycontextselinuxoptions)
      * [`fn withLevel(level)`](#fn-specsecuritycontextselinuxoptionswithlevel)
      * [`fn withRole(role)`](#fn-specsecuritycontextselinuxoptionswithrole)
      * [`fn withType(type)`](#fn-specsecuritycontextselinuxoptionswithtype)
      * [`fn withUser(user)`](#fn-specsecuritycontextselinuxoptionswithuser)
    * [`obj spec.securityContext.seccompProfile`](#obj-specsecuritycontextseccompprofile)
      * [`fn withLocalhostProfile(localhostProfile)`](#fn-specsecuritycontextseccompprofilewithlocalhostprofile)
      * [`fn withType(type)`](#fn-specsecuritycontextseccompprofilewithtype)
    * [`obj spec.securityContext.windowsOptions`](#obj-specsecuritycontextwindowsoptions)
      * [`fn withGmsaCredentialSpec(gmsaCredentialSpec)`](#fn-specsecuritycontextwindowsoptionswithgmsacredentialspec)
      * [`fn withGmsaCredentialSpecName(gmsaCredentialSpecName)`](#fn-specsecuritycontextwindowsoptionswithgmsacredentialspecname)
      * [`fn withRunAsUserName(runAsUserName)`](#fn-specsecuritycontextwindowsoptionswithrunasusername)
  * [`obj spec.storage`](#obj-specstorage)
    * [`fn withDisableMountSubPath(disableMountSubPath)`](#fn-specstoragewithdisablemountsubpath)
    * [`obj spec.storage.emptyDir`](#obj-specstorageemptydir)
      * [`fn withMedium(medium)`](#fn-specstorageemptydirwithmedium)
      * [`fn withSizeLimit(sizeLimit)`](#fn-specstorageemptydirwithsizelimit)
    * [`obj spec.storage.volumeClaimTemplate`](#obj-specstoragevolumeclaimtemplate)
      * [`fn withKind(kind)`](#fn-specstoragevolumeclaimtemplatewithkind)
      * [`obj spec.storage.volumeClaimTemplate.metadata`](#obj-specstoragevolumeclaimtemplatemetadata)
        * [`fn withAnnotations(annotations)`](#fn-specstoragevolumeclaimtemplatemetadatawithannotations)
        * [`fn withAnnotationsMixin(annotations)`](#fn-specstoragevolumeclaimtemplatemetadatawithannotationsmixin)
        * [`fn withLabels(labels)`](#fn-specstoragevolumeclaimtemplatemetadatawithlabels)
        * [`fn withLabelsMixin(labels)`](#fn-specstoragevolumeclaimtemplatemetadatawithlabelsmixin)
        * [`fn withName(name)`](#fn-specstoragevolumeclaimtemplatemetadatawithname)
      * [`obj spec.storage.volumeClaimTemplate.spec`](#obj-specstoragevolumeclaimtemplatespec)
        * [`fn withAccessModes(accessModes)`](#fn-specstoragevolumeclaimtemplatespecwithaccessmodes)
        * [`fn withAccessModesMixin(accessModes)`](#fn-specstoragevolumeclaimtemplatespecwithaccessmodesmixin)
        * [`fn withStorageClassName(storageClassName)`](#fn-specstoragevolumeclaimtemplatespecwithstorageclassname)
        * [`fn withVolumeMode(volumeMode)`](#fn-specstoragevolumeclaimtemplatespecwithvolumemode)
        * [`fn withVolumeName(volumeName)`](#fn-specstoragevolumeclaimtemplatespecwithvolumename)
        * [`obj spec.storage.volumeClaimTemplate.spec.dataSource`](#obj-specstoragevolumeclaimtemplatespecdatasource)
          * [`fn withApiGroup(apiGroup)`](#fn-specstoragevolumeclaimtemplatespecdatasourcewithapigroup)
          * [`fn withKind(kind)`](#fn-specstoragevolumeclaimtemplatespecdatasourcewithkind)
          * [`fn withName(name)`](#fn-specstoragevolumeclaimtemplatespecdatasourcewithname)
        * [`obj spec.storage.volumeClaimTemplate.spec.resources`](#obj-specstoragevolumeclaimtemplatespecresources)
          * [`fn withLimits(limits)`](#fn-specstoragevolumeclaimtemplatespecresourceswithlimits)
          * [`fn withLimitsMixin(limits)`](#fn-specstoragevolumeclaimtemplatespecresourceswithlimitsmixin)
          * [`fn withRequests(requests)`](#fn-specstoragevolumeclaimtemplatespecresourceswithrequests)
          * [`fn withRequestsMixin(requests)`](#fn-specstoragevolumeclaimtemplatespecresourceswithrequestsmixin)
        * [`obj spec.storage.volumeClaimTemplate.spec.selector`](#obj-specstoragevolumeclaimtemplatespecselector)
          * [`fn withMatchExpressions(matchExpressions)`](#fn-specstoragevolumeclaimtemplatespecselectorwithmatchexpressions)
          * [`fn withMatchExpressionsMixin(matchExpressions)`](#fn-specstoragevolumeclaimtemplatespecselectorwithmatchexpressionsmixin)
          * [`fn withMatchLabels(matchLabels)`](#fn-specstoragevolumeclaimtemplatespecselectorwithmatchlabels)
          * [`fn withMatchLabelsMixin(matchLabels)`](#fn-specstoragevolumeclaimtemplatespecselectorwithmatchlabelsmixin)

## Fields

### fn new

```ts
new(name)
```

new returns an instance of Grafanaagent

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

Spec holds the specification of the desired behavior for the Grafana Agent cluster.

### fn spec.withConfigMaps

```ts
withConfigMaps(configMaps)
```

ConfigMaps is a liset of config maps in the same namespace as the GrafanaAgent object which will be mounted into each running Grafana Agent pod. The secrets are mounted into /etc/grafana-agent/configmaps/<configmap-name>.

### fn spec.withConfigMapsMixin

```ts
withConfigMapsMixin(configMaps)
```

ConfigMaps is a liset of config maps in the same namespace as the GrafanaAgent object which will be mounted into each running Grafana Agent pod. The secrets are mounted into /etc/grafana-agent/configmaps/<configmap-name>.

**Note:** This function appends passed data to existing values

### fn spec.withContainers

```ts
withContainers(containers)
```

Containers allows injecting additional containers or modifying operator generated containers. This can be used to allow adding an authentication proxy to a Grafana Agent pod or to change the behavior of an operator-generated container. Containers described here modify an operator generated container if they share the same name and modifications are done via a strategic merge patch. The current container names are: `grafana-agent` and `config-reloader`. Overriding containers is entirely outside the scope of what the Grafana Agent team will support and by doing so, you accept that this behavior may break at any time without notice.

### fn spec.withContainersMixin

```ts
withContainersMixin(containers)
```

Containers allows injecting additional containers or modifying operator generated containers. This can be used to allow adding an authentication proxy to a Grafana Agent pod or to change the behavior of an operator-generated container. Containers described here modify an operator generated container if they share the same name and modifications are done via a strategic merge patch. The current container names are: `grafana-agent` and `config-reloader`. Overriding containers is entirely outside the scope of what the Grafana Agent team will support and by doing so, you accept that this behavior may break at any time without notice.

**Note:** This function appends passed data to existing values

### fn spec.withImage

```ts
withImage(image)
```

Image, when specified, overrides the image used to run the Agent. It should be specified along with a tag. Version must still be set to ensure the Grafana Agent Operator knows which version of Grafana Agent is being configured.

### fn spec.withImagePullSecrets

```ts
withImagePullSecrets(imagePullSecrets)
```

ImagePullSecrets holds an optional list of references to secrets within the same namespace to use for pulling the Grafana Agent image from registries. More info: https://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod

### fn spec.withImagePullSecretsMixin

```ts
withImagePullSecretsMixin(imagePullSecrets)
```

ImagePullSecrets holds an optional list of references to secrets within the same namespace to use for pulling the Grafana Agent image from registries. More info: https://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod

**Note:** This function appends passed data to existing values

### fn spec.withInitContainers

```ts
withInitContainers(initContainers)
```

InitContainers allows adding initContainers to the pod definition. These can be used to, for example, fetch secrets for injection into the Grafana Agent configuration from external sources. Any errors during the execution of an initContainer will lead to a restart of the pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ Using initContainers for any use case other than secret fetching is entirely outside the scope of what the Grafana Agent maintainers will support and by doing so, you accept that this behavior may break at any time without notice.

### fn spec.withInitContainersMixin

```ts
withInitContainersMixin(initContainers)
```

InitContainers allows adding initContainers to the pod definition. These can be used to, for example, fetch secrets for injection into the Grafana Agent configuration from external sources. Any errors during the execution of an initContainer will lead to a restart of the pod. More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/ Using initContainers for any use case other than secret fetching is entirely outside the scope of what the Grafana Agent maintainers will support and by doing so, you accept that this behavior may break at any time without notice.

**Note:** This function appends passed data to existing values

### fn spec.withLogFormat

```ts
withLogFormat(logFormat)
```

LogFormat controls the logging format of the generated pods. Defaults to "logfmt" if not set.

### fn spec.withLogLevel

```ts
withLogLevel(logLevel)
```

LogLevel controls the log level of the generated pods. Defaults to "info" if not set.

### fn spec.withNodeSelector

```ts
withNodeSelector(nodeSelector)
```

NodeSelector defines which nodes pods should be scheduling on.

### fn spec.withNodeSelectorMixin

```ts
withNodeSelectorMixin(nodeSelector)
```

NodeSelector defines which nodes pods should be scheduling on.

**Note:** This function appends passed data to existing values

### fn spec.withPaused

```ts
withPaused(paused)
```

Paused prevents actions except for deletion to be performed on the underlying managed objects.

### fn spec.withPortName

```ts
withPortName(portName)
```

Port name used for the pods and governing service. This defaults to agent-metrics.

### fn spec.withPriorityClassName

```ts
withPriorityClassName(priorityClassName)
```

PriorityClassName is the priority class assigned to pods.

### fn spec.withSecrets

```ts
withSecrets(secrets)
```

Secrets is a list of secrets in the same namespace as the GrafanaAgent object which will be mounted into each running Grafana Agent pod. The secrets are mounted into /etc/grafana-agent/secrets/<secret-name>.

### fn spec.withSecretsMixin

```ts
withSecretsMixin(secrets)
```

Secrets is a list of secrets in the same namespace as the GrafanaAgent object which will be mounted into each running Grafana Agent pod. The secrets are mounted into /etc/grafana-agent/secrets/<secret-name>.

**Note:** This function appends passed data to existing values

### fn spec.withServiceAccountName

```ts
withServiceAccountName(serviceAccountName)
```

ServiceAccountName is the name of the ServiceAccount to use for running Grafana Agent pods.

### fn spec.withTolerations

```ts
withTolerations(tolerations)
```

Tolerations, if specified, controls the pod's tolerations.

### fn spec.withTolerationsMixin

```ts
withTolerationsMixin(tolerations)
```

Tolerations, if specified, controls the pod's tolerations.

**Note:** This function appends passed data to existing values

### fn spec.withTopologySpreadConstraints

```ts
withTopologySpreadConstraints(topologySpreadConstraints)
```

TopologySpreadConstraints, if specified, controls the pod's topology spread constraints.

### fn spec.withTopologySpreadConstraintsMixin

```ts
withTopologySpreadConstraintsMixin(topologySpreadConstraints)
```

TopologySpreadConstraints, if specified, controls the pod's topology spread constraints.

**Note:** This function appends passed data to existing values

### fn spec.withVersion

```ts
withVersion(version)
```

Version of Grafana Agent to be deployed.

### fn spec.withVolumeMounts

```ts
withVolumeMounts(volumeMounts)
```

VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition. VolumEMounts specified will be appended to other VolumeMounts in the Grafana Agent container that are generated as a result of StorageSpec objects.

### fn spec.withVolumeMountsMixin

```ts
withVolumeMountsMixin(volumeMounts)
```

VolumeMounts allows configuration of additional VolumeMounts on the output StatefulSet definition. VolumEMounts specified will be appended to other VolumeMounts in the Grafana Agent container that are generated as a result of StorageSpec objects.

**Note:** This function appends passed data to existing values

### fn spec.withVolumes

```ts
withVolumes(volumes)
```

Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will be appended to other volumes that are generated as a result of StorageSpec objects.

### fn spec.withVolumesMixin

```ts
withVolumesMixin(volumes)
```

Volumes allows configuration of additional volumes on the output StatefulSet definition. Volumes specified will be appended to other volumes that are generated as a result of StorageSpec objects.

**Note:** This function appends passed data to existing values

## obj spec.affinity

Affinity, if specified, controls pod scheduling constraints.

## obj spec.affinity.nodeAffinity

Describes node affinity scheduling rules for the pod.

### fn spec.affinity.nodeAffinity.withPreferredDuringSchedulingIgnoredDuringExecution

```ts
withPreferredDuringSchedulingIgnoredDuringExecution(preferredDuringSchedulingIgnoredDuringExecution)
```

The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.

### fn spec.affinity.nodeAffinity.withPreferredDuringSchedulingIgnoredDuringExecutionMixin

```ts
withPreferredDuringSchedulingIgnoredDuringExecutionMixin(preferredDuringSchedulingIgnoredDuringExecution)
```

The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node matches the corresponding matchExpressions; the node(s) with the highest sum are the most preferred.

**Note:** This function appends passed data to existing values

## obj spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution

If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to an update), the system may or may not try to eventually evict the pod from its node.

### fn spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.withNodeSelectorTerms

```ts
withNodeSelectorTerms(nodeSelectorTerms)
```

Required. A list of node selector terms. The terms are ORed.

### fn spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.withNodeSelectorTermsMixin

```ts
withNodeSelectorTermsMixin(nodeSelectorTerms)
```

Required. A list of node selector terms. The terms are ORed.

**Note:** This function appends passed data to existing values

## obj spec.affinity.podAffinity

Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)).

### fn spec.affinity.podAffinity.withPreferredDuringSchedulingIgnoredDuringExecution

```ts
withPreferredDuringSchedulingIgnoredDuringExecution(preferredDuringSchedulingIgnoredDuringExecution)
```

The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.

### fn spec.affinity.podAffinity.withPreferredDuringSchedulingIgnoredDuringExecutionMixin

```ts
withPreferredDuringSchedulingIgnoredDuringExecutionMixin(preferredDuringSchedulingIgnoredDuringExecution)
```

The scheduler will prefer to schedule pods to nodes that satisfy the affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.

**Note:** This function appends passed data to existing values

### fn spec.affinity.podAffinity.withRequiredDuringSchedulingIgnoredDuringExecution

```ts
withRequiredDuringSchedulingIgnoredDuringExecution(requiredDuringSchedulingIgnoredDuringExecution)
```

If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.

### fn spec.affinity.podAffinity.withRequiredDuringSchedulingIgnoredDuringExecutionMixin

```ts
withRequiredDuringSchedulingIgnoredDuringExecutionMixin(requiredDuringSchedulingIgnoredDuringExecution)
```

If the affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.

**Note:** This function appends passed data to existing values

## obj spec.affinity.podAntiAffinity



### fn spec.affinity.podAntiAffinity.withPreferredDuringSchedulingIgnoredDuringExecution

```ts
withPreferredDuringSchedulingIgnoredDuringExecution(preferredDuringSchedulingIgnoredDuringExecution)
```

The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.

### fn spec.affinity.podAntiAffinity.withPreferredDuringSchedulingIgnoredDuringExecutionMixin

```ts
withPreferredDuringSchedulingIgnoredDuringExecutionMixin(preferredDuringSchedulingIgnoredDuringExecution)
```

The scheduler will prefer to schedule pods to nodes that satisfy the anti-affinity expressions specified by this field, but it may choose a node that violates one or more of the expressions. The node that is most preferred is the one with the greatest sum of weights, i.e. for each node that meets all of the scheduling requirements (resource request, requiredDuringScheduling anti-affinity expressions, etc.), compute a sum by iterating through the elements of this field and adding "weight" to the sum if the node has pods which matches the corresponding podAffinityTerm; the node(s) with the highest sum are the most preferred.

**Note:** This function appends passed data to existing values

### fn spec.affinity.podAntiAffinity.withRequiredDuringSchedulingIgnoredDuringExecution

```ts
withRequiredDuringSchedulingIgnoredDuringExecution(requiredDuringSchedulingIgnoredDuringExecution)
```

If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.

### fn spec.affinity.podAntiAffinity.withRequiredDuringSchedulingIgnoredDuringExecutionMixin

```ts
withRequiredDuringSchedulingIgnoredDuringExecutionMixin(requiredDuringSchedulingIgnoredDuringExecution)
```

If the anti-affinity requirements specified by this field are not met at scheduling time, the pod will not be scheduled onto the node. If the anti-affinity requirements specified by this field cease to be met at some point during pod execution (e.g. due to a pod label update), the system may or may not try to eventually evict the pod from its node. When there are multiple elements, the lists of nodes corresponding to each podAffinityTerm are intersected, i.e. all terms must be satisfied.

**Note:** This function appends passed data to existing values

## obj spec.apiServer



### fn spec.apiServer.withBearerToken

```ts
withBearerToken(bearerToken)
```

Bearer token for accessing apiserver.

### fn spec.apiServer.withBearerTokenFile

```ts
withBearerTokenFile(bearerTokenFile)
```

File to read bearer token for accessing apiserver.

### fn spec.apiServer.withHost

```ts
withHost(host)
```

Host of apiserver. A valid string consisting of a hostname or IP followed by an optional port number

## obj spec.apiServer.basicAuth



## obj spec.apiServer.basicAuth.password

The secret in the service monitor namespace that contains the password for authentication.

### fn spec.apiServer.basicAuth.password.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.apiServer.basicAuth.password.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.apiServer.basicAuth.password.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.apiServer.basicAuth.username



### fn spec.apiServer.basicAuth.username.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.apiServer.basicAuth.username.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.apiServer.basicAuth.username.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.apiServer.tlsConfig

TLS Config to use for accessing apiserver.

### fn spec.apiServer.tlsConfig.withCaFile

```ts
withCaFile(caFile)
```

Path to the CA cert in the Prometheus container to use for the targets.

### fn spec.apiServer.tlsConfig.withCertFile

```ts
withCertFile(certFile)
```

Path to the client cert file in the Prometheus container for the targets.

### fn spec.apiServer.tlsConfig.withInsecureSkipVerify

```ts
withInsecureSkipVerify(insecureSkipVerify)
```

Disable target certificate validation.

### fn spec.apiServer.tlsConfig.withKeyFile

```ts
withKeyFile(keyFile)
```

Path to the client key file in the Prometheus container for the targets.

### fn spec.apiServer.tlsConfig.withServerName

```ts
withServerName(serverName)
```

Used to verify the hostname for the targets.

## obj spec.apiServer.tlsConfig.ca

Struct containing the CA cert to use for the targets.

## obj spec.apiServer.tlsConfig.ca.configMap



### fn spec.apiServer.tlsConfig.ca.configMap.withKey

```ts
withKey(key)
```

The key to select.

### fn spec.apiServer.tlsConfig.ca.configMap.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.apiServer.tlsConfig.ca.configMap.withOptional

```ts
withOptional(optional)
```

Specify whether the ConfigMap or its key must be defined

## obj spec.apiServer.tlsConfig.ca.secret



### fn spec.apiServer.tlsConfig.ca.secret.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.apiServer.tlsConfig.ca.secret.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.apiServer.tlsConfig.ca.secret.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.apiServer.tlsConfig.cert

Struct containing the client cert file for the targets.

## obj spec.apiServer.tlsConfig.cert.configMap



### fn spec.apiServer.tlsConfig.cert.configMap.withKey

```ts
withKey(key)
```

The key to select.

### fn spec.apiServer.tlsConfig.cert.configMap.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.apiServer.tlsConfig.cert.configMap.withOptional

```ts
withOptional(optional)
```

Specify whether the ConfigMap or its key must be defined

## obj spec.apiServer.tlsConfig.cert.secret



### fn spec.apiServer.tlsConfig.cert.secret.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.apiServer.tlsConfig.cert.secret.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.apiServer.tlsConfig.cert.secret.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.apiServer.tlsConfig.keySecret

Secret containing the client key file for the targets.

### fn spec.apiServer.tlsConfig.keySecret.withKey

```ts
withKey(key)
```

The key of the secret to select from.  Must be a valid secret key.

### fn spec.apiServer.tlsConfig.keySecret.withName

```ts
withName(name)
```

Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?

### fn spec.apiServer.tlsConfig.keySecret.withOptional

```ts
withOptional(optional)
```

Specify whether the Secret or its key must be defined

## obj spec.podMetadata

PodMetadata configures Labels and Annotations which are propagated to created Grafana Agent pods.

### fn spec.podMetadata.withAnnotations

```ts
withAnnotations(annotations)
```

Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations

### fn spec.podMetadata.withAnnotationsMixin

```ts
withAnnotationsMixin(annotations)
```

Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations

**Note:** This function appends passed data to existing values

### fn spec.podMetadata.withLabels

```ts
withLabels(labels)
```

Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels

### fn spec.podMetadata.withLabelsMixin

```ts
withLabelsMixin(labels)
```

Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels

**Note:** This function appends passed data to existing values

### fn spec.podMetadata.withName

```ts
withName(name)
```

Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names

## obj spec.prometheus



### fn spec.prometheus.withEnforcedNamespaceLabel

```ts
withEnforcedNamespaceLabel(enforcedNamespaceLabel)
```

EnforcedNamepsaceLabel enforces adding a namespace label of origin for each metric that is user-created. The label value will always be the namespace of the object that is being created.

### fn spec.prometheus.withEnforcedSampleLimit

```ts
withEnforcedSampleLimit(enforcedSampleLimit)
```

EnforcedSampleLimit defines global limit on the number of scraped samples that will be accepted. This overrides any SampleLimit set per ServiceMonitor and/or PodMonitor. It is meant to be used by admins to enforce the SampleLimit to keep the overall number of samples and series under the desired limit. Note that if a SampleLimit from a ServiceMonitor or PodMonitor is lower, that value will be used instead.

### fn spec.prometheus.withEnforcedTargetLimit

```ts
withEnforcedTargetLimit(enforcedTargetLimit)
```

EnforcedTargetLimit defines a global limit on the number of scraped targets. This overrides any TargetLimit set per ServiceMonitor and/or PodMonitor. It is meant to be used by admins to enforce the TargetLimit to keep the overall number of targets under the desired limit. Note that if a TargetLimit from a ServiceMonitor or PodMonitor is higher, that value will be used instead.

### fn spec.prometheus.withExternalLabels

```ts
withExternalLabels(externalLabels)
```

ExternalLabels are labels to add to any time series when sending data over remote_write.

### fn spec.prometheus.withExternalLabelsMixin

```ts
withExternalLabelsMixin(externalLabels)
```

ExternalLabels are labels to add to any time series when sending data over remote_write.

**Note:** This function appends passed data to existing values

### fn spec.prometheus.withIgnoreNamespaceSelectors

```ts
withIgnoreNamespaceSelectors(ignoreNamespaceSelectors)
```

IgnoreNamespaceSelectors, if true, will ignore NamespaceSelector settings from the PodMonitor and ServiceMonitor configs, and they will only discover endpoints within their current namespace.

### fn spec.prometheus.withOverrideHonorLabels

```ts
withOverrideHonorLabels(overrideHonorLabels)
```

OverrideHonorLabels, if true, overrides all configured honor_labels read from ServiceMonitor or PodMonitor to false.

### fn spec.prometheus.withOverrideHonorTimestamps

```ts
withOverrideHonorTimestamps(overrideHonorTimestamps)
```

OverrideHonorTimestamps allows to globally enforce honoring timestamps in all scrape configs.

### fn spec.prometheus.withPrometheusExternalLabelName

```ts
withPrometheusExternalLabelName(prometheusExternalLabelName)
```

PrometheusExternalLabelName is the name of the external label used to denote Grafana Agent cluster. Defaults to "cluster." External label will _not_ be added when value is set to the empty string.

### fn spec.prometheus.withRemoteWrite

```ts
withRemoteWrite(remoteWrite)
```

RemoteWrite controls default remote_write settings for all instances. If an instance does not provide its own remoteWrite settings, these will be used instead.

### fn spec.prometheus.withRemoteWriteMixin

```ts
withRemoteWriteMixin(remoteWrite)
```

RemoteWrite controls default remote_write settings for all instances. If an instance does not provide its own remoteWrite settings, these will be used instead.

**Note:** This function appends passed data to existing values

### fn spec.prometheus.withReplicaExternalLabelName

```ts
withReplicaExternalLabelName(replicaExternalLabelName)
```

ReplicaExternalLabelName is the name of the Prometheus external label used to denote replica name. Defaults to __replica__. External label will _not_ be added when value is set to the empty string.

### fn spec.prometheus.withReplicas

```ts
withReplicas(replicas)
```

Replicas of each shard to deploy for metrics pods. Number of replicas multiplied by the number of shards is the total number of pods created.

### fn spec.prometheus.withScrapeInterval

```ts
withScrapeInterval(scrapeInterval)
```

ScrapeInterval is the time between consecutive scrapes.

### fn spec.prometheus.withScrapeTimeout

```ts
withScrapeTimeout(scrapeTimeout)
```

ScrapeTimeout is the time to wait for a target to respond before marking a scrape as failed.

### fn spec.prometheus.withShards

```ts
withShards(shards)
```

Shards to distribute targets onto. Number of replicas multiplied by the number of shards is the total number of pods created. Note that scaling down shards will not reshard data onto remaining instances, it must be manually moved. Increasing shards will not reshard data either but it will continue to be available from the same instances. Sharding is performed on the content of the __address__ target meta-label.

## obj spec.prometheus.arbitraryFSAccessThroughSMs

ArbitraryFSAccessThroughSMs configures whether configuration based on a ServiceMonitor can access arbitrary files on the file system of the Grafana Agent container e.g. bearer token files.

### fn spec.prometheus.arbitraryFSAccessThroughSMs.withDeny

```ts
withDeny(deny)
```



## obj spec.prometheus.instanceNamespaceSelector

InstanceNamespaceSelector are the set of labels to determine which namespaces to watch for PrometheusInstances. If not provided, only checks own namespace.

### fn spec.prometheus.instanceNamespaceSelector.withMatchExpressions

```ts
withMatchExpressions(matchExpressions)
```

matchExpressions is a list of label selector requirements. The requirements are ANDed.

### fn spec.prometheus.instanceNamespaceSelector.withMatchExpressionsMixin

```ts
withMatchExpressionsMixin(matchExpressions)
```

matchExpressions is a list of label selector requirements. The requirements are ANDed.

**Note:** This function appends passed data to existing values

### fn spec.prometheus.instanceNamespaceSelector.withMatchLabels

```ts
withMatchLabels(matchLabels)
```

matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

### fn spec.prometheus.instanceNamespaceSelector.withMatchLabelsMixin

```ts
withMatchLabelsMixin(matchLabels)
```

matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

**Note:** This function appends passed data to existing values

## obj spec.prometheus.instanceSelector

InstanceSelector determines which PrometheusInstances should be selected for running. Each instance runs its own set of Prometheus components, including service discovery, scraping, and remote_write.

### fn spec.prometheus.instanceSelector.withMatchExpressions

```ts
withMatchExpressions(matchExpressions)
```

matchExpressions is a list of label selector requirements. The requirements are ANDed.

### fn spec.prometheus.instanceSelector.withMatchExpressionsMixin

```ts
withMatchExpressionsMixin(matchExpressions)
```

matchExpressions is a list of label selector requirements. The requirements are ANDed.

**Note:** This function appends passed data to existing values

### fn spec.prometheus.instanceSelector.withMatchLabels

```ts
withMatchLabels(matchLabels)
```

matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

### fn spec.prometheus.instanceSelector.withMatchLabelsMixin

```ts
withMatchLabelsMixin(matchLabels)
```

matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

**Note:** This function appends passed data to existing values

## obj spec.resources

Resources holds requests and limits for individual pods.

### fn spec.resources.withLimits

```ts
withLimits(limits)
```

Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

### fn spec.resources.withLimitsMixin

```ts
withLimitsMixin(limits)
```

Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

**Note:** This function appends passed data to existing values

### fn spec.resources.withRequests

```ts
withRequests(requests)
```

Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

### fn spec.resources.withRequestsMixin

```ts
withRequestsMixin(requests)
```

Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

**Note:** This function appends passed data to existing values

## obj spec.securityContext

SecurityContext holds pod-level security attributes and common container settings. When unspecified, defaults to the default PodSecurityContext.

### fn spec.securityContext.withFsGroup

```ts
withFsGroup(fsGroup)
```

A special supplemental group that applies to all containers in a pod. Some volume types allow the Kubelet to change the ownership of that volume to be owned by the pod: 
 1. The owning GID will be the FSGroup 2. The setgid bit is set (new files created in the volume will be owned by FSGroup) 3. The permission bits are OR'd with rw-rw---- 
 If unset, the Kubelet will not modify the ownership and permissions of any volume.

### fn spec.securityContext.withFsGroupChangePolicy

```ts
withFsGroupChangePolicy(fsGroupChangePolicy)
```

fsGroupChangePolicy defines behavior of changing ownership and permission of the volume before being exposed inside Pod. This field will only apply to volume types which support fsGroup based ownership(and permissions). It will have no effect on ephemeral volume types such as: secret, configmaps and emptydir. Valid values are "OnRootMismatch" and "Always". If not specified defaults to "Always".

### fn spec.securityContext.withRunAsGroup

```ts
withRunAsGroup(runAsGroup)
```

The GID to run the entrypoint of the container process. Uses runtime default if unset. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.

### fn spec.securityContext.withRunAsNonRoot

```ts
withRunAsNonRoot(runAsNonRoot)
```

Indicates that the container must run as a non-root user. If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. If unset or false, no such validation will be performed. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.

### fn spec.securityContext.withRunAsUser

```ts
withRunAsUser(runAsUser)
```

The UID to run the entrypoint of the container process. Defaults to user specified in image metadata if unspecified. May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.

### fn spec.securityContext.withSupplementalGroups

```ts
withSupplementalGroups(supplementalGroups)
```

A list of groups applied to the first process run in each container, in addition to the container's primary GID.  If unspecified, no groups will be added to any container.

### fn spec.securityContext.withSupplementalGroupsMixin

```ts
withSupplementalGroupsMixin(supplementalGroups)
```

A list of groups applied to the first process run in each container, in addition to the container's primary GID.  If unspecified, no groups will be added to any container.

**Note:** This function appends passed data to existing values

### fn spec.securityContext.withSysctls

```ts
withSysctls(sysctls)
```

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported sysctls (by the container runtime) might fail to launch.

### fn spec.securityContext.withSysctlsMixin

```ts
withSysctlsMixin(sysctls)
```

Sysctls hold a list of namespaced sysctls used for the pod. Pods with unsupported sysctls (by the container runtime) might fail to launch.

**Note:** This function appends passed data to existing values

## obj spec.securityContext.seLinuxOptions

The SELinux context to be applied to all containers. If unspecified, the container runtime will allocate a random SELinux context for each container.  May also be set in SecurityContext.  If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence for that container.

### fn spec.securityContext.seLinuxOptions.withLevel

```ts
withLevel(level)
```

Level is SELinux level label that applies to the container.

### fn spec.securityContext.seLinuxOptions.withRole

```ts
withRole(role)
```

Role is a SELinux role label that applies to the container.

### fn spec.securityContext.seLinuxOptions.withType

```ts
withType(type)
```

Type is a SELinux type label that applies to the container.

### fn spec.securityContext.seLinuxOptions.withUser

```ts
withUser(user)
```

User is a SELinux user label that applies to the container.

## obj spec.securityContext.seccompProfile



### fn spec.securityContext.seccompProfile.withLocalhostProfile

```ts
withLocalhostProfile(localhostProfile)
```

localhostProfile indicates a profile defined in a file on the node should be used. The profile must be preconfigured on the node to work. Must be a descending path, relative to the kubelet's configured seccomp profile location. Must only be set if type is 'Localhost'.

### fn spec.securityContext.seccompProfile.withType

```ts
withType(type)
```

type indicates which kind of seccomp profile will be applied. Valid options are: 
 Localhost - a profile defined in a file on the node should be used. RuntimeDefault - the container runtime default profile should be used. Unconfined - no profile should be applied.

## obj spec.securityContext.windowsOptions



### fn spec.securityContext.windowsOptions.withGmsaCredentialSpec

```ts
withGmsaCredentialSpec(gmsaCredentialSpec)
```

GMSACredentialSpec is where the GMSA admission webhook (https://github.com/kubernetes-sigs/windows-gmsa) inlines the contents of the GMSA credential spec named by the GMSACredentialSpecName field.

### fn spec.securityContext.windowsOptions.withGmsaCredentialSpecName

```ts
withGmsaCredentialSpecName(gmsaCredentialSpecName)
```

GMSACredentialSpecName is the name of the GMSA credential spec to use.

### fn spec.securityContext.windowsOptions.withRunAsUserName

```ts
withRunAsUserName(runAsUserName)
```

The UserName in Windows to run the entrypoint of the container process. Defaults to the user specified in image metadata if unspecified. May also be set in PodSecurityContext. If set in both SecurityContext and PodSecurityContext, the value specified in SecurityContext takes precedence.

## obj spec.storage

Storage spec to specify how storage will be used.

### fn spec.storage.withDisableMountSubPath

```ts
withDisableMountSubPath(disableMountSubPath)
```

Deprecated: subPath usage will be disabled by default in a future release, this option will become unnecessary. DisableMountSubPath allows to remove any subPath usage in volume mounts.

## obj spec.storage.emptyDir

EmptyDirVolumeSource to be used by the Prometheus StatefulSets. If specified, used in place of any volumeClaimTemplate. More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir

### fn spec.storage.emptyDir.withMedium

```ts
withMedium(medium)
```

What type of storage medium should back this directory. The default is '' which means to use the node's default medium. Must be an empty string (default) or Memory. More info: https://kubernetes.io/docs/concepts/storage/volumes#emptydir

### fn spec.storage.emptyDir.withSizeLimit

```ts
withSizeLimit(sizeLimit)
```

Total amount of local storage required for this EmptyDir volume. The size limit is also applicable for memory medium. The maximum usage on memory medium EmptyDir would be the minimum value between the SizeLimit specified here and the sum of memory limits of all containers in a pod. The default is nil which means that the limit is undefined. More info: http://kubernetes.io/docs/user-guide/volumes#emptydir

## obj spec.storage.volumeClaimTemplate

A PVC spec to be used by the Prometheus StatefulSets.

### fn spec.storage.volumeClaimTemplate.withKind

```ts
withKind(kind)
```

Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds

## obj spec.storage.volumeClaimTemplate.metadata



### fn spec.storage.volumeClaimTemplate.metadata.withAnnotations

```ts
withAnnotations(annotations)
```

Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations

### fn spec.storage.volumeClaimTemplate.metadata.withAnnotationsMixin

```ts
withAnnotationsMixin(annotations)
```

Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations

**Note:** This function appends passed data to existing values

### fn spec.storage.volumeClaimTemplate.metadata.withLabels

```ts
withLabels(labels)
```

Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels

### fn spec.storage.volumeClaimTemplate.metadata.withLabelsMixin

```ts
withLabelsMixin(labels)
```

Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels

**Note:** This function appends passed data to existing values

### fn spec.storage.volumeClaimTemplate.metadata.withName

```ts
withName(name)
```

Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names

## obj spec.storage.volumeClaimTemplate.spec



### fn spec.storage.volumeClaimTemplate.spec.withAccessModes

```ts
withAccessModes(accessModes)
```

AccessModes contains the desired access modes the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1

### fn spec.storage.volumeClaimTemplate.spec.withAccessModesMixin

```ts
withAccessModesMixin(accessModes)
```

AccessModes contains the desired access modes the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#access-modes-1

**Note:** This function appends passed data to existing values

### fn spec.storage.volumeClaimTemplate.spec.withStorageClassName

```ts
withStorageClassName(storageClassName)
```

Name of the StorageClass required by the claim. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#class-1

### fn spec.storage.volumeClaimTemplate.spec.withVolumeMode

```ts
withVolumeMode(volumeMode)
```

volumeMode defines what type of volume is required by the claim. Value of Filesystem is implied when not included in claim spec.

### fn spec.storage.volumeClaimTemplate.spec.withVolumeName

```ts
withVolumeName(volumeName)
```

VolumeName is the binding reference to the PersistentVolume backing this claim.

## obj spec.storage.volumeClaimTemplate.spec.dataSource

This field can be used to specify either: * An existing VolumeSnapshot object (snapshot.storage.k8s.io/VolumeSnapshot - Beta) * An existing PVC (PersistentVolumeClaim) * An existing custom resource/object that implements data population (Alpha) In order to use VolumeSnapshot object types, the appropriate feature gate must be enabled (VolumeSnapshotDataSource or AnyVolumeDataSource) If the provisioner or an external controller can support the specified data source, it will create a new volume based on the contents of the specified data source. If the specified data source is not supported, the volume will not be created and the failure will be reported as an event. In the future, we plan to support more data source types and the behavior of the provisioner may change.

### fn spec.storage.volumeClaimTemplate.spec.dataSource.withApiGroup

```ts
withApiGroup(apiGroup)
```

APIGroup is the group for the resource being referenced. If APIGroup is not specified, the specified Kind must be in the core API group. For any other third-party types, APIGroup is required.

### fn spec.storage.volumeClaimTemplate.spec.dataSource.withKind

```ts
withKind(kind)
```

Kind is the type of resource being referenced

### fn spec.storage.volumeClaimTemplate.spec.dataSource.withName

```ts
withName(name)
```

Name is the name of resource being referenced

## obj spec.storage.volumeClaimTemplate.spec.resources

Resources represents the minimum resources the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources

### fn spec.storage.volumeClaimTemplate.spec.resources.withLimits

```ts
withLimits(limits)
```

Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

### fn spec.storage.volumeClaimTemplate.spec.resources.withLimitsMixin

```ts
withLimitsMixin(limits)
```

Limits describes the maximum amount of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

**Note:** This function appends passed data to existing values

### fn spec.storage.volumeClaimTemplate.spec.resources.withRequests

```ts
withRequests(requests)
```

Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

### fn spec.storage.volumeClaimTemplate.spec.resources.withRequestsMixin

```ts
withRequestsMixin(requests)
```

Requests describes the minimum amount of compute resources required. If Requests is omitted for a container, it defaults to Limits if that is explicitly specified, otherwise to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/

**Note:** This function appends passed data to existing values

## obj spec.storage.volumeClaimTemplate.spec.selector

A label query over volumes to consider for binding.

### fn spec.storage.volumeClaimTemplate.spec.selector.withMatchExpressions

```ts
withMatchExpressions(matchExpressions)
```

matchExpressions is a list of label selector requirements. The requirements are ANDed.

### fn spec.storage.volumeClaimTemplate.spec.selector.withMatchExpressionsMixin

```ts
withMatchExpressionsMixin(matchExpressions)
```

matchExpressions is a list of label selector requirements. The requirements are ANDed.

**Note:** This function appends passed data to existing values

### fn spec.storage.volumeClaimTemplate.spec.selector.withMatchLabels

```ts
withMatchLabels(matchLabels)
```

matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

### fn spec.storage.volumeClaimTemplate.spec.selector.withMatchLabelsMixin

```ts
withMatchLabelsMixin(matchLabels)
```

matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels map is equivalent to an element of matchExpressions, whose key field is "key", the operator is "In", and the values array contains only "value". The requirements are ANDed.

**Note:** This function appends passed data to existing values