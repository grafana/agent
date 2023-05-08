// Package k8s spins up a Kubernetes cluster for testing.
package k8s

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	docker_types "github.com/docker/docker/api/types"
	docker_nat "github.com/docker/go-connections/nat"
	gragent "github.com/grafana/agent/pkg/operator/apis/monitoring/v1alpha1"
	k3d_client "github.com/k3d-io/k3d/v5/pkg/client"
	config "github.com/k3d-io/k3d/v5/pkg/config"
	k3d_cfgtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	k3d_config "github.com/k3d-io/k3d/v5/pkg/config/v1alpha4"
	k3d_log "github.com/k3d-io/k3d/v5/pkg/logger"
	k3d_runtime "github.com/k3d-io/k3d/v5/pkg/runtimes"
	k3d_docker "github.com/k3d-io/k3d/v5/pkg/runtimes/docker"
	k3d_types "github.com/k3d-io/k3d/v5/pkg/types"
	k3d_version "github.com/k3d-io/k3d/v5/version"
	promop_v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apiextensions_v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	k8s_clientcmd "k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Cluster is a Kubernetes cluster that runs inside of a k3s Docker container.
// Call GetConfig to retrieve a Kubernetes *rest.Config to use to connect to
// the cluster.
//
// Cluster also runs an NGINX ingress controller which is exposed to the host.
// Call GetHTTPAddr to get the address for making requests against the server.
//
// Set K8S_USE_DOCKER_NETWORK in your environment variables if you are
// running tests from inside of a Docker container. This environment variable
// configures the k3s Docker container to join the same network as the
// container tests are running in. When this environment variable isn't set,
// the exposed ports on the Docker host are used for cluster communication.
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
	kubeClient client.Client
	nginxAddr  string
}

// Options control creation of a cluster.
type Options struct {
	// Scheme is the Kubernetes scheme used for the generated Kubernetes client.
	// If nil, a generated scheme that contains all known Kubernetes API types
	// will be generated.
	Scheme *runtime.Scheme
}

func (o *Options) applyDefaults() error {
	if o.Scheme == nil {
		o.Scheme = runtime.NewScheme()

		for _, add := range []func(*runtime.Scheme) error{
			scheme.AddToScheme,
			apiextensions_v1.AddToScheme,
			gragent.AddToScheme,
			promop_v1.AddToScheme,
		} {
			if err := add(o.Scheme); err != nil {
				return fmt.Errorf("unable to register scheme: %w", err)
			}
		}
	}
	return nil
}

