package sumdb_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/keyset"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pseudomuto/pacman/internal/crypto"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/enttest"
	. "github.com/pseudomuto/pacman/internal/sumdb"
	"github.com/pseudomuto/sumdb"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	kh, err := keyset.NewHandle(aead.AES256GCMKeyTemplate())
	require.NoError(t, err)

	cipher, err := aead.New(kh)
	require.NoError(t, err)
	crypto.SetCipher(cipher)

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	initTestDB(t, client)

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
		require.Equal(t, int64(3), recs[0].ID)
		require.Equal(t, "github.com/pseudomuto/where", recs[0].Path)
	})

	t.Run("AddRecord", func(t *testing.T) {
		rec := mkRecord([]string{
			"github.com/pseudomuto/protoc-gen-doc",
			"v1.5.1",
			"h1:Ah259kcrio7Ix1Rhb6u8FCaOkzf9qRBqXnvAufg061w=",
			"h1:XpMKYg6zkcpgfpCfQ8GcWBDRtRxOmMR5w7pz4Xo+dYM=",
		})

		id, err := store2.AddRecord(ctx, &sumdb.Record{
			Path:    rec.Path,
			Version: rec.Version,
			Data:    rec.Data,
		})

		require.NoError(t, err)
		require.Equal(t, int64(4), id)

		recs, err := store2.Records(ctx, 1, 10)
		require.NoError(t, err)
		require.Len(t, recs, 2)
	})

	t.Run("ReadHashes", func(t *testing.T) {
		hashes, err := store.ReadHashes(ctx, []int64{1, 2, 3})
		require.NoError(t, err)
		require.Len(t, hashes, 3)

		hashes, err = store2.ReadHashes(ctx, []int64{1, 3, 4})
		require.NoError(t, err)
		require.Len(t, hashes, 2)
	})
}

func initTestDB(t *testing.T, c *ent.Client) {
	t.Helper()

	modules := [][]string{
		{
			"github.com/pseudomuto/protoc-gen-doc",
			"v1.5.1",
			"h1:Ah259kcrio7Ix1Rhb6u8FCaOkzf9qRBqXnvAufg061w=", // zip hash
			"h1:XpMKYg6zkcpgfpCfQ8GcWBDRtRxOmMR5w7pz4Xo+dYM=", // mod hash
		},
		{
			"github.com/pseudomuto/where",
			"v0.1.0",
			"h1:NQtb1jgHaYMb6SwH3AnzF/Y/WnurmUInfbVJSMCWLrc=",
			"h1:ZNSWY7FiJI2r+6CMS2XKpUc0ah+N03cRok15QGpsHKw=",
		},
	}

	trees := []ent.SumDBTree{
		{
			Name:        "test.example.com",
			SignerKey:   crypto.Secret("secret"),
			VerifierKey: "notSecret",
			Edges: ent.SumDBTreeEdges{
				Hashes: []*ent.SumDBHash{
					{Index: 1, Hash: bytes.Repeat([]byte{1}, 32)},
					{Index: 2, Hash: bytes.Repeat([]byte{2}, 32)},
					{Index: 3, Hash: bytes.Repeat([]byte{3}, 32)},
				},
				Records: []*ent.SumDBRecord{
					mkRecord(modules[0]),
					mkRecord(modules[1]),
				},
			},
		},
		{
			Name:        "test2.example.com",
			SignerKey:   crypto.Secret("secret"),
			VerifierKey: "notSecret",
			Edges: ent.SumDBTreeEdges{
				Hashes: []*ent.SumDBHash{
					{Index: 1, Hash: bytes.Repeat([]byte{4}, 32)},
					{Index: 3, Hash: bytes.Repeat([]byte{5}, 32)},
				},
				Records: []*ent.SumDBRecord{
					mkRecord(modules[1]),
				},
			},
		},
	}

	for _, tree := range trees {
		st := c.SumDBTree.Create().
			SetName(tree.Name).
			SetSignerKey(tree.SignerKey).
			SetVerifierKey(tree.VerifierKey).
			SaveX(t.Context())

		hashes := make([]*ent.SumDBHashCreate, len(tree.Edges.Hashes))
		for i := range tree.Edges.Hashes {
			hashes[i] = c.SumDBHash.Create().
				SetTree(st).
				SetIndex(tree.Edges.Hashes[i].Index).
				SetHash(tree.Edges.Hashes[i].Hash)
		}

		c.SumDBHash.CreateBulk(hashes...).SaveX(t.Context())

		records := make([]*ent.SumDBRecordCreate, len(tree.Edges.Records))
		for i, rec := range tree.Edges.Records {
			records[i] = c.SumDBRecord.Create().
				SetPath(rec.Path).
				SetVersion(rec.Version).
				SetData(rec.Data).
				SetTree(st)
		}

		c.SumDBRecord.CreateBulk(records...).SaveX(t.Context())
	}
}

func mkRecord(vals []string) *ent.SumDBRecord {
	return &ent.SumDBRecord{
		Path:    vals[0],
		Version: vals[1],
		Data: fmt.Appendf(
			nil,
			"%s %s %s\n%s %s/go.mod %s\n",
			vals[0],
			vals[1],
			vals[2],
			vals[0],
			vals[1],
			vals[3],
		),
	}
}
