package yamlio

// PyYAML implements YAML 1.1; gopkg.in/yaml.v3 implements YAML 1.2. They
// resolve untagged (plain) scalars differently, and the two libraries also
// render the resulting values differently even where they agree on the type.
// Both differences change behaviour, not just bytes, so every scalar the port
// reads goes through Resolve here instead of through yaml.v3's own resolution.
//
// The rules below are transliterated from the vendored PyYAML 6.0.2 at
// company-os-starter/vendor/yaml: the implicit-resolver table in resolver.py
// and the construct_yaml_* functions in constructor.py. Measured divergences
// are recorded in internal/yamlio/testdata/pyyaml-truth.json, which is the
// table the tests run against.

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Kind classifies a scalar the way PyYAML's implicit resolver classifies it.
type Kind int

const (
	KindStr Kind = iota
	KindNull
	KindBool
	KindInt
	KindFloat
	KindTimestamp
	// KindMerge and KindValue are the "<<" and "=" tags. PyYAML's SafeLoader
	// resolves them but has no constructor for either, so anywhere but a
	// mapping KEY they raise ConstructorError. Resolution still succeeds here,
	// as it does in PyYAML; construct is what refuses them, and it is the only
	// caller that has the position information needed to.
	KindMerge
	KindValue
)

// Tag returns the YAML tag PyYAML assigns to this kind, in "!!name" short form.
func (k Kind) Tag() string {
	switch k {
	case KindNull:
		return "!!null"
	case KindBool:
		return "!!bool"
	case KindInt:
		return "!!int"
	case KindFloat:
		return "!!float"
	case KindTimestamp:
		return "!!timestamp"
	case KindMerge:
		return "!!merge"
	case KindValue:
		return "!!value"
	default:
		return "!!str"
	}
}

// Scalar is a YAML scalar resolved the way PyYAML's safe_load resolves it.
//
// Only the field matching Kind carries meaning; the rest are zero. String
// renders the scalar exactly as Python's str() renders the corresponding
// value, which is what the Python CLI's output is built from.
type Scalar struct {
	Kind Kind

	// Raw is the scalar text exactly as authored, after quote and escape
	// processing. For KindStr it is the value.
	Raw string

	Bool  bool     // KindBool
	Int   *big.Int // KindInt — Python ints are arbitrary precision
	Float float64  // KindFloat
	Time  time.Time

	// DateOnly reports that a KindTimestamp had no time component, so PyYAML
	// built a datetime.date rather than a datetime.datetime.
	DateOnly bool
	// HasZone reports that a KindTimestamp carried an explicit Z or ±HH:MM.
	// Python distinguishes aware from naive datetimes when rendering.
	HasZone bool
}

// The implicit-resolver regexps, transliterated from PyYAML's resolver.py.
//
// PyYAML indexes resolvers by first character and tries them in registration
// order. Every regexp below only matches strings starting with a character in
// its own declared first-character set, so trying them in registration order
// without the index gives identical results.
var (
	reBool = regexp.MustCompile(`^(?:yes|Yes|YES|no|No|NO` +
		`|true|True|TRUE|false|False|FALSE` +
		`|on|On|ON|off|Off|OFF)$`)

	reFloat = regexp.MustCompile(`^(?:[-+]?(?:[0-9][0-9_]*)\.[0-9_]*(?:[eE][-+][0-9]+)?` +
		`|\.[0-9][0-9_]*(?:[eE][-+][0-9]+)?` +
		`|[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\.[0-9_]*` +
		`|[-+]?\.(?:inf|Inf|INF)` +
		`|\.(?:nan|NaN|NAN))$`)

	reInt = regexp.MustCompile(`^(?:[-+]?0b[0-1_]+` +
		`|[-+]?0[0-7_]+` +
		`|[-+]?(?:0|[1-9][0-9_]*)` +
		`|[-+]?0x[0-9a-fA-F_]+` +
		`|[-+]?[1-9][0-9_]*(?::[0-5]?[0-9])+)$`)

	reMerge = regexp.MustCompile(`^(?:<<)$`)

	reNull = regexp.MustCompile(`^(?:~|null|Null|NULL|)$`)

	reTimestampTag = regexp.MustCompile(`^(?:[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]` +
		`|[0-9][0-9][0-9][0-9]-[0-9][0-9]?-[0-9][0-9]?` +
		`(?:[Tt]|[ \t]+)[0-9][0-9]?` +
		`:[0-9][0-9]:[0-9][0-9](?:\.[0-9]*)?` +
		`(?:[ \t]*(?:Z|[-+][0-9][0-9]?(?::[0-9][0-9])?))?)$`)

	reValue = regexp.MustCompile(`^(?:=)$`)

	// The constructor's own, looser regexp — construct_yaml_timestamp
	// re-matches with this one to pull the fields out.
	reTimestampParts = regexp.MustCompile(`^(?P<year>[0-9][0-9][0-9][0-9])` +
		`-(?P<month>[0-9][0-9]?)` +
		`-(?P<day>[0-9][0-9]?)` +
		`(?:(?:[Tt]|[ \t]+)` +
		`(?P<hour>[0-9][0-9]?)` +
		`:(?P<minute>[0-9][0-9])` +
		`:(?P<second>[0-9][0-9])` +
		`(?:\.(?P<fraction>[0-9]*))?` +
		`(?:[ \t]*(?P<tz>Z|(?P<tz_sign>[-+])(?P<tz_hour>[0-9][0-9]?)` +
		`(?::(?P<tz_minute>[0-9][0-9]))?))?)?$`)
)

