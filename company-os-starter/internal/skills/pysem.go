package skills

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// This file holds the three CPython string behaviours the skills cluster leans
// on. They are written out rather than approximated with the nearest Go stdlib
// call because each of them decides whether a line is a skill STEP, and a step
// that Python sees and Go does not is a silently short merged view.
//
// internal/frontmatter took the same position on universal-newline translation
// for the same reason.

// pyIsSpace reports whether CPython considers r whitespace.
//
// One predicate covers two call sites that must agree: str.strip() (which
// strips every c where c.isspace()) and the `\s` class of a `str` regex, which
// compiles to the same Py_UNICODE_ISSPACE test. Go has no single equivalent —
// regexp's `\s` is ASCII-only and misses \v, and unicode.IsSpace omits the
// 0x1C-0x1F separator block that CPython counts.
func pyIsSpace(r rune) bool {
	switch {
	case r >= 0x09 && r <= 0x0D: // \t \n \v \f \r
		return true
	case r >= 0x1C && r <= 0x1F: // file/group/record/unit separators
		return true
	case r == 0x85: // NEL
		return true
	}
	return unicode.Is(unicode.Zs, r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r)
}

// pyStrip is str.strip() with no argument.
func pyStrip(s string) string { return strings.TrimFunc(s, pyIsSpace) }

// pySplitLines is str.splitlines().
//
// The bodies reaching here have already been through internal/frontmatter's
// universal-newline translation, so "\r\n" and "\r" are gone; what remains and
// what strings.Split(s, "\n") would still miss is \v, \f, the 0x1C-0x1E
// separators, NEL, and the two Unicode line/paragraph separators. Unlike
// Split, splitlines() also yields nothing for the empty string and drops the
// trailing empty element after a final terminator — both of which change how
// many candidate step lines a body offers.
func pySplitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); {
		r, size := decodeRune(s[i:])
		width := 0
		switch r {
		case '\r':
			width = size
			if i+size < len(s) && s[i+size] == '\n' {
				width++
			}
		case '\n', '\v', '\f', 0x1C, 0x1D, 0x1E, 0x85, 0x2028, 0x2029:
			width = size
		}
		if width == 0 {
			i += size
			continue
		}
		out = append(out, s[start:i])
		i += width
		start = i
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// tiers are the three markers STEP_RE accepts (bin/company-os:772), in the
// alternation's own order.
var tiers = []string{"mandatory", "default", "guidance"}

// isStep is STEP_RE.match(line): `^\s*\d+\.\s*\((mandatory|default|guidance)\)`.
//
// Written as a scan rather than a Go regexp because two of its classes are
// Unicode-wide in Python and ASCII-only in Go: `\s` (see pyIsSpace) and `\d`,
// which matches any Nd digit — an Arabic-Indic numbered step is a step to the
// oracle. Anchored at the start only; the pattern has no trailing anchor, so
// anything may follow the closing paren.
func isStep(line string) bool {
	i := skipSpace(line, 0)
	digits := 0
	for i < len(line) {
		r, size := decodeRune(line[i:])
		if !unicode.IsDigit(r) {
			break
		}
		digits++
		i += size
	}
	if digits == 0 || i >= len(line) || line[i] != '.' {
		return false
	}
	i = skipSpace(line, i+1)
	if i >= len(line) || line[i] != '(' {
		return false
	}
	rest := line[i+1:]
	for _, tier := range tiers {
		if strings.HasPrefix(rest, tier+")") {
			return true
		}
	}
	return false
}

func skipSpace(s string, i int) int {
	for i < len(s) {
		r, size := decodeRune(s[i:])
		if !pyIsSpace(r) {
			return i
		}
		i += size
	}
	return i
}

// decodeRune is utf8.DecodeRuneInString, guaranteed to advance: a malformed
// byte reports width 1, which keeps the scanners above from looping forever on
// input internal/frontmatter would have rejected anyway.
func decodeRune(s string) (rune, int) {
	r, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return utf8.RuneError, 1
	}
	return r, size
}
