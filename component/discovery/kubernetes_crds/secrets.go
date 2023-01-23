package kubernetes_crds

import (
	"context"
	"fmt"

	"github.com/psanford/memfs"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type secretManager struct {
	fs     *memfs.FS
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

// save the secret to the file system and return the file path
func (sm *secretManager) StoreSecretData(ctx context.Context, namespace string, secretName string, field string) (string, error) {
	content, err := sm.GetSecretData(ctx, namespace, secretName, field)
	if err != nil {
		return "", err
	}
	fname := fmt.Sprintf("secret_%s_%s_%s", namespace, secretName, field)
	if err = sm.fs.WriteFile(fname, []byte(content), 0600); err != nil {
		return "", err
	}
	return fname, nil
}

// save the config map field to the file system and return the file path
func (sm *secretManager) StoreConfigMapData(ctx context.Context, namespace string, name string, field string) (string, error) {
	content, err := sm.GetConfigMapData(ctx, namespace, name, field)
	if err != nil {
		return "", err
	}
	fname := fmt.Sprintf("configmap_%s_%s_%s", namespace, name, field)
	if err = sm.fs.WriteFile(fname, []byte(content), 0600); err != nil {
		return "", err
	}
	return fname, nil
}
