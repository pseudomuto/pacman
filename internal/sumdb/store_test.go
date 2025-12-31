package sumdb_test

import (
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pseudomuto/pacman/internal/ent/enttest"
	. "github.com/pseudomuto/pacman/internal/sumdb"
	"github.com/pseudomuto/sumdb"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	loadFixture(t, client)

	ctx := t.Context()
	tree1, _ := client.SumDBTree.Get(ctx, 1)
	tree2, _ := client.SumDBTree.Get(ctx, 2)

	store := NewStore(tree1, client)
	store2 := NewStore(tree2, client)

	t.Run("RecordID", func(t *testing.T) {
		id, err := store.RecordID(ctx, "github.com/pseudomuto/protoc-gen-doc", "v1.5.1")
		require.Equal(t, int64(1), id)
		require.NoError(t, err)

		id, err = store2.RecordID(ctx, "github.com/pseudomuto/protoc-gen-doc", "v1.5.1")
		require.Zero(t, id)
		require.ErrorIs(t, err, sumdb.ErrNotFound)
	})

	t.Run("Records", func(t *testing.T) {
		recs, err := store.Records(ctx, 1, 10)
		require.NoError(t, err)
		require.Len(t, recs, 2)
		require.Equal(t, int64(1), recs[0].ID)
		require.Equal(t, "github.com/pseudomuto/protoc-gen-doc", recs[0].Path)
		require.Equal(t, int64(2), recs[1].ID)
		require.Equal(t, "github.com/pseudomuto/where", recs[1].Path)

		recs, err = store2.Records(ctx, 1, 10)
		require.NoError(t, err)
		require.Len(t, recs, 1)
		require.Equal(t, int64(1), recs[0].ID)
		require.Equal(t, "github.com/pseudomuto/where", recs[0].Path)
	})

	t.Run("AddRecord", func(t *testing.T) {
		id, err := store2.AddRecord(ctx, &sumdb.Record{
			Path:    "github.com/pseudomuto/protoc-gen-doc",
			Version: "v1.5.1",
			Data: fmt.Appendf(
				nil,
				"%s %s %s\n%s %s/go.mod %s\n",
				"github.com/pseudomuto/protoc-gen-doc",
				"v1.5.1",
				"h1:Ah259kcrio7Ix1Rhb6u8FCaOkzf9qRBqXnvAufg061w=",
				"github.com/pseudomuto/protoc-gen-doc",
				"v1.5.1",
				"h1:XpMKYg6zkcpgfpCfQ8GcWBDRtRxOmMR5w7pz4Xo+dYM=",
			),
		})

		require.NoError(t, err)
		require.Equal(t, int64(2), id)

		recs, err := store2.Records(ctx, 1, 10)
		require.NoError(t, err)
		require.Len(t, recs, 2)
	})

	t.Run("ReadHashes", func(t *testing.T) {
		hashes, err := store.ReadHashes(ctx, []int64{0, 1, 2, 3})
		require.NoError(t, err)
		require.Len(t, hashes, 2)

		hashes, err = store2.ReadHashes(ctx, []int64{0, 1, 3, 4})
		require.NoError(t, err)
		require.Len(t, hashes, 1)
	})
}
