package yamlio

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The corpus is deliberately weighted to the shapes the scaffolding commands
// actually emit, plus the decisions that are easy to get wrong and invisible
// until a file-tree diff:
//
//   - '1.0' and '2026.2' are STRINGS that the implicit resolver would reclaim as
//     floats, so safe_dump single-quotes them. Emitting them plain would make
//     every scaffolded platform.yaml and company.yaml read back wrong.
//   - a block sequence that is a mapping VALUE is indentless (`tags:\n- x`),
//     which is the single largest visual difference from yaml.v3.
//   - the 85-column `precedence:` line in a scaffolded team.yaml is NOT folded,
//     because no space on it sits past column 80.
//   - empty collections go to flow style: `requirements: []`, `components: {}`.
//   - a float loses its TYPE if it re-emits without a fraction, so `1.0` must
//     not come back as `1`.
//   - a key of 123 runes or more, an empty key, and a multiline key each fall
//     back to the explicit `? key\n: value` form.
//   - a double-quoted scalar long enough to fold AFTER an escape used to panic.
var emitCorpus = []struct{ name, yaml string }{
	{"company", `
schemaVersion: '1.0'
kind: Company
metadata:
  id: acme-inc
  name: Acme Inc.
tags:
- kind/company
`},
	{"baseline-78-column-plain", `
controls:
- id: security-baseline
  version: '1.0'
  level: default
  requirement: Services authenticate inbound calls and encrypt data in transit.
`},
	{"team-85-column-unfolded", `
agentSkills:
  canonicalPath: skills/
  precedence: canonical-mandatory > personal > canonical-default > canonical-guidance
  onConflict: prefer-canonical-and-inform-user
`},
	{"empty-collections", `
requirements: []
components: {}
platform:
  id: my-platform
`},
	{"registry-flow-in-block-out", `
schemaVersion: '1.0'
kind: IdRegistry
ids:
- {id: 'platform://communications', definedIn: platforms/communications/platform.yaml}
- {id: 'component://customer-notification-service', definedIn: platforms/communications/components/customer-notification-service.yaml}
tags: [ontology/registry]
`},
	{"scalars-needing-quotes", `
a: '1.0'
b: '2026.2'
c: 'yes'
d: 'null'
e: ''
f: '2035-01-15'
g: '0755'
h: '12:30'
`},
	{"scalars-staying-plain", `
a: platform://communications
b: repo://svc
c: skills/
d: scratchpad/personal-rules/
e: canonical-mandatory > personal
f: Acme Inc.
g: team://TODO
`},
	{"non-string-scalars", `
a: true
b: false
c: 42
d: -7
e: 3.5
f: null
g: 2035-01-15
h: 1e30
`},
	{"indicator-scalars", `
a: '#hash'
b: '- dash'
c: 'key: value'
d: 'trailing '
e: ' leading'
f: '---fence'
g: "quote'inside"
h: '@at'
`},
	{"nested-and-long", `
platformRelationships:
- platform: platform://communications
  relationship: belongs-to
deep:
  one:
    two:
      three: a much longer sentence that runs past the eightieth column so the fold is exercised here
`},
	{"unicode-forces-double-quotes", `
name: Café Ñandú
`},
	{"long-single-quoted", `
value: '1.0 aaaaaaaa bbbbbbbb cccccccc dddddddd eeeeeeee ffffffff gggggggg hhhhhhhh iiiiiiii'
`},
	// P1: integral floats must keep their fraction or safe_load reclaims them as
	// ints, and the exponent threshold is CPython's, not Go's.
	{"floats-keep-their-type", `
a: 1.0
b: -1.0
c: 1234567.0
d: 1000000000000000.0
e: 1.0e+16
f: 0.0001
g: 1.0e-05
h: 6.02e+23
i: .inf
j: -.inf
k: .nan
`},
	// P2: check_simple_key's three fallbacks to the explicit `? key` form.
	{"explicit-key-too-long", `
kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk: v
`},
	{"simple-key-just-under-the-bound", `
kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk: v
`},
	{"explicit-key-empty", `
'': v
`},
	{"explicit-key-multiline", `
"a\nb": v
`},
	{"explicit-key-collection-values", `
? kkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkkk
: - a
  - b
? aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
: {}
`},
	// P0: a non-ASCII rune escapes to six characters, which pushes the column
	// past best_width one iteration after the escape set start = end+1.
	{"double-quoted-fold-right-after-an-escape", `
name: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaébb
`},
	{"double-quoted-fold-with-many-escapes", `
name: "ééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééééé"
`},
}

