package sumdb_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/crypto"
	"github.com/pseudomuto/pacman/internal/ent/enttest"
	. "github.com/pseudomuto/pacman/internal/sumdb"
	"github.com/pseudomuto/sumdb"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func TestSumDB(t *testing.T) {
	t.Parallel()

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	skey, vkey, err := sumdb.GenerateKeys("test.sumdb.com")
	require.NoError(t, err)

	tree := client.SumDBTree.Create().
		SetName("test.sumdb.com").
		SetSize(0).
		SetSignerKey(crypto.Secret(skey)).
		SetVerifierKey(vkey).
		SaveX(t.Context())

	r, err := recorder.New(filepath.Join("testdata", "cassettes", "goproxy"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, r.Stop()) })

	db, err := NewSumDB(tree, client, sumdb.WithHTTPClient(r.GetDefaultClient()))
	require.NoError(t, err)

	svr := gin.Default()
	db.RegisterRoutes(svr)

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(
		t.Context(),
		"GET",
		"/sumdb/test.sumdb.com/lookup/github.com/pseudomuto/where@v0.1.0",
		nil,
	)
	svr.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequestWithContext(
		t.Context(),
		"GET",
		"/sumdb/test.sumdb.com/lookup/github.com/pseudomuto/protoc-gen-doc@v1.5.1",
		nil,
	)
	svr.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
}
