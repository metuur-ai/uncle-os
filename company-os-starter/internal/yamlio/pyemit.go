package yamlio

// A PyYAML-`safe_dump`-compatible emitter, for the value space the scaffolding
// commands write.
//
// Why this exists rather than yaml.v3's encoder: every YAML source file `init`,
// `add` and `reality new` produce goes through `canonical_or_block`
// (bin/company-os:1888) or `dump_yaml` (:65), both of which are
// `yaml.safe_dump(..., sort_keys=False, default_flow_style=False)`. The 0.3
// differential harness compares the resulting file tree byte-for-byte, so the
// emitter's signature is part of the contract. yaml.v3 diverges on three things
// at once — it double-quotes where PyYAML single-quotes, indents block
// sequences where PyYAML emits them indentless, and defaults to a 4-space
// indent — and this package's round-trip measurement (task 1.3) already froze
// those as unfixable through yaml.v3's knobs.
//
// `register_id` (:1815) is what makes a real emitter unavoidable: it loads
// company-ontology/ids/registry.yaml, appends one entry, and re-dumps the WHOLE
// file. Measured on examples/workspace, that rewrites seven flow-style entries
// into block style and `tags: [ontology/registry]` into an indentless block
// sequence — arbitrary authored content flowing back out through safe_dump.
//
// The writer primitives (writeIndent, writeIndicator, writePlain,
// writeSingleQuoted, writeDoubleQuoted), analyzeScalar and checkSimpleKey are
// transliterated from vendor/yaml/emitter.py. The 80-column fold in writePlain
// is the reason they are transliterated rather than approximated: PyYAML only
// breaks at a space whose column ALREADY exceeds best_width, so
// `precedence: canonical-mandatory > … > canonical-guidance` in a scaffolded
// team.yaml stays on one 85-column line. A "wrap at 80" reimplementation gets
// that file wrong.
//
// Scope, deliberate: this covers block style, non-canonical output, default
// flow style off, and the scalar types PyYAML's SafeRepresenter emits for str,
// bool, int, float, None, date and datetime. It does NOT implement literal or
// folded block scalars — nothing in the represented data selects them, since
// represent_str always emits style=None.
//
// It lives here, next to the loader, because internal/scaffold, internal/
// governance (`deviation declare`) and internal/federation (the lock) all need
// it and none of them may reach it through another command package.

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"gopkg.in/yaml.v3"
)

// PyValue is one Python object as PyYAML's SafeRepresenter would see it. The
// closed set below is exactly what safe_load can produce for the artifacts the
// command packages read and what their scaffolds construct.
type PyValue interface {
	pyRepr() (text string, isStr bool, err error)
}

type (
	// PyStr is a Python str: the only kind whose plain form may need quoting
	// to survive a round trip.
	PyStr string
	// PyBool is a Python bool.
	PyBool bool
	// PyInt is a Python int. Arbitrary precision, as PyYAML's ints are.
	PyInt struct{ N *big.Int }
	// PyFloat is a Python float.
	PyFloat float64
	// PyNull is Python None.
	PyNull struct{}
	// PyTime is a datetime.date or datetime.datetime, pre-rendered by
	// isoformat exactly as represent_date/represent_datetime render it.
	PyTime string
	// PySeq is a Python list.
	PySeq []PyValue
	// PyPair is one ordered mapping entry.
	PyPair struct {
		K string
		V PyValue
	}
	// PyMap is a Python dict. Insertion order is preserved because safe_dump
	// runs with sort_keys=False and the authored order is what the oracle
	// emits.
	PyMap []PyPair
)

func (s PyStr) pyRepr() (string, bool, error) { return string(s), true, nil }

func (b PyBool) pyRepr() (string, bool, error) {
	if bool(b) {
		return "true", false, nil
	}
	return "false", false, nil
}

func (i PyInt) pyRepr() (string, bool, error) { return i.N.String(), false, nil }

// pyRepr is represent_float (vendor/yaml/representer.py:171): repr() lowercased,
// with an exponent-only form given an explicit ".0" mantissa separator.
func (f PyFloat) pyRepr() (string, bool, error) {
	v := float64(f)
	switch {
	case math.IsNaN(v):
		return ".nan", false, nil
	case math.IsInf(v, 1):
		return ".inf", false, nil
	case math.IsInf(v, -1):
		return "-.inf", false, nil
	}
	// repr(), which scalar.go already reproduces for Scalar.String() — the same
	// question, asked on the way in rather than on the way out.
	//
	// strconv.FormatFloat(v,'g',-1,64) was NOT a substitute, and the difference
	// was a type change rather than a cosmetic one: Go's 'g' drops the trailing
	// ".0", so the float 1.0 emitted as `1` and safe_load read it back as an
	// INT. Its exponent threshold is Go's, not CPython's decpt > 16, so
	// 1234567.0 emitted as 1.234567e+06. 104 of 114 sampled floats mismatched.
	text := strings.ToLower(pyFloat(v))
	if !strings.Contains(text, ".") && strings.Contains(text, "e") {
		text = strings.Replace(text, "e", ".0e", 1)
	}
	return text, false, nil
}

