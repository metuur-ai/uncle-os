package scaffold

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
// indent — and `internal/yamlio`'s round-trip measurement (task 1.3) already
// froze those as unfixable through yaml.v3's knobs.
//
// `register_id` (:1815) is what makes a real emitter unavoidable: it loads
// company-ontology/ids/registry.yaml, appends one entry, and re-dumps the WHOLE
// file. Measured on examples/workspace, that rewrites seven flow-style entries
// into block style and `tags: [ontology/registry]` into an indentless block
// sequence — arbitrary authored content flowing back out through safe_dump.
//
// The writer primitives (writeIndent, writeIndicator, writePlain,
// writeSingleQuoted, writeDoubleQuoted) and analyzeScalar are transliterated
// from vendor/yaml/emitter.py. The 80-column fold in writePlain is the reason
// they are transliterated rather than approximated: PyYAML only breaks at a
// space whose column ALREADY exceeds best_width, so `precedence: canonical-
// mandatory > … > canonical-guidance` in a scaffolded team.yaml stays on one
// 85-column line. A "wrap at 80" reimplementation gets that file wrong.
//
// Scope, deliberate: this covers block style, non-canonical output, default
// flow style off, and the scalar types PyYAML's SafeRepresenter emits for str,
// bool, int, float, None, date and datetime. It does NOT implement literal or
// folded block scalars — nothing in the represented data selects them, since
// represent_str always emits style=None. It belongs in internal/yamlio next to
// the loader once internal/governance (deviation declare) and
// internal/federation (the lock) need it; it lives here while that package is
// under concurrent edit.

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
	"gopkg.in/yaml.v3"
)

// pyValue is one Python object as PyYAML's SafeRepresenter would see it. The
// closed set below is exactly what safe_load can produce for the artifacts this
// package reads and what the scaffolds construct.
type pyValue interface {
	pyRepr() (text string, isStr bool, err error)
}

type (
	// pyStr is a Python str: the only kind whose plain form may need quoting
	// to survive a round trip.
	pyStr string
	// pyBool is a Python bool.
	pyBool bool
	// pyInt is a Python int. Arbitrary precision, as PyYAML's ints are.
	pyInt struct{ N *big.Int }
	// pyFloat is a Python float.
	pyFloat float64
	// pyNull is Python None.
	pyNull struct{}
	// pyTime is a datetime.date or datetime.datetime, pre-rendered by
	// isoformat exactly as represent_date/represent_datetime render it.
	pyTime string
	// pySeq is a Python list.
	pySeq []pyValue
	// pyPair is one ordered mapping entry.
	pyPair struct {
		K string
		V pyValue
	}
	// pyMap is a Python dict. Insertion order is preserved because safe_dump
	// runs with sort_keys=False and the authored order is what the oracle
	// emits.
	pyMap []pyPair
)

func (s pyStr) pyRepr() (string, bool, error) { return string(s), true, nil }

func (b pyBool) pyRepr() (string, bool, error) {
	if bool(b) {
		return "true", false, nil
	}
	return "false", false, nil
}

func (i pyInt) pyRepr() (string, bool, error) { return i.N.String(), false, nil }

// pyRepr is represent_float (vendor/yaml/representer.py): repr() lowercased,
// with an exponent-only form given an explicit ".0" mantissa separator.
func (f pyFloat) pyRepr() (string, bool, error) {
	v := float64(f)
	switch {
	case math.IsNaN(v):
		return ".nan", false, nil
	case math.IsInf(v, 1):
		return ".inf", false, nil
	case math.IsInf(v, -1):
		return "-.inf", false, nil
	}
	text := strings.ToLower(strconv.FormatFloat(v, 'g', -1, 64))
	if !strings.Contains(text, ".") && strings.Contains(text, "e") {
		text = strings.Replace(text, "e", ".0e", 1)
	}
	return text, false, nil
}

func (pyNull) pyRepr() (string, bool, error)   { return "null", false, nil }
func (t pyTime) pyRepr() (string, bool, error) { return string(t), false, nil }

func (pySeq) pyRepr() (string, bool, error) {
	return "", false, fmt.Errorf("scaffold: pyRepr called on a sequence")
}