// quoted marks the styles that suppress implicit resolution. An explicit tag
// suppresses it too, but is handled apart from these because it also selects a
// constructor — see resolveTagged.
const quoted = yaml.DoubleQuotedStyle | yaml.SingleQuotedStyle |
	yaml.LiteralStyle | yaml.FoldedStyle

// Resolve resolves a scalar node the way PyYAML's safe_load resolves it.
//
// Quoted scalars keep their string value; only plain scalars are resolved
// implicitly. That is what lets an unquoted `reviewDate: 2035-01-15` and a
// quoted `reviewDate: '2035-01-14'` flow through this one path and come out
// rendering the same way. An explicitly tagged scalar is neither: PyYAML
// suppresses only the implicit RESOLVER for it and still runs the constructor
// its tag names, so it goes to resolveTagged.
//
// An alias is followed to its anchor first, because PyYAML's composer expands
// aliases before construction ever sees them — measured: safe_load("a: &x 1\nb:
// *x") is {'a': 1, 'b': 1}.
//
// It returns an error for a scalar PyYAML resolves but then fails to
// construct, which today means an out-of-range calendar date or a tagged
// scalar whose text the tag's constructor rejects.
func Resolve(n *yaml.Node) (Scalar, error) {
	n = Deref(n)
	if n == nil || n.Kind != yaml.ScalarNode {
		return Scalar{}, fmt.Errorf("yamlio: Resolve wants a scalar node, got kind %d", kindOf(n))
	}
	if n.Style&yaml.TaggedStyle != 0 {
		return resolveTagged(n)
	}
	if n.Style&quoted != 0 {
		return Scalar{Kind: KindStr, Raw: n.Value}, nil
	}
	return ResolvePlain(n.Value)
}