func (PyNull) pyRepr() (string, bool, error)   { return "null", false, nil }
func (t PyTime) pyRepr() (string, bool, error) { return string(t), false, nil }

func (PySeq) pyRepr() (string, bool, error) {
	return "", false, fmt.Errorf("yamlio: pyRepr called on a sequence")
}

func (PyMap) pyRepr() (string, bool, error) {
	return "", false, fmt.Errorf("yamlio: pyRepr called on a mapping")
}

// Get returns the value at key, or nil.
func (m PyMap) Get(key string) PyValue {
	for _, p := range m {
		if p.K == key {
			return p.V
		}
	}
	return nil
}

// Set replaces the value at key in place, or appends it, which is what a Python
// dict assignment does to an insertion-ordered mapping.
func (m PyMap) Set(key string, v PyValue) PyMap {
	for i := range m {
		if m[i].K == key {
			m[i].V = v
			return m
		}
	}
	return append(m, PyPair{K: key, V: v})
}

// PyFalsy is `if not value:` over a loaded Python object — the truthiness test
// that `load_yaml(path, None) or {…}` (bin/company-os:63, :1817) branches on.
//
// It is the object-level twin of this package's node-level isFalsy: PyLoadFile
// hands back Python objects rather than nodes, so the same question has to be
// answerable on both sides. TestPyFalsyAgreesWithIsFalsy pins the two together
// so they cannot drift.
func PyFalsy(v PyValue) bool {
	switch t := v.(type) {
	case nil:
		return true
	case PyNull:
		return true
	case PyBool:
		return !bool(t)
	case PyInt:
		return t.N == nil || t.N.Sign() == 0
	case PyFloat:
		return float64(t) == 0
	case PyStr:
		return t == ""
	case PySeq:
		return len(t) == 0
	case PyMap:
		return len(t) == 0
	}
	// datetime.date and datetime.datetime are always truthy.
	return false
}

// ---------------------------------------------------------------- emitting

const (
	pyBestIndent = 2  // safe_dump's indent default
	pyBestWidth  = 80 // safe_dump's width default
	// pySimpleKeyMax is check_simple_key's `length < 128` (emitter.py:454). The
	// length it bounds is the prepared TAG plus the analyzed scalar, and every
	// key safe_dump writes here is a str, whose prepared tag is the five
	// characters "!!str" — so the effective bound on a key is 123 runes.
	// Measured: a 122-rune key stays simple, a 123-rune key emits as
	// `? key\n: value`.
	pySimpleKeyMax = 128
	pyStrTagLen    = len("!!str")
)

// PyDump is yaml.safe_dump(data, sort_keys=False, default_flow_style=False).
func PyDump(v PyValue) (string, error) {
	return pyDump(v, pyOptions{})
}

// PyDumpAutoFlow is yaml.safe_dump(data, sort_keys=False,
// default_flow_style=None) — the serialization rewrite_frontmatter_tags
// (bin/company-os:1354) uses, and the ONLY caller that needs flow style.
//
// `None` is not a third style: it hands the choice back to the representer,
// which sets a collection's flow_style to `best_style` — True when every child
// is a plain scalar (represent_sequence/represent_mapping,
// vendor/yaml/representer.py:114-149). That is why committed frontmatter reads
// `tags: [kind/reality, platform/communications]` and `pointers:` stays block
// (its items are mappings) while each pointer mapping is flow (its values are
// scalars). Emitting those in block style would rewrite every doc `graph build`
// touches and break R-0.6 on the first pass.
func PyDumpAutoFlow(v PyValue) (string, error) {
	return pyDump(v, pyOptions{flowAuto: true})
}

// PyDumpCanonical is canonical_yaml (bin/company-os:96-99):
// yaml.safe_dump(data, sort_keys=True, default_flow_style=False,
// allow_unicode=True).
//
// It is the SEMANTIC comparison form for derived YAML artifacts (R-0.7c) — two
// documents that parse to the same structure produce the same text here no
// matter how either was laid out — and the single writer for them, so a
// write and the drift gate that re-derives it cannot disagree.
func PyDumpCanonical(v PyValue) (string, error) {
	return pyDump(pySortKeys(v), pyOptions{allowUnicode: true})
}

// PyWriteCanonical is dump_canonical_yaml (bin/company-os:102-106).
func PyWriteCanonical(path string, v PyValue) error {
	text, err := PyDumpCanonical(v)
	if err != nil {
		return model.Errorf(model.ExitArtifact, "cannot serialize %s: %v", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(text), 0o666); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
	}
	return nil
}

