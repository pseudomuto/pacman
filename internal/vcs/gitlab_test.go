package vcs_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pseudomuto/pacman/internal/types"
	. "github.com/pseudomuto/pacman/internal/vcs"
	"github.com/stretchr/testify/require"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	gitlabtesting "gitlab.com/gitlab-org/api/client-go/testing"
)

func TestGitLab_FetchArchive(t *testing.T) {
	client := gitlabtesting.NewTestClient(t)
	client.MockRepositories.EXPECT().
		Archive(
			"test/repo",
			&gitlab.ArchiveOptions{
				Format: gitlab.Ptr("tar.gz"),
				Path:   gitlab.Ptr("some/sub/dir"),
				SHA:    gitlab.Ptr("c12345d"),
			},
		).
		Return([]byte("testdata"), nil, nil)

	buf := new(bytes.Buffer)

	gl := NewGitLab(client.Repositories)
	require.NoError(t, gl.FetchArchive(buf, "test/repo", types.VCSOptions{
		Dir: "some/sub/dir",
		Ref: "c12345d",
	}))

	require.Equal(t, "testdata", buf.String())

	t.Run("on VCS failure", func(t *testing.T) {
		client.MockRepositories.EXPECT().
			Archive(
				"test/repo",
				&gitlab.ArchiveOptions{
					Format: gitlab.Ptr("tar.gz"),
					Path:   gitlab.Ptr("some/sub/dir"),
					SHA:    gitlab.Ptr("c12345d"),
				},
			).
			Return(nil, nil, fmt.Errorf("Error: %s", "boom"))

		err := gl.FetchArchive(buf, "test/repo", types.VCSOptions{
			Dir: "some/sub/dir",
			Ref: "c12345d",
		})
		require.ErrorContains(t, err, "Error: boom")
	})
}
