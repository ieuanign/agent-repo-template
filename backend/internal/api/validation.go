package api

import (
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/apperr"
)

// validator rejects requests the spec does not allow; its text never reaches the client, as it can echo request values.
// The servers warning is silenced because it prints through log, and the relative /api entry has no host to check.
func validator(spec *openapi3.T) gin.HandlerFunc {
	return ginmiddleware.OapiRequestValidatorWithOptions(spec, &ginmiddleware.Options{
		SilenceServersWarning: true,
		// 404 is the spec lacking a route gin has; only a spec out of step with the generated code reaches it.
		ErrorHandler: func(c *gin.Context, _ string, status int) {
			if status == http.StatusNotFound {
				fail(c, apperr.NotFound)
				return
			}
			fail(c, apperr.Validation)
		},
	})
}