// pySortKeys is representer.represent_mapping's `sorted(mapping)` under
// sort_keys=True, applied at every level. Keys are always str here (fromNode
// rejects anything else), so the tuple comparison Python performs reduces to a
// key comparison and cannot raise the TypeError safe_dump swallows.
func pySortKeys(v PyValue) PyValue {
	switch t := v.(type) {
	case PyMap:
		out := make(PyMap, len(t))
		copy(out, t)
		sort.SliceStable(out, func(i, j int) bool { return out[i].K < out[j].K })
		for i := range out {
			out[i].V = pySortKeys(out[i].V)
		}
		return out
	case PySeq:
		out := make(PySeq, len(t))
		for i, item := range t {
			out[i] = pySortKeys(item)
		}
		return out
	}
	return v
}

func pyDump(v PyValue, opt pyOptions) (string, error) {
	e := &pyEmitter{indent: -1, whitespace: true, indention: true, opt: opt}
	if err := e.node(v, false, false); err != nil {
		return "", err
	}
	// expect_document_end's write_indent, which is what terminates the final
	// line. It is unconditional in PyYAML and produces the trailing newline.
	e.indent = -1
	e.writeIndent()
	return e.b.String(), nil
}

// pyOptions are the two safe_dump keyword arguments this port varies. sort_keys
// is not here because it is a representer concern, applied by pySortKeys before
// the emitter ever runs — exactly where PyYAML applies it.
type pyOptions struct {
	// flowAuto is default_flow_style=None: let each collection pick.
	flowAuto bool
	// allowUnicode keeps non-ASCII characters as themselves instead of
	// escaping them into a double-quoted scalar.
	allowUnicode bool
}

type pyEmitter struct {
	b          strings.Builder
	column     int
	whitespace bool
	indention  bool
	// indent is Python's self.indent, where -1 stands for None.
	indent int
	// flowLevel is Python's self.flow_level: non-zero anywhere inside a flow
	// collection, which forces every nested collection flow too.
	flowLevel int
	opt       pyOptions
}

func (e *pyEmitter) writeLineBreak() {
	e.whitespace = true
	e.indention = true
	e.column = 0
	e.b.WriteByte('\n')
}

func (e *pyEmitter) writeIndent() {
	indent := e.indent
	if indent < 0 {
		indent = 0
	}
	if !e.indention || e.column > indent || (e.column == indent && !e.whitespace) {
		e.writeLineBreak()
	}
	if e.column < indent {
		e.whitespace = true
		e.b.WriteString(strings.Repeat(" ", indent-e.column))
		e.column = indent
	}
}

func (e *pyEmitter) writeIndicator(indicator string, needWhitespace, whitespace, indention bool) {
	data := indicator
	if !e.whitespace && needWhitespace {
		data = " " + indicator
	}
	e.whitespace = whitespace
	e.indention = e.indention && indention
	e.column += len([]rune(data))
	e.b.WriteString(data)
}

// increaseIndent is Emitter.increase_indent; it returns the previous value so
// the caller can restore it the way PyYAML pops self.indents.
func (e *pyEmitter) increaseIndent(flow, indentless bool) int {
	saved := e.indent
	switch {
	case e.indent < 0:
		if flow {
			e.indent = pyBestIndent
		} else {
			e.indent = 0
		}
	case !indentless:
		e.indent += pyBestIndent
	}
	return saved
}

// node is expect_node for the block-style, non-canonical case.
//
// mappingCtx and simpleKey carry the two pieces of Emitter state that change
// what a nested node does: a sequence that is a mapping VALUE is emitted
// indentless (`tags:\n- kind/company`), and a scalar in key position is never
// line-folded.
func (e *pyEmitter) node(v PyValue, mappingCtx, simpleKey bool) error {
	switch t := v.(type) {
	case PySeq:
		// expect_node's sequence branch (emitter.py:262-267): flow when already
		// inside one, when the representer chose it, or when it is empty
		// (check_empty_sequence).
		if e.flowLevel > 0 || len(t) == 0 || (e.opt.flowAuto && pyBestStyleSeq(t)) {
			return e.flowSeq(t)
		}
		indentless := mappingCtx && !e.indention
		saved := e.increaseIndent(false, indentless)
		for _, item := range t {
			e.writeIndent()
			e.writeIndicator("-", true, false, true)
			if err := e.node(item, false, false); err != nil {
				return err
			}
		}
		e.indent = saved
		return nil

	case PyMap:
		if e.flowLevel > 0 || len(t) == 0 || (e.opt.flowAuto && pyBestStyleMap(t)) {
			return e.flowMap(t)
		}
		saved := e.increaseIndent(false, false)
		for _, pair := range t {
			e.writeIndent()
			if checkSimpleKey(pair.K) {
				// expect_block_mapping_key's simple branch (emitter.py:401-403)
				// followed by expect_block_mapping_simple_value (:409-412).
				if err := e.node(PyStr(pair.K), true, true); err != nil {
					return err
				}
				e.writeIndicator(":", false, false, false)
			} else {
				// expect_block_mapping_key's explicit branch (:405-407) followed
				// by expect_block_mapping_value (:414-418).
				e.writeIndicator("?", true, false, true)
				if err := e.node(PyStr(pair.K), true, false); err != nil {
					return err
				}
				e.writeIndent()
				e.writeIndicator(":", true, false, true)
			}
			if err := e.node(pair.V, true, false); err != nil {
				return err
			}
		}
		e.indent = saved
		return nil

	default:
		saved := e.increaseIndent(true, false)
		err := e.scalar(v, simpleKey)
		e.indent = saved
		return err
	}
}

