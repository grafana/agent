//go:build linux

package flow_test

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPullUpdating(t *testing.T) {
	// Previously we used fetch instead of pull, which would set the FETCH_HEAD but not HEAD
	// This caused changes not to propagate if there were changes, since HEAD was pinned to whatever it was on the initial download.
	// Switching to pull removes this problem at the expense of network bandwidth.
	// Tried switching to FETCH_HEAD but FETCH_HEAD is only set on fetch and not initial repo clone so we would need to
	// remember to always call fetch after clone.
	//
	// This test ensures we can pull the correct values down if they update no matter what, it works by creating a local
	// file based git repo then committing a file, running the component, then updating the file in the repo.
	testRepo := t.TempDir()

	main := `
import.git "testImport" {
	repository = "` + testRepo + `"
  	path = "math.river"
    pull_frequency = "5s"
}

testImport.add "cc" {
	a = 1
    b = 1
}
`
	// Create our git repository.
	runCommand(t, "git", "init", testRepo)

	// Change directory so the git commands work.
	os.Chdir(testRepo)

	// Add the file we want.
	math := filepath.Join(testRepo, "math.river")
	err := os.WriteFile(math, []byte(contents), 0666)
	require.NoError(t, err)

	runCommand(t, "git", "add", ".")

	runCommand(t, "git", "commit", "-m \"test\"")

	defer verifyNoGoroutineLeaks(t)
	ctrl, f := setup(t, main)
	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.Run(ctx)
	}()

	// Check for initial condition
	require.Eventually(t, func() bool {
		export := getExport[map[string]interface{}](t, ctrl, "", "testImport.add.cc")
		return export["sum"] == 2
	}, 3*time.Second, 10*time.Millisecond)

	err = os.WriteFile(math, []byte(contentsMore), 0666)
	require.NoError(t, err)

	runCommand(t, "git", "add", ".")

	commit2 := exec.Command("git", "commit", "-m \"test2\"")
	commit2.Dir = testRepo
	commit2.Run()

	// Check for final condition.
	require.Eventually(t, func() bool {
		export := getExport[map[string]interface{}](t, ctrl, "", "testImport.add.cc")
		return export["sum"] == 3
	}, 20*time.Second, 1*time.Millisecond)
}

func TestPullUpdatingFromBranch(t *testing.T) {

	testRepo := t.TempDir()

	main := `
import.git "testImport" {
	repository = "` + testRepo + `"
  	path = "math.river"
    pull_frequency = "1s"
    revision = "testor"
}

testImport.add "cc" {
	a = 1
    b = 1
}
`
	runCommand(t, "git", "init", testRepo)

	os.Chdir(testRepo)

	runCommand(t, "git", "checkout", "-b", "testor")

	math := filepath.Join(testRepo, "math.river")
	err := os.WriteFile(math, []byte(contents), 0666)
	require.NoError(t, err)

	runCommand(t, "git", "add", ".")

	runCommand(t, "git", "commit", "-m \"test\"")

	defer verifyNoGoroutineLeaks(t)
	ctrl, f := setup(t, main)
	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.Run(ctx)
	}()

	// Check for initial condition
	require.Eventually(t, func() bool {
		export := getExport[map[string]interface{}](t, ctrl, "", "testImport.add.cc")
		return export["sum"] == 2
	}, 3*time.Second, 10*time.Millisecond)

	err = os.WriteFile(math, []byte(contentsMore), 0666)
	require.NoError(t, err)

	runCommand(t, "git", "add", ".")

	runCommand(t, "git", "commit", "-m \"test2\"")

	// Check for final condition.
	require.Eventually(t, func() bool {
		export := getExport[map[string]interface{}](t, ctrl, "", "testImport.add.cc")
		return export["sum"] == 3
	}, 20*time.Second, 1*time.Millisecond)
}

