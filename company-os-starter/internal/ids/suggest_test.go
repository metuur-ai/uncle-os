package ids

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// The difflib port is measured, not eyeballed: testdata/difflib-oracle.json is
// CPython's own difflib answering the same questions. Regenerate with
//
//	python3 - > internal/ids/testdata/difflib-oracle.json <<'PY'
//	import difflib, json, random
//	random.seed(7)
//	pool = [...]; words = [...]
//	matches=[{"word": w, "want": difflib.get_close_matches(w, pool, n=3, cutoff=0.3)}
//	         for w in words]
//	alph="abcdefg-"; ratios=[]
//	for _ in range(200):
//	    a="".join(random.choice(alph) for _ in range(random.randint(0,14)))
//	    b="".join(random.choice(alph) for _ in range(random.randint(0,14)))
//	    sm=difflib.SequenceMatcher(None,a,b)
//	    ratios.append({"a":a,"b":b,"ratio":sm.ratio(),
//	                   "quick":sm.quick_ratio(),"real":sm.real_quick_ratio()})
//	print(json.dumps({"pool":pool,"matches":matches,"ratios":ratios}, indent=1))
//	PY
type difflibOracle struct {
	Pool    []string `json:"pool"`
	Matches []struct {
		Word string   `json:"word"`
		Want []string `json:"want"`
	} `json:"matches"`
	Ratios []struct {
		A     string  `json:"a"`
		B     string  `json:"b"`
		Ratio float64 `json:"ratio"`
		Quick float64 `json:"quick"`
		Real  float64 `json:"real"`
	} `json:"ratios"`
}

func loadOracle(t *testing.T) difflibOracle {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "difflib-oracle.json"))
	if err != nil {
		t.Fatalf("reading oracle: %v", err)
	}
	var o difflibOracle
	if err := json.Unmarshal(data, &o); err != nil {
		t.Fatalf("parsing oracle: %v", err)
	}
	return o
}

// TestRatiosMatchCPython pins the three SequenceMatcher ratios against CPython
// on 200 random pairs. The two quick ratios are upper bounds used only as a
// prefilter, but a wrong one silently drops a candidate before ratio ever runs.
func TestRatiosMatchCPython(t *testing.T) {
	for _, c := range loadOracle(t).Ratios {
		a, b := []rune(c.A), []rune(c.B)
		if got := ratio(a, b); !closeEnough(got, c.Ratio) {
			t.Errorf("ratio(%q,%q) = %v, want %v", c.A, c.B, got, c.Ratio)
		}
		if got := quickRatio(a, b); !closeEnough(got, c.Quick) {
			t.Errorf("quickRatio(%q,%q) = %v, want %v", c.A, c.B, got, c.Quick)
		}
		if got := realQuickRatio(a, b); !closeEnough(got, c.Real) {
			t.Errorf("realQuickRatio(%q,%q) = %v, want %v", c.A, c.B, got, c.Real)
		}
	}
}

// TestCloseMatchesMatchCPython pins the whole get_close_matches contract: the
// cutoff, the cap at three, and — the part a reimplementation gets wrong — the
// tie-break, which is heapq.nlargest over (score, candidate) tuples and so is
// descending by candidate string, not input order.
func TestCloseMatchesMatchCPython(t *testing.T) {
	o := loadOracle(t)
	for _, c := range o.Matches {
		got := closeMatches(c.Word, o.Pool, 3, 0.3)
		if len(got) != len(c.Want) {
			t.Errorf("closeMatches(%q) = %v, want %v", c.Word, got, c.Want)
			continue
		}
		for i := range got {
			if got[i] != c.Want[i] {
				t.Errorf("closeMatches(%q) = %v, want %v", c.Word, got, c.Want)
				break
			}
		}
	}
}

// closeEnough compares two float ratios. Both sides compute 2*M/T from integers,
// so they agree bit for bit in practice; the epsilon guards the comparison
// rather than papering over an algorithmic difference.
func closeEnough(got, want float64) bool { return math.Abs(got-want) < 1e-12 }

// TestLocalNameTakesTheLastSegment pins `s.split("://")[-1]` on the shapes the
// registry actually holds, including an id with no scheme at all.
func TestLocalNameTakesTheLastSegment(t *testing.T) {
	cases := map[string]string{
		"component://customer-notification-service":    "customer-notification-service",
		"capability://communications/message-delivery": "communications/message-delivery",
		"bare-name": "bare-name",
		"":          "",
	}
	for in, want := range cases {
		if got := localName(in); got != want {
			t.Errorf("localName(%q) = %q, want %q", in, got, want)
		}
	}
}
