package sumdb_test

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"slices"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/pseudomuto/pacman/internal/crypto"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/sumdb"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/fixtures.yaml
var fixtureBytes []byte

type (
	fixture struct {
		Trees []treeFixture
	}

	treeFixture struct {
		Name    string
		Records []recordFixture
	}

	recordFixture struct {
		ID      int64
		Path    string
		Version string
		Zip     string
		Mod     string
	}
)

func loadFixture(t *testing.T, db *ent.Client) {
	t.Helper()

	var fix fixture
	require.NoError(t, yaml.Unmarshal(fixtureBytes, &fix))

	for _, tree := range fix.Trees {
		skey, vkey, err := sumdb.GenerateKeys(tree.Name)
		require.NoError(t, err)

		maxID := slices.MaxFunc(tree.Records, func(a, b recordFixture) int {
			return cmp.Compare(a.ID, b.ID)
		})

		st := db.SumDBTree.Create().
			SetName(tree.Name).
			SetSize(maxID.ID + 1).
			SetSignerKey(crypto.Secret(skey)).
			SetVerifierKey(vkey).
			SaveX(t.Context())

		hashes := make([]*ent.SumDBHashCreate, len(tree.Records))
		records := make([]*ent.SumDBRecordCreate, len(tree.Records))
		for i, rec := range tree.Records {
			hashes[i] = db.SumDBHash.Create().
				SetIndex(int64(i)).
				SetHash(bytes.Repeat([]byte{byte(i)}, 32)).
				SetTree(st)

			records[i] = db.SumDBRecord.Create().
				SetRecordID(rec.ID).
				SetPath(rec.Path).
				SetVersion(rec.Version).
				SetData(fmt.Appendf(
					nil,
					"%s %s %s\n%s %s/go.mod %s\n",
					rec.Path,
					rec.Version,
					rec.Zip,
					rec.Path,
					rec.Version,
					rec.Mod,
				)).
				SetTree(st)
		}

		db.SumDBHash.CreateBulk(hashes...).SaveX(t.Context())
		db.SumDBRecord.CreateBulk(records...).SaveX(t.Context())
	}
}
