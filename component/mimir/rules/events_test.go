package rules

import (
	"os"
	"testing"

	"github.com/go-kit/log"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
)

func TestQueueEventHandler(t *testing.T) {
	handler := Component{
		log:   log.NewLogfmtLogger(os.Stdout),
		queue: workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
	}

	ns := &v1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "name",
			Namespace: "namespace",
		},
	}

	handler.OnAdd(ns)
	event, _ := handler.queue.Get()
	require.Equal(t, EventTypeResourceChanged, event.(Event).Type)
	require.Equal(t, "namespace/name", event.(Event).ObjectKey)
	handler.queue.Forget(event)
	handler.queue.Done(event)

	handler.OnDelete(ns)
	event, _ = handler.queue.Get()
	require.Equal(t, EventTypeResourceChanged, event.(Event).Type)
	require.Equal(t, "namespace/name", event.(Event).ObjectKey)
	handler.queue.Forget(event)
	handler.queue.Done(event)

	handler.OnUpdate(ns, ns)
	event, _ = handler.queue.Get()
	require.Equal(t, EventTypeResourceChanged, event.(Event).Type)
	require.Equal(t, "namespace/name", event.(Event).ObjectKey)
	handler.queue.Forget(event)
	handler.queue.Done(event)
}
