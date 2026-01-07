package goproxy_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pseudomuto/pacman/internal/ent/enttest"
	"github.com/pseudomuto/pacman/internal/ent/schema"
	. "github.com/pseudomuto/pacman/internal/goproxy"
	"github.com/pseudomuto/pacman/internal/types"
	"github.com/stretchr/testify/require"
)

func TestUpstreamProxy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		url   string
		code  int
		cType string
		body  string
	}{
		{
			name:  "module file",
			url:   "go.example.com/module/@v/v0.1.0.mod",
			code:  http.StatusOK,
			cType: "text/plain; charset=utf-8",
			body:  "gs://test-bucket/go.mod",
		},
		{
			name:  "zip file",
			url:   "go.example.com/module/@v/v0.1.0.zip",
			code:  http.StatusOK,
			cType: "application/octet-stream",
			body:  "gs://test-bucket/go.zip",
		},
		{
			name: "upstream call",
			url:  "go.example.com/module/@v/v0.1.1.zip",
			code: http.StatusOK,
			body: "proxied: /go.example.com/module/@v/v0.1.1.zip",
		},
		{
			name: "malformed module",
			url:  "thing!/@v/v0.1.0.mod",
			code: http.StatusBadRequest,
		},
		{
			name: "malformed version",
			url:  "go.example.com/module/@v/v0.1.0!.zip",
			code: http.StatusBadRequest,
		},
		{
			name: "invalid extension",
			url:  "go.example.com/module/@v/v0.1.0.info",
			code: http.StatusNotFound,
		},
	}

	client := enttest.Open(t, "sqlite3", "file:ent?mode=memory&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	client.Archive.Create().
		SetAssets([]schema.AssetURL{
			{
				Type: types.TextFile,
				URL:  "gs://test-bucket/go.mod",
			},
			{
				Type: types.Archive,
				URL:  "gs://test-bucket/go.zip",
			},
		}).
		SetCoordinate("go.example.com/module@v0.1.0").
		SetType(types.GoModule).
		SaveX(t.Context())

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "proxied: %s", r.URL.Path)
	}))
	t.Cleanup(svr.Close)

	up := NewUpstreamProxyWithHost(
		client,
		ReaderFunc(func(ctx context.Context, w io.Writer, s string) error {
			fmt.Fprint(w, s)
			return nil
		}),
		svr.URL,
	)

	req := func(p string) *http.Request {
		return httptest.NewRequestWithContext(
			t.Context(),
			http.MethodGet,
			"/goproxy/proxy.golang.org/"+p,
			nil,
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			up.ServeHTTP(w, req(tt.url))
			require.Equal(t, tt.code, w.Code)

			if tt.cType != "" {
				require.Equal(t, tt.cType, w.Header().Get("Content-Type"))
			}

			if tt.body != "" {
				require.Equal(t, tt.body, w.Body.String())
			}
		})
	}
}
