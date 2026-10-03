package api

// messages holds each route's success message, keyed "METHOD /mounted/pattern".
var messages = map[string]string{
	"GET /api/health":       "Backend is up.",
	"GET /api/health/ready": "Readiness checked.",
}