// pyBestStyleSeq and pyBestStyleMap are represent_sequence's and
// represent_mapping's `best_style` (representer.py:114-149): a collection is
// laid out flow only when every child is a style-less scalar. Nothing here
// produces a styled scalar node — represent_str always passes style=None — so
// "is a scalar" is the whole test.
func pyBestStyleSeq(t PySeq) bool {
	for _, item := range t {
		if pyIsCollection(item) {
			return false
		}
	}
	return true
}

func pyBestStyleMap(t PyMap) bool {
	// Keys are always str, hence always scalars; only the values can veto.
	for _, pair := range t {
		if pyIsCollection(pair.V) {
			return false
		}
	}
	return true
}

func pyIsCollection(v PyValue) bool {
	switch v.(type) {
	case PySeq, PyMap:
		return true
	}
	return false
}

// flowSeq is expect_flow_sequence plus expect_(first_)flow_sequence_item
// (emitter.py:296-325), with the canonical branches dropped. The indent is
// popped and flow_level decremented BEFORE the closing indicator, which is what
// keeps a folded flow collection's `]` at the item indent rather than two
// columns further in.
func (e *pyEmitter) flowSeq(t PySeq) error {
	e.writeIndicator("[", true, true, false)
	e.flowLevel++
	saved := e.increaseIndent(true, false)
	for i, item := range t {
		if i > 0 {
			e.writeIndicator(",", false, false, false)
		}
		if e.column > pyBestWidth {
			e.writeIndent()
		}
		if err := e.node(item, false, false); err != nil {
			return err
		}
	}
	e.indent = saved
	e.flowLevel--
	e.writeIndicator("]", false, false, false)
	return nil
}

// flowMap is expect_flow_mapping and its four item states (emitter.py:327-370).
func (e *pyEmitter) flowMap(t PyMap) error {
	e.writeIndicator("{", true, true, false)
	e.flowLevel++
	saved := e.increaseIndent(true, false)
	for i, pair := range t {
		if i > 0 {
			e.writeIndicator(",", false, false, false)
		}
		if e.column > pyBestWidth {
			e.writeIndent()
		}
		if checkSimpleKey(pair.K) {
			if err := e.node(PyStr(pair.K), true, true); err != nil {
				return err
			}
			// expect_flow_mapping_simple_value (:360-363).
			e.writeIndicator(":", false, false, false)
		} else {
			e.writeIndicator("?", true, false, false)
			if err := e.node(PyStr(pair.K), true, false); err != nil {
				return err
			}
			// expect_flow_mapping_value (:365-370).
			if e.column > pyBestWidth {
				e.writeIndent()
			}
			e.writeIndicator(":", true, false, false)
		}
		if err := e.node(pair.V, true, false); err != nil {
			return err
		}
	}
	e.indent = saved
	e.flowLevel--
	e.writeIndicator("}", false, false, false)
	return nil
}

// checkSimpleKey is Emitter.check_simple_key (emitter.py:437-455) for the only
// key shape safe_dump produces here: an implicitly tagged str scalar.
//
// Omitting it is invisible until a document carries a long, empty or multiline
// key: PyYAML falls back to the explicit `? key\n: value` form there, and a
// `key: value` emission is then not what safe_load reads back. Measured, it
// diverged on 201 of 1200 fuzzed structural documents.
func checkSimpleKey(key string) bool {
	runes := []rune(key)
	// length is the prepared tag plus the analyzed scalar. check_simple_key
	// runs before process_tag has the chance to suppress the tag, so it counts
	// even though nothing writes it.
	if pyStrTagLen+len(runes) >= pySimpleKeyMax {
		return false
	}
	// allow_unicode reaches only allow_*_plain and allow_single_quoted through
	// special_characters; empty and multiline are independent of it, so the
	// flag cannot change this answer.
	a := analyzeScalar(runes, false)
	return !a.empty && !a.multiline
}