// NewCluster creates a new Cluster. NewCluster won't return with success until
// the cluster is running, but things like the ingress controller might not be
// running right away. You should never assume that any resource in the cluster
// is running and utilize exponential backoffs to allow time for things to spin
// up.
func NewCluster(ctx context.Context, o Options) (cluster *Cluster, err error) {
	var (
		// We force the Docker runtime so we can create a Docker client for getting
		// the exposed ports for the API server and NGINX.
		runtime = k3d_runtime.Docker

		// Running in docker indicates that we should configure k3s to connect to
		// the same docker network as the current container.
		runningInDocker = os.Getenv("K8S_USE_DOCKER_NETWORK") == "1"
	)

	if err := o.applyDefaults(); err != nil {
		return nil, fmt.Errorf("failed to apply defaults to options: %w", err)
	}

	k3dConfig := k3d_config.SimpleConfig{
		TypeMeta: k3d_cfgtypes.TypeMeta{
			Kind:       "Simple",
			APIVersion: config.DefaultConfigApiVersion,
		},
		ObjectMeta: k3d_cfgtypes.ObjectMeta{
			Name: randomClusterName(),
		},
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
	if runningInDocker {
		err := injectCurrentDockerNetwork(ctx, &k3dConfig)
		if err != nil {
			return nil, fmt.Errorf("could not connect k3d to current docker network: %w", err)
		}
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

	var (
		httpAddr      string
		apiServerAddr string
	)

	// If we're currently running inside of Docker, we can connect directly to
	// our container. Otherwise, we have to find what the bound host IPs are.
	if runningInDocker {
		httpAddr, apiServerAddr, err = clusterInternalAddrs(ctx, clusterConfig.Cluster)
	} else {
		httpAddr, apiServerAddr, err = loadBalancerExposedAddrs(ctx, clusterConfig.Cluster)
	}
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

	kubeClient, err := client.New(restCfg, client.Options{
		Scheme: o.Scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate client: %w", err)
	}

	return &Cluster{
		runtime:    runtime,
		k3dCluster: clusterConfig.Cluster,
		restConfig: restCfg,
		nginxAddr:  httpAddr,
		kubeClient: kubeClient,
	}, nil
}

// injectCurrentDockerNetwork reconfigures config to join the Docker network of
// the current container. Fails if the function is not being called from inside
// of a Docker container.
func injectCurrentDockerNetwork(ctx context.Context, config *k3d_config.SimpleConfig) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("could not get hostname: %w", err)
	}

	cli, err := k3d_docker.GetDockerClient()
	if err != nil {
		return fmt.Errorf("failed to get docker client: %w", err)
	}
	info, err := cli.ContainerInspect(ctx, hostname)
	if err != nil {
		return fmt.Errorf("failed to find current docker container: %w", err)
	}

	networks := make([]string, 0, len(info.NetworkSettings.Networks))
	for nw := range info.NetworkSettings.Networks {
		networks = append(networks, nw)
	}
	sort.Strings(networks)

	if len(networks) == 0 {
		return fmt.Errorf("no networks")
	}
	config.Network = networks[0]
	return nil
}

func randomClusterName() string {
	return "grafana-agent-e2e-" + rand.String(5)
}

func clusterInternalAddrs(ctx context.Context, cluster k3d_types.Cluster) (httpAddr, serverAddr string, err error) {
	var lb, server *k3d_types.Node
	for _, n := range cluster.Nodes {
		switch n.Role {
		case k3d_types.LoadBalancerRole:
			if lb == nil {
				lb = n
			}
		case k3d_types.ServerRole:
			if server == nil {
				server = n
			}
		}
	}
	if lb == nil {
		return "", "", fmt.Errorf("no loadbalancer node")
	} else if server == nil {
		return "", "", fmt.Errorf("no server node")
	}

	cli, err := k3d_docker.GetDockerClient()
	if err != nil {
		return "", "", fmt.Errorf("failed to get docker client: %w", err)
	}

	lbInfo, err := cli.ContainerInspect(ctx, lb.Name)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect loadbalancer: %w", err)
	} else if nw, found := lbInfo.NetworkSettings.Networks[cluster.Network.Name]; !found {
		return "", "", fmt.Errorf("loadbalancer not connected to expected network %q", cluster.Network.Name)
	} else {
		httpAddr = fmt.Sprintf("%s:80", nw.IPAddress)
	}

	serverInfo, err := cli.ContainerInspect(ctx, server.Name)
	if err != nil {
		return "", "", fmt.Errorf("failed to inspect worker: %w", err)
	} else if nw, found := serverInfo.NetworkSettings.Networks[cluster.Network.Name]; !found {
		return "", "", fmt.Errorf("worker not connected to expected network %q", cluster.Network.Name)
	} else {
		serverAddr = fmt.Sprintf("%s:6443", nw.IPAddress)
	}

	return httpAddr, serverAddr, nil
}

func loadBalancerExposedAddrs(ctx context.Context, cluster k3d_types.Cluster) (httpAddr, apiServerAddr string, err error) {
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

// Client returns the Kubernetes client for this Cluster. Client is handling
// objects registered to the Scheme passed to Options when creating the
// cluster.
func (c *Cluster) Client() client.Client {
	return c.kubeClient
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
func (c *Cluster) Stop() {
	err := k3d_client.ClusterDelete(context.Background(), c.runtime, &c.k3dCluster, k3d_types.ClusterDeleteOpts{})
	if err != nil {
		k3d_log.Log().Errorf("failed to shut down cluster, docker containers may have leaked: %s", err)
	}
}
