package smoke

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// The Task interface represents some unit of work performed concurrently.
type Task interface {
	Run(context.Context) error
}

type repeatingTask struct {
	Task
	frequency time.Duration
}

type deletePodTask struct {
	logger    log.Logger
	clientset *kubernetes.Clientset
	namespace string
	pod       string
}

var _zero int64

func (t *deletePodTask) Run(ctx context.Context) error {
	if err := t.clientset.CoreV1().Pods(t.namespace).Delete(ctx, t.pod, metav1.DeleteOptions{
		GracePeriodSeconds: &_zero,
	}); err != nil {
		const msg = "failed to delete %s pod"
		level.Error(t.logger).Log("msg", fmt.Sprintf(msg, t.pod), "err", err)
	}
	return nil
}

type scaleDeploymentTask struct {
	logger      log.Logger
	clientset   *kubernetes.Clientset
	namespace   string
	deployment  string
	maxReplicas int
	minReplicas int
}

func (t *scaleDeploymentTask) Run(ctx context.Context) error {
	newReplicas := rand.Intn(t.maxReplicas-t.minReplicas) + t.minReplicas
	const msg = "scaling %s replicas"
	level.Debug(t.logger).Log("msg", fmt.Sprintf(msg, t.deployment), "replicas", newReplicas)

	scale, err := t.clientset.AppsV1().Deployments(t.namespace).
		GetScale(ctx, t.deployment, metav1.GetOptions{})
	if err != nil {
		const msg = "failed to get autoscalingv1.Scale object for %s deployment"
		level.Error(t.logger).Log("msg", fmt.Sprintf(msg, t.deployment), "err", err)
	}

	sc := *scale
	sc.Spec.Replicas = int32(newReplicas)
	_, err = t.clientset.AppsV1().Deployments(t.namespace).
		UpdateScale(ctx, t.deployment, &sc, metav1.UpdateOptions{})
	if err != nil {
		const msg = "failed to scale %s deployment"
		level.Error(t.logger).Log("msg", fmt.Sprintf(msg, t.deployment), "err", err)
	}
	return nil
}
