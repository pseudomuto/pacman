package vcs

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pseudomuto/pacman/internal/types"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type (
	GitLab struct {
		repo GitLabRepo
	}

	GitLabRepo interface {
		Archive(any, *gitlab.ArchiveOptions, ...gitlab.RequestOptionFunc) ([]byte, *gitlab.Response, error)
	}
)

func NewGitLab(repo GitLabRepo) *GitLab {
	return &GitLab{repo: repo}
}

func (g *GitLab) Name() string {
	return "gitlab"
}

func (g *GitLab) FetchArchive(w io.Writer, repo string, opts types.VCSOptions) error {
	data, _, err := g.repo.Archive(
		repo,
		&gitlab.ArchiveOptions{
			Format: gitlab.Ptr("tar.gz"),
			Path:   &opts.Dir,
			SHA:    &opts.Ref,
		},
	)
	if err != nil {
		return fmt.Errorf("failed fetching VCS archive: %s:%s, %w", repo, opts.Dir, err)
	}

	_, err = io.Copy(w, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to write VCS archive: %s:%s, %w", repo, opts.Dir, err)
	}

	return nil
}
