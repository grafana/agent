// Package k8s spins up a Kubernetes cluster for testing.
package k8s

import (
	"context"
	"fmt"
	"time"

	docker_types "github.com/docker/docker/api/types"
	docker_nat "github.com/docker/go-connections/nat"
	k3d_client "github.com/rancher/k3d/v5/pkg/client"
	config "github.com/rancher/k3d/v5/pkg/config"
	k3d_cfgtypes "github.com/rancher/k3d/v5/pkg/config/types"
	k3d_config "github.com/rancher/k3d/v5/pkg/config/v1alpha3"
	k3d_runtime "github.com/rancher/k3d/v5/pkg/runtimes"
	k3d_docker "github.com/rancher/k3d/v5/pkg/runtimes/docker"
	k3d_types "github.com/rancher/k3d/v5/pkg/types"
	k3d_version "github.com/rancher/k3d/v5/version"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"
	k8s_clientcmd "k8s.io/client-go/tools/clientcmd"
)

// Cluster is a Kubernetes cluster that runs inside of a k3s Docker container.
// Call GetConfig to retrieve a Kubernetes *rest.Config to use to connect to
// the cluster.
//
// Cluster also runs an NGINX ingress controller which is exposed to the host.
// Call GetHTTPAddr to get the address for making requests against the server.
//
// Note that k3s uses containerd as its runtime, which means local Docker
// images are not immediately available for use. To push local images to a
// container, call PushImages. It's recommended that tests use image names that
// are not available on Docker Hub to avoid accidentally testing against the
// wrong image.
//
// Cluster should be stopped by calling Stop, otherwise running Docker
// containers will leak.
type Cluster struct {
	runtime    k3d_runtime.Runtime
	k3dCluster k3d_types.Cluster
	restConfig *rest.Config
	nginxAddr  string
}

// NewCluster creates a new Cluster.
func NewCluster() (cluster *Cluster, err error) {
	var (
		ctx = context.Background()

		// We force the Docker runtime so we can create a Docker client for getting
		// the exposed ports for the API server and NGINX.
		runtime = k3d_runtime.Docker
	)

	k3dConfig := k3d_config.SimpleConfig{
		TypeMeta: k3d_cfgtypes.TypeMeta{
			Kind:       "Simple",
			APIVersion: config.DefaultConfigApiVersion,
		},
		Name:    randomClusterName(),
		Servers: 1,
		Ports: []k3d_config.PortWithNodeFilters{{
			// Bind NGINX (container port 80) to 127.0.0.1:0
			Port:        "127.0.0.1:0:80",
			NodeFilters: []string{"loadbalancer"},
		}},
		ExposeAPI: k3d_config.SimpleExposureOpts{
			// Bind API sever to 127.0.0.1:0
			Host:     "127.0.0.1",
			HostIP:   "127.0.0.1",
			HostPort: "0",
		},
		Image: fmt.Sprintf("%s:%s", k3d_types.DefaultK3sImageRepo, k3d_version.K3sVersion),
		Options: k3d_config.SimpleConfigOptions{
			K3dOptions: k3d_config.SimpleConfigOptionsK3d{
				Wait:    true,
				Timeout: time.Minute,
			},
		},
	}
	clusterConfig, err := config.TransformSimpleToClusterConfig(ctx, runtime, k3dConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cluster config: %w", err)
	}

	err = k3d_client.ClusterRun(ctx, runtime, clusterConfig)

	defer func() {
		// We don't want to leak the cluster here, and we can't really be sure how
		// many resources exist, even if ClusterRun fails. If we never set our
		// cluster return argument, we'll delete the k3d cluster. This also
		// gracefully handles panics.
		if cluster == nil {
			_ = k3d_client.ClusterDelete(ctx, runtime, &clusterConfig.Cluster, k3d_types.ClusterDeleteOpts{})
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("failed to run cluster: %w", err)
	}

	// Retrieve the actual local addresses for NGINX and the API server rather
	// than the 127.0.0.1:0 addresses still present in the cluster config.
	httpAddr, apiServerAddr, err := loadBalancerAddrs(ctx, clusterConfig.Cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to discover exposed cluster addresses: %w", err)
	}

	kubeconfig, err := k3d_client.KubeconfigGet(ctx, runtime, &clusterConfig.Cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve kubeconfig: %w", err)
	}
	if c, ok := kubeconfig.Clusters[kubeconfig.CurrentContext]; ok && c != nil {
		// The generated kubeconfig will set https://127.0.0.1:0 as the address. We
		// need to replace it with the actual exposed port that Docker generated
		// for us.
		c.Server = "https://" + apiServerAddr
	} else {
		return nil, fmt.Errorf("generated kubeconfig missing context set")
	}
	restCfg, err := k8s_clientcmd.NewDefaultClientConfig(*kubeconfig, nil).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not generate k8s REST API config: %w", err)
	}

	return &Cluster{
		runtime:    runtime,
		k3dCluster: clusterConfig.Cluster,
		restConfig: restCfg,
		nginxAddr:  httpAddr,
	}, nil
}