func (e *pyEmitter) scalar(v PyValue, simpleKey bool) error {
	text, isStr, err := v.pyRepr()
	if err != nil {
		return err
	}
	runes := []rune(text)
	a := analyzeScalar(runes, e.opt.allowUnicode)

	// serialize_node computes implicit[0] as "the plain form resolves back to
	// this node's tag". Only a str can fail it: every other representer emits
	// text its own resolver reclaims.
	plainRoundTrips := true
	if isStr {
		plainRoundTrips = resolvesToStr(text)
	}

	// choose_scalar_style consults allow_flow_plain inside a flow collection and
	// allow_block_plain outside one (emitter.py:194-197). The two differ on
	// `,`, `[`, `]`, `{` and `}`, which are indicators only in flow context —
	// so a value that is plain in a block mapping becomes single-quoted the
	// moment the same value appears inside `[...]`.
	allowPlain := a.allowBlockPlain
	if e.flowLevel > 0 {
		allowPlain = a.allowFlowPlain
	}

	split := !simpleKey
	switch {
	case plainRoundTrips && !(simpleKey && (a.empty || a.multiline)) && allowPlain:
		e.writePlain(runes, split)
	case a.allowSingleQuoted && !(simpleKey && a.multiline):
		e.writeSingleQuoted(runes, split)
	default:
		e.writeDoubleQuoted(runes, split)
	}
	return nil
}

// resolvesToStr reports whether PyYAML's implicit resolver classifies this
// plain text as a string. Anything else — `1.0`, `2026.2`, `yes`, `2035-01-15`,
// the empty string — must be quoted or safe_load would not read it back as the
// str it is. A resolution ERROR means the resolver matched (an unconstructible
// date is still a !!timestamp match), so it is equally not a str.
func resolvesToStr(text string) bool {
	s, err := ResolvePlain(text)
	return err == nil && s.Kind == KindStr
}

// ------------------------------------------------------- analyze_scalar

type scalarAnalysis struct {
	empty             bool
	multiline         bool
	allowFlowPlain    bool
	allowBlockPlain   bool
	allowSingleQuoted bool
}

// analyzeScalar is Emitter.analyze_scalar (vendor/yaml/emitter.py:626),
// restricted to the flags the flow- and block-style writers consult. allow_block
// is dropped because nothing here emits literal or folded block scalars, and
// allow_double_quoted is unconditionally true in PyYAML.
//
// allowUnicode is False under plain safe_dump, so any non-ASCII character is a
// "special character" and forces double quotes — which is why a company name
// with an accent round-trips as an escaped \uXXXX rather than as UTF-8.
// canonical_yaml (bin/company-os:96) passes allow_unicode=True, where the same
// character stays a plain scalar.
func analyzeScalar(scalar []rune, allowUnicode bool) scalarAnalysis {
	if len(scalar) == 0 {
		return scalarAnalysis{empty: true, allowFlowPlain: true,
			allowBlockPlain: true, allowSingleQuoted: true}
	}

	var (
		blockIndicators   bool
		flowIndicators    bool
		lineBreaks        bool
		specialCharacters bool

		leadingSpace  bool
		leadingBreak  bool
		trailingSpace bool
		trailingBreak bool
		breakSpace    bool
		spaceBreak    bool
	)

	text := string(scalar)
	if strings.HasPrefix(text, "---") || strings.HasPrefix(text, "...") {
		blockIndicators = true
	}

	precededByWhitespace := true
	followedByWhitespace := len(scalar) == 1 || isWS(scalar[1])
	previousSpace, previousBreak := false, false

	for index := 0; index < len(scalar); index++ {
		ch := scalar[index]

		if index == 0 {
			if strings.ContainsRune("#,[]{}&*!|>'\"%@`", ch) {
				flowIndicators = true
				blockIndicators = true
			}
			if ch == '?' || ch == ':' {
				flowIndicators = true
				if followedByWhitespace {
					blockIndicators = true
				}
			}
			if ch == '-' && followedByWhitespace {
				flowIndicators = true
				blockIndicators = true
			}
		} else {
			if strings.ContainsRune(",?[]{}", ch) {
				flowIndicators = true
			}
			if ch == ':' {
				flowIndicators = true
				if followedByWhitespace {
					blockIndicators = true
				}
			}
			if ch == '#' && precededByWhitespace {
				flowIndicators = true
				blockIndicators = true
			}
		}

		if isBreak(ch) {
			lineBreaks = true
		}
		if !isSpecialOK(ch, allowUnicode) {
			specialCharacters = true
		}

		switch {
		case ch == ' ':
			if index == 0 {
				leadingSpace = true
			}
			if index == len(scalar)-1 {
				trailingSpace = true
			}
			if previousBreak {
				breakSpace = true
			}
			previousSpace, previousBreak = true, false
		case isBreak(ch):
			if index == 0 {
				leadingBreak = true
			}
			if index == len(scalar)-1 {
				trailingBreak = true
			}
			if previousSpace {
				spaceBreak = true
			}
			previousSpace, previousBreak = false, true
		default:
			previousSpace, previousBreak = false, false
		}

		precededByWhitespace = isWS(ch)
		followedByWhitespace = index+2 >= len(scalar) || isWS(scalar[index+2])
	}

	a := scalarAnalysis{multiline: lineBreaks, allowFlowPlain: true,
		allowBlockPlain: true, allowSingleQuoted: true}
	if leadingSpace || leadingBreak || trailingSpace || trailingBreak {
		a.allowFlowPlain, a.allowBlockPlain = false, false
	}
	if breakSpace {
		a.allowFlowPlain, a.allowBlockPlain, a.allowSingleQuoted = false, false, false
	}
	if spaceBreak || specialCharacters {
		a.allowFlowPlain, a.allowBlockPlain, a.allowSingleQuoted = false, false, false
	}
	if lineBreaks {
		a.allowFlowPlain, a.allowBlockPlain = false, false
	}
	if flowIndicators {
		a.allowFlowPlain = false
	}
	if blockIndicators {
		a.allowBlockPlain = false
	}
	return a
}

