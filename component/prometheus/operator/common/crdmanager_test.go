package common

import (
	"testing"
	"golang.org/x/exp/maps"

	"k8s.io/apimachinery/pkg/util/intstr"
	"github.com/go-kit/log"
	"github.com/grafana/agent/component"
	"github.com/grafana/agent/service/cluster"
	"github.com/grafana/agent/service/labelstore"
	"github.com/grafana/agent/component/prometheus/operator"
	"github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/scrape"

	promopv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/require"
)

func TestClearConfigsSameNsSamePrefix(t *testing.T) {
	logger := log.NewNopLogger()
	m := newCrdManager(
		component.Options{
			Logger: logger,
			GetServiceData: func(name string) (interface{}, error) { return nil, nil },
		},
		cluster.Mock(), 
		logger, 
		&operator.DefaultArguments, 
		KindServiceMonitor, 
		labelstore.New(logger),
	)

	m.discoveryManager = newMockDiscoveryManager()
	m.scrapeManager = newMockScrapeManager()

	targetPort := intstr.FromInt(9090)
	m.onAddServiceMonitor(&promopv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "monitoring",
			Name:      "svcmonitor",
		},
		Spec: promopv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": "my-group",
				},
			},
			Endpoints: []promopv1.Endpoint{
				{
					TargetPort: &targetPort,
					ScrapeTimeout:   "5s",
					Interval:        "10s",
				},
			},
		},
	})
	m.onAddServiceMonitor(&promopv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "monitoring",
			Name:      "svcmonitor-another",
		},
		Spec: promopv1.ServiceMonitorSpec{
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"group": "my-group",
				},
			},
			Endpoints: []promopv1.Endpoint{
				{
					TargetPort: &targetPort,
					ScrapeTimeout:   "5s",
					Interval:        "10s",
				},
			},
		},	})

	require.ElementsMatch(t, []string{"serviceMonitor/monitoring/svcmonitor-another/0", "serviceMonitor/monitoring/svcmonitor/0"}, maps.Keys(m.discoveryConfigs))
	m.clearConfigs("monitoring", "svcmonitor")
	require.ElementsMatch(t, []string{"serviceMonitor/monitoring/svcmonitor-another/0"}, maps.Keys(m.discoveryConfigs))
}

type mockDiscoveryManager struct {
	
}

func newMockDiscoveryManager() *mockDiscoveryManager {
	return &mockDiscoveryManager{
		
	}
}

func (m *mockDiscoveryManager) Run() error {
	return nil
}

func (m *mockDiscoveryManager) SyncCh() <-chan map[string][]*targetgroup.Group {
	return nil
}

func (m *mockDiscoveryManager) ApplyConfig(cfg map[string]discovery.Configs) error {
	return nil
}

type mockScrapeManager struct {
	
}

func newMockScrapeManager() *mockScrapeManager {
	return &mockScrapeManager{
		
	}
}

func (m *mockScrapeManager) Run(tsets <-chan map[string][]*targetgroup.Group) error {
	return nil
}

func (m *mockScrapeManager) Stop() {
	
}

func (m *mockScrapeManager) TargetsActive() map[string][]*scrape.Target {
	return nil
}

func (m *mockScrapeManager) ApplyConfig(cfg *config.Config) error {
	return nil
}
