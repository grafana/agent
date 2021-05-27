package clientutil

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateOrUpdateService applies the given svc against the client.
func CreateOrUpdateService(ctx context.Context, c client.Client, svc *v1.Service) error {
	var exist v1.Service
	err := c.Get(ctx, client.ObjectKeyFromObject(svc), &exist)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("failed to retrieve existing service: %w", err)
	}

	if k8s_errors.IsNotFound(err) {
		err := c.Create(ctx, svc)
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}
	} else {
		svc.ResourceVersion = exist.ResourceVersion
		svc.Spec.IPFamily = exist.Spec.IPFamily
		svc.SetOwnerReferences(mergeOwnerReferences(svc.GetOwnerReferences(), exist.GetOwnerReferences()))
		svc.SetLabels(mergeMaps(svc.Labels, exist.Labels))
		svc.SetAnnotations(mergeMaps(svc.Annotations, exist.Annotations))

		err := c.Update(ctx, svc)
		if err != nil && !k8s_errors.IsNotFound(err) {
			return fmt.Errorf("failed to update service: %w", err)
		}
	}

	return nil
}

func mergeOwnerReferences(new, old []meta_v1.OwnerReference) []meta_v1.OwnerReference {
	existing := make(map[types.UID]bool)
	for _, ref := range old {
		existing[ref.UID] = true
	}
	for _, ref := range new {
		if _, ok := existing[ref.UID]; !ok {
			old = append(old, ref)
		}
	}
	return old
}

func mergeMaps(new, old map[string]string) map[string]string {
	if old == nil {
		old = make(map[string]string, len(new))
	}
	for k, v := range new {
		old[k] = v
	}
	return old
}