// isSpecialOK is analyze_scalar's printability test (emitter.py:733-742): plain
// ASCII always passes, and the wider Unicode ranges pass only when
// allow_unicode is set. U+FEFF never passes — a byte-order mark inside a scalar
// is escaped whatever the setting.
func isSpecialOK(ch rune, allowUnicode bool) bool {
	if ch == '\n' || (ch >= 0x20 && ch <= 0x7E) {
		return true
	}
	if !allowUnicode || ch == 0xFEFF {
		return false
	}
	return ch == 0x85 || (ch >= 0xA0 && ch <= 0xD7FF) ||
		(ch >= 0xE000 && ch <= 0xFFFD) || (ch >= 0x10000 && ch < 0x10FFFF)
}

func isWS(ch rune) bool {
	return ch == 0 || ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' ||
		ch == 0x85 || ch == 0x2028 || ch == 0x2029
}

func isBreak(ch rune) bool {
	return ch == '\n' || ch == 0x85 || ch == 0x2028 || ch == 0x2029
}

// ------------------------------------------------------------- writers

func (e *pyEmitter) write(data string) {
	e.column += len([]rune(data))
	e.b.WriteString(data)
}

// writePlain is Emitter.write_plain. The fold is the load-bearing part: a break
// happens only at a SINGLE space whose column already exceeds best_width, which
// is why an 85-column line with no space past column 67 is never folded.
func (e *pyEmitter) writePlain(text []rune, split bool) {
	if len(text) == 0 {
		return
	}
	if !e.whitespace {
		e.write(" ")
	}
	e.whitespace, e.indention = false, false

	spaces, breaks := false, false
	start, end := 0, 0
	for end <= len(text) {
		var ch rune
		hasCh := end < len(text)
		if hasCh {
			ch = text[end]
		}
		switch {
		case spaces:
			if !hasCh || ch != ' ' {
				if start+1 == end && e.column > pyBestWidth && split {
					e.writeIndent()
					e.whitespace, e.indention = false, false
				} else {
					e.write(string(text[start:end]))
				}
				start = end
			}
		case breaks:
			if !hasCh || !isBreak(ch) {
				e.writeBreakRun(text[start:end])
				e.writeIndent()
				e.whitespace, e.indention = false, false
				start = end
			}
		default:
			if !hasCh || ch == ' ' || isBreak(ch) {
				e.write(string(text[start:end]))
				start = end
			}
		}
		if hasCh {
			spaces = ch == ' '
			breaks = isBreak(ch)
		}
		end++
	}
}

// writeSingleQuoted is Emitter.write_single_quoted.
func (e *pyEmitter) writeSingleQuoted(text []rune, split bool) {
	e.writeIndicator("'", true, false, false)
	spaces, breaks := false, false
	start, end := 0, 0
	for end <= len(text) {
		var ch rune
		hasCh := end < len(text)
		if hasCh {
			ch = text[end]
		}
		switch {
		case spaces:
			if !hasCh || ch != ' ' {
				if start+1 == end && e.column > pyBestWidth && split &&
					start != 0 && end != len(text) {
					e.writeIndent()
				} else {
					e.write(string(text[start:end]))
				}
				start = end
			}
		case breaks:
			if !hasCh || !isBreak(ch) {
				e.writeBreakRun(text[start:end])
				e.writeIndent()
				start = end
			}
		default:
			if !hasCh || ch == ' ' || isBreak(ch) || ch == '\'' {
				if start < end {
					e.write(string(text[start:end]))
					start = end
				}
			}
		}
		if hasCh && ch == '\'' {
			e.write("''")
			start = end + 1
		}
		if hasCh {
			spaces = ch == ' '
			breaks = isBreak(ch)
		}
		end++
	}
	e.writeIndicator("'", false, false, false)
}

