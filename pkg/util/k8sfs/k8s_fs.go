// package k8sfs implements an fs.FS that lazily reads data from secrets and configmaps in a kubernetes cluster.
package k8sfs

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type FS struct {
	client *kubernetes.Clientset
}

func New(client *kubernetes.Clientset) *FS {
	return &FS{
		client: client,
	}
}

func (f *FS) Open(name string) (fs.File, error) {
	parts := strings.Split(name, "_")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid file name")
	}
	ns := parts[1]
	objname := parts[2]
	key := parts[3]
	switch parts[0] {
	case "secret":
		return f.openSecret(ns, objname, key)
	case "configmap":
		return f.openConfigMap(ns, objname, key)
	default:
	}
	return nil, fmt.Errorf("invalid object type")
}

// TODO: hook this all up to a caching informer

func (f *FS) openSecret(ns, name, key string) (*file, error) {
	dat, err := f.ReadSecret(ns, name, key)
	if err != nil {
		return nil, err
	}
	return newFile(dat), nil
}

func (f *FS) openConfigMap(ns, name, key string) (*file, error) {
	dat, err := f.ReadConfigMap(ns, name, key)
	if err != nil {
		return nil, err
	}
	return newFile(dat), nil
}

func (f *FS) ReadSecret(ns, name, key string) (string, error) {
	secret, err := f.client.CoreV1().Secrets(ns).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	dat, ok := secret.Data[key]
	if !ok {
		return "", fmt.Errorf("secret %s/%s has no field %s", ns, name, key)
	}
	return string(dat), nil
}

func (f *FS) ReadConfigMap(ns, name, key string) (string, error) {
	cmap, err := f.client.CoreV1().ConfigMaps(ns).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return "", err
	}
	dat, ok := cmap.Data[key]
	if !ok {
		return "", fmt.Errorf("configmap %s/%s has no field %s", ns, name, key)
	}
	return string(dat), nil
}

type file struct {
	r strings.Reader
}

func newFile(content string) *file {
	return &file{r: *strings.NewReader(content)}
}

func (f *file) Stat() (fs.FileInfo, error) { return nil, fmt.Errorf("stat not implemented") }
func (f *file) Read(b []byte) (int, error) {
	return f.r.Read(b)
}
func (f *file) Close() error { return nil }

func SecretFilename(namespace, name, key string) string {
	return fmt.Sprintf("secret_%s_%s_%s", namespace, name, key)
}

func ConfigMapFilename(namespace, name, key string) string {
	return fmt.Sprintf("configmap_%s_%s_%s", namespace, name, key)
}