func (pyMap) pyRepr() (string, bool, error) {
	return "", false, fmt.Errorf("scaffold: pyRepr called on a mapping")
}

// get returns the value at key, or nil.
func (m pyMap) get(key string) pyValue {
	for _, p := range m {
		if p.K == key {
			return p.V
		}
	}
	return nil
}

// set replaces the value at key in place, or appends it, which is what a Python
// dict assignment does to an insertion-ordered mapping.
func (m pyMap) set(key string, v pyValue) pyMap {
	for i := range m {
		if m[i].K == key {
			m[i].V = v
			return m
		}
	}
	return append(m, pyPair{K: key, V: v})
}

// ---------------------------------------------------------------- emitting

const (
	pyBestIndent = 2  // safe_dump's indent default
	pyBestWidth  = 80 // safe_dump's width default
)

// pyDump is yaml.safe_dump(data, sort_keys=False, default_flow_style=False).
func pyDump(v pyValue) (string, error) {
	e := &pyEmitter{indent: -1, whitespace: true, indention: true}
	if err := e.node(v, false, false); err != nil {
		return "", err
	}
	// expect_document_end's write_indent, which is what terminates the final
	// line. It is unconditional in PyYAML and produces the trailing newline.
	e.indent = -1
	e.writeIndent()
	return e.b.String(), nil
}