func TestPullUpdatingFromHash(t *testing.T) {
	testRepo := t.TempDir()

	runCommand(t, "git", "init", testRepo)

	os.Chdir(testRepo)

	math := filepath.Join(testRepo, "math.river")
	err := os.WriteFile(math, []byte(contents), 0666)
	require.NoError(t, err)

	runCommand(t, "git", "add", ".")

	runCommand(t, "git", "commit", "-m \"test\"")

	getHead := exec.Command("git", "rev-parse", "HEAD")
	var stdBuffer bytes.Buffer
	getHead.Stdout = bufio.NewWriter(&stdBuffer)
	err = getHead.Run()
	require.NoError(t, err)
	hash := stdBuffer.String()
	hash = strings.TrimSpace(hash)

	main := `
import.git "testImport" {
	repository = "` + testRepo + `"
  	path = "math.river"
    pull_frequency = "10s"
    revision = "` + hash + `"
}

testImport.add "cc" {
	a = 1
    b = 1
}
`

	// After this update the sum should still be 2 and not 3 since it is pinned to the initial hash.
	require.NoError(t, err)

	err = os.WriteFile(math, []byte(contentsMore), 0666)
	require.NoError(t, err)

	runCommand(t, "git", "add", ".")

	runCommand(t, "git", "commit", "-m \"test2\"")

	defer verifyNoGoroutineLeaks(t)
	ctrl, f := setup(t, main)
	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.Run(ctx)
	}()

	// Check for final condition.
	require.Eventually(t, func() bool {
		export := getExport[map[string]interface{}](t, ctrl, "", "testImport.add.cc")
		return export["sum"] == 2
	}, 20*time.Second, 1*time.Millisecond)
}

func TestPullUpdatingFromTag(t *testing.T) {
	testRepo := t.TempDir()

	runCommand(t, "git", "init", testRepo)

	os.Chdir(testRepo)

	math := filepath.Join(testRepo, "math.river")
	err := os.WriteFile(math, []byte(contents), 0666)
	require.NoError(t, err)

	runCommand(t, "git", "add", ".")

	runCommand(t, "git", "commit", "-m \"test\"")

	runCommand(t, "git", "tag", "-a", "tagtest", "-m", "testtag")

	main := `
import.git "testImport" {
	repository = "` + testRepo + `"
  	path = "math.river"
    pull_frequency = "10s"
    revision = "tagtest"
}

testImport.add "cc" {
	a = 1
    b = 1
}
`

	// After this update the sum should still be 2 and not 3 since it is pinned to the tag.

	require.NoError(t, err)

	err = os.WriteFile(math, []byte(contentsMore), 0666)
	require.NoError(t, err)

	runCommand(t, "git", "add", ".")

	runCommand(t, "git", "commit", "-m \"test2\"")

	defer verifyNoGoroutineLeaks(t)

	ctrl, f := setup(t, main)
	err = ctrl.LoadSource(f, nil)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	defer func() {
		cancel()
		wg.Wait()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ctrl.Run(ctx)
	}()

	// Check for final condition.
	require.Eventually(t, func() bool {
		export := getExport[map[string]interface{}](t, ctrl, "", "testImport.add.cc")
		return export["sum"] == 2
	}, 20*time.Second, 1*time.Millisecond)
}

func runCommand(t *testing.T, command string, args ...string) {
	exe := exec.Command(command, args...)
	var stdErr bytes.Buffer
	exe.Stderr = bufio.NewWriter(&stdErr)
	err := exe.Run()
	errTxt := stdErr.String()
	if err != nil {
		t.Error(errTxt)
	}
	require.NoErrorf(t, err, "command %s %v failed", command, args)
}

const contents = `declare "add" {
    argument "a" {}
    argument "b" {}

    export "sum" {
        value = argument.a.value + argument.b.value
    }
}`

const contentsMore = `declare "add" {
    argument "a" {}
    argument "b" {}

    export "sum" {
        value = argument.a.value + argument.b.value + 1
    }
}`
