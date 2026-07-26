package validate_test

// R-7.7: the text renderer and the JSON encoder report the same findings, in the
// same order, for `validate` on every committed fixture.
//
// The assertion is deliberately made against the RENDERED TEXT rather than
// against the record slice both renderers were handed. Comparing the input to
// itself would pass no matter what either renderer dropped; parsing the text back
// out is the only version of this test that can fail. It is cheap to do because
// R-3.4b made the JSON side one encoder over the record types: there is exactly
// one place for the two to disagree.
//
// The tuple is {gate, code, severity, subject}. Text carries no code — that is
// what codes are FOR — so code is asserted as non-empty and carried through from
// the JSON side, and the shared columns are matched position by position within
// each gate.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/render"
)

type tuple struct {
	gate     int
	code     string
	severity string
	subject  string
}

func (t tuple) String() string {
	return fmt.Sprintf("[%d] %s %s %q", t.gate, t.severity, t.code, t.subject)
}

var (
	headerLine  = regexp.MustCompile(`^\[(\d+)/(\d+)\] `)
	findingLine = regexp.MustCompile(`^  \[(ok|warn|FAIL)\] (.*)$`)
)

// textTuples re-reads render.Validate's own output. severity comes from the
// bracketed marker; subject is everything before the first ": " when the JSON
// side says there is a subject, which is why the JSON tuples are threaded in
// rather than guessed at — "a: b" is ambiguous in text and is not ambiguous in
// the records.
func textTuples(t *testing.T, text string, want []tuple) []tuple {
	t.Helper()
	var got []tuple
	gate := 0
	i := 0
	for _, ln := range strings.Split(text, "\n") {
		if m := headerLine.FindStringSubmatch(ln); m != nil {
			gate = atoi(t, m[1])
			continue
		}
		m := findingLine.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		sev := m[1]
		if sev == "FAIL" {
			sev = "fail"
		}
		subject := ""
		if i < len(want) && want[i].subject != "" {
			if rest, ok := strings.CutPrefix(m[2], want[i].subject+": "); ok {
				subject = want[i].subject
				_ = rest
			}
		}
		code := ""
		if i < len(want) {
			code = want[i].code
		}
		got = append(got, tuple{gate: gate, code: code, severity: sev, subject: subject})
		i++
	}
	return got
}

// jsonTuples reads the same findings out of the `--json` document, through
// encoding/json, which also proves the payload is valid JSON (R-3.1).
func jsonTuples(t *testing.T, sections []model.GateResult) []tuple {
	t.Helper()
	var buf bytes.Buffer
	err := render.JSON(&buf, render.Result{
		Command: "validate", Root: "/ws", Sections: sections,
	})
	if err != nil {
		t.Fatalf("render.JSON: %v", err)
	}
	var doc struct {
		SchemaVersion int `json:"schemaVersion"`
		Build         struct {
			Version string `json:"version"`
		} `json:"build"`
		Sections []struct {
			Ordinal  int    `json:"ordinal"`
			Slug     string `json:"slug"`
			Findings []struct {
				Severity string `json:"severity"`
				Code     string `json:"code"`
				Subject  string `json:"subject"`
			} `json:"findings"`
		} `json:"sections"`
	}
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("--json payload is not valid JSON: %v\n%s", err, buf.String())
	}
	if doc.SchemaVersion != render.SchemaVersion {
		t.Errorf("schemaVersion = %d, want %d", doc.SchemaVersion, render.SchemaVersion)
	}
	if doc.Build.Version == "" {
		t.Error("payload carries no build identifier (R-3.5)")
	}
	var out []tuple
	for _, s := range doc.Sections {
		if s.Slug == model.SlugWorkspace {
			// The banner is a section but not a gate and has no text line of its
			// own below the header, so it is outside the comparison on both sides.
			continue
		}
		for _, f := range s.Findings {
			if f.Code == "" {
				t.Errorf("gate %d emitted a finding with no code (R-2.4)", s.Ordinal)
			}
			out = append(out, tuple{
				gate: s.Ordinal, code: f.Code, severity: f.Severity, subject: f.Subject,
			})
		}
	}
	return out
}

// TestTextAndJSONReportTheSameFindings is R-7.7 and, through it, R-2.9.
func TestTextAndJSONReportTheSameFindings(t *testing.T) {
	dir := examplesDir(t)
	for _, fixture := range []string{
		"workspace", "federated",
		"failing-workspace", "failing-federated", "failing-federated-nolock",
	} {
		t.Run(fixture, func(t *testing.T) {
			text, _, sections := runValidate(t, filepath.Join(dir, fixture))
			want := jsonTuples(t, sections)
			got := textTuples(t, text, want)
			if len(got) != len(want) {
				t.Fatalf("text reported %d findings, --json reported %d\n%s",
					len(got), len(want), text)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Errorf("finding %d: text %s, json %s", i, got[i], want[i])
				}
			}
			if len(want) == 0 {
				t.Fatal("fixture produced no findings — the comparison is vacuous")
			}
		})
	}
}

// TestAbortedRunIsStillValidJSON is R-3.8 at the record level: the mid-gate abort
// R-2.6a defined must encode, carry its completed gates, and say which gate count
// they were numbered against.
func TestAbortedRunIsStillValidJSON(t *testing.T) {
	sections := []model.GateResult{
		{Slug: model.SlugWorkspace, Findings: []model.Finding{{
			Code:   model.CodeValidateRoot,
			Fields: model.Fields{"root": "/ws", "complete": false, "gates": 7},
		}}},
		{Ordinal: 1, Slug: "ownership-reconciliation", Findings: []model.Finding{
			{Severity: model.SevOK, Code: model.CodeOwnershipAgrees, Subject: "svc"},
		}},
	}
	var buf bytes.Buffer
	err := render.JSON(&buf, render.Result{
		Command:  "validate",
		Root:     "/ws",
		Sections: sections,
		Err:      model.Errorf(model.ExitValidation, "not an ISO-8601 date"),
		ExitCode: model.ExitValidation,
	})
	if err != nil {
		t.Fatalf("render.JSON: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("aborted run did not emit valid JSON: %v\n%s", err, buf.String())
	}
	if doc["error"] != "not an ISO-8601 date" {
		t.Errorf("error field = %v", doc["error"])
	}
	if doc["exitCode"] != float64(model.ExitValidation) {
		t.Errorf("exitCode = %v, want %d", doc["exitCode"], model.ExitValidation)
	}
	if n := len(doc["sections"].([]any)); n != 2 {
		t.Errorf("aborted run encoded %d sections, want the banner plus gate 1", n)
	}
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
	}
	return n
}