// TestEmitterMatchesPyYAML re-runs yaml.safe_dump over the same documents and
// asserts byte equality. It is the only thing standing between this port and a
// silently different file tree: the differential harness compares file BYTES,
// and every YAML artifact the scaffolding commands write is produced here.
//
// It follows the oracle pattern of internal/frontmatter and internal/workspace —
// skip, never pass, when python3 or the vendored PyYAML is unavailable.
func TestEmitterMatchesPyYAML(t *testing.T) {
	env := oracleEnv(t)
	for _, c := range emitCorpus {
		t.Run(c.name, func(t *testing.T) {
			src := strings.TrimLeft(c.yaml, "\n")
			path := filepath.Join(t.TempDir(), "doc.yaml")
			if err := os.WriteFile(path, []byte(src), 0o666); err != nil {
				t.Fatal(err)
			}
			loaded, err := PyLoadFile(path)
			if err != nil {
				t.Fatalf("PyLoadFile: %v", err)
			}
			got, err := PyDump(loaded)
			if err != nil {
				t.Fatalf("PyDump: %v", err)
			}
			want := safeDump(t, env, src)
			if got != want {
				t.Fatalf("emitter diverged from safe_dump\n--- python\n%s--- go\n%s", want, got)
			}
		})
	}
}

// TestEmitterIsAFixedPoint is what `graph build; graph build` and repeated
// `add` runs depend on: re-dumping an emitter's own output must not change it.
func TestEmitterIsAFixedPoint(t *testing.T) {
	for _, c := range emitCorpus {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			first := filepath.Join(dir, "a.yaml")
			if err := os.WriteFile(first, []byte(strings.TrimLeft(c.yaml, "\n")), 0o666); err != nil {
				t.Fatal(err)
			}
			loaded, err := PyLoadFile(first)
			if err != nil {
				t.Fatalf("PyLoadFile: %v", err)
			}
			once, err := PyDump(loaded)
			if err != nil {
				t.Fatalf("PyDump: %v", err)
			}
			second := filepath.Join(dir, "b.yaml")
			if err := os.WriteFile(second, []byte(once), 0o666); err != nil {
				t.Fatal(err)
			}
			reloaded, err := PyLoadFile(second)
			if err != nil {
				t.Fatalf("PyLoadFile (round 2): %v", err)
			}
			twice, err := PyDump(reloaded)
			if err != nil {
				t.Fatalf("PyDump (round 2): %v", err)
			}
			if once != twice {
				t.Fatalf("not a fixed point\n--- first\n%s--- second\n%s", once, twice)
			}
		})
	}
}

// --------------------------------------------------------------------- P0

// TestWriteDoubleQuotedFoldsAfterEscape is the panic repro reduced to the
// emitter. write_double_quoted's escape branch leaves start one PAST end, and
// the fold test on that same iteration then sliced text[end+1:end].
//
// Python tolerates it — text[start:end] is "" when start > end
// (emitter.py:959) — where Go panicked with "slice bounds out of range". Reached
// by `company-os init --company "<80 ASCII>ébb"` and by 151 of 700 fuzzed
// documents.
func TestWriteDoubleQuotedFoldsAfterEscape(t *testing.T) {
	cases := []string{
		strings.Repeat("a", 80) + "ébb",
		strings.Repeat("é", 40),
		strings.Repeat("a", 78) + "é " + strings.Repeat("b", 20),
		"é" + strings.Repeat("a", 200),
		strings.Repeat("aé", 60),
	}
	for i, s := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			// The panic was inside PyDump; returning at all is the assertion.
			out, err := PyDump(PyMap{{K: "name", V: PyStr(s)}})
			if err != nil {
				t.Fatalf("PyDump: %v", err)
			}
			if !strings.HasPrefix(out, "name: \"") {
				t.Fatalf("expected a double-quoted scalar, got:\n%s", out)
			}
		})
	}
}