// writeDoubleQuoted is Emitter.write_double_quoted. Under allow_unicode False
// every non-printable-ASCII rune is escaped; under True the printable Unicode
// ranges pass through as themselves.
func (e *pyEmitter) writeDoubleQuoted(text []rune, split bool) {
	e.writeIndicator(`"`, true, false, false)
	start, end := 0, 0
	for end <= len(text) {
		var ch rune
		hasCh := end < len(text)
		if hasCh {
			ch = text[end]
		}
		if !hasCh || ch == '"' || ch == '\\' || ch == 0x85 || ch == 0x2028 ||
			ch == 0x2029 || ch == 0xFEFF || !isDoubleQuotedLiteral(ch, e.opt.allowUnicode) {
			if start < end {
				e.write(string(text[start:end]))
				start = end
			}
			if hasCh {
				e.write(escapeRune(ch))
				start = end + 1
			}
		}
		if 0 < end && end < len(text)-1 && (ch == ' ' || start >= end) &&
			e.column+(end-start) > pyBestWidth && split {
			// The escape branch above leaves start == end+1, so on the very next
			// iteration start is one PAST end. Python's text[start:end] yields
			// "" for start > end (emitter.py:959); slicing a Go slice that way
			// panics, and it was reachable from any double-quoted scalar long
			// enough to fold after an escape — 151 of 700 fuzzed documents, and
			// `init --company "<80 chars>ébb"` on the CLI. See
			// TestWriteDoubleQuotedFoldsAfterEscape.
			data := `\`
			if start < end {
				data = string(text[start:end]) + `\`
				start = end
			}
			e.write(data)
			e.writeIndent()
			e.whitespace, e.indention = false, false
			if text[start] == ' ' {
				e.write(`\`)
			}
		}
		end++
	}
	e.writeIndicator(`"`, false, false, false)
}

// isDoubleQuotedLiteral is write_double_quoted's pass-through test
// (emitter.py:1010-1016). It differs from isSpecialOK: a raw newline is escaped
// inside a double-quoted scalar, and U+0085 is handled by the explicit list
// above rather than here.
func isDoubleQuotedLiteral(ch rune, allowUnicode bool) bool {
	if ch >= 0x20 && ch <= 0x7E {
		return true
	}
	if !allowUnicode {
		return false
	}
	return (ch >= 0xA0 && ch <= 0xD7FF) || (ch >= 0xE000 && ch <= 0xFFFD) ||
		(ch >= 0x10000 && ch < 0x10FFFF)
}

var escapeReplacements = map[rune]string{
	0x00: "0", 0x07: "a", 0x08: "b", 0x09: "t", 0x0A: "n", 0x0B: "v",
	0x0C: "f", 0x0D: "r", 0x1B: "e", '"': `"`, '\\': `\`,
	0x85: "N", 0xA0: "_", 0x2028: "L", 0x2029: "P",
}

func escapeRune(ch rune) string {
	if rep, ok := escapeReplacements[ch]; ok {
		return `\` + rep
	}
	switch {
	case ch <= 0xFF:
		return fmt.Sprintf(`\x%02X`, ch)
	case ch <= 0xFFFF:
		return fmt.Sprintf(`\u%04X`, ch)
	default:
		return fmt.Sprintf(`\U%08X`, ch)
	}
}

// writeBreakRun reproduces the break-flushing block shared by write_plain and
// write_single_quoted: a run of breaks emits one line break per break, plus a
// leading one when the run starts with \n.
func (e *pyEmitter) writeBreakRun(run []rune) {
	if len(run) == 0 {
		return
	}
	if run[0] == '\n' {
		e.writeLineBreak()
	}
	for range run {
		e.writeLineBreak()
	}
}

// ------------------------------------------------------------- loading

// PyLoadFile is load_yaml (bin/company-os:56-62): the parsed document, or nil
// when the file does not exist. The `or default` half is the caller's, because
// Python applies it as truthiness rather than as a nil check — see PyFalsy
// (R-1.7a).
func PyLoadFile(path string) (PyValue, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot read %s: %v", path, err)
	}
	doc, err := Load(raw)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
	}
	root := doc.Root()
	if root == nil {
		return nil, nil
	}
	return fromNode(root, path)
}

// PyLoadBytes is yaml.safe_load over an in-memory document — what
// frontmatter() (bin/company-os:76-82) does to the text between the fences,
// where there is no file to reopen. name only labels diagnostics.
func PyLoadBytes(raw []byte, name string) (PyValue, error) {
	doc, err := Load(raw)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", name, err)
	}
	root := doc.Root()
	if root == nil {
		return nil, nil
	}
	return fromNode(root, name)
}

// PyEqual is Python's `==` over two loaded objects: structural for containers,
// value-wise for scalars, and — like `==` and unlike a Go type switch — true
// across the numeric types, so `5 == 5.0`.
//
// A mapping compares by CONTENT, not by order, which is what makes it the
// right test for `meta.get("tags") == tags` and for a pointer's `p not in seen`.
func PyEqual(a, b PyValue) bool {
	if PyIsNone(a) || PyIsNone(b) {
		return PyIsNone(a) && PyIsNone(b)
	}
	switch x := a.(type) {
	case PySeq:
		y, ok := b.(PySeq)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !PyEqual(x[i], y[i]) {
				return false
			}
		}
		return true
	case PyMap:
		y, ok := b.(PyMap)
		if !ok || len(x) != len(y) {
			return false
		}
		for _, p := range x {
			if !pyHasKey(y, p.K) || !PyEqual(p.V, y.Get(p.K)) {
				return false
			}
		}
		return true
	case PyStr:
		y, ok := b.(PyStr)
		return ok && x == y
	case PyBool:
		// bool is a subclass of int in Python, so True == 1. Nothing in this
		// corpus relies on it and reproducing it would make `False == 0` true
		// as well; the numeric cases below stay closed over int and float.
		y, ok := b.(PyBool)
		return ok && x == y
	case PyTime:
		y, ok := b.(PyTime)
		return ok && x == y
	case PyInt:
		switch y := b.(type) {
		case PyInt:
			return x.N != nil && y.N != nil && x.N.Cmp(y.N) == 0
		case PyFloat:
			f, _ := new(big.Float).SetInt(x.N).Float64()
			return x.N != nil && f == float64(y)
		}
		return false
	case PyFloat:
		switch y := b.(type) {
		case PyFloat:
			return x == y
		case PyInt:
			f, _ := new(big.Float).SetInt(y.N).Float64()
			return y.N != nil && float64(x) == f
		}
		return false
	}
	return false
}

// PyIsNone is `v is None`, distinct from PyFalsy: an empty list is falsy but is
// not None, and the two questions are asked at different sites.
func PyIsNone(v PyValue) bool {
	if v == nil {
		return true
	}
	_, ok := v.(PyNull)
	return ok
}

func pyHasKey(m PyMap, key string) bool {
	for _, p := range m {
		if p.K == key {
			return true
		}
	}
	return false
}

// fromNode converts a parsed node tree into the Python objects safe_load builds,
// resolving every scalar so YAML 1.1 rules apply.
func fromNode(n *yaml.Node, path string) (PyValue, error) {
	n = Deref(n)
	switch n.Kind {
	case yaml.MappingNode:
		out := PyMap{}
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, err := fromNode(n.Content[i], path)
			if err != nil {
				return nil, err
			}
			key, ok := k.(PyStr)
			if !ok {
				return nil, model.Errorf(model.ExitArtifact,
					"%s: only string mapping keys are supported", path)
			}
			v, err := fromNode(n.Content[i+1], path)
			if err != nil {
				return nil, err
			}
			out = out.Set(string(key), v)
		}
		return out, nil
	case yaml.SequenceNode:
		out := make(PySeq, 0, len(n.Content))
		for _, c := range n.Content {
			v, err := fromNode(c, path)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case yaml.ScalarNode:
		s, err := Resolve(n)
		if err != nil {
			return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
		}
		return fromScalar(s, path)
	}
	return nil, model.Errorf(model.ExitArtifact, "%s: unsupported YAML node", path)
}

func fromScalar(s Scalar, path string) (PyValue, error) {
	switch s.Kind {
	case KindStr:
		return PyStr(s.Raw), nil
	case KindNull:
		return PyNull{}, nil
	case KindBool:
		return PyBool(s.Bool), nil
	case KindInt:
		return PyInt{N: s.Int}, nil
	case KindFloat:
		return PyFloat(s.Float), nil
	case KindTimestamp:
		return PyTime(isoformat(s)), nil
	}
	return nil, model.Errorf(model.ExitArtifact,
		"%s: cannot re-emit a %s scalar", path, s.Kind.Tag())
}

// isoformat renders a timestamp the way represent_date and represent_datetime
// do — date.isoformat(), or datetime.isoformat(' ') with a space separator.
func isoformat(s Scalar) string {
	if s.DateOnly {
		return s.Time.Format("2006-01-02")
	}
	text := s.Time.Format("2006-01-02 15:04:05")
	if s.Time.Nanosecond() != 0 {
		text += "." + fmt.Sprintf("%06d", s.Time.Nanosecond()/1000)
	}
	if s.HasZone {
		text += s.Time.Format("-07:00")
	}
	return text
}

// PyWriteFile is dump_yaml (bin/company-os:65-69): create the parent directory,
// then overwrite the file with safe_dump's output.
func PyWriteFile(path string, v PyValue) error {
	text, err := PyDump(v)
	if err != nil {
		return model.Errorf(model.ExitArtifact, "cannot serialize %s: %v", path, err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot create %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(text), 0o666); err != nil {
		return model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
	}
	return nil
}
