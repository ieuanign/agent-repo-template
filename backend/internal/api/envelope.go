package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type meta struct {
	Path      string `json:"path"`
	Timestamp string `json:"timestamp"`
}

type envelopeBody struct {
	Meta    meta            `json:"meta"`
	Status  int             `json:"status"`
	Code    *string         `json:"code"`
	Message string          `json:"message"`
	Payload json.RawMessage `json:"payload"`
}

// captureWriter buffers the status and body a typed strict response writes.
type captureWriter struct {
	gin.ResponseWriter
	status int
	body   bytes.Buffer
}

func (w *captureWriter) WriteHeader(code int)              { w.status = code }
func (w *captureWriter) WriteHeaderNow()                   {}
func (w *captureWriter) Write(b []byte) (int, error)       { return w.body.Write(b) }
func (w *captureWriter) WriteString(s string) (int, error) { return w.body.WriteString(s) }
func (w *captureWriter) Status() int                       { return w.status }
func (w *captureWriter) Size() int                         { return w.body.Len() }
func (w *captureWriter) Written() bool                     { return false }

// envelope wraps each operation's response in {meta, status, code, message, payload}.
func envelope() gin.HandlerFunc {
	return func(c *gin.Context) {
		orig := c.Writer
		cw := &captureWriter{ResponseWriter: orig, status: http.StatusOK}
		c.Writer = cw
		// Restored on panic too, so recovery answers on the real writer.
		defer func() { c.Writer = orig }()
		c.Next()
		c.Writer = orig

		if err := c.Errors.Last(); err != nil {
			status, code, message := mapError(c, err.Err)
			writeFailure(c, status, code, message)
			return
		}
		if cw.status == http.StatusNoContent || cw.status == http.StatusNotModified {
			orig.Header().Del("Content-Type")
			orig.WriteHeader(cw.status)
			orig.WriteHeaderNow()
			return
		}
		payload := json.RawMessage(bytes.TrimSpace(cw.body.Bytes()))
		if len(payload) == 0 {
			payload = json.RawMessage("null")
		}
		c.JSON(cw.status, envelopeBody{
			Meta:    metaFor(c),
			Status:  cw.status,
			Message: messages[c.Request.Method+" "+c.FullPath()],
			Payload: payload,
		})
	}
}

// writeFailure answers on the real writer; recovery uses it too, from outside any operation.
func writeFailure(c *gin.Context, status int, code, message string) {
	c.JSON(status, envelopeBody{Meta: metaFor(c), Status: status, Code: &code, Message: message, Payload: json.RawMessage("null")})
}

func metaFor(c *gin.Context) meta {
	return meta{Path: c.Request.URL.Path, Timestamp: time.Now().UTC().Format(time.RFC3339)}
}
