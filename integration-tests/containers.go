package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

const (
	mimirPort                      = "9009"
	prometheusMetricsGeneratorPort = "9001"
	agentPortOTLP                  = "4318"
	lokiPort                       = "3100"
)

func startContainer(ctx context.Context, containerRequest testcontainers.ContainerRequest) testcontainers.Container {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerRequest,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start container %s: %s", containerRequest.Image, err)
	}
	return container
}

func getAbsFilePath(path string) string {
	configPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Failed to get absolute path for config: %s", err)
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}
	return configPath
}

// bindSamePort uses the same port for the container and the host.
func bindSamePort(ip string, port string) func(*container.HostConfig) {
	return bindPort(ip, port, port)
}

func bindPort(ip string, containerPort string, hostPort string) func(*container.HostConfig) {
	return func(hostConfig *container.HostConfig) {
		hostConfig.PortBindings = nat.PortMap{
			nat.Port(containerPort + "/tcp"): []nat.PortBinding{
				{
					HostIP:   ip,
					HostPort: hostPort,
				},
			},
		}
	}
}

func createMimirContainer(image string) testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{mimirPort},
		Cmd:          []string{"/bin/mimir", "-config.file=/etc/mimir-config/mimir.yaml"},
		Mounts: []testcontainers.ContainerMount{
			{
				Source: testcontainers.GenericBindMountSource{HostPath: getAbsFilePath("./configs/mimir/mimir.yaml")},
				Target: "/etc/mimir-config/mimir.yaml",
			},
		},
		Networks: []string{
			networkName,
		},
		NetworkAliases: map[string][]string{
			networkName: {"mimir"},
		},
		HostConfigModifier: bindSamePort("0.0.0.0", mimirPort),
	}
}

func createPrometheusMetricGeneratorContainer(image string) testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{prometheusMetricsGeneratorPort},
		Networks: []string{
			networkName,
		},
		NetworkAliases: map[string][]string{
			networkName: {"prom-metrics-gen"},
		},
	}
}

func createLokiContainer(image string) testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{lokiPort},
		Networks: []string{
			networkName,
		},
		NetworkAliases: map[string][]string{
			networkName: {"loki"},
		},
		HostConfigModifier: bindSamePort("0.0.0.0", lokiPort),
	}
}

func createOTLPMetricsGeneratorContainer(image string) testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		Image: image,
		Networks: []string{
			networkName,
		},
		NetworkAliases: map[string][]string{
			networkName: {"otlp-metrics-gen"},
		},
	}
}

func createAgentContainer(image string, testName string) testcontainers.ContainerRequest {
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{agentPortOTLP},
		Cmd:          []string{"run", "/etc/agent/config.river"},
		Env:          map[string]string{"AGENT_MODE": "flow"},
		Mounts: []testcontainers.ContainerMount{
			{
				Source: testcontainers.GenericBindMountSource{HostPath: getAbsFilePath("tests/" + testName + "/config.river")},
				Target: "/etc/agent/config.river",
			},
		},
		Networks: []string{
			networkName,
		},
		NetworkAliases: map[string][]string{
			networkName: {testName},
		},
	}

	if testName == "read-log-file" {
		req.Mounts = append(req.Mounts, testcontainers.ContainerMount{
			Source: testcontainers.GenericBindMountSource{HostPath: getAbsFilePath("./tests/read-log-file/logs.txt")},
			Target: "/etc/agent/logs.txt",
		})
	}

	return req
}