// TestWriteDoubleQuotedFoldMatchesPyYAML pins the bytes, not just the absence of
// a panic — a fold that lands one character early is still a wrong file.
func TestWriteDoubleQuotedFoldMatchesPyYAML(t *testing.T) {
	env := oracleEnv(t)
	for n := 70; n <= 90; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			value := strings.Repeat("a", n) + "ébb"
			got, err := PyDump(PyMap{{K: "name", V: PyStr(value)}})
			if err != nil {
				t.Fatalf("PyDump: %v", err)
			}
			if want := safeDumpValue(t, env, value); got != want {
				t.Fatalf("fold diverged\n--- python\n%s--- go\n%s", want, got)
			}
		})
	}
}

// --------------------------------------------------------------------- P1

// pyFloatSamples is `yaml.safe_dump(<float>)` under the vendored PyYAML 6.0.2,
// keyed by the IEEE-754 bit pattern so a sample survives being read by a human
// who might otherwise "tidy" 1.0 into 1.
//
// Generated once from the vendored oracle; the strings below are its output, not
// a transcription of what CPython is believed to do. Regenerate one with
//
//	PYTHONPATH=vendor python3 -c 'import yaml,struct;
//	  print(yaml.safe_dump(struct.unpack(">d", bytes.fromhex("<bits>"))[0]))'
//
// The head is every boundary the two implementations disagree on — integral
// values, the 1e16 and 1e-05 exponent thresholds, subnormals, the extremes — and
// the tail is a fixed-seed sample of arbitrary bit patterns.
var pyFloatSamples = []struct {
	bits uint64
	want string
}{
	{bits: 0x0000000000000000, want: "0.0"},
	{bits: 0x8000000000000000, want: "-0.0"},
	{bits: 0x3ff0000000000000, want: "1.0"},
	{bits: 0xbff0000000000000, want: "-1.0"},
	{bits: 0x3fe0000000000000, want: "0.5"},
	{bits: 0x3ff8000000000000, want: "1.5"},
	{bits: 0x4000000000000000, want: "2.0"},
	{bits: 0x4024000000000000, want: "10.0"},
	{bits: 0x4059000000000000, want: "100.0"},
	{bits: 0x412e848000000000, want: "1000000.0"},
	{bits: 0x4132d68700000000, want: "1234567.0"},
	{bits: 0x41678c29c0000000, want: "12345678.0"},
	{bits: 0x430c6bf526340000, want: "1000000000000000.0"},
	{bits: 0x4341c37937e08000, want: "1.0e+16"},
	{bits: 0x4376345785d8a000, want: "1.0e+17"},
	{bits: 0x444b1ae4d6e2ef50, want: "1.0e+21"},
	{bits: 0x4480f0cf064dd592, want: "1.0e+22"},
	{bits: 0x3f1a36e2eb1c432d, want: "0.0001"},
	{bits: 0x3ee4f8b588e368f1, want: "1.0e-05"},
	{bits: 0x3eb0c6f7a0b5ed8d, want: "1.0e-06"},
	{bits: 0x3de12e0be826d695, want: "1.25e-10"},
	{bits: 0x400921fb54442d18, want: "3.141592653589793"},
	{bits: 0x4005bf0a8b145769, want: "2.718281828459045"},
	{bits: 0x54b249ad2594c37d, want: "1.0e+100"},
	{bits: 0x2b2bff2ee48e0530, want: "1.0e-100"},
	{bits: 0x7fefffffffffffff, want: "1.7976931348623157e+308"},
	{bits: 0x0000000000000001, want: "5.0e-324"},
	{bits: 0x0010000000000000, want: "2.2250738585072014e-308"},
	{bits: 0x437b69b4ba630f35, want: "1.2345678901234568e+17"},
	{bits: 0x44dfde9f10a8d361, want: "6.02e+23"},
	{bits: 0x7e37e43c8800759c, want: "1.0e+300"},
	{bits: 0xfe37e43c8800759c, want: "-1.0e+300"},
	{bits: 0x3fb999999999999a, want: "0.1"},
	{bits: 0x3fd3333333333333, want: "0.3"},
	{bits: 0x3fd5555555555555, want: "0.3333333333333333"},
	{bits: 0x3f50624dd2f1a9fc, want: "0.001"},
	{bits: 0x4341c37937e07fff, want: "9999999999999998.0"},
	{bits: 0x42a2309ce5400000, want: "10000000000000.0"},
	{bits: 0x405edd2f1a9fbe77, want: "123.456"},
	{bits: 0x416312d000000000, want: "10000000.0"},
	{bits: 0xbf1a36e2eb1c432d, want: "-0.0001"},
	{bits: 0xbee4f8b588e368f1, want: "-1.0e-05"},
	{bits: 0xdda1494c73cf256d, want: "-1.0539756485530688e+143"},
	{bits: 0xdb5b5fab8f4d3e27, want: "-1.2143721666352482e+132"},
	{bits: 0xc7fde805ec99108d, want: "-6.360375184993316e+38"},
	{bits: 0x73ab48767734d7c1, want: "1.526087629730313e+249"},
	{bits: 0xdae445508201e2bd, want: "-7.02551513787579e+129"},
	{bits: 0x309d6b79965eda32, want: "1.6260771785726701e-74"},
	{bits: 0xcdcc69292f45e678, want: "-5.984009685093657e+66"},
	{bits: 0x79cb9e86830c71c2, want: "4.8959585264576934e+278"},
	{bits: 0x9d2c67eda13ffe79, want: "-3.7634144828029947e-168"},
	{bits: 0x2fa91425cb008853, want: "4.2301541954837196e-79"},
	{bits: 0x7253edc618187993, want: "5.315422088462547e+242"},
	{bits: 0x244caf9c4dabb481, want: "7.893354548061464e-134"},
	{bits: 0x89e7d15f17362f25, want: "-6.0511296129096725e-261"},
	{bits: 0xe3eff9c0cf44dd3f, want: "-2.471417988383651e+173"},
	{bits: 0xa26b7f62b1852f27, want: "-7.046717376279707e-143"},
	{bits: 0x986e86cb0ab8ab67, want: "-5.352667702512672e-191"},
	{bits: 0x656abd72fb710734, want: "3.467443581611296e+180"},
	{bits: 0x73f778aaf6fa5db8, want: "4.201212419249933e+250"},
	{bits: 0xbd299753a7677796, want: "-4.545896140860994e-14"},
	{bits: 0xa66b0d389d95847e, want: "-1.278808376866442e-123"},
	{bits: 0x9f8558a628518867, want: "-7.773821756658801e-157"},
	{bits: 0xd4ea65d003d71684, want: "-1.1547680013076696e+101"},
	{bits: 0x102b938b8743feb6, want: "8.88116792998917e-231"},
	{bits: 0x09208a650f3ebdd3, want: "1.0259476444142278e-264"},
	{bits: 0xe12b2b8f30b17d0b, want: "-1.1937126864956989e+160"},
	{bits: 0x998092253deffa38, want: "-7.616900394922395e-186"},
	{bits: 0xc7321cc007b37e14, want: "-9.404446511312006e+34"},
	{bits: 0x5387f61376c468ae, want: "2.4990667113411633e+94"},
}

