package k8s

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceSet struct {
	log        log.Logger
	kubeClient client.Client
	objects    []client.Object
}

// NewResourceSet returns a new resource set.
func NewResourceSet(l log.Logger, cluster *Cluster) *ResourceSet {
	return &ResourceSet{
		log:        l,
		kubeClient: cluster.Client(),
	}
}

// Add will read from r and deploy the resources into the cluster.
func (rs *ResourceSet) Add(ctx context.Context, r io.Reader) error {
	readObjects, err := ReadObjects(r, rs.kubeClient)
	if err != nil {
		return fmt.Errorf("error reading fixture: %w", err)
	}
	err = CreateObjects(ctx, rs.kubeClient, readObjects...)
	if err != nil {
		return err
	}

	rs.objects = append(rs.objects, readObjects...)
	return nil
}

// AddFile will open filename and deploy it into the cluster.
func (rs *ResourceSet) AddFile(ctx context.Context, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", filename, err)
	}
	defer f.Close()
	return rs.Add(ctx, f)
}

// Stop removes deployed resources from the cluster.
func (rs *ResourceSet) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for _, obj := range rs.objects {
		err := rs.kubeClient.Delete(ctx, obj)
		if err != nil {
			level.Error(rs.log).Log("msg", "failed to delete object", "obj", client.ObjectKeyFromObject(obj), "err", err)
		}
	}
}
