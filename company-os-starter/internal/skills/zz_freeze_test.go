package skills_test

// Temporary generator: walks gateCases, builds each synthesized workspace, runs
// the Python reference CLI recovered from tag python-cli-final, and records its
// gate-7 block. Removed after use.

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestZZFreezeGate7(t *testing.T) {
	kit := os.Getenv("FREEZE_KIT") // .../company-os-starter with bin/ vendor/ templates/
	out := os.Getenv("FREEZE_OUT")
	if kit == "" || out == "" {
		t.Skip("generator")
	}
	cli := filepath.Join(kit, "bin", "company-os")
	if _, err := os.Stat(cli); err != nil {
		t.Fatalf("recovered CLI absent: %v", err)
	}

	answers := map[string]string{}
	for _, tc := range gateCases {
		ws := fourLayerWorkspace(t)
		tc.build(t, ws)

		cmd := exec.Command("python3", cli, "--root", ws.Root, "validate")
		cmd.Env = append(os.Environ(), "PYTHONPATH="+filepath.Join(kit, "vendor"))
		var buf bytes.Buffer
		cmd.Stdout, cmd.Stderr = &buf, &buf
		_ = cmd.Run()

		var block []string
		in := false
		for _, line := range strings.Split(buf.String(), "\n") {
			if strings.HasPrefix(line, "[7/") {
				in = true
			} else if in && line == "" {
				break
			}
			if in {
				block = append(block, line)
			}
		}
		if len(block) == 0 {
			t.Fatalf("%s: reference validate produced no gate 7 block:\n%s", tc.name, buf.String())
		}
		answers[tc.name] = strings.Join(block, "\n") + "\n"
	}

	payload := map[string]any{
		"provenance": "Gate-7 blocks captured from the Python reference CLI recovered " +
			"from tag python-cli-final, in the change that removed the last Python " +
			"from the repo. One entry per gateCases case, keyed by case name.",
		"gate7": answers,
	}
	b, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(out, append(b, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Logf("froze %d gate-7 answers", len(answers))
}