// TestPyFloatMatchesPyYAML is the round-trip-corruption guard. Go's
// strconv.FormatFloat(v,'g',-1,64) drops the fraction from an integral float, so
// `version: 1.0` in an authored registry re-emitted as `version: 1` and came
// back from safe_load as an INT. 104 of 114 sampled floats mismatched.
func TestPyFloatMatchesPyYAML(t *testing.T) {
	for _, c := range pyFloatSamples {
		v := math.Float64frombits(c.bits)
		got, _, err := PyFloat(v).pyRepr()
		if err != nil {
			t.Fatalf("%#016x: %v", c.bits, err)
		}
		if got != c.want {
			t.Errorf("%#016x: got %q, want %q", c.bits, got, c.want)
		}
	}
}

// TestPyFloatRoundTripsAsAFloat asserts the consequence the sample table exists
// to prevent, directly: whatever the emitter writes for a float must resolve
// back to a float.
func TestPyFloatRoundTripsAsAFloat(t *testing.T) {
	for _, c := range pyFloatSamples {
		v := math.Float64frombits(c.bits)
		text, _, err := PyFloat(v).pyRepr()
		if err != nil {
			t.Fatalf("%#016x: %v", c.bits, err)
		}
		s, err := ResolvePlain(text)
		if err != nil {
			t.Fatalf("%#016x: emitted %q, which does not resolve: %v", c.bits, text, err)
		}
		if s.Kind != KindFloat {
			t.Errorf("%#016x: emitted %q, which resolves as %s, not a float",
				c.bits, text, s.Kind.Tag())
		}
	}
}

