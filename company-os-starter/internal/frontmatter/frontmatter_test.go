package frontmatter

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Differential test for R-1.5. Python's frontmatter() (bin/company-os:76) is the
// oracle. Every expectation below was MEASURED by running it, not reasoned about
// — see .devlocal/go-port/frontmatter-truth.md for the transcript and the
// surprises (`---\n---\n` rejects; CRLF accepts because read_text() translates
// newlines before the regex runs).
//
// Two layers assert the same corpus:
//   1. TestFrozenPythonTruth — the recorded measurement, always runs.
//   2. TestDifferentialAgainstPythonOracle — re-runs the real Python on the same
//      corpus, so the frozen table cannot quietly rot.

type doc struct {
	name string
	raw  []byte
}

// corpus covers, at minimum: no frontmatter; empty frontmatter; missing closing
// fence; closing fence without a trailing newline; CRLF and lone-CR line
// endings; BOM and whitespace before the opening fence; "---" in the body; "---"
// inside a YAML string; a document that is exactly "---"; trailing whitespace
// after a fence; and a very large body.
var corpus = []doc{
	{"no_frontmatter", []byte("# Title\n\nbody text\n")},
	{"empty_file", []byte("")},
	{"only_newline", []byte("\n")},
	{"exactly_three_dashes", []byte("---")},
	{"three_dashes_newline", []byte("---\n")},
	{"empty_fm_adjacent_fences", []byte("---\n---\n")},
	{"empty_fm_blank_line", []byte("---\n\n---\n")},
	{"fm_only_whitespace", []byte("---\n   \n---\nbody\n")},
	{"fm_only_comment", []byte("---\n# just a comment\n---\nbody\n")},
	{"simple_ok", []byte("---\ntype: prd\nid: x\n---\nbody\n")},
	{"ok_no_body", []byte("---\ntype: prd\n---\n")},
	{"ok_body_no_trailing_nl", []byte("---\ntype: prd\n---\nbody")},
	{"missing_closing_fence", []byte("---\ntype: prd\n")},
	{"closing_fence_no_trailing_nl", []byte("---\ntype: prd\n---")},
	{"triple_fence", []byte("---\n---\n---\nbody\n")},
	{"crlf_everywhere", []byte("---\r\ntype: prd\r\n---\r\nbody\r\n")},
	{"crlf_fences_only", []byte("---\r\ntype: prd\n---\r\nbody\n")},
	{"cr_only_old_mac", []byte("---\rtype: prd\r---\rbody\r")},
	{"bom_before_fence", []byte("\ufeff---\ntype: prd\n---\nbody\n")},
	{"leading_blank_line", []byte("\n---\ntype: prd\n---\nbody\n")},
	{"leading_space", []byte(" ---\ntype: prd\n---\nbody\n")},
	{"open_fence_trailing_space", []byte("--- \ntype: prd\n---\nbody\n")},
	{"close_fence_trailing_space", []byte("---\ntype: prd\n--- \nbody\n")},
	{"close_fence_trailing_tab", []byte("---\ntype: prd\n---\t\nbody\n")},
	{"four_dashes_open", []byte("----\ntype: prd\n---\nbody\n")},
	{"four_dashes_close", []byte("---\ntype: prd\n----\nbody\n")},
	{"dashes_in_body", []byte("---\ntype: prd\n---\nbody\n---\nmore\n")},
	{"dashes_in_yaml_string", []byte("---\ntitle: \"a\n---\nb\"\nx: 1\n---\nbody\n")},
	{"dashes_indented_in_yaml", []byte("---\ntype: prd\ndesc: |\n  ---\n  still yaml\n---\nbody\n")},
	{"body_starts_with_fence", []byte("---\ntype: prd\n---\n---\nbody\n")},
	{"utf8_multibyte", []byte("---\ntype: prd\ntitle: caf\u00e9 \u2014 \u65e5\u672c\n---\nbody \u00fc\n")},
	{"invalid_utf8", []byte("---\ntype: prd\n---\nbody \xff\xfe\n")},
	{"large_body", []byte("---\ntype: prd\n---\n" + strings.Repeat("x", 200000) + "\n")},
	{"large_fm", []byte("---\n" + strings.Repeat("k: v\n", 20000) + "---\nbody\n")},
	{"null_byte_in_body", []byte("---\ntype: prd\n---\nbo\x00dy\n")},
	{"windows_bom_and_crlf", []byte("\ufeff---\r\ntype: prd\r\n---\r\nbody\r\n")},
	{"fm_with_trailing_blank", []byte("---\ntype: prd\n\n---\nbody\n")},
	{"only_fences_no_newline_end", []byte("---\n\n---")},
	{"body_is_only_newlines", []byte("---\ntype: prd\n---\n\n\n")},
	{"tabs_in_fm", []byte("---\ntype:\tprd\n---\nbody\n")}}

