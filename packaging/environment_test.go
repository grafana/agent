//go:build !nonetwork && !nodocker && !race && packaging
// +build !nonetwork,!nodocker,!race,packaging

package packaging_test

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

type Environment struct {
	Install    func() ExecResult
	Uninstall  func() ExecResult
	ExecScript func(string) ExecResult
}

type ExecResult struct {
	Stdout, Stderr string
	ExitCode       int
}

// RPMEnvironment creates an Environment to install an RPM against.
func RPMEnvironment(t *testing.T, packageName string, pool *dockertest.Pool) Environment {
	t.Helper()

	container := environmentContainer(
		t,
		pool,
		"testdata/centos-systemd.Dockerfile",
		packageName+"-test-centos-systemd",
		fmt.Sprintf("../dist/%s-0.0.0-1.%s.rpm", packageName, runtime.GOARCH),
	)

	return Environment{
		Install: func() ExecResult {
			filename := fmt.Sprintf("/tmp/%s-0.0.0-1.%s.rpm", packageName, runtime.GOARCH)
			return containerExec(t, container, "rpm", "-i", filename)
		},
		Uninstall: func() ExecResult {
			return containerExec(t, container, "rpm", "-e", packageName)
		},
		ExecScript: func(script string) ExecResult {
			return containerExec(t, container, "/bin/bash", "-c", script)
		},
	}
}

// DEBEnvironment creates an Environment to install a DEB against.
func DEBEnvironment(t *testing.T, packageName string, pool *dockertest.Pool) Environment {
	t.Helper()

	container := environmentContainer(
		t,
		pool,
		"testdata/debian-systemd.Dockerfile",
		packageName+"-test-debian-systemd",
		fmt.Sprintf("../dist/%s-0.0.0-1.%s.deb", packageName, runtime.GOARCH),
	)

	return Environment{
		Install: func() ExecResult {
			filename := fmt.Sprintf("/tmp/%s-0.0.0-1.%s.deb", packageName, runtime.GOARCH)
			return containerExec(t, container, "dpkg", "--force-confold", "-i", filename)
		},
		Uninstall: func() ExecResult {
			return containerExec(t, container, "dpkg", "-r", packageName)
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
	if err != nil {
		return nil, err
	}
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