// Deref follows an alias node to the node its anchor names, which is what
// PyYAML's composer does before construction runs.
//
// The chain cannot cycle: yaml.v3 resolves an alias only against an anchor it
// has already composed, so every link points strictly backwards in the
// document.
func Deref(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

// resolveTagged constructs a scalar carrying an explicit tag.
//
// PyYAML runs resolution and construction as two separate passes, and an
// explicit tag suppresses only the first: construction then dispatches on the
// tag. Measured against the vendored PyYAML 6.0.2, `a: !!int 0` is `{'a': 0}` —
// an int, and so falsy, which `or {}` collapses — and `a: !!int '0'` is `{'a':
// 0}` too, so the tag also outranks quoting. Returning a string for either got
// the kind, the rendering and the truthiness wrong.
//
// A tag with no construct* counterpart here — !!binary, !!set, a local
// !custom — keeps the string behaviour this function replaced rather than
// gaining a new one. PyYAML constructs the first two and raises on the third;
// closing that is a separate change, and no committed artifact carries one.
func resolveTagged(n *yaml.Node) (Scalar, error) {
	switch n.Tag {
	case "!!null":
		// construct_yaml_null ignores the text entirely.
		return Scalar{Kind: KindNull, Raw: n.Value}, nil
	case "!!bool":
		b, ok := constructBool(n.Value)
		if !ok {
			// PyYAML's bool_values lookup raises KeyError here.
			return Scalar{}, fmt.Errorf("yamlio: %q is not a !!bool", n.Value)
		}
		return Scalar{Kind: KindBool, Raw: n.Value, Bool: b}, nil
	case "!!int":
		i, err := constructInt(n.Value)
		if err != nil {
			return Scalar{}, err
		}
		return Scalar{Kind: KindInt, Raw: n.Value, Int: i}, nil
	case "!!float":
		f, err := constructFloat(n.Value)
		if err != nil {
			return Scalar{}, err
		}
		return Scalar{Kind: KindFloat, Raw: n.Value, Float: f}, nil
	case "!!timestamp":
		return constructTimestamp(n.Value)
	case "!!merge":
		return Scalar{Kind: KindMerge, Raw: n.Value}, nil
	case "!!value":
		return Scalar{Kind: KindValue, Raw: n.Value}, nil
	default:
		return Scalar{Kind: KindStr, Raw: n.Value}, nil
	}
}

func kindOf(n *yaml.Node) yaml.Kind {
	if n == nil {
		return 0
	}
	return n.Kind
}

// ResolvePlain resolves the text of a plain (unquoted, untagged) scalar.
//
// Each regexp has already validated the syntax its constructor parses, so the
// construct* errors below cannot fire on this path.
func ResolvePlain(text string) (Scalar, error) {
	switch {
	case reBool.MatchString(text):
		b, _ := constructBool(text)
		return Scalar{Kind: KindBool, Raw: text, Bool: b}, nil
	case reFloat.MatchString(text):
		f, err := constructFloat(text)
		if err != nil {
			return Scalar{}, err
		}
		return Scalar{Kind: KindFloat, Raw: text, Float: f}, nil
	case reInt.MatchString(text):
		i, err := constructInt(text)
		if err != nil {
			return Scalar{}, err
		}
		return Scalar{Kind: KindInt, Raw: text, Int: i}, nil
	case reMerge.MatchString(text):
		return Scalar{Kind: KindMerge, Raw: text}, nil
	case reNull.MatchString(text):
		return Scalar{Kind: KindNull, Raw: text}, nil
	case reTimestampTag.MatchString(text):
		return constructTimestamp(text)
	case reValue.MatchString(text):
		return Scalar{Kind: KindValue, Raw: text}, nil
	default:
		return Scalar{Kind: KindStr, Raw: text}, nil
	}
}

// boolValues is SafeConstructor.bool_values, verbatim.
var boolValues = map[string]bool{
	"yes": true, "no": false,
	"true": true, "false": false,
	"on": true, "off": false,
}

// constructBool mirrors construct_yaml_bool, a bool_values lookup keyed on
// value.lower(). The second result is that dict lookup's success; PyYAML raises
// KeyError when it fails, which only an explicitly tagged scalar can reach.
func constructBool(text string) (value, ok bool) {
	value, ok = boolValues[strings.ToLower(text)]
	return value, ok
}

// splitSign strips PyYAML's underscores and leading sign, as
// construct_yaml_int and construct_yaml_float both do before parsing.
func splitSign(text string) (neg bool, rest string) {
	rest = strings.ReplaceAll(text, "_", "")
	if rest == "" {
		return false, rest
	}
	switch rest[0] {
	case '-':
		return true, rest[1:]
	case '+':
		return false, rest[1:]
	}
	return false, rest
}

// constructInt mirrors construct_yaml_int.
//
// On the plain path reInt has already validated the syntax and the error cannot
// fire; on the tagged path the text is arbitrary, and PyYAML raises ValueError
// (or IndexError, on empty text) exactly where this returns one.
func constructInt(text string) (*big.Int, error) {
	neg, v := splitSign(text)
	if v == "" {
		return nil, fmt.Errorf("yamlio: %q is not a !!int", text)
	}
	out := new(big.Int)

	var ok bool
	switch {
	case v == "0":
		return out, nil // Python returns a bare 0 here, sign and all.
	case strings.HasPrefix(v, "0b"):
		_, ok = out.SetString(v[2:], 2)
	case strings.HasPrefix(v, "0x"):
		_, ok = out.SetString(v[2:], 16)
	case v[0] == '0':
		// Python's int(v, 8) accepts an explicit 0o prefix as well as a bare
		// leading zero; Go's SetString accepts neither prefix at a fixed base.
		digits := v
		if len(digits) > 1 && (digits[1] == 'o' || digits[1] == 'O') {
			digits = digits[2:]
		}
		_, ok = out.SetString(digits, 8)
	case strings.Contains(v, ":"):
		base := big.NewInt(1)
		sixty := big.NewInt(60)
		parts := strings.Split(v, ":")
		ok = true
		for i := len(parts) - 1; i >= 0 && ok; i-- {
			var digit *big.Int
			if digit, ok = new(big.Int).SetString(parts[i], 10); ok {
				out.Add(out, new(big.Int).Mul(digit, base))
				base = new(big.Int).Mul(base, sixty)
			}
		}
	default:
		_, ok = out.SetString(v, 10)
	}
	if !ok {
		return nil, fmt.Errorf("yamlio: %q is not a !!int", text)
	}
	if neg {
		out.Neg(out)
	}
	return out, nil
}

// constructFloat mirrors construct_yaml_float, which lowercases before
// dispatching, so ".Inf" and ".NaN" fall out for free. The error is reachable
// only from the tagged path, where PyYAML raises ValueError or IndexError.
func constructFloat(text string) (float64, error) {
	neg, v := splitSign(text)
	if v == "" {
		return 0, fmt.Errorf("yamlio: %q is not a !!float", text)
	}
	v = strings.ToLower(v)

	var f float64
	switch {
	case v == ".inf":
		f = math.Inf(1)
	case v == ".nan":
		return math.NaN(), nil // Python's nan has no meaningful sign here.
	case strings.Contains(v, ":"):
		base := 1.0
		parts := strings.Split(v, ":")
		for i := len(parts) - 1; i >= 0; i-- {
			digit, err := strconv.ParseFloat(parts[i], 64)
			if err != nil {
				return 0, fmt.Errorf("yamlio: %q is not a !!float", text)
			}
			f += digit * base
			base *= 60
		}
	default:
		var err error
		if f, err = strconv.ParseFloat(v, 64); err != nil {
			return 0, fmt.Errorf("yamlio: %q is not a !!float", text)
		}
	}
	if neg {
		f = -f
	}
	return f, nil
}

// constructTimestamp mirrors construct_yaml_timestamp, including the range
// checks Python's datetime constructors impose. yaml.v3 silently normalises
// 2035-02-30 into 2035-03-02; PyYAML raises, so this returns an error.
func constructTimestamp(text string) (Scalar, error) {
	m := reTimestampParts.FindStringSubmatch(text)
	if m == nil {
		// The tag regexp matched but the field regexp did not; PyYAML would
		// crash on the None match, so treat it as unconstructible.
		return Scalar{}, fmt.Errorf("yamlio: %q is not a constructible timestamp", text)
	}
	get := func(name string) string {
		return m[reTimestampParts.SubexpIndex(name)]
	}
	atoi := func(s string) int { n, _ := strconv.Atoi(s); return n }

	year, month, day := atoi(get("year")), atoi(get("month")), atoi(get("day"))
	if year < 1 {
		return Scalar{}, fmt.Errorf("yamlio: %q: year %d is out of range", text, year)
	}
	if month < 1 || month > 12 {
		return Scalar{}, fmt.Errorf("yamlio: %q: month must be in 1..12", text)
	}
	if day < 1 || day > daysInMonth(year, month) {
		return Scalar{}, fmt.Errorf("yamlio: %q: day is out of range for month", text)
	}

	if get("hour") == "" {
		return Scalar{
			Kind:     KindTimestamp,
			Raw:      text,
			Time:     time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC),
			DateOnly: true,
		}, nil
	}

	hour, minute, second := atoi(get("hour")), atoi(get("minute")), atoi(get("second"))
	if hour > 23 {
		return Scalar{}, fmt.Errorf("yamlio: %q: hour must be in 0..23", text)
	}
	if minute > 59 {
		return Scalar{}, fmt.Errorf("yamlio: %q: minute must be in 0..59", text)
	}
	if second > 59 {
		return Scalar{}, fmt.Errorf("yamlio: %q: second must be in 0..59", text)
	}

	// PyYAML truncates the fraction to 6 digits and right-pads to microseconds.
	micros := 0
	if frac := get("fraction"); frac != "" {
		if len(frac) > 6 {
			frac = frac[:6]
		}
		micros = atoi(frac + strings.Repeat("0", 6-len(frac)))
	}

	loc, hasZone := time.UTC, false
	if sign := get("tz_sign"); sign != "" {
		offset := atoi(get("tz_hour"))*3600 + atoi(get("tz_minute"))*60
		if sign == "-" {
			offset = -offset
		}
		if offset <= -24*3600 || offset >= 24*3600 {
			return Scalar{}, fmt.Errorf(
				"yamlio: %q: offset must be strictly between -24h and 24h", text)
		}
		loc, hasZone = time.FixedZone("", offset), true
	} else if get("tz") != "" {
		hasZone = true // a bare Z
	}

	return Scalar{
		Kind:    KindTimestamp,
		Raw:     text,
		Time:    time.Date(year, time.Month(month), day, hour, minute, second, micros*1000, loc),
		HasZone: hasZone,
	}, nil
}

