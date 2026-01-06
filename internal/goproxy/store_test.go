package goproxy_test

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pseudomuto/goproxy"
	"github.com/pseudomuto/pacman/internal/crypto"
	"github.com/pseudomuto/pacman/internal/ent/enttest"
	. "github.com/pseudomuto/pacman/internal/goproxy"
	"github.com/pseudomuto/pacman/internal/types"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	tree := client.SumDBTree.Create().
		SetName("test.example.com").
		SetSize(0).
		SetSignerKey(crypto.Secret("shh")).
		SetVerifierKey("good").
		SaveX(t.Context())

	assets := client.Asset.CreateBulk(
		client.Asset.Create().SetType(types.TextFile).SetURI("go.mod"),
		client.Asset.Create().SetType(types.Archive).SetURI("go.zip"),
	).SaveX(t.Context())

	mod := client.SumDBRecord.Create().
		AddAssetIDs(assets[0].ID, assets[1].ID).
		SetTree(tree).
		SetRecordID(1).
		SetPath("github.com/pseudomuto/where").
		SetVersion("v0.1.0").
		SetData([]byte("data right hur")).
		SaveX(t.Context())

	store := NewStore(client, tree.ID, nil)

	t.Run("Get", func(t *testing.T) {
		v, err := store.Get(t.Context(), mod.Path, mod.Version)
		require.NoError(t, err)
		require.Equal(t, mod.Path, v.Path)
		require.Equal(t, mod.Version, v.Version)
		require.Equal(t, mod.CreatedAt, v.CreatedAt)

		v, err = store.Get(t.Context(), mod.Path, mod.Version+"1")
		require.Nil(t, v)
		require.ErrorIs(t, err, goproxy.ErrModuleNotFound)
	})

	t.Run("GetVersions", func(t *testing.T) {
		vs, err := store.GetVersions(t.Context(), mod.Path)
		require.NoError(t, err)
		require.Len(t, vs, 1)
		require.Equal(t, mod.Path, vs[0].Path)
		require.Equal(t, mod.Version, vs[0].Version)
		require.Equal(t, mod.CreatedAt, vs[0].CreatedAt)
	})
}