type want struct {
	accept      bool
	yaml        string
	body        string
	decodeError bool
}

// pythonTruth: measured 2026-07-26 against Python 3.12.11 with the vendored
// PyYAML. yaml/body are m.group(1)/m.group(2); on reject, body is the whole
// newline-translated text, which is what Python returns alongside {}.
var pythonTruth = map[string]want{
	"no_frontmatter":               {accept: false, body: "# Title\n\nbody text\n"},
	"empty_file":                   {accept: false, body: ""},
	"only_newline":                 {accept: false, body: "\n"},
	"exactly_three_dashes":         {accept: false, body: "---"},
	"three_dashes_newline":         {accept: false, body: "---\n"},
	"empty_fm_adjacent_fences":     {accept: false, body: "---\n---\n"},
	"empty_fm_blank_line":          {accept: true, yaml: "", body: ""},
	"fm_only_whitespace":           {accept: true, yaml: "   ", body: "body\n"},
	"fm_only_comment":              {accept: true, yaml: "# just a comment", body: "body\n"},
	"simple_ok":                    {accept: true, yaml: "type: prd\nid: x", body: "body\n"},
	"ok_no_body":                   {accept: true, yaml: "type: prd", body: ""},
	"ok_body_no_trailing_nl":       {accept: true, yaml: "type: prd", body: "body"},
	"missing_closing_fence":        {accept: false, body: "---\ntype: prd\n"},
	"closing_fence_no_trailing_nl": {accept: false, body: "---\ntype: prd\n---"},
	"triple_fence":                 {accept: true, yaml: "---", body: "body\n"},
	"crlf_everywhere":              {accept: true, yaml: "type: prd", body: "body\n"},
	"crlf_fences_only":             {accept: true, yaml: "type: prd", body: "body\n"},
	"cr_only_old_mac":              {accept: true, yaml: "type: prd", body: "body\n"},
	"bom_before_fence":             {accept: false, body: "\ufeff---\ntype: prd\n---\nbody\n"},
	"leading_blank_line":           {accept: false, body: "\n---\ntype: prd\n---\nbody\n"},
	"leading_space":                {accept: false, body: " ---\ntype: prd\n---\nbody\n"},
	"open_fence_trailing_space":    {accept: false, body: "--- \ntype: prd\n---\nbody\n"},
	"close_fence_trailing_space":   {accept: false, body: "---\ntype: prd\n--- \nbody\n"},
	"close_fence_trailing_tab":     {accept: false, body: "---\ntype: prd\n---\t\nbody\n"},
	"four_dashes_open":             {accept: false, body: "----\ntype: prd\n---\nbody\n"},
	"four_dashes_close":            {accept: false, body: "---\ntype: prd\n----\nbody\n"},
	"dashes_in_body":               {accept: true, yaml: "type: prd", body: "body\n---\nmore\n"},
	"dashes_in_yaml_string":        {accept: true, yaml: "title: \"a", body: "b\"\nx: 1\n---\nbody\n"},
	"dashes_indented_in_yaml":      {accept: true, yaml: "type: prd\ndesc: |\n  ---\n  still yaml", body: "body\n"},
	"body_starts_with_fence":       {accept: true, yaml: "type: prd", body: "---\nbody\n"},
	"utf8_multibyte":               {accept: true, yaml: "type: prd\ntitle: caf\u00e9 \u2014 \u65e5\u672c", body: "body \u00fc\n"},
	"invalid_utf8":                 {decodeError: true},
	"large_body":                   {accept: true, yaml: "type: prd", body: strings.Repeat("x", 200000) + "\n"},
	"large_fm":                     {accept: true, yaml: strings.TrimSuffix(strings.Repeat("k: v\n", 20000), "\n"), body: "body\n"},
	"null_byte_in_body":            {accept: true, yaml: "type: prd", body: "bo\x00dy\n"},
	"windows_bom_and_crlf":         {accept: false, body: "\ufeff---\ntype: prd\n---\nbody\n"},
	"fm_with_trailing_blank":       {accept: true, yaml: "type: prd\n", body: "body\n"},
	"only_fences_no_newline_end":   {accept: false, body: "---\n\n---"},
	"body_is_only_newlines":        {accept: true, yaml: "type: prd", body: "\n\n"},
	"tabs_in_fm":                   {accept: true, yaml: "type:\tprd", body: "body\n"}}

