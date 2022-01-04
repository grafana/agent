package smoke

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// The Task interface represents units of work performed by some function on
// some frequency interval by the smoke framework.
type Task interface {
	Task() (func(context.Context, *Smoke) error, time.Duration)
}

type deletePodTask struct {
	namespace string
	pod       string
	duration  time.Duration
}

var _zero int64

func (t *deletePodTask) Task() (func(context.Context, *Smoke) error, time.Duration) {
	return func(ctx context.Context, s *Smoke) error {
		if err := s.clientset.CoreV1().Pods(t.namespace).Delete(ctx, t.pod, metav1.DeleteOptions{
			GracePeriodSeconds: &_zero,
		}); err != nil {
			const msg = "failed to delete %s pod"
			s.logError("msg", fmt.Sprintf(msg, t.pod), "err", err)
		}
		return nil
	}, t.duration
}

type scaleDeploymentTask struct {
	namespace   string
	deployment  string
	maxReplicas int
	minReplicas int
	duration    time.Duration
}

func (t *scaleDeploymentTask) Task() (func(context.Context, *Smoke) error, time.Duration) {
	return func(ctx context.Context, s *Smoke) error {
		newReplicas := rand.Intn(t.maxReplicas-t.minReplicas) + t.minReplicas
		const msg = "scaling %s replicas"
		s.logDebug("msg", fmt.Sprintf(msg, t.deployment), "replicas", newReplicas)

		scale, err := s.clientset.AppsV1().Deployments(t.namespace).
			GetScale(ctx, t.deployment, metav1.GetOptions{})
		if err != nil {
			const msg = "failed to get autoscalingv1.Scale object for %s deployment"
			s.logError("msg", fmt.Sprintf(msg, t.deployment), "err", err)
			// TODO: return error here? intermittent failure could cause test to exit
		}

		sc := *scale
		sc.Spec.Replicas = int32(newReplicas)
		_, err = s.clientset.AppsV1().Deployments(t.namespace).
			UpdateScale(ctx, t.deployment, &sc, metav1.UpdateOptions{})
		if err != nil {
			const msg = "failed to scale %s deployment"
			s.logError("msg", fmt.Sprintf(msg, t.deployment), "err", err)
			// TODO: same here
		}
		return nil
	}, t.duration
}
