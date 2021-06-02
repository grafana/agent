package clientutil

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	apps_v1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var invalidDNS1123Characters = regexp.MustCompile("[^-a-z0-9]+")

// SanitizeVolumeName ensures that the given volume name is a valid DNS-1123 label
// accepted by Kubernetes.
//
// Copied from github.com/prometheus-operator/prometheus-operator/pkg/k8sutil.
func SanitizeVolumeName(name string) string {
	name = strings.ToLower(name)
	name = invalidDNS1123Characters.ReplaceAllString(name, "-")
	if len(name) > validation.DNS1123LabelMaxLength {
		name = name[0:validation.DNS1123LabelMaxLength]
	}
	return strings.Trim(name, "-")
}

// CreateOrUpdateSecret applies the given secret against the client.
func CreateOrUpdateSecret(ctx context.Context, c client.Client, s *v1.Secret) error {
	var exist v1.Secret
	err := c.Get(ctx, client.ObjectKeyFromObject(s), &exist)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("failed to retrieve existing service: %w", err)
	}

	if k8s_errors.IsNotFound(err) {
		err := c.Create(ctx, s)
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}
	} else {
		s.ResourceVersion = exist.ResourceVersion
		s.SetOwnerReferences(mergeOwnerReferences(s.GetOwnerReferences(), exist.GetOwnerReferences()))
		s.SetLabels(mergeMaps(s.Labels, exist.Labels))
		s.SetAnnotations(mergeMaps(s.Annotations, exist.Annotations))

		err := c.Update(ctx, s)
		if err != nil && !k8s_errors.IsNotFound(err) {
			return fmt.Errorf("failed to update service: %w", err)
		}
	}

	return nil
}

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
		svc.Spec.IPFamilies = exist.Spec.IPFamilies
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

// CreateOrUpdateStatefulSet applies the given StatefulSet against the client.
func CreateOrUpdateStatefulSet(ctx context.Context, c client.Client, ss *apps_v1.StatefulSet) error {
	var exist apps_v1.StatefulSet
	err := c.Get(ctx, client.ObjectKeyFromObject(ss), &exist)
	if err != nil && !k8s_errors.IsNotFound(err) {
		return fmt.Errorf("failed to retrieve existing statefulset: %w", err)
	}

	if k8s_errors.IsNotFound(err) {
		err := c.Create(ctx, ss)
		if err != nil {
			return fmt.Errorf("failed to create statefulset: %w", err)
		}
	} else {
		ss.ResourceVersion = exist.ResourceVersion
		ss.SetOwnerReferences(mergeOwnerReferences(ss.GetOwnerReferences(), exist.GetOwnerReferences()))
		ss.SetLabels(mergeMaps(ss.Labels, exist.Labels))
		ss.SetAnnotations(mergeMaps(ss.Annotations, exist.Annotations))

		err := c.Update(ctx, ss)
		if k8s_errors.IsNotAcceptable(err) {
			err = c.Delete(ctx, ss)
			if err != nil {
				return fmt.Errorf("failed to update statefulset: deleting old statefulset: %w", err)
			}
			err = c.Create(ctx, ss)
			if err != nil {
				return fmt.Errorf("failed to update statefulset: statefulset: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to update statefulset: %w", err)
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
