package smoke

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
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
	clientset kubernetes.Interface
	namespace string
	pod       string
}

func (t *deletePodTask) Run(ctx context.Context) error {
	level.Debug(t.logger).Log("msg", "deleting pod")
	if err := t.clientset.CoreV1().Pods(t.namespace).Delete(ctx, t.pod, metav1.DeleteOptions{
		GracePeriodSeconds: pointer.Int64(0),
	}); err != nil {
		level.Error(t.logger).Log("msg", "failed to delete pod", "err", err)
	}
	return nil
}

type scaleDeploymentTask struct {
	logger      log.Logger
	clientset   kubernetes.Interface
	namespace   string
	deployment  string
	maxReplicas int
	minReplicas int
}

func (t *scaleDeploymentTask) Run(ctx context.Context) error {
	newReplicas := rand.Intn(t.maxReplicas-t.minReplicas) + t.minReplicas
	level.Debug(t.logger).Log("msg", "scaling replicas", "replicas", newReplicas)

	scale, err := t.clientset.AppsV1().Deployments(t.namespace).
		GetScale(ctx, t.deployment, metav1.GetOptions{})
	if err != nil {
		level.Error(t.logger).Log("msg", "failed to get autoscalingv1.Scale object", "err", err)
		return nil
	}

	sc := *scale
	sc.Spec.Replicas = int32(newReplicas)
	_, err = t.clientset.AppsV1().Deployments(t.namespace).
		UpdateScale(ctx, t.deployment, &sc, metav1.UpdateOptions{})
	if err != nil {
		level.Error(t.logger).Log("msg", "failed to scale deployment", "err", err)
	}
	return nil
}

type deletePodBySelectorTask struct {
	logger    log.Logger
	clientset kubernetes.Interface
	namespace string
	selector  string
}

func (t *deletePodBySelectorTask) Run(ctx context.Context) error {
	list, err := t.clientset.CoreV1().Pods(t.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: t.selector,
	})
	if err != nil {
		level.Error(t.logger).Log("msg", "failed to list pods", "err", err)
		return nil
	}

	l := len(list.Items)
	if l > 0 {
		i := rand.Intn(l)
		pod := list.Items[i].Name
		level.Debug(t.logger).Log("msg", "deleting pod", "pod", pod)
		if err := t.clientset.CoreV1().Pods(t.namespace).Delete(ctx, pod, metav1.DeleteOptions{
			GracePeriodSeconds: pointer.Int64(0),
		}); err != nil {
			level.Error(t.logger).Log("msg", "failed to delete pod", "pod", pod, "err", err)
		}
	}

	return nil
}
