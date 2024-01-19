package receivers

import (
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/component/common/config"
	"github.com/grafana/agent/component/prometheus/remotewrite"
	"github.com/grafana/river/rivertypes"
	"github.com/prometheus/client_golang/prometheus"
)

// newManagedPrometheusRemoteWrite creates the new prometheus.remote_write managed component.
func (c *Component) newManagedPrometheusRemoteWrite(url string, username string, password rivertypes.Secret, slug string) (*remotewrite.Component, error) {
	rwArgs := remotewrite.Arguments{}
	rwArgs.SetToDefault()
	rwEndpoint := &remotewrite.EndpointOptions{}
	rwEndpoint.SetToDefault()
	rwEndpoint.URL = url + "/api/prom/push"
	rwEndpoint.HTTPClientConfig.BasicAuth = &config.BasicAuth{}
	rwEndpoint.HTTPClientConfig.BasicAuth.Username = username
	rwEndpoint.HTTPClientConfig.BasicAuth.Password = password
	rwArgs.Endpoints = append(rwArgs.Endpoints, rwEndpoint)

	newOpts := c.opts
	newOpts.DataPath = c.opts.DataPath + "/prometheus/" + slug
	newOpts.ID = c.opts.ID + ".prometheus.remote_write." + slug
	newOpts.Registerer = prometheus.WrapRegistererWith(prometheus.Labels{
		"sub_component_id": "prometheus.remote_write." + slug,
	}, c.opts.Registerer)

	// TODO Logger and Tracer
	// Logger: log.With(globals.Logger, "component", cn.globalID),
	// Tracer: tracing.WrapTracer(globals.TraceProvider, cn.globalID),

	newOpts.OnStateChange = func(e component.Exports) {
		c.exportsMut.Lock()
		defer c.exportsMut.Unlock()

		c.exports.Stacks[slug] = Stack{
			PrometheusReceiver: e.(remotewrite.Exports).Receiver,
		}

		c.opts.OnStateChange(c.exports)
	}

	return remotewrite.New(newOpts, rwArgs)
}