func TestFrozenPythonTruth(t *testing.T) {
	if len(corpus) != len(pythonTruth) {
		t.Fatalf("corpus has %d docs, truth table has %d", len(corpus), len(pythonTruth))
	}
	for _, d := range corpus {
		d := d
		t.Run(d.name, func(t *testing.T) {
			w, ok := pythonTruth[d.name]
			if !ok {
				t.Fatalf("no measured truth for %q", d.name)
			}
			got, err := Parse(d.raw)
			if w.decodeError {
				if err == nil {
					t.Fatalf("Python raised UnicodeDecodeError; Go returned ok=%v", got.HasFrontmatter)
				}
				return
			}
			if err != nil {
				t.Fatalf("Python returned normally; Go errored: %v", err)
			}
			if got.HasFrontmatter != w.accept {
				t.Fatalf("accept: Go=%v Python=%v", got.HasFrontmatter, w.accept)
			}
			if w.accept && string(got.YAML) != w.yaml {
				t.Errorf("yaml block:\n Go     = %q\n Python = %q", got.YAML, w.yaml)
			}
			if !w.accept && got.YAML != nil {
				t.Errorf("rejected doc must carry no yaml block, got %q", got.YAML)
			}
			if string(got.Body) != w.body {
				t.Errorf("body:\n Go     = %q\n Python = %q", trunc(got.Body), trunc([]byte(w.body)))
			}
		})
	}
}

// TestDifferentialAgainstPythonOracle runs the real Python frontmatter() over
// the same corpus and asserts Go made the identical decision for every document.
// It SKIPS (never silently passes) when python3 or the vendored PyYAML is
// unavailable, so a green run on a stripped machine is not mistaken for a
// verified one.
func TestDifferentialAgainstPythonOracle(t *testing.T) {
	py, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not on PATH; oracle cannot run")
	}
	dir := t.TempDir()
	for _, d := range corpus {
		if err := os.WriteFile(filepath.Join(dir, d.name+".md"), d.raw, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	out, err := exec.Command(py, filepath.Join("testdata", "oracle.py"), dir).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Skipf("oracle unavailable (bin/company-os or vendored PyYAML not loadable):\n%s", ee.Stderr)
		}
		t.Skipf("oracle unavailable: %v", err)
	}
	var oracle map[string]struct {
		Decision string `json:"decision"`
		YAMLB64  string `json:"yaml_b64"`
		BodyB64  string `json:"body_b64"`
	}
	if err := json.Unmarshal(out, &oracle); err != nil {
		t.Fatalf("oracle output not JSON: %v", err)
	}
	if len(oracle) != len(corpus) {
		t.Fatalf("oracle saw %d docs, corpus has %d", len(oracle), len(corpus))
	}
	for _, d := range corpus {
		d := d
		t.Run(d.name, func(t *testing.T) {
			o, ok := oracle[d.name]
			if !ok {
				t.Fatalf("oracle produced no record for %q", d.name)
			}
			got, err := Parse(d.raw)
			if o.Decision == "decode-error" {
				if err == nil {
					t.Fatalf("Python raised on decode; Go returned a document")
				}
				return
			}
			if err != nil {
				t.Fatalf("Python decided %q; Go errored: %v", o.Decision, err)
			}
			wantAccept := o.Decision == "accept"
			if got.HasFrontmatter != wantAccept {
				t.Fatalf("accept: Go=%v Python=%v", got.HasFrontmatter, wantAccept)
			}
			if wantAccept {
				if want := b64(t, o.YAMLB64); string(got.YAML) != want {
					t.Errorf("yaml block:\n Go     = %q\n Python = %q", got.YAML, want)
				}
			}
			if want := b64(t, o.BodyB64); string(got.Body) != want {
				t.Errorf("body:\n Go     = %q\n Python = %q", trunc(got.Body), trunc([]byte(want)))
			}
		})
	}
}

// TestParseFileMatchesParse pins the path-taking entry point to the byte-taking
// one, since Python's frontmatter() takes a path and its read_text() decode is
// part of the measured contract.
func TestParseFileMatchesParse(t *testing.T) {
	dir := t.TempDir()
	for _, d := range corpus {
		p := filepath.Join(dir, d.name+".md")
		if err := os.WriteFile(p, d.raw, 0o644); err != nil {
			t.Fatal(err)
		}
		fromBytes, errB := Parse(d.raw)
		fromFile, errF := ParseFile(p)
		if (errB == nil) != (errF == nil) {
			t.Errorf("%s: Parse err=%v but ParseFile err=%v", d.name, errB, errF)
			continue
		}
		if errB != nil {
			continue
		}
		if fromFile.HasFrontmatter != fromBytes.HasFrontmatter ||
			string(fromFile.YAML) != string(fromBytes.YAML) ||
			string(fromFile.Body) != string(fromBytes.Body) {
			t.Errorf("%s: ParseFile and Parse disagree", d.name)
		}
	}
}

func b64(t *testing.T, s string) string {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("bad base64 from oracle: %v", err)
	}
	return string(b)
}

func trunc(b []byte) string {
	if len(b) > 120 {
		return string(b[:120]) + "…(" + itoa(len(b)) + " bytes)"
	}
	return string(b)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var d []byte
	for n > 0 {
		d = append([]byte{byte('0' + n%10)}, d...)
		n /= 10
	}
	return string(d)
}

var _ = strings.Repeat