func randomClusterName() string {
	return "grafana-agent-e2e-" + rand.String(5)
}

func loadBalancerAddrs(ctx context.Context, cluster k3d_types.Cluster) (httpAddr, apiServerAddr string, err error) {
	var lb *k3d_types.Node
	for _, n := range cluster.Nodes {
		if n.Role == k3d_types.LoadBalancerRole {
			lb = n
			break
		}
	}
	if lb == nil {
		return "", "", fmt.Errorf("no loadbalancer node")
	}

	cli, err := k3d_docker.GetDockerClient()
	if err != nil {
		return "", "", fmt.Errorf("failed to get docker client: %w", err)
	}
	info, err := cli.ContainerInspect(ctx, lb.Name)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect loadbalancer: %w", err)
	}

	httpAddr, err = hostBinding(info, 80)
	if err != nil {
		return "", "", fmt.Errorf("failed to discover NGINX HTTP addr: %w", err)
	}
	apiServerAddr, err = hostBinding(info, 6443)
	if err != nil {
		return "", "", fmt.Errorf("failed to discover API server addr: %w", err)
	}
	return httpAddr, apiServerAddr, nil
}

func hostBinding(containerInfo docker_types.ContainerJSON, containerPort int) (string, error) {
	for rawPort, bindings := range containerInfo.NetworkSettings.Ports {
		_, portString := docker_nat.SplitProtoPort(string(rawPort))
		port, _ := docker_nat.ParsePort(portString)
		if port != containerPort {
			continue
		}
		if len(bindings) == 0 {
			return "", fmt.Errorf("no exposed bindings for port %d", containerPort)
		}
		return fmt.Sprintf("%s:%s", bindings[0].HostIP, bindings[0].HostPort), nil
	}

	return "", fmt.Errorf("no container port %d exposed", containerPort)
}

// GetConfig returns a *rest.Config that can be used to connect to the
// Kubernetes cluster. The returned Config is a copy and is safe for
// modification.
func (c *Cluster) GetConfig() *rest.Config {
	return rest.CopyConfig(c.restConfig)
}

// GetHTTPAddr returns the host:port address that can be used to connect to the
// cluster's NGINX server.
func (c *Cluster) GetHTTPAddr() string {
	return c.nginxAddr
}

// PushImages push images from the local Docker host into the Cluster. If the
// specified image does not have a tag, `:latest` is assumed.
func (c *Cluster) PushImages(images ...string) error {
	return k3d_client.ImageImportIntoClusterMulti(
		context.Background(),
		c.runtime,
		images,
		&c.k3dCluster,
		k3d_types.ImageImportOpts{},
	)
}

// Stop shuts down and deletes the cluster. Stop must be called to clean up
// created Docker resources.
func (c *Cluster) Stop() error {
	return k3d_client.ClusterDelete(context.Background(), c.runtime, &c.k3dCluster, k3d_types.ClusterDeleteOpts{})
}
