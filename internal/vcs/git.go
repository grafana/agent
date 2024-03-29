package vcs

import (
	"context"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-getter"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
)

type GitRepoOptions struct {
	Repository string
	Revision   string
	Auth       GitAuthConfig
}

// GitRepo manages a Git repository for the purposes of retrieving a file from
// it.
type GitRepo struct {
	opts        GitRepoOptions
	storagePath string
}

// NewGitRepo creates a new instance of a GitRepo, where the Git repository is
// managed at storagePath.
//
// If storagePath is empty on disk, NewGitRepo initializes GitRepo by cloning
// the repository. Otherwise, NewGitRepo will do a pull.
//
// After GitRepo is initialized, it checks out to the Revision specified in
// GitRepoOptions.
func NewGitRepo(ctx context.Context, storagePath string, opts GitRepoOptions) (*GitRepo, error) {
	err := pull(ctx, storagePath, opts)
	if err != nil {
		return nil, err
	}
	return &GitRepo{
		opts:        opts,
		storagePath: storagePath,
	}, nil
}

func pull(ctx context.Context, storagePath string, opts GitRepoOptions) error {
	// Create query string
	base := fmt.Sprintf("git::%s", opts.Repository)
	v := url.Values{}
	if opts.Revision != "" {
		v.Set("ref", opts.Revision)
	}
	if opts.Auth.SSHKey != nil {
		v.Set("sshkey", string(opts.Auth.SSHKey.Key))
	}
	query := base + "?" + v.Encode()
	err := getter.Get(storagePath, query, getter.WithContext(ctx))
	if err != nil {
		return DownloadFailedError{
			Repository: opts.Repository,
			Inner:      err,
		}
	}
	return nil
}

// Update updates the repository by pulling new content and re-checking out to
// latest version of Revision.
func (repo *GitRepo) Update(ctx context.Context) error {
	return pull(ctx, repo.storagePath, repo.opts)
}

// ReadFile returns a file from the repository specified by path.
func (repo *GitRepo) ReadFile(path string) ([]byte, error) {
	fullpath := filepath.Join(repo.storagePath, path)
	return os.ReadFile(fullpath)
}

// Stat returns info from the repository specified by path.
func (repo *GitRepo) Stat(path string) (fs.FileInfo, error) {
	fullpath := filepath.Join(repo.storagePath, path)
	return os.Stat(fullpath)
}

// ReadDir returns info about the content of the directory in the repository.
func (repo *GitRepo) ReadDir(path string) ([]fs.FileInfo, error) {
	fullpath := filepath.Join(repo.storagePath, path)
	dirEntries, err := os.ReadDir(fullpath)
	if err != nil {
		return nil, err
	}
	files := make([]os.FileInfo, 0)
	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
	}
	return os.Stat(fullpath)
}

// CurrentRevision returns the current revision of the repository (by SHA).
func (repo *GitRepo) CurrentRevision() (string, error) {
	ref, err := repo.repo.Head()
	if err != nil {
		return "", err
	}
	return ref.Hash().String(), nil
}

func findRevision(rev string, repo *git.Repository) (plumbing.Hash, error) {
	// Try looking for the revision in the following order:
	//
	// 1. Search by tag name.
	// 2. Search by remote ref name.
	// 3. Try to resolve the revision directly.

	if tagRef, err := repo.Tag(rev); err == nil {
		return tagRef.Hash(), nil
	}

	if remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", rev), true); err == nil {
		return remoteRef.Hash(), nil
	}

	if hash, err := repo.ResolveRevision(plumbing.Revision(rev)); err == nil {
		return *hash, nil
	}

	return plumbing.ZeroHash, plumbing.ErrReferenceNotFound
}
