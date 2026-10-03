package api

import (
	"testing"

	"go.uber.org/zap"
)

func TestEveryRouteHasASuccessMessage(t *testing.T) {
	routes := NewRouter(stub{}, liveSpec(t), zap.NewNop()).Routes()
	if len(routes) == 0 {
		t.Fatal("router has no routes")
	}
	for _, rt := range routes {
		if _, ok := messages[rt.Method+" "+rt.Path]; !ok {
			t.Errorf("no success message for %q", rt.Method+" "+rt.Path)
		}
	}
}
