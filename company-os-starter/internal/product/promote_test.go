package product

import (
	"strings"
	"testing"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// meta loads a frontmatter block the way the promotion path will hand it over.
func meta(t *testing.T, yamlText string) yamlio.PyMap {
	t.Helper()
	v, err := yamlio.PyLoadBytes([]byte(yamlText), "test")
	if err != nil {
		t.Fatal(err)
	}
	m, ok := v.(yamlio.PyMap)
	if !ok {
		t.Fatalf("fixture is not a mapping: %#v", v)
	}
	return m
}

// TestPromoteTargetResolves is R-4.1: the target is read from
// `promoteTo.platform` and resolved to the catalog directory.
func TestPromoteTargetResolves(t *testing.T) {
	ws := fixture(t)
	m := meta(t, "platform: payments\npromoteTo:\n  platform: payments\n")
	got, err := ResolvePromoteTarget(ws, m, "draft.md")
	if err != nil {
		t.Fatal(err)
	}
	if got.Platform != "payments" {
		t.Errorf("Platform = %q, want payments", got.Platform)
	}
	if !strings.HasSuffix(got.Dir, "payments") {
		t.Errorf("Dir = %q, want it to end in payments", got.Dir)
	}
}

// TestPromoteTargetIgnoresScope is R-4.5: `promoteTo.scope` is reserved and is
// not read — its presence changes nothing.
func TestPromoteTargetIgnoresScope(t *testing.T) {
	ws := fixture(t)
	m := meta(t, "platform: payments\npromoteTo:\n  platform: payments\n  scope: component\n")
	got, err := ResolvePromoteTarget(ws, m, "draft.md")
	if err != nil {
		t.Fatal(err)
	}
	if got.Platform != "payments" {
		t.Errorf("Platform = %q, want payments", got.Platform)
	}
}

// TestPromoteTargetRefusals covers R-4.2 (absent/placeholder -> artifact),
// R-4.3 (unknown platform -> workspace, named) and R-4.4 (disagreement ->
// artifact, both values reported).
func TestPromoteTargetRefusals(t *testing.T) {
	ws := fixture(t)
	cases := []struct {
		name string
		yaml string
		code model.ExitCode
		want []string
	}{
		{
			name: "absent",
			yaml: "platform: payments\n",
			code: model.ExitArtifact,
			want: []string{"no target platform", "promoteTo.platform"},
		},
		{
			name: "placeholder",
			yaml: "platform: TODO\npromoteTo:\n  platform: TODO\n",
			code: model.ExitArtifact,
			want: []string{"no target platform"},
		},
		{
			name: "empty block",
			yaml: "platform: payments\npromoteTo: {}\n",
			code: model.ExitArtifact,
			want: []string{"no target platform"},
		},
		{
			name: "unknown platform",
			yaml: "platform: ledger\npromoteTo:\n  platform: ledger\n",
			code: model.ExitWorkspace,
			want: []string{"ledger"},
		},
		{
			name: "disagreement",
			yaml: "platform: billing\npromoteTo:\n  platform: payments\n",
			code: model.ExitArtifact,
			want: []string{"payments", "billing"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ResolvePromoteTarget(ws, meta(t, tc.yaml), "draft.md")
			if err == nil {
				t.Fatal("expected a refusal")
			}
			if got := model.CodeOf(err); got != tc.code {
				t.Errorf("exit code = %d, want %d (%v)", got, tc.code, err)
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("message %q does not mention %q", err, w)
				}
			}
		})
	}
}

// TestPromoteTargetUnsetProcessFieldIsNotDisagreement: a draft whose process
// `platform` is still TODO is Unit 5's omission to report, not a disagreement
// naming TODO as a platform.
func TestPromoteTargetUnsetProcessFieldIsNotDisagreement(t *testing.T) {
	ws := fixture(t)
	m := meta(t, "platform: TODO\npromoteTo:\n  platform: payments\n")
	got, err := ResolvePromoteTarget(ws, m, "draft.md")
	if err != nil {
		t.Fatal(err)
	}
	if got.Platform != "payments" {
		t.Errorf("Platform = %q, want payments", got.Platform)
	}
}

// ---------------------------------------------------------------- R-6.1/6.2

// TestPromotedFromRoundTrip is R-6.1: origin team, draft id and digest survive
// a write/read cycle through the shared shape.
func TestPromotedFromRoundTrip(t *testing.T) {
	want := PromotedFrom{Team: "core", Draft: "2026-thing", Digest: "sha256:abc"}
	m := yamlio.PyMap{}.Set(KeyPromotedFrom, want.PyMap())
	got, ok := ReadPromotedFrom(m)
	if !ok || got != want {
		t.Errorf("ReadPromotedFrom = %+v ok=%v, want %+v", got, ok, want)
	}
}

