package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swaggest/swgui/v5emb"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

// Option adds optional routes to the router.
type Option func(r *gin.Engine)

// Docs serves Swagger UI at /api/docs and specJSON at /api/openapi.json, outside the envelope and
// validation. Allowlisted so an unknown environment fails closed: production never registers them.
func Docs(env config.AppEnv, specJSON []byte) Option {
	return func(r *gin.Engine) {
		if env != config.Dev && env != config.Staging {
			return
		}
		ui := gin.WrapH(v5emb.New("agent-repo-template backend", "/api/openapi.json", "/api/docs/"))
		r.GET("/api/docs", ui)
		r.GET("/api/docs/*filepath", ui)
		r.GET("/api/openapi.json", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/json", specJSON)
		})
	}
}
