package main

import (
	"os"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/graph"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

func TestZZProbe(t *testing.T) {
	root := os.Getenv("PROBE_ROOT")
	if root == "" {
		t.Skip("probe")
	}
	ws := workspace.New(root)
	g, err := graph.NodeGate(ws, 5)
	t.Logf("NodeGate err=%v", err)
	for _, f := range g.Findings {
		t.Logf("  sev=%v code=%s subject=%q", f.Severity, f.Code, f.Subject)
	}
}