// TestPromotedToRoundTrip is R-6.2: target platform and record id.
func TestPromotedToRoundTrip(t *testing.T) {
	want := PromotedTo{Platform: "payments", Record: "2026-thing"}
	m := yamlio.PyMap{}.Set(KeyPromotedTo, want.PyMap())
	got, ok := ReadPromotedTo(m)
	if !ok || got != want {
		t.Errorf("ReadPromotedTo = %+v ok=%v, want %+v", got, ok, want)
	}
}

// TestPromotedBlocksAbsentIsNotAnError is R-6.6: a record that was never
// promoted has neither key, and that is a non-event.
func TestPromotedBlocksAbsentIsNotAnError(t *testing.T) {
	m := meta(t, "type: prd\nstatus: proposed\n")
	if _, ok := ReadPromotedFrom(m); ok {
		t.Error("ReadPromotedFrom reported a block on a never-promoted record")
	}
	if _, ok := ReadPromotedTo(m); ok {
		t.Error("ReadPromotedTo reported a block on a never-promoted record")
	}
}

// ------------------------------------------------------------------- digest

const digestDraft = "---\n" +
	"type: prd\n" +
	"id: 2026-thing\n" +
	"status: draft\n" +
	"platform: payments\n" +
	"---\n" +
	"# Thing\n" +
	"\n" +
	"## Problem\n" +
	"Bodies drift.\n"

// TestDigestIsStable: the same bytes digest identically every time, which is
// what makes a recompute meaningful at all.
func TestDigestIsStable(t *testing.T) {
	first := DraftDigest([]byte(digestDraft))
	if first != DraftDigest([]byte(digestDraft)) {
		t.Error("the same draft digested differently twice")
	}
	if !strings.HasPrefix(first, DigestPrefix) {
		t.Errorf("digest %q is not labelled with %q", first, DigestPrefix)
	}
	if len(first) != len(DigestPrefix)+64 {
		t.Errorf("digest %q is not a sha256 hex", first)
	}
}

// TestDigestNormalizes is R-6.3: CRLF line endings, trailing whitespace and the
// derived `tags:` block are all outside what the digest sees.
func TestDigestNormalizes(t *testing.T) {
	base := DraftDigest([]byte(digestDraft))
	variants := map[string]string{
		"crlf":              strings.ReplaceAll(digestDraft, "\n", "\r\n"),
		"trailing spaces":   strings.ReplaceAll(digestDraft, "# Thing\n", "# Thing   \t\n"),
		"trailing newlines": digestDraft + "\n\n\n",
		"tags added":        strings.Replace(digestDraft, "status: draft\n", "status: draft\ntags: [kind/prd, platform/payments]\n", 1),
		"tags reordered":    strings.Replace(digestDraft, "status: draft\n", "status: draft\ntags: [platform/payments, kind/prd]\n", 1),
		"tags block style":  strings.Replace(digestDraft, "status: draft\n", "status: draft\ntags:\n- kind/prd\n", 1),
		"tags last not mid": strings.Replace(digestDraft, "platform: payments\n", "platform: payments\ntags: [kind/prd]\n", 1),
		"tags and crlf":     strings.ReplaceAll(strings.Replace(digestDraft, "status: draft\n", "status: draft\ntags: [kind/prd]\n", 1), "\n", "\r\n"),
	}
	for name, text := range variants {
		if got := DraftDigest([]byte(text)); got != base {
			t.Errorf("%s changed the digest: %s != %s", name, got, base)
		}
	}
}

// TestDigestSeesContent: normalization must not be so aggressive that a real
// edit slips through.
func TestDigestSeesContent(t *testing.T) {
	edited := strings.Replace(digestDraft, "Bodies drift.", "Bodies drifted.", 1)
	if DraftDigest([]byte(edited)) == DraftDigest([]byte(digestDraft)) {
		t.Error("an edited body digested the same as the original")
	}
	restatused := strings.Replace(digestDraft, "status: draft", "status: promoted", 1)
	if DraftDigest([]byte(restatused)) == DraftDigest([]byte(digestDraft)) {
		t.Error("a changed status digested the same as the original")
	}
}

// TestDigestFileMatchesBytes is R-6.4: the file-reading entry point and the
// bytes entry point are the same helper, so promotion (which has the file) and
// the gate (which may have either) cannot diverge.
func TestDigestFileMatchesBytes(t *testing.T) {
	ws := fixture(t)
	dir, err := DraftDir(ws, "core")
	if err != nil {
		t.Fatal(err)
	}
	path := DraftPath(dir, "2026-thing")
	write(t, path, strings.ReplaceAll(digestDraft, "\n", "\r\n"))

	fromFile, err := DraftDigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := DraftDigest([]byte(digestDraft)); fromFile != want {
		t.Errorf("DraftDigestFile = %s, DraftDigest = %s", fromFile, want)
	}
}
