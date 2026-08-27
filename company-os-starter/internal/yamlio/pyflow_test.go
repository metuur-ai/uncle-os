package yamlio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// flowCorpus is the value space rewrite_frontmatter_tags and canonical_yaml
// actually meet: committed frontmatter (where default_flow_style=None puts
// scalar-only collections inline and folds them at 80 columns), the derived
// feature-index, and the shapes that separate flow style from block style.
var flowCorpus = []struct{ name, yaml string }{
	{"reality-frontmatter", `
id: reality-customer-notification-service
type: component-reality
authority: canonical
updated: 2026-07-18
pointers:
  - {label: Delivery runbook, system: confluence, url: 'https://example.invalid/cns-runbook'}
tags: [authority/canonical, component/customer-notification-service, kind/reality,
  platform/communications]
`},
	{"prd-frontmatter", `
type: prd
id: 2026-per-channel-quiet-hours
title: Per-channel quiet hours
status: completed
components: [customer-notification-service]
created: 2026-07-18
tags: [component/customer-notification-service, discovery/2026-per-channel-quiet-hours,
  kind/prd, platform/communications, status/completed, team/customer-engagement]
`},
	{"feature-index", `
components:
  customer-notification-service:
    archivedPrds:
    - 2026-per-channel-quiet-hours
    externalPointers:
    - label: Delivery runbook
      system: confluence
      url: https://example.invalid/cns-runbook
    outcomes:
    - due: '2026-10-16'
      prd: 2026-per-channel-quiet-hours
      status: pending
    reality: reality/components/customer-notification-service.md
platform: communications
`},
	{"empty-collections", `
a: []
b: {}
c: [[]]
d: [1, 2, 3]
`},
	{"nested-blocks-stay-block", `
outer:
  - inner: [1, 2]
  - other: {k: v}
`},
	{"flow-indicators-force-quotes", `
tags: ['a,b', 'x[y]', 'p{q}', 'r?s', 'no,comma here']
`},
	{"long-scalar-in-flow-seq", `
tags: [aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa, bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb,
  cccccccccccccccccccccccccccccccccccccccc]
`},
	{"typed-scalars-in-flow", `
mixed: [1, 1.5, true, null, 2026-07-18, '2026-07-18', 'yes']
`},
	{"unicode-value", `
name: Café Ünïcode
tags: [a/é]
`},
	{"deep-map-in-flow", `
row: {name: Ada, role: Engineer}
rows:
  - {name: Ada, role: Engineer}
  - {name: Grace, role: Admiral}
`},
}

// TestAutoFlowMatchesPyYAML pins PyDumpAutoFlow to
// safe_dump(default_flow_style=None). It is what keeps `graph build` from
// rewriting every frontmatter document it scans: a block-style `tags:` would
// differ from the committed bytes on the first pass and never settle (R-0.6).
// The expected bytes are frozen in testdata/pyyaml_safedump.json; see
// TestEmitterMatchesPyYAML for why they are frozen rather than re-derived.
func TestAutoFlowMatchesPyYAML(t *testing.T) {
	want := frozenFlow(t, "autoflow")
	for i, c := range flowCorpus {
		t.Run(c.name, func(t *testing.T) {
			src := strings.TrimLeft(c.yaml, "\n")
			checkFrozenFlowCase(t, want, i, c.name, src)
			loaded := loadString(t, src)
			got, err := PyDumpAutoFlow(loaded)
			if err != nil {
				t.Fatalf("PyDumpAutoFlow: %v", err)
			}
			if got != want[i].Want {
				t.Fatalf("auto-flow emitter diverged\n--- python\n%s--- go\n%s",
					want[i].Want, got)
			}
		})
	}
}

