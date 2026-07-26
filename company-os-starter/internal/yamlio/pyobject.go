package yamlio

// Python built-ins applied to a loaded document, for the read views that
// interpolate or measure whatever an artifact happens to hold.
//
// The Python CLI reads its own artifacts with no schema in between: cmd_today
// f-strings `m.get('due')` and calls `len(v)` on `requirements['platform']`'s
// values. Both are total over the loaded object graph in a way a Go type switch
// is not — `len` counts a str's characters and a dict's keys, and an f-string
// renders a dict through repr(). Approximating either produced a silently wrong
// NUMBER on a path that exits 0, which is worse than a refusal.
//
// Scalar.String() already answers str() for scalars; these two answer it for
// the collections, and answer len().

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// PyLen is len(value) over a loaded document. ok is false for the objects
// Python has no length for — int, float, bool, None and datetime — where
// `len(v)` raises TypeError.
//
// A str's length is its CHARACTER count, which is why the rune count and not
// len(node.Value) is the answer.
func PyLen(n *yaml.Node) (int, bool) {
	n = Deref(n)
	if n == nil {
		return 0, false
	}
	switch n.Kind {
	case yaml.SequenceNode:
		return len(n.Content), true
	case yaml.MappingNode:
		return len(n.Content) / 2, true
	case yaml.ScalarNode:
		s, err := Resolve(n)
		if err != nil || s.Kind != KindStr {
			return 0, false
		}
		return utf8.RuneCountInString(s.Raw), true
	}
	return 0, false
}

// PyText is str(value) over a loaded document: Scalar.String() for a scalar, and
// Python's repr for a list or a dict, which is what an f-string interpolates for
// them.
func PyText(n *yaml.Node) string {
	n = Deref(n)
	if n == nil {
		return "None"
	}
	if n.Kind == yaml.ScalarNode {
		s, err := Resolve(n)
		if err != nil {
			return n.Value
		}
		return s.String()
	}
	return pyRepr(n)
}

// pyRepr is repr(value). str() and repr() differ only for a str — a container's
// str() is its repr(), and its ELEMENTS are always rendered with repr(), which
// is why `[x, y]` shows quotes that `str(x)` alone would not.
func pyRepr(n *yaml.Node) string {
	n = Deref(n)
	if n == nil {
		return "None"
	}
	switch n.Kind {
	case yaml.SequenceNode:
		parts := make([]string, 0, len(n.Content))
		for _, c := range n.Content {
			parts = append(parts, pyRepr(c))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case yaml.MappingNode:
		parts := make([]string, 0, len(n.Content)/2)
		for i := 0; i+1 < len(n.Content); i += 2 {
			parts = append(parts, pyRepr(n.Content[i])+": "+pyRepr(n.Content[i+1]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case yaml.ScalarNode:
		s, err := Resolve(n)
		if err != nil {
			return pyStrRepr(n.Value)
		}
		switch s.Kind {
		case KindStr:
			return pyStrRepr(s.Raw)
		case KindTimestamp:
			return s.pyTimestampRepr()
		default:
			// None, True/False, ints and floats repr() exactly as they str().
			return s.String()
		}
	}
	return pyStrRepr(n.Value)
}

// pyStrRepr is repr(str): single quotes unless the value contains one and no
// double quote, backslash and the chosen quote escaped, and the C0 controls
// escaped. Non-ASCII printable characters are NOT escaped — Python 3's repr
// emits them as themselves.
func pyStrRepr(s string) string {
	quote := byte('\'')
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		quote = '"'
	}
	var b strings.Builder
	b.WriteByte(quote)
	for _, r := range s {
		switch {
		case r == rune(quote) || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7F:
			fmt.Fprintf(&b, `\x%02x`, r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// pyTimestampRepr is repr(datetime.date) / repr(datetime.datetime), which name
// the constructor rather than rendering the ISO form str() gives. Trailing
// zero arguments are omitted, so midnight is `(2026, 1, 1, 0, 0)` and a
// microsecond forces the seconds argument back in.
func (s Scalar) pyTimestampRepr() string {
	t := s.Time
	if s.DateOnly {
		return fmt.Sprintf("datetime.date(%d, %d, %d)", t.Year(), int(t.Month()), t.Day())
	}
	out := fmt.Sprintf("datetime.datetime(%d, %d, %d, %d, %d",
		t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute())
	micro := t.Nanosecond() / 1000
	if t.Second() != 0 || micro != 0 {
		out += fmt.Sprintf(", %d", t.Second())
	}
	if micro != 0 {
		out += fmt.Sprintf(", %d", micro)
	}
	if s.HasZone {
		out += ", tzinfo=" + pyTimezoneRepr(t)
	}
	return out + ")"
}

// pyTimezoneRepr is repr(datetime.timezone). timedelta normalises to a
// non-negative seconds field, so a west-of-Greenwich offset reprs as
// `days=-1, seconds=<86400+offset>` rather than as a negative second count.
func pyTimezoneRepr(t time.Time) string {
	_, offset := t.Zone()
	if offset == 0 {
		return "datetime.timezone.utc"
	}
	if offset > 0 {
		return fmt.Sprintf("datetime.timezone(datetime.timedelta(seconds=%d))", offset)
	}
	return fmt.Sprintf("datetime.timezone(datetime.timedelta(days=-1, seconds=%d))",
		86400+offset)
}
