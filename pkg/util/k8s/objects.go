package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/grafana/dskit/backoff"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetaFromKey creates object metadata from the namespace and name pair. This
// is useful for packages to avoid importing
// k8s.io/apimachinery/pkg/apis/meta/v1
func MetaFromKey(namespace, name string) meta_v1.ObjectMeta {
	return meta_v1.ObjectMeta{Namespace: namespace, Name: name}
}

// CreateObjects will create the provided set of objects. If any object
// couldn't be created, an error will be returned and created objects will be
// deleted.
func CreateObjects(ctx context.Context, cli client.Client, objs ...client.Object) (err error) {
	// Index offset into objs for objects we managed to create.
	createdOffset := -1

	defer func() {
		if err == nil {
			return
		}
		// Delete the subset of objs we managed to create
		for i := 0; i <= createdOffset; i++ {
			_ = cli.Delete(context.Background(), objs[i])
		}
	}()

	for i, obj := range objs {
		if err := cli.Create(ctx, obj); err != nil {
			return fmt.Errorf("failed to create %s: %w", client.ObjectKeyFromObject(obj), err)
		}
		createdOffset = i
	}
	return nil
}

// ReadObjects will read the set of objects from r and convert them into
// client.Object based on the scheme of the provided Kubernetes client.
//
// The data of r may be YAML or JSON.
func ReadObjects(r io.Reader, cli client.Client) ([]client.Object, error) {
	var (
		objects []client.Object

		scheme     = cli.Scheme()
		rawDecoder = yaml.NewYAMLOrJSONDecoder(r, 4096)
		decoder    = serializer.NewCodecFactory(scheme).UniversalDecoder(scheme.PrioritizedVersionsAllGroups()...)
	)

NextObject:
	for {
		var raw json.RawMessage

		err := rawDecoder.Decode(&raw)
		switch {
		case errors.Is(err, io.EOF):
			break NextObject
		case err != nil:
			return nil, fmt.Errorf("error parsing object: %w", err)
		case len(raw) == 0:
			// Skip over empty objects. This can happen when --- is used at the top
			// of YAML files.
			continue NextObject
		}

		obj, _, err := decoder.Decode(raw, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to decode object: %w", err)
		}
		clientObj, ok := obj.(client.Object)
		if !ok {
			return nil, fmt.Errorf("decoded object %T is not a controller-runtime object", obj)
		}
		objects = append(objects, clientObj)
	}

	return objects, nil
}

// DefaultBackoff is a default backoff config that retries forever until ctx is
// canceled.
var DefaultBackoff = backoff.Config{
	MinBackoff: 100 * time.Millisecond,
	MaxBackoff: 1 * time.Second,
}

// WaitReady will return with no error if obj becomes ready before ctx cancels
// or the backoff fails.
//
// obj may be one of: DaemonSet, StatefulSet, Deployment, Pod. obj must have
// namespace and name set so it can be found. obj will be updated with the
// state of the object in the cluster as WaitReady runs.
//
// The final state of the object will be returned when it is ready.
func WaitReady(ctx context.Context, cli client.Client, obj client.Object, bc backoff.Config) error {
	bo := backoff.New(ctx, bc)

	key := client.ObjectKeyFromObject(obj)

	var readyCheck func() bool
	switch obj := obj.(type) {
	case *apps_v1.DaemonSet:
		readyCheck = func() bool {
			return obj.Status.NumberReady >= obj.Status.UpdatedNumberScheduled
		}
	case *apps_v1.StatefulSet:
		readyCheck = func() bool {
			return obj.Status.ReadyReplicas >= obj.Status.UpdatedReplicas
		}
	case *apps_v1.Deployment:
		readyCheck = func() bool {
			return obj.Status.ReadyReplicas >= obj.Status.UpdatedReplicas
		}
	case *core_v1.Pod:
		readyCheck = func() bool {
			phase := obj.Status.Phase
			return phase == core_v1.PodRunning || phase == core_v1.PodSucceeded
		}
	}

	for bo.Ongoing() {
		err := cli.Get(ctx, key, obj)
		if err == nil && readyCheck() {
			break
		}
		bo.Wait()
	}

	return bo.Err()
}
