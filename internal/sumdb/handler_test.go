package sumdb_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/ent/enttest"
	. "github.com/pseudomuto/pacman/internal/sumdb"
	"github.com/pseudomuto/pacman/internal/sumdb/api"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	loadFixture(t, client)
	h := NewHandler(client)

	t.Run("ListTrees", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		h.ListTrees(ctx)
		var trees api.TreeList
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &trees), w.Body.String)
		require.Len(t, trees, 2)
	})

	t.Run("ListTreeHashes", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		h.ListTreeHashes(ctx, "test2.example.com")
		var hashes api.HashList
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &hashes), w.Body.String())
		require.Len(t, hashes, 1)
	})

	t.Run("ListTreeRecords", func(t *testing.T) {
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		h.ListTreeRecords(ctx, "test.example.com")
		var records api.RecordList
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &records), w.Body.String())
		require.Len(t, records, 2)
	})
}