// TestPyFloatSpecials covers the three values represent_float branches on before
// it reaches repr().
func TestPyFloatSpecials(t *testing.T) {
	cases := []struct {
		v    float64
		want string
	}{
		{math.NaN(), ".nan"},
		{math.Inf(1), ".inf"},
		{math.Inf(-1), "-.inf"},
	}
	for _, c := range cases {
		got, _, err := PyFloat(c.v).pyRepr()
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("%v: got %q, want %q", c.v, got, c.want)
		}
	}
}

// --------------------------------------------------------------------- P2

// TestCheckSimpleKey pins the three fallbacks and the exact rune bound.
// Measured against the vendored PyYAML: a 122-rune key stays simple and a
// 123-rune key does not, because check_simple_key bounds
// `len("!!str") + len(key)` at 128.
func TestCheckSimpleKey(t *testing.T) {
	cases := []struct {
		name string
		key  string
		want bool
	}{
		{"short", "id", true},
		{"122-runes", strings.Repeat("k", 122), true},
		{"123-runes", strings.Repeat("k", 123), false},
		{"124-runes", strings.Repeat("k", 124), false},
		// The bound counts runes, not bytes: 123 two-byte runes is 123, not 246.
		{"122-non-ascii-runes", strings.Repeat("é", 122), true},
		{"123-non-ascii-runes", strings.Repeat("é", 123), false},
		{"empty", "", false},
		{"multiline", "a\nb", false},
		{"trailing-space", "a ", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := checkSimpleKey(c.key); got != c.want {
				t.Fatalf("checkSimpleKey(%d runes) = %v, want %v",
					len([]rune(c.key)), got, c.want)
			}
		})
	}
}