func daysInMonth(year, month int) int {
	// Day 0 of the next month is the last day of this one.
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// String renders the scalar exactly as Python's str() renders the value
// PyYAML's safe_load produces for it.
func (s Scalar) String() string {
	switch s.Kind {
	case KindNull:
		return "None"
	case KindBool:
		if s.Bool {
			return "True"
		}
		return "False"
	case KindInt:
		if s.Int == nil {
			return "0"
		}
		return s.Int.String()
	case KindFloat:
		return pyFloat(s.Float)
	case KindTimestamp:
		return s.pyTimestamp()
	default:
		// KindStr, and the merge/value tags, which have no Python value.
		return s.Raw
	}
}

// pyTimestamp reproduces str(datetime.date) and str(datetime.datetime), which
// are isoformat() with a space separator, microseconds only when non-zero, and
// a ±HH:MM offset only on an aware datetime.
func (s Scalar) pyTimestamp() string {
	if s.DateOnly {
		return s.Time.Format("2006-01-02")
	}
	out := s.Time.Format("2006-01-02 15:04:05")
	if micro := s.Time.Nanosecond() / 1000; micro != 0 {
		out += fmt.Sprintf(".%06d", micro)
	}
	if s.HasZone {
		out += s.Time.Format("-07:00")
	}
	return out
}

// pyFloat reproduces Python 3's repr of a float: the shortest decimal that
// round-trips, switched to exponent form when the decimal point would land at
// or before -4 or after 16, and always carrying a ".0" in positional form so
// the value cannot be mistaken for an int. Go's %v differs on all three counts
// — it renders 3.0 as "3" and 1e15 as "1e+15".
func pyFloat(f float64) string {
	switch {
	case math.IsNaN(f):
		return "nan"
	case math.IsInf(f, 1):
		return "inf"
	case math.IsInf(f, -1):
		return "-inf"
	}

	// Shortest round-trip, normalised to d.dddde±dd so the digits and the
	// exponent can be read off directly.
	sci := strconv.FormatFloat(f, 'e', -1, 64)
	sign := ""
	if sci[0] == '-' {
		sign, sci = "-", sci[1:]
	}
	mantissa, expPart, _ := strings.Cut(sci, "e")
	exp, _ := strconv.Atoi(expPart)
	digits := strings.Replace(mantissa, ".", "", 1)

	// decpt is the decimal point's offset into digits: value = 0.digits * 10^decpt.
	decpt := exp + 1
	if decpt <= -4 || decpt > 16 {
		out := digits[:1]
		if len(digits) > 1 {
			out += "." + digits[1:]
		}
		return fmt.Sprintf("%s%se%s%02d", sign, out, expSign(exp), abs(exp))
	}

	switch {
	case decpt <= 0:
		return sign + "0." + strings.Repeat("0", -decpt) + digits
	case decpt >= len(digits):
		return sign + digits + strings.Repeat("0", decpt-len(digits)) + ".0"
	default:
		return sign + digits[:decpt] + "." + digits[decpt:]
	}
}

func expSign(e int) string {
	if e < 0 {
		return "-"
	}
	return "+"
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
