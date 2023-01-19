package kubernetes_crds

import (
	"context"
	"io/fs"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type secretManager struct {
	fs     fs.FS
	client *kubernetes.Clientset
}

func (sm *secretManager) GetSecretData(ctx context.Context, namespace string, secretName string, field string) (string, error) {
	secret, err := sm.client.CoreV1().Secrets(namespace).Get(ctx, secretName, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data[field]), nil
}

func (sm *secretManager) GetConfigMapData(ctx context.Context, namespace string, name string, field string) (string, error) {
	cmap, err := sm.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	return cmap.Data[field], nil
}