// TestCanonicalMatchesPyYAML pins PyDumpCanonical to canonical_yaml
// (bin/company-os:96-99) — the semantic comparison form R-0.7c guards every
// derived write on, and the one gate 6 re-derives against.
func TestCanonicalMatchesPyYAML(t *testing.T) {
	want := frozenFlow(t, "canonical")
	for i, c := range flowCorpus {
		t.Run(c.name, func(t *testing.T) {
			src := strings.TrimLeft(c.yaml, "\n")
			checkFrozenFlowCase(t, want, i, c.name, src)
			got, err := PyDumpCanonical(loadString(t, src))
			if err != nil {
				t.Fatalf("PyDumpCanonical: %v", err)
			}
			if got != want[i].Want {
				t.Fatalf("canonical emitter diverged\n--- python\n%s--- go\n%s",
					want[i].Want, got)
			}
		})
	}
}

// frozenFlowCase is one flowCorpus document and what the vendored PyYAML 6.0.2
// emitted for it under a named dump mode.
type frozenFlowCase struct {
	Name string `json:"name"`
	Src  string `json:"src"`
	Want string `json:"want"`
}

// frozenFlow returns the frozen answers for one dump mode ("autoflow" or
// "canonical"), failing loudly if the file does not cover the live corpus.
func frozenFlow(t *testing.T, mode string) []frozenFlowCase {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "pyyaml_safedump.json"))
	if err != nil {
		t.Fatalf("reading the frozen PyYAML answers: %v", err)
	}
	var f struct {
		Flow map[string][]frozenFlowCase `json:"flow"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("decoding the frozen PyYAML answers: %v", err)
	}
	cases, ok := f.Flow[mode]
	if !ok {
		t.Fatalf("no frozen answers for dump mode %q", mode)
	}
	if len(cases) != len(flowCorpus) {
		t.Fatalf("frozen %s answers cover %d documents, corpus has %d; re-capture "+
			"testdata/pyyaml_safedump.json", mode, len(cases), len(flowCorpus))
	}
	return cases
}

// checkFrozenFlowCase guards against the frozen answers silently drifting out of
// alignment with the corpus, which would compare Go against the wrong document.
func checkFrozenFlowCase(t *testing.T, cases []frozenFlowCase, i int, name, src string) {
	t.Helper()
	if cases[i].Name != name || cases[i].Src != src {
		t.Fatalf("frozen answer %d is for a different document (%q) than the corpus "+
			"holds (%q); re-capture testdata/pyyaml_safedump.json", i, cases[i].Name, name)
	}
}

// TestCanonicalIsOrderBlind is the property the guard depends on: two documents
// that parse to the same structure must serialize identically here whatever
// order their keys were authored in, or a re-sync would rewrite files it should
// have skipped.
func TestCanonicalIsOrderBlind(t *testing.T) {
	a := loadString(t, "b: 2\na: 1\nc: {z: 1, y: 2}\n")
	b := loadString(t, "a: 1\nc: {y: 2, z: 1}\nb: 2\n")
	ta, err := PyDumpCanonical(a)
	if err != nil {
		t.Fatal(err)
	}
	tb, err := PyDumpCanonical(b)
	if err != nil {
		t.Fatal(err)
	}
	if ta != tb {
		t.Fatalf("canonical form is order-sensitive:\n%s\n---\n%s", ta, tb)
	}
}

// TestAutoFlowIsAFixedPoint is R-0.6 at the emitter level: a document already
// written by this emitter must re-emit unchanged.
func TestAutoFlowIsAFixedPoint(t *testing.T) {
	for _, c := range flowCorpus {
		t.Run(c.name, func(t *testing.T) {
			once, err := PyDumpAutoFlow(loadString(t, strings.TrimLeft(c.yaml, "\n")))
			if err != nil {
				t.Fatal(err)
			}
			twice, err := PyDumpAutoFlow(loadString(t, once))
			if err != nil {
				t.Fatal(err)
			}
			if once != twice {
				t.Fatalf("not a fixed point\n--- once\n%s--- twice\n%s", once, twice)
			}
		})
	}
}

// frontmatterQuoteCases pin PyDumpFrontmatter's one divergence from safe_dump:
// a TOP-LEVEL `title:`/`description:` string carrying any rune outside
// [A-Za-z0-9 ] is emitted double-quoted on one unfolded line. Everything else —
// another key, a nested title, a non-string value, a value of only letters,
// digits and spaces — must come out exactly as PyDumpAutoFlow emits it, which
// the companion assertion in TestFrontmatterQuotesTitleAndDescription checks
// case by case rather than by eye.
var frontmatterQuoteCases = []struct {
	name     string
	yaml     string
	want     string
	diverges bool
}{
	{
		name:     "description-with-percent-stays-one-line",
		yaml:     "type: prd\ndescription: \"Targets push opt-out below 3.0% within 60 days, with urgent messages deliberately exempt from the window.\"\ntags: [kind/prd]\n",
		want:     "type: prd\ndescription: \"Targets push opt-out below 3.0% within 60 days, with urgent messages deliberately exempt from the window.\"\ntags: [kind/prd]\n",
		diverges: true,
	},
	{
		name:     "title-with-hyphen",
		yaml:     "title: Per-channel quiet hours\ntags: [kind/prd]\n",
		want:     "title: \"Per-channel quiet hours\"\ntags: [kind/prd]\n",
		diverges: true,
	},
	{
		name:     "description-with-double-quote-is-escaped",
		yaml:     "description: 'He said \"no\" twice'\ntags: [kind/prd]\n",
		want:     "description: \"He said \\\"no\\\" twice\"\ntags: [kind/prd]\n",
		diverges: true,
	},
	{
		name:     "title-alphanumeric-and-spaces-unchanged",
		yaml:     "title: Quiet hours v2\ntags: [kind/prd]\n",
		want:     "title: Quiet hours v2\ntags: [kind/prd]\n",
		diverges: false,
	},
	{
		name:     "other-key-with-the-same-value-unchanged",
		yaml:     "summary: Per-channel quiet hours\ntags: [kind/prd]\n",
		want:     "summary: Per-channel quiet hours\ntags: [kind/prd]\n",
		diverges: false,
	},
	{
		name:     "nested-title-unchanged",
		yaml:     "pointers:\n  - {title: Per-channel quiet hours, url: 'https://example.invalid/x'}\ntags: [kind/prd]\n",
		want:     "pointers:\n- {title: Per-channel quiet hours, url: 'https://example.invalid/x'}\ntags: [kind/prd]\n",
		diverges: false,
	},
	{
		name:     "non-string-title-unchanged",
		yaml:     "title: 2026-07-18\ntags: [kind/prd]\n",
		want:     "title: 2026-07-18\ntags: [kind/prd]\n",
		diverges: false,
	},
	{
		name:     "empty-description-unchanged",
		yaml:     "description: ''\ntags: [kind/prd]\n",
		want:     "description: ''\ntags: [kind/prd]\n",
		diverges: false,
	},
	{
		name:     "unicode-title-escapes-like-safe-dump",
		yaml:     "title: Café hours\ntags: [kind/prd]\n",
		want:     "title: \"Caf\\xE9 hours\"\ntags: [kind/prd]\n",
		diverges: false, // safe_dump already double-quotes it; only the fold could differ.
	},
}

// TestFrontmatterQuotesTitleAndDescription is the rule itself, plus the scope
// claim: every non-diverging case must be byte-identical to PyDumpAutoFlow, so a
// future widening of the rule cannot pass unnoticed.
func TestFrontmatterQuotesTitleAndDescription(t *testing.T) {
	for _, c := range frontmatterQuoteCases {
		t.Run(c.name, func(t *testing.T) {
			loaded := loadString(t, c.yaml)
			got, err := PyDumpFrontmatter(loaded)
			if err != nil {
				t.Fatalf("PyDumpFrontmatter: %v", err)
			}
			if got != c.want {
				t.Fatalf("PyDumpFrontmatter\n--- want\n%s--- got\n%s", c.want, got)
			}
			auto, err := PyDumpAutoFlow(loaded)
			if err != nil {
				t.Fatalf("PyDumpAutoFlow: %v", err)
			}
			if c.diverges && got == auto {
				t.Fatalf("expected a divergence from safe_dump, got the same bytes:\n%s", got)
			}
			if !c.diverges && got != auto {
				t.Fatalf("out-of-scope case diverged from safe_dump\n--- auto\n%s--- got\n%s",
					auto, got)
			}
		})
	}
}

// TestFrontmatterQuotingRoundTrips is the correctness constraint: a style change
// may not become a value change. Re-loading the emitted document must give back
// the identical object.
func TestFrontmatterQuotingRoundTrips(t *testing.T) {
	for _, c := range frontmatterQuoteCases {
		t.Run(c.name, func(t *testing.T) {
			loaded := loadString(t, c.yaml)
			got, err := PyDumpFrontmatter(loaded)
			if err != nil {
				t.Fatal(err)
			}
			if back := loadString(t, got); !PyEqual(loaded, back) {
				t.Fatalf("value changed across the emit\n--- in\n%s--- out\n%s", c.yaml, got)
			}
		})
	}
}

// TestFrontmatterQuotingIsAFixedPoint is TestAutoFlowIsAFixedPoint for the new
// entry point: a document this emitter wrote must re-emit unchanged, or
// `graph build; graph build` would stop being a no-op diff (R-0.6).
func TestFrontmatterQuotingIsAFixedPoint(t *testing.T) {
	cases := append([]struct {
		name     string
		yaml     string
		want     string
		diverges bool
	}{}, frontmatterQuoteCases...)
	for _, c := range flowCorpus {
		cases = append(cases, struct {
			name     string
			yaml     string
			want     string
			diverges bool
		}{name: c.name, yaml: strings.TrimLeft(c.yaml, "\n")})
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			once, err := PyDumpFrontmatter(loadString(t, c.yaml))
			if err != nil {
				t.Fatal(err)
			}
			twice, err := PyDumpFrontmatter(loadString(t, once))
			if err != nil {
				t.Fatal(err)
			}
			if once != twice {
				t.Fatalf("not a fixed point\n--- once\n%s--- twice\n%s", once, twice)
			}
		})
	}
}

func TestPyEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want bool
	}{
		{"same-list", "v: [a, b]", "v: [a, b]", true},
		{"reordered-list", "v: [a, b]", "v: [b, a]", false},
		{"reordered-map", "v: {a: 1, b: 2}", "v: {b: 2, a: 1}", true},
		{"int-vs-float", "v: 5", "v: 5.0", true},
		{"str-vs-int", "v: '5'", "v: 5", false},
		{"null-vs-empty", "v: null", "v: ''", false},
		{"nested", "v: {a: [1, {b: c}]}", "v: {a: [1, {b: c}]}", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			a := loadString(t, c.a+"\n").(PyMap).Get("v")
			b := loadString(t, c.b+"\n").(PyMap).Get("v")
			if got := PyEqual(a, b); got != c.want {
				t.Fatalf("PyEqual = %v, want %v", got, c.want)
			}
		})
	}
}

func loadString(t *testing.T, src string) PyValue {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.yaml")
	if err := os.WriteFile(path, []byte(src), 0o666); err != nil {
		t.Fatal(err)
	}
	v, err := PyLoadFile(path)
	if err != nil {
		t.Fatalf("PyLoadFile: %v", err)
	}
	return v
}

// The two dump modes whose answers are frozen in
// testdata/pyyaml_safedump.json under "flow". Kept as source rather than prose
// so regenerating the file does not require reconstructing them from memory:
//
//	autoflow:  yaml.safe_dump(yaml.safe_load(src),
//	                          sort_keys=False, default_flow_style=None)
//	canonical: yaml.safe_dump(yaml.safe_load(src),
//	                          sort_keys=True, default_flow_style=False,
//	                          allow_unicode=True)