type pyEmitter struct {
	b          strings.Builder
	column     int
	whitespace bool
	indention  bool
	// indent is Python's self.indent, where -1 stands for None.
	indent int
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
func (e *pyEmitter) node(v pyValue, mappingCtx, simpleKey bool) error {
	switch t := v.(type) {
	case pySeq:
		if len(t) == 0 {
			// check_empty_sequence routes an empty list to flow style.
			e.writeIndicator("[", true, true, false)
			e.writeIndicator("]", false, false, false)
			return nil
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

	case pyMap:
		if len(t) == 0 {
			e.writeIndicator("{", true, true, false)
			e.writeIndicator("}", false, false, false)
			return nil
		}
		saved := e.increaseIndent(false, false)
		for _, pair := range t {
			e.writeIndent()
			if err := e.node(pyStr(pair.K), true, true); err != nil {
				return err
			}
			e.writeIndicator(":", false, false, false)
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

func (e *pyEmitter) scalar(v pyValue, simpleKey bool) error {
	text, isStr, err := v.pyRepr()
	if err != nil {
		return err
	}
	runes := []rune(text)
	a := analyzeScalar(runes)

	// serialize_node computes implicit[0] as "the plain form resolves back to
	// this node's tag". Only a str can fail it: every other representer emits
	// text its own resolver reclaims.
	plainRoundTrips := true
	if isStr {
		plainRoundTrips = resolvesToStr(text)
	}

	split := !simpleKey
	switch {
	case plainRoundTrips && !(simpleKey && (a.empty || a.multiline)) && a.allowBlockPlain:
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
	s, err := yamlio.ResolvePlain(text)
	return err == nil && s.Kind == yamlio.KindStr
}

// ------------------------------------------------------- analyze_scalar

type scalarAnalysis struct {
	empty             bool
	multiline         bool
	allowBlockPlain   bool
	allowSingleQuoted bool
}

// analyzeScalar is Emitter.analyze_scalar (vendor/yaml/emitter.py:626),
// restricted to the flags the block-style writer consults. allow_flow_plain and
// allow_block are dropped because nothing here emits flow or block scalars, and
// allow_double_quoted is unconditionally true in PyYAML.
//
// allow_unicode is False under safe_dump, so any non-ASCII character is a
// "special character" and forces double quotes — which is why a company name
// with an accent round-trips as an escaped \uXXXX rather than as UTF-8.
func analyzeScalar(scalar []rune) scalarAnalysis {
	if len(scalar) == 0 {
		return scalarAnalysis{empty: true, allowBlockPlain: true, allowSingleQuoted: true}
	}

	var (
		blockIndicators   bool
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
				blockIndicators = true
			}
			if (ch == '?' || ch == ':') && followedByWhitespace {
				blockIndicators = true
			}
			if ch == '-' && followedByWhitespace {
				blockIndicators = true
			}
		} else {
			if ch == ':' && followedByWhitespace {
				blockIndicators = true
			}
			if ch == '#' && precededByWhitespace {
				blockIndicators = true
			}
		}

		if isBreak(ch) {
			lineBreaks = true
		}
		if !(ch == '\n' || (ch >= 0x20 && ch <= 0x7E)) {
			// allow_unicode is False, so every non-printable-ASCII character
			// lands in the special bucket.
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

	a := scalarAnalysis{multiline: lineBreaks, allowBlockPlain: true, allowSingleQuoted: true}
	if leadingSpace || leadingBreak || trailingSpace || trailingBreak {
		a.allowBlockPlain = false
	}
	if breakSpace {
		a.allowBlockPlain, a.allowSingleQuoted = false, false
	}
	if spaceBreak || specialCharacters {
		a.allowBlockPlain, a.allowSingleQuoted = false, false
	}
	if lineBreaks {
		a.allowBlockPlain = false
	}
	if blockIndicators {
		a.allowBlockPlain = false
	}
	return a
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

// writeDoubleQuoted is Emitter.write_double_quoted, with allow_unicode False so
// every non-printable-ASCII rune is escaped.
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
			ch == 0x2029 || ch == 0xFEFF || !(ch >= 0x20 && ch <= 0x7E) {
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
			data := string(text[start:end]) + `\`
			if start < end {
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

// loadPy is load_yaml (bin/company-os:56-62): the parsed document, or nil when
// the file does not exist. The `or default` half is the caller's, because
// Python applies it as truthiness rather than as a nil check (R-1.7a).
func loadPy(path string) (pyValue, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "cannot read %s: %v", path, err)
	}
	doc, err := yamlio.Load(raw)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
	}
	root := doc.Root()
	if root == nil {
		return nil, nil
	}
	return fromNode(root, path)
}

// fromNode converts a parsed node tree into the Python objects safe_load builds,
// resolving every scalar through internal/yamlio so YAML 1.1 rules apply.
func fromNode(n *yaml.Node, path string) (pyValue, error) {
	n = yamlio.Deref(n)
	switch n.Kind {
	case yaml.MappingNode:
		out := pyMap{}
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, err := fromNode(n.Content[i], path)
			if err != nil {
				return nil, err
			}
			key, ok := k.(pyStr)
			if !ok {
				return nil, model.Errorf(model.ExitArtifact,
					"%s: only string mapping keys are supported", path)
			}
			v, err := fromNode(n.Content[i+1], path)
			if err != nil {
				return nil, err
			}
			out = out.set(string(key), v)
		}
		return out, nil
	case yaml.SequenceNode:
		out := make(pySeq, 0, len(n.Content))
		for _, c := range n.Content {
			v, err := fromNode(c, path)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	case yaml.ScalarNode:
		s, err := yamlio.Resolve(n)
		if err != nil {
			return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
		}
		return fromScalar(s, path)
	}
	return nil, model.Errorf(model.ExitArtifact, "%s: unsupported YAML node", path)
}

func fromScalar(s yamlio.Scalar, path string) (pyValue, error) {
	switch s.Kind {
	case yamlio.KindStr:
		return pyStr(s.Raw), nil
	case yamlio.KindNull:
		return pyNull{}, nil
	case yamlio.KindBool:
		return pyBool(s.Bool), nil
	case yamlio.KindInt:
		return pyInt{N: s.Int}, nil
	case yamlio.KindFloat:
		return pyFloat(s.Float), nil
	case yamlio.KindTimestamp:
		return pyTime(isoformat(s)), nil
	}
	return nil, model.Errorf(model.ExitArtifact,
		"%s: cannot re-emit a %s scalar", path, s.Kind.Tag())
}

// isoformat renders a timestamp the way represent_date and represent_datetime
// do — date.isoformat(), or datetime.isoformat(' ') with a space separator.
func isoformat(s yamlio.Scalar) string {
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

// writePyYAML is dump_yaml (bin/company-os:65-69): create the parent directory,
// then overwrite the file with safe_dump's output.
func writePyYAML(path string, v pyValue) error {
	text, err := pyDump(v)
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
