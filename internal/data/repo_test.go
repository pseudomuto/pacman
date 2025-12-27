package data_test

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	. "github.com/pseudomuto/pacman/internal/data"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/enttest"
	"github.com/stretchr/testify/require"
)

func TestRepo_CreateArtifact(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	repo := NewRepo(client)

	t.Run("with versions", func(t *testing.T) {
		art := &ent.Artifact{
			Name:        "test with versions",
			Description: "An artifact with versions",
			Edges: ent.ArtifactEdges{
				Versions: []*ent.ArtifactVersion{
					{Version: "v1", URI: "/code/v1"},
					{Version: "v2", URI: "/code/v2"},
				},
			},
		}

		res, err := repo.CreateArtifact(t.Context(), art)
		require.NoError(t, err)

		require.NotZero(t, res.ID)
		require.Len(t, res.Edges.Versions, 2)
		for _, v := range res.Edges.Versions {
			require.NotZero(t, v.ID)
			require.Equal(t, res.ID, v.Edges.Artifact.ID)
		}
	})
}
