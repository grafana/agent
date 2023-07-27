package consumer

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/grafana/agent/component/discovery"
	"github.com/prometheus/common/model"
	"go.opentelemetry.io/collector/client"
	otelcomponent "go.opentelemetry.io/collector/component"
	otelconsumer "go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/collector/semconv/v1.5.0"
)

const (
	// OperationTypeInsert inserts a new k/v if it isn't already present
	OperationTypeInsert = "insert"
	// OperationTypeUpdate only modifies an existing k/v
	OperationTypeUpdate = "update"
	// OperationTypeUpsert does both of above
	OperationTypeUpsert = "upsert"

	//TODO: It'd be cleaner to get these from the otel semver package?
	//      Not all are in semver though. E.g. "k8s.pod.ip" is internal inside the k8sattributesprocessor.
	PodAssociationIPLabel       = "ip"
	PodAssociationOTelIPLabel   = "net.host.ip"
	PodAssociationk8sIPLabel    = "k8s.pod.ip"
	PodAssociationHostnameLabel = "hostname"
	PodAssociationConnectionIP  = "connection"
)

func ValidateOperationType(operationType string) error {
	switch operationType {
	case
		OperationTypeUpsert,
		OperationTypeInsert,
		OperationTypeUpdate:
		// Valid configuration, do nothing.
	default:
		return fmt.Errorf("unknown operation type %s", operationType)
	}
	return nil
}

func ValidatePodAssociations(podAssociations []string) error {
	for _, podAssociation := range podAssociations {
		switch podAssociation {
		case
			PodAssociationIPLabel,
			PodAssociationOTelIPLabel,
			PodAssociationk8sIPLabel,
			PodAssociationHostnameLabel,
			PodAssociationConnectionIP:
			// Valid configuration, do nothing.
		default:
			return fmt.Errorf("unknown pod association %s", podAssociation)
		}
	}
	return nil
}

// TODO: Put a private member so that this can't be created without calling NewConsumer?
type Consumer struct {
	nextConsumer otelconsumer.Traces

	mtx        sync.Mutex
	hostLabels map[string]discovery.Target

	operationType   string
	podAssociations []string

	logger log.Logger
}

var _ otelconsumer.Traces = (*Consumer)(nil)

func NewConsumer(nextConsumer otelconsumer.Traces, operationType string, podAssociations []string, logger log.Logger) (*Consumer, error) {
	if nextConsumer == nil {
		return nil, otelcomponent.ErrNilNextConsumer
	}

	err := ValidateOperationType(operationType)
	if err != nil {
		return nil, err
	}

	err = ValidatePodAssociations(podAssociations)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		nextConsumer:    nextConsumer,
		hostLabels:      make(map[string]discovery.Target),
		operationType:   operationType,
		podAssociations: podAssociations,
		logger:          logger,
	}, nil
}

func (c *Consumer) SetHostLabels(hostLabels map[string]discovery.Target) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.hostLabels = hostLabels
}

func (c *Consumer) GetHostLabels(ip string) (map[string]string, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	labels, ok := c.hostLabels[ip]
	if !ok {
		return nil, fmt.Errorf("unable to find labels for ip %q", ip)
	}

	return labels, nil
}

func (c *Consumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)

		c.processAttributes(ctx, rs.Resource().Attributes())
	}

	return c.nextConsumer.ConsumeTraces(ctx, td)
}

func (c *Consumer) Capabilities() otelconsumer.Capabilities {
	return otelconsumer.Capabilities{MutatesData: true}
}

func (c *Consumer) processAttributes(ctx context.Context, attrs pcommon.Map) {
	ip := c.getPodIP(ctx, attrs)
	// have to have an ip for labels lookup
	if ip == "" {
		level.Debug(c.logger).Log("msg", "unable to find ip in span attributes, skipping attribute addition")
		return
	}

	labels, err := c.GetHostLabels(ip)
	if err != nil {
		level.Debug(c.logger).Log("msg", err)
		return
	}

	for k, v := range labels {
		switch c.operationType {
		case OperationTypeUpsert:
			attrs.PutStr(string(k), string(v))
		case OperationTypeInsert:
			if _, ok := attrs.Get(string(k)); !ok {
				attrs.PutStr(string(k), string(v))
			}
		case OperationTypeUpdate:
			if toVal, ok := attrs.Get(string(k)); ok {
				toVal.SetStr(string(v))
			}
		}
	}
}

func (c *Consumer) getPodIP(ctx context.Context, attrs pcommon.Map) string {
	for _, podAssociation := range c.podAssociations {
		switch podAssociation {
		case PodAssociationIPLabel, PodAssociationOTelIPLabel, PodAssociationk8sIPLabel:
			ip := stringAttributeFromMap(attrs, podAssociation)
			if ip != "" {
				return ip
			}
		case PodAssociationHostnameLabel:
			hostname := stringAttributeFromMap(attrs, semconv.AttributeHostName)
			if net.ParseIP(hostname) != nil {
				return hostname
			}
		case PodAssociationConnectionIP:
			ip := c.getConnectionIP(ctx)
			if ip != "" {
				return ip
			}
		}
	}
	return ""
}

func stringAttributeFromMap(attrs pcommon.Map, key string) string {
	if attr, ok := attrs.Get(key); ok {
		if attr.Type() == pcommon.ValueTypeStr {
			return attr.Str()
		}
	}
	return ""
}

func (c *Consumer) getConnectionIP(ctx context.Context) string {
	cl := client.FromContext(ctx)
	if cl.Addr == nil {
		return ""
	}

	host := cl.Addr.String()
	if strings.Contains(host, ":") {
		var err error
		splitHost, _, err := net.SplitHostPort(host)
		if err != nil {
			// It's normal for this to fail for IPv6 address strings that don't actually include a port.
			level.Debug(c.logger).Log("msg", "unable to split connection host and port", "host", host, "err", err)
		} else {
			host = splitHost
		}
	}

	return host
}

func GetHostFromLabels(labels discovery.Target) (string, error) {
	address, ok := labels[model.AddressLabel]
	if !ok {
		return "", fmt.Errorf("unable to find address in labels %q", labels.Labels())
	}

	host := string(address)
	if strings.Contains(host, ":") {
		var err error
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return "", fmt.Errorf("unable to split host and port in address %q: %w", address, err)
		}
	}

	return host, nil
}

func CleanupLabels(labels discovery.Target) {
	for k := range labels {
		if strings.HasPrefix(string(k), "__") {
			delete(labels, k)
		}
	}
}
