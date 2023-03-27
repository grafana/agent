//go:build !nonetwork && !nodocker && !race && packaging
// +build !nonetwork,!nodocker,!race,packaging

package packaging_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

// TestLinuxPackages runs the entire test suite for the Linux packages.
func TestLinuxPackages(t *testing.T) {
	fmt.Println("Building packages (this may take a while...)")
	buildPackages(t)

	dockerPool, err := dockertest.NewPool("")
	require.NoError(t, err)

	tt := []struct {
		name string
		f    func(*testing.T, Environment)
	}{
		{"install package", EnvironmentTestInstall},
		{"ensure existing config doesn't get overridden", EnvironmentTestConfigPersistence},
		{"test data folder permissions", EnvironmentTestDataFolderPermissions},

		// TODO: a test to verify that the systemd service works would be nice, but not
		// required.
		//
		// An implementation of the test would have to consider what host platforms it
		// works on; bind mounting /sys/fs/cgroup and using the host systemd wouldn't
		// work on macOS or Windows.
	}

	for _, tc := range tt {
		t.Run(tc.name+"/rpm", func(t *testing.T) {
			tc.f(t, RPMEnvironment(t, dockerPool))
		})
		t.Run(tc.name+"/deb", func(t *testing.T) {
			tc.f(t, DEBEnvironment(t, dockerPool))
		})
	}
}

func buildPackages(t *testing.T) {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	root, err := filepath.Abs(filepath.Join(wd, "..", ".."))
	require.NoError(t, err)

	cmd := exec.Command("make", fmt.Sprintf("dist-agent-packages-%s", runtime.GOARCH))
	cmd.Env = append(
		os.Environ(),
		"VERSION=v0.0.0",
		"DOCKER_OPTS=",
	)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run())
}

func EnvironmentTestInstall(t *testing.T, env Environment) {
	res := env.Install()
	require.Equal(t, 0, res.ExitCode, "installing failed")

	res = env.ExecScript(`[ -f /usr/bin/grafana-agent ]`)
	require.Equal(t, 0, res.ExitCode, "expected grafana-agent to be installed")
	res = env.ExecScript(`[ -f /usr/bin/grafana-agentctl ]`)
	require.Equal(t, 0, res.ExitCode, "expected grafana-agentctl to be installed")
	res = env.ExecScript(`[ -f /etc/grafana-agent.yaml ]`)
	require.Equal(t, 0, res.ExitCode, "expected grafana agent configuration file to exist")

	res = env.Uninstall()
	require.Equal(t, 0, res.ExitCode, "uninstalling failed")

	res = env.ExecScript(`[ -f /usr/bin/grafana-agent ]`)
	require.Equal(t, 1, res.ExitCode, "expected grafana-agent to be uninstalled")
	res = env.ExecScript(`[ -f /usr/bin/grafana-agentctl ]`)
	require.Equal(t, 1, res.ExitCode, "expected grafana-agentctl to be uninstalled")
	// NOTE(rfratto): we don't check for what happens to the config file here,
	// sicne the behavior is inconsistent: rpm uninstalls it, but deb doesn't.
}

func EnvironmentTestConfigPersistence(t *testing.T, env Environment) {
	res := env.ExecScript(`echo -n "keepalive" > /etc/grafana-agent.yaml`)
	require.Equal(t, 0, res.ExitCode, "failed to write config file")

	res = env.Install()
	require.Equal(t, 0, res.ExitCode, "installation failed")

	res = env.ExecScript(`cat /etc/grafana-agent.yaml`)
	require.Equal(t, "keepalive", res.Stdout, "Expected existing file to not be overridden")
}

func EnvironmentTestDataFolderPermissions(t *testing.T, env Environment) {
	// Installing should create /var/lib/grafana-agent, assign it to the
	// grafana-agent user and group, and set its permissions to 0770.
	res := env.Install()
	require.Equal(t, 0, res.ExitCode, "installation failed")

	res = env.ExecScript(`[ -d /var/lib/grafana-agent ]`)
	require.Equal(t, 0, res.ExitCode, "Expected /var/lib/grafana-agent to have been created during install")

	res = env.ExecScript(`stat -c '%a:%U:%G' /var/lib/grafana-agent`)
	require.Equal(t, "770:grafana-agent:grafana-agent\n", res.Stdout, "wrong permissions for data folder")
	require.Equal(t, 0, res.ExitCode, "stat'ing data folder failed")
}

