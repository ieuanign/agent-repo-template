package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ieuanign/agent-repo-template/backend/internal/platform/config"
)

func TestNewEncoding(t *testing.T) {
	tests := []struct {
		env      config.AppEnv
		wantJSON bool
	}{
		{config.Dev, false},
		{config.Staging, true},
		{config.Production, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.env), func(t *testing.T) {
			var out bytes.Buffer
			New(tt.env, &out).Info("hello")
			var m map[string]any
			isJSON := json.Unmarshal(out.Bytes(), &m) == nil
			if isJSON != tt.wantJSON {
				t.Fatalf("JSON = %v, want %v: %q", isJSON, tt.wantJSON, out.String())
			}
			if !bytes.Contains(out.Bytes(), []byte("hello")) {
				t.Fatalf("output %q lacks message", out.String())
			}
		})
	}
}
