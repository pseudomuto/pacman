package goproxy

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pseudomuto/pacman/internal/ent"
	"github.com/pseudomuto/pacman/internal/ent/archive"
	"github.com/pseudomuto/pacman/internal/ent/schema"
	"github.com/pseudomuto/pacman/internal/types"
	"golang.org/x/mod/module"
)

const defaultProxy = "https://proxy.golang.org"

// UpstreamProxy is used by our SumDB server to fetch unknown modules.
//
// This proxy will check for published Archives in the archives table before proxying upstream to proxy.golang.org. When
// the archive exists, the SumDB server will receive the values from the archive and use those to populate SumDBRecords
// for the appropriate tree. After which, it will never ask again for the particular path/version.
//
// When not found, the request will be proxied upstream to the supplied goProxy handler. The only exception here is if
// the path is specified in Config.Go.NoSumPatterns, in which case no upstream proxying will be done.
//
// NB: This should not be in your GOPROXY list. This is an implementation detail for pacman and only responds for .mod
// and .zip endpoints.
type UpstreamProxy struct {
	prefix string
	db     *ent.Client
	rp     http.Handler
	rdr    Reader
}

func NewUpstreamProxy(db *ent.Client, rdr Reader) *UpstreamProxy {
	return NewUpstreamProxyWithHost(db, rdr, defaultProxy)
}

func NewUpstreamProxyWithHost(db *ent.Client, rdr Reader, host string) *UpstreamProxy {
	url, _ := url.Parse(host)
	rp := httputil.NewSingleHostReverseProxy(url)
	rp.Director = func(req *http.Request) {
		req.Host = url.Host
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/goproxy/proxy.golang.org")
	}

	return &UpstreamProxy{
		prefix: "/goproxy/proxy.golang.org",
		db:     db,
		rp:     rp,
		rdr:    rdr,
	}
}

func (s *UpstreamProxy) RegisterRoutes(g *gin.Engine) {
	g.GET(s.prefix+"/*action", gin.WrapH(s))
}

func (s *UpstreamProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// NB: We only need to respond to .mod and .zip requests from the sumdb
	if !slices.Contains([]string{".mod", ".zip"}, path.Ext(req.URL.Path)) {
		http.NotFound(w, req)
		return
	}

	mod, err := parseModule(strings.TrimPrefix(req.URL.Path, s.prefix))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	arch, err := s.db.Archive.Query().
		Where(
			archive.TypeEQ(types.GoModule),
			archive.Coordinate(mod.String()),
		).
		Only(req.Context())
	if err != nil {
		var nfe *ent.NotFoundError
		if errors.As(err, &nfe) {
			// TODO: check for NoSumPatterns

			// Fallback to proxying upstream.
			s.rp.ServeHTTP(w, req)
			return
		}

		// bad error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	at := types.TextFile
	ct := "text/plain; charset=utf-8"
	if strings.HasSuffix(req.URL.Path, ".zip") {
		at = types.Archive
		ct = "application/octet-stream"
	}

	idx := slices.IndexFunc(arch.Assets, func(e schema.AssetURL) bool { return e.Type == at })
	if idx < 0 {
		http.Error(w, "asset not found", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", ct)
	if err := s.rdr.Read(req.Context(), w, arch.Assets[idx].URL); err != nil {
		http.Error(w, "failed writing asset: %v"+err.Error(), http.StatusInternalServerError)
		return
	}
}

func parseModule(path string) (module.Version, error) {
	// <prefix>/<mod>/@v/<version>.<ext>

	var v module.Version
	mPath, route, ok := strings.Cut(path, "/@")
	if !ok {
		return v, errors.New("unknown route: " + path)
	}

	modPath, err := module.UnescapePath(strings.TrimPrefix(mPath, "/"))
	if err != nil {
		return v, fmt.Errorf(
			"invalid module: %s, %w",
			mPath,
			err,
		)
	}

	ver := strings.TrimPrefix(route, "v/")
	modVersion := ver
	if pos := strings.LastIndexByte(ver, '.'); pos > -1 {
		modVersion = ver[:pos]
	}

	modVersion, err = module.UnescapeVersion(modVersion)
	if err != nil {
		return v, fmt.Errorf(
			"invalid version: %s, %w",
			ver,
			err,
		)
	}

	v.Path = modPath
	v.Version = modVersion
	return v, nil
}
