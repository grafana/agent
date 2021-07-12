package config

import (
	grafana "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	prom "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// AssetReference is a namespaced Secret or ConfigMap selector.
type AssetReference struct {
	Namespace string
	Reference prom.SecretOrConfigMap
}

// AssetReferences returns all secret or configmap selectors used throughout
// the deployment. Every used secret and configmap should then be loaded into
// an assets.SecretStore.
func (d *Deployment) AssetReferences() []AssetReference {
	var res []AssetReference

	// Retrieve referenences from Agent
	if d.Agent.Spec.APIServerConfig != nil {
		res = append(res, apiServerAssetReferences(d.Agent.Namespace, d.Agent.Spec.APIServerConfig)...)
	}
	for _, rw := range d.Agent.Spec.Prometheus.RemoteWrite {
		res = append(res, remoteWriteAssetReferences(d.Agent.Namespace, &rw)...)
	}

	// Retrieve references from each PrometheusInstance
	for _, inst := range d.Prometheis {
		// Retrieve from inner PrometheusInstance
		res = append(res, AssetReference{
			Namespace: inst.Instance.Namespace,
			Reference: prom.SecretOrConfigMap{Secret: inst.Instance.Spec.AdditionalScrapeConfigs},
		})
		for _, rw := range inst.Instance.Spec.RemoteWrite {
			res = append(res, remoteWriteAssetReferences(inst.Instance.Namespace, &rw)...)
		}

		// Retrieve from ServiceMonitors
		for _, monitor := range inst.ServiceMonitors {
			for _, endpoint := range monitor.Spec.Endpoints {
				res = append(res, basicAuthAssetReferences(monitor.Namespace, endpoint.BasicAuth)...)

				if endpoint.TLSConfig != nil {
					res = append(res, tlsConfigReferences(monitor.Namespace, endpoint.TLSConfig)...)
				}
				res = append(res, AssetReference{
					Namespace: monitor.Namespace,
					Reference: prom.SecretOrConfigMap{Secret: &endpoint.BearerTokenSecret},
				})
			}
		}

		// Retrieve from PodMonitors
		for _, monitor := range inst.PodMonitors {
			for _, endpoint := range monitor.Spec.PodMetricsEndpoints {
				res = append(res, basicAuthAssetReferences(monitor.Namespace, endpoint.BasicAuth)...)

				if endpoint.TLSConfig != nil {
					res = append(res, tlsConfigReferences(monitor.Namespace, &prom.TLSConfig{
						SafeTLSConfig: endpoint.TLSConfig.SafeTLSConfig,
					})...)
				}
				res = append(res, AssetReference{
					Namespace: monitor.Namespace,
					Reference: prom.SecretOrConfigMap{Secret: &endpoint.BearerTokenSecret},
				})
			}
		}

		// Retrieve from Probes
		for _, probe := range inst.Probes {
			res = append(res, basicAuthAssetReferences(probe.Namespace, probe.Spec.BasicAuth)...)

			if probe.Spec.TLSConfig != nil {
				res = append(res, tlsConfigReferences(probe.Namespace, &prom.TLSConfig{
					SafeTLSConfig: probe.Spec.TLSConfig.SafeTLSConfig,
				})...)
			}
			res = append(res, AssetReference{
				Namespace: probe.Namespace,
				Reference: prom.SecretOrConfigMap{Secret: &probe.Spec.BearerTokenSecret},
			})
		}
	}

	return filterEmptyReferences(res)
}

func apiServerAssetReferences(namespace string, apiServer *prom.APIServerConfig) []AssetReference {
	var res []AssetReference

	res = append(res, basicAuthAssetReferences(namespace, apiServer.BasicAuth)...)

	if apiServer.TLSConfig != nil {
		res = append(res, tlsConfigReferences(namespace, apiServer.TLSConfig)...)
	}

	return filterEmptyReferences(res)
}

func basicAuthAssetReferences(namespace string, ba *prom.BasicAuth) []AssetReference {
	if ba == nil {
		return nil
	}

	return []AssetReference{{
		Namespace: namespace,
		Reference: prom.SecretOrConfigMap{Secret: &ba.Username},
	}, {
		Namespace: namespace,
		Reference: prom.SecretOrConfigMap{Secret: &ba.Password},
	}}
}

func remoteWriteAssetReferences(namespace string, rw *grafana.RemoteWriteSpec) []AssetReference {
	var res []AssetReference

	res = append(res, basicAuthAssetReferences(namespace, rw.BasicAuth)...)

	if rw.SigV4 != nil {
		res = append(res, AssetReference{
			Namespace: namespace,
			Reference: prom.SecretOrConfigMap{Secret: rw.SigV4.AccessKey},
		})
		res = append(res, AssetReference{
			Namespace: namespace,
			Reference: prom.SecretOrConfigMap{Secret: rw.SigV4.SecretKey},
		})
	}

	if rw.TLSConfig != nil {
		res = append(res, tlsConfigReferences(namespace, rw.TLSConfig)...)
	}

	return filterEmptyReferences(res)
}

func tlsConfigReferences(namespace string, cfg *prom.TLSConfig) []AssetReference {
	return filterEmptyReferences([]AssetReference{
		{Namespace: namespace, Reference: cfg.CA},
		{Namespace: namespace, Reference: cfg.Cert},
		{Namespace: namespace, Reference: prom.SecretOrConfigMap{Secret: cfg.KeySecret}},
	})
}

// filterEmptyReferences is a post-processing step of retrieving references to
// remove any references that are empty or nil. This makes the preceding code
// a little cleaner, since it won't have to check if each individual reference
// is defined.
func filterEmptyReferences(refs []AssetReference) []AssetReference {
	res := make([]AssetReference, 0, len(refs))

Outer:
	for _, ref := range refs {
		var (
			configMap = ref.Reference.ConfigMap
			secret    = ref.Reference.Secret
		)
		switch {
		case configMap == nil && secret == nil:
			continue Outer
		case configMap != nil && configMap.LocalObjectReference.Name == "":
			continue Outer
		case secret != nil && secret.LocalObjectReference.Name == "":
			continue Outer
		}
		res = append(res, ref)
	}

	return res
}
