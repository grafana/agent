package git_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	git_component "github.com/grafana/agent/component/remote/git"
	"github.com/grafana/agent/pkg/flow/componenttest"
	"github.com/grafana/agent/pkg/util"
	"github.com/grafana/river"
	"github.com/stretchr/testify/require"
)

func Test_(t *testing.T) {
	ctx := componenttest.TestContext(t)
	origRepo := initRepository(t)

	expectedContent := "Hello, world!"

	// Write a file into the repository and commit it.
	{
		err := origRepo.WriteFile("a.txt", []byte(expectedContent))
		require.NoError(t, err)

		_, err = origRepo.Worktree.Add(".")
		require.NoError(t, err)

		_, err = origRepo.Worktree.Commit("initial commit", &git.CommitOptions{})
		require.NoError(t, err)
	}

	origRef, err := origRepo.CurrentRef()
	require.NoError(t, err)

	ctrl, err := componenttest.NewControllerFromID(util.TestLogger(t), "remote.git")
	require.NoError(t, err)

	cfg := fmt.Sprintf(`
		repository = "%s"
		revision = "%s"
		path = "%s"
		pull_frequency = "50ms"
	`, origRepo.Directory, origRef, "a.txt")
	var args git_component.Arguments
	require.NoError(t, river.Unmarshal([]byte(cfg), &args))

	go func() {
		err := ctrl.Run(ctx, args)
		require.NoError(t, err)
	}()

	require.NoError(t, ctrl.WaitExports(time.Second), "component didn't update exports")

	require.Eventually(t, func() bool {
		return ctrl.Exports().(git_component.Exports).Content.Value == expectedContent
	}, 100*time.Millisecond, 10*time.Millisecond, "timed out waiting for git content")

	expectedContent = "See you later!"

	// Update the file.
	{
		err := origRepo.WriteFile("a.txt", []byte(expectedContent))
		require.NoError(t, err)

		_, err = origRepo.Worktree.Add(".")
		require.NoError(t, err)

		_, err = origRepo.Worktree.Commit("commit 2", &git.CommitOptions{})
		require.NoError(t, err)
	}

	require.Eventually(t, func() bool {
		return ctrl.Exports().(git_component.Exports).Content.Value == expectedContent
	}, 100*time.Millisecond, 10*time.Millisecond, "timed out waiting for git content")
}

type testRepository struct {
	Directory string
	Repo      *git.Repository
	Worktree  *git.Worktree
}

func (repo *testRepository) CurrentRef() (string, error) {
	ref, err := repo.Repo.Head()
	if err != nil {
		return "", nil
	}
	return ref.Name().Short(), nil
}

func (repo *testRepository) WriteFile(path string, contents []byte) error {
	f, err := repo.Worktree.Filesystem.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(contents)
	return err
}

// initRepository creates a new, uninitialized Git repository in a temporary
// directory. The Git repository is deleted when the test exits.
func initRepository(t *testing.T) *testRepository {
	t.Helper()

	worktreeDir := t.TempDir()
	repo, err := git.PlainInit(worktreeDir, false)
	require.NoError(t, err)

	// Create a placeholder config for the repo.
	{
		cfg := config.NewConfig()
		cfg.User.Name = "Go test"
		cfg.User.Email = "go-test@example.com"

		err := repo.SetConfig(cfg)
		require.NoError(t, err)
	}

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	return &testRepository{
		Directory: worktreeDir,
		Repo:      repo,
		Worktree:  worktree,
	}
}