type Environment struct {
	Install    func() ExecResult
	Uninstall  func() ExecResult
	ExecScript func(string) ExecResult
}

type ExecResult struct {
	Stdout, Stderr string
	ExitCode       int
}

// RPMEnvironment creates an Environment to install the agent RPM against.
func RPMEnvironment(t *testing.T, pool *dockertest.Pool) Environment {
	t.Helper()

	container := environmentContainer(
		t,
		pool,
		"../testdata/centos-systemd.Dockerfile",
		"agent-test-centos-systemd",
		fmt.Sprintf("../../dist/grafana-agent-0.0.0-1.%s.rpm", runtime.GOARCH),
	)

	return Environment{
		Install: func() ExecResult {
			filename := fmt.Sprintf("/tmp/grafana-agent-0.0.0-1.%s.rpm", runtime.GOARCH)
			return containerExec(t, container, "rpm", "-i", filename)
		},
		Uninstall: func() ExecResult {
			return containerExec(t, container, "rpm", "-e", "grafana-agent")
		},
		ExecScript: func(script string) ExecResult {
			return containerExec(t, container, "/bin/bash", "-c", script)
		},
	}
}

// DEBEnvironment creates an Environment to install the agent RPM against.
func DEBEnvironment(t *testing.T, pool *dockertest.Pool) Environment {
	t.Helper()

	container := environmentContainer(
		t,
		pool,
		"../testdata/debian-systemd.Dockerfile",
		"agent-test-debian-systemd",
		fmt.Sprintf("../../dist/grafana-agent-0.0.0-1.%s.deb", runtime.GOARCH),
	)

	return Environment{
		Install: func() ExecResult {
			filename := fmt.Sprintf("/tmp/grafana-agent-0.0.0-1.%s.deb", runtime.GOARCH)
			return containerExec(t, container, "dpkg", "--force-confold", "-i", filename)
		},
		Uninstall: func() ExecResult {
			return containerExec(t, container, "dpkg", "-r", "grafana-agent")
		},
		ExecScript: func(script string) ExecResult {
			return containerExec(t, container, "/bin/bash", "-c", script)
		},
	}
}

func environmentContainer(t *testing.T, pool *dockertest.Pool, dockerfile string, name string, packagePath string) *dockertest.Resource {
	t.Helper()

	container, err := pool.BuildAndRunWithOptions(
		dockerfile,
		&dockertest.RunOptions{
			Name:       name,
			Entrypoint: []string{"/bin/bash"},
			Tty:        true,
			Mounts:     []string{"/sys/fs/cgroup:/sys/fs/cgroup:ro"},
			PortBindings: map[docker.Port][]docker.PortBinding{
				"9009/tcp": {{HostIP: "0.0.0.0", HostPort: "0"}},
			},
		},
		func(hc *docker.HostConfig) {
			hc.Tmpfs = map[string]string{
				"/run":      "rw",
				"/run/lock": "rw",
			}
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = container.Close()
	})

	packageFile, err := buildTar(packagePath)
	require.NoError(t, err)
	err = pool.Client.UploadToContainer(container.Container.ID, docker.UploadToContainerOptions{
		InputStream: packageFile,
		Path:        "/tmp",
	})
	require.NoError(t, err)

	return container
}

func buildTar(path string) (io.Reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	defer w.Close()

	err = w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     filepath.Base(path),
		Size:     fi.Size(),
		ModTime:  fi.ModTime(),
		Mode:     0600,
	})
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(w, f)
	return &buf, err
}

func containerExec(t *testing.T, res *dockertest.Resource, cmd ...string) ExecResult {
	t.Helper()

	var stdout, stderr bytes.Buffer

	exitCode, err := res.Exec(cmd, dockertest.ExecOptions{
		StdOut: &stdout,
		StdErr: &stderr,
	})
	require.NoError(t, err)

	return ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}
