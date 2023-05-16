package configgen

import (
	"context"
	"fmt"

	promopv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SecretFetcher interface {
	GetSecretValue(namespace, name, field string) (string, error)
	GetConfigMapValue(namespace, name, field string) (string, error)
	SecretOrConfigMapValue(namespace string, socm promopv1.SecretOrConfigMap) (string, error)
}

// secretManager fetches secrets from kubernetes and stores a short-term cache of values.
// lifetime is intended to be a single conversion.
type secretManager struct {
	secretCache map[string]map[string][]byte
	configCache map[string]map[string]string
	client      *kubernetes.Clientset
}

func NewSecretManager(client *kubernetes.Clientset) SecretFetcher {
	return &secretManager{
		secretCache: make(map[string]map[string][]byte),
		configCache: make(map[string]map[string]string),
		client:      client,
	}
}

func (s *secretManager) GetSecretValue(namespace, name, field string) (string, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	var data map[string][]byte
	if m, ok := s.secretCache[key]; ok {
		data = m
	} else {
		secret, err := s.client.CoreV1().Secrets(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return "", err
		}
		data = secret.Data
		s.secretCache[key] = data
	}

	if dat, ok := data[field]; ok {
		return string(dat), nil
	} else {
		return "", fmt.Errorf("secret %s/%s has no field %s", namespace, name, field)
	}
}

func (s *secretManager) GetConfigMapValue(namespace, name, field string) (string, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	var data map[string]string
	if m, ok := s.configCache[key]; ok {
		data = m
	} else {
		cMap, err := s.client.CoreV1().ConfigMaps(namespace).Get(context.Background(), name, v1.GetOptions{})
		if err != nil {
			return "", err
		}
		data = cMap.Data
		s.configCache[key] = data
	}
	if dat, ok := data[field]; ok {
		return dat, nil
	} else {
		return "", fmt.Errorf("configmap %s/%s has no field %s", namespace, name, field)
	}
}

func (s *secretManager) SecretOrConfigMapValue(namespace string, socm promopv1.SecretOrConfigMap) (string, error) {
	if socm.Secret != nil {
		return s.GetSecretValue(namespace, socm.Secret.Name, socm.Secret.Key)
	} else if socm.ConfigMap != nil {
		return s.GetConfigMapValue(namespace, socm.ConfigMap.Name, socm.ConfigMap.Key)
	}
}