// TestExplicitKeyForm asserts the emitted shape, since checkSimpleKey returning
// false is only half the fix — the `?` indicator, the line break before `:`, and
// the space after it are all part of what safe_load reads back.
func TestExplicitKeyForm(t *testing.T) {
	cases := []struct {
		name string
		in   PyMap
		want string
	}{
		{"long", PyMap{{K: strings.Repeat("k", 123), V: PyStr("v")}},
			"? " + strings.Repeat("k", 123) + "\n: v\n"},
		{"empty", PyMap{{K: "", V: PyStr("v")}}, "? ''\n: v\n"},
		{"multiline", PyMap{{K: "a\nb", V: PyStr("v")}}, "? 'a\n\n  b'\n: v\n"},
		{"nested", PyMap{{K: "outer", V: PyMap{{K: strings.Repeat("k", 130), V: PyStr("v")}}}},
			"outer:\n  ? " + strings.Repeat("k", 130) + "\n  : v\n"},
		{"seq-value", PyMap{{K: strings.Repeat("k", 130), V: PySeq{PyStr("a"), PyStr("b")}}},
			"? " + strings.Repeat("k", 130) + "\n: - a\n  - b\n"},
		{"just-under", PyMap{{K: strings.Repeat("k", 122), V: PyStr("v")}},
			strings.Repeat("k", 122) + ": v\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := PyDump(c.in)
			if err != nil {
				t.Fatalf("PyDump: %v", err)
			}
			if got != c.want {
				t.Fatalf("got:\n%q\nwant:\n%q", got, c.want)
			}
		})
	}
}

// ------------------------------------------------------------- truthiness

// TestPyFalsyAgreesWithIsFalsy keeps the object-level and node-level truthiness
// tests from drifting apart. registerID branches on the first and `today` on the
// second, over the same `or default` in the same Python source.
func TestPyFalsyAgreesWithIsFalsy(t *testing.T) {
	docs := []string{
		"", "null\n", "[]\n", "{}\n", "''\n", "0\n", "0.0\n", "false\n",
		"a\n", "1\n", "0.5\n", "true\n", "[1]\n", "{a: 1}\n", "2026-01-01\n",
	}
	dir := t.TempDir()
	for i, src := range docs {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			path := filepath.Join(dir, fmt.Sprintf("d%d.yaml", i))
			if err := os.WriteFile(path, []byte(src), 0o666); err != nil {
				t.Fatal(err)
			}
			v, err := PyLoadFile(path)
			if err != nil {
				t.Fatalf("PyLoadFile: %v", err)
			}
			doc, err := Load([]byte(src))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got, want := PyFalsy(v), doc.IsFalsy(); got != want {
				t.Fatalf("%q: PyFalsy=%v, IsFalsy=%v", src, got, want)
			}
		})
	}
}

// ---------------------------------------------------------------- oracle

// oracleEnv locates the vendored PyYAML and skips when it or python3 is absent.
func oracleEnv(t *testing.T) []string {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available; cannot re-run the oracle")
	}
	moduleRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Skipf("cannot locate the module root: %v", err)
	}
	vendor := filepath.Join(moduleRoot, "vendor")
	if _, err := os.Stat(filepath.Join(vendor, "yaml")); err != nil {
		t.Skipf("vendored PyYAML unavailable: %v", err)
	}
	return append(os.Environ(), "PYTHONPATH="+vendor)
}

const safeDumpScript = `
import sys, yaml
sys.stdout.write(yaml.safe_dump(yaml.safe_load(sys.stdin.read()),
                                sort_keys=False, default_flow_style=False))
`

func safeDump(t *testing.T, env []string, src string) string {
	t.Helper()
	return runOracle(t, env, safeDumpScript, src)
}

// safeDumpValueScript dumps `{"name": <stdin verbatim>}` so the value reaches
// PyYAML without a YAML parse in between — the fold cases above carry runes a
// loader would otherwise be free to re-encode.
const safeDumpValueScript = `
import sys, yaml
sys.stdout.write(yaml.safe_dump({"name": sys.stdin.read()},
                                sort_keys=False, default_flow_style=False))
`

func safeDumpValue(t *testing.T, env []string, value string) string {
	t.Helper()
	return runOracle(t, env, safeDumpValueScript, value)
}

func runOracle(t *testing.T, env []string, script, stdin string) string {
	t.Helper()
	cmd := exec.Command("python3", "-c", script)
	cmd.Env = env
	cmd.Stdin = strings.NewReader(stdin)
	var out, errb strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		t.Fatalf("safe_dump oracle failed: %v\n%s", err, errb.String())
	}
	return out.String()
}
