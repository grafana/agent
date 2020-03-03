package prometheus

import (
	"context"
	"net"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/pkg/labels"
)

var (
	kubernetesNodeNameLabel = "__meta_kubernetes_pod_node_name"
)

// DiscoveredGroups is a set of groups found via service discovery.
type DiscoveredGroups = map[string][]*targetgroup.Group

// GroupChannel is a channel that provides discovered target groups.
type GroupChannel = <-chan DiscoveredGroups

// HostFilter acts as a MITM between the discovery manager and the
// scrape manager, filtering out discovered targets that are not
// running on the same node as the agent itself.
type HostFilter struct {
	ctx    context.Context
	cancel context.CancelFunc

	host string

	inputCh  GroupChannel
	outputCh chan map[string][]*targetgroup.Group
}

// NewHostFilter creates a new HostFilter
func NewHostFilter(host string) *HostFilter {
	ctx, cancel := context.WithCancel(context.Background())
	f := &HostFilter{
		ctx:    ctx,
		cancel: cancel,

		host: host,

		outputCh: make(chan map[string][]*targetgroup.Group),
	}
	return f
}

func (f *HostFilter) Run(syncCh GroupChannel) {
	f.inputCh = syncCh

	for {
		select {
		case <-f.ctx.Done():
			return
		case data := <-f.inputCh:
			f.outputCh <- FilterGroups(data, f.host)
		}
	}
}

// Stop stops the host filter from processing more target updates.
func (f *HostFilter) Stop() {
	f.cancel()
}

// SyncCh returns a read only channel used by all the clients to receive
// target updates.
func (f *HostFilter) SyncCh() GroupChannel {
	return f.outputCh
}

// FilterGroups takes a set of DiscoveredGroups as input and filters out
// any Target that is not running on the host machine provided by host.
//
// This is done by looking at two labels:
//
//   1. __meta_kubernetes_pod_node_name is used first to represent the host
//      machine of networked containers. Primarily useful for Kubernetes. This
//      label is automatically added when using Kubernetes service discovery.
//
//   2. __address__ is used next to represent the address of the service
//      to scrape. In a containerized envirment, __address__ will be the
//      address of the container.
//
// If the discovered address is localhost or 127.0.0.1, the group is never
// filtered out.
func FilterGroups(in DiscoveredGroups, host string) DiscoveredGroups {
	out := make(DiscoveredGroups, len(in))

	for name, groups := range in {
		groupList := make([]*targetgroup.Group, 0, len(groups))

		for _, group := range groups {
			newGroup := &targetgroup.Group{
				Targets: make([]model.LabelSet, 0, len(group.Targets)),
				Labels:  group.Labels,
				Source:  group.Source,
			}

			for _, target := range group.Targets {
				if !shouldFilterTarget(target, group.Labels, host) {
					level.Debug(util.Logger).Log("msg", "including target", "target_labels", target.String(), "common_labels", group.Labels.String(), "host", host)

					newGroup.Targets = append(newGroup.Targets, target)
				} else {
					level.Debug(util.Logger).Log("msg", "ignoring target", "target_labels", target.String(), "common_labels", group.Labels.String(), "host", host)
				}
			}

			groupList = append(groupList, newGroup)
		}

		out[name] = groupList
	}

	return out
}

// shouldFilterTarget returns true when the target labels (combined with the set of common
// labels) should be filtered out by FilterGroups.
func shouldFilterTarget(target model.LabelSet, common model.LabelSet, host string) bool {
	lbls := make([]labels.Label, 0, len(target)+len(common))

	for name, value := range target {
		lbls = append(lbls, labels.Label{Name: string(name), Value: string(value)})
	}

	for name, value := range common {
		if _, ok := target[name]; !ok {
			lbls = append(lbls, labels.Label{Name: string(name), Value: string(value)})
		}
	}

	shouldFilterTargetByLabelValue := func(labelValue string) bool {
		if addr, _, err := net.SplitHostPort(labelValue); err == nil {
			labelValue = addr
		}

		// Special case: always allow localhost/127.0.0.1
		if labelValue == "localhost" || labelValue == "127.0.0.1" {
			return false
		}

		return labelValue != host
	}

	lset := labels.New(lbls...)
	addressLabel := lset.Get(model.AddressLabel)
	if addressLabel == "" {
		// No address label. This is invalid and will generate an error by the scrape
		// manager, so we'll pass it on for now.
		return false
	}

	// If the __address__ label matches, we can quit early.
	if !shouldFilterTargetByLabelValue(addressLabel) {
		return false
	}

	// Fall back to testing __meta_kubernetes_pod_node_name
	if hostAddress := lset.Get(kubernetesNodeNameLabel); hostAddress != "" {
		return shouldFilterTargetByLabelValue(hostAddress)
	}

	// Nothing matches, filter it out.
	return true
}
