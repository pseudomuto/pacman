package sumdb

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md
const sumDBURL = "https://sum.golang.org"

type SumDBProxy struct {
	h gin.HandlerFunc
}

func NewSumDBProxy() (*SumDBProxy, error) {
	url, err := url.Parse(sumDBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SumDB URL: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Director = func(r *http.Request) {
		r.Host = url.Host
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
	}

	return &SumDBProxy{
		h: gin.WrapH(proxy),
	}, nil
}

func (s *SumDBProxy) RegisterRoutes(g *gin.Engine) {
	group := g.Group("/sumdb/sum.golang.org")
	group.GET("/supported", ok)
	group.GET("/lookup/*path", s.handler)
}

func ok(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{})
}

func (s *SumDBProxy) handler(ctx *gin.Context) {
	path := strings.TrimPrefix(ctx.Param("path"), "/")
	idx := strings.LastIndex(path, "@")
	if idx == -1 {
		ctx.JSON(400, gin.H{"error": "missing version"})
		return
	}

	ctx.Request.URL.Path = fmt.Sprintf(
		"/lookup/%s@%s",
		path[:idx],
		path[idx+1:],
	)

	s.h(ctx)
}
