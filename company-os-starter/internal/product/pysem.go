package product

// The Python expressions this cluster leans on that internal/yamlio does not
// already answer, plus the one construct no other cluster needs: str.format.
//
// Everything scalar-shaped here is a thin adapter over yamlio (PyFalsy,
// PyString) rather than a reimplementation — the scalar semantics and the
// truthiness rule live there, and a second copy of either is how the two
// answers drift.

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

type (
	pyMap = yamlio.PyMap
	pySeq = yamlio.PySeq
	pyVal = yamlio.PyValue
)

// today is TODAY (`bin/company-os:31`) — datetime.date.today(), the local
// calendar date, truncated to midnight so AddDate below is pure CALENDAR
// arithmetic exactly as `date + timedelta(days=90)` is. Adding 90*24h to a wall
// clock lands on the previous day across a DST spring-forward, which would be a
// one-day-wrong `due:` written into an archived outcome review.
func today() time.Time {
	n := time.Now()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, time.UTC)
}

// isoDate renders a date the way `str(datetime.date)` and an f-string do.
const isoDate = "2006-01-02"

// relTo is Path.relative_to(root) in POSIX form (R-1.12). It falls back to the
// absolute path where Python raises ValueError, which no call site reaches.
func relTo(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// truthy is `if meta.get(field):` over a loaded frontmatter value: absent and
// falsy are the same answer, which is what core_field_errors and the
// process-contract loop both branch on.
func truthy(m pyMap, key string) bool {
	if m == nil {
		return false
	}
	return !yamlio.PyFalsy(m.Get(key))
}

// strOf is `str(meta.get(key))` with Python's rendering of None, of a bool, and
// — the case that matters here — of a YAML-parsed datetime.date.
func strOf(m pyMap, key string) string {
	if m == nil {
		return "None"
	}
	return yamlio.PyString(m.Get(key))
}

// splitComponents is `[c.strip() for c in args.components.split(",")]`.
//
// Python's split(",") on "" yields [""], so an EMPTY --components resolves to
// one component whose id is the empty string rather than to none — which is
// what makes `prd new --components ""` warn about a component called ” rather
// than silently skipping the governance block. Reproduced deliberately.
func splitComponents(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

// sectionBody is `re.search(rf"## {section}\n(.*?)(\n## |\Z)", body, re.DOTALL)`
// — the grep `discover validate`, `prd validate` and `prd new --from-discovery`
// all run against a document body.
//
// The section name is interpolated INTO the pattern in Python, so a name
// carrying regex metacharacters would change the match. Every name comes from
// the two frozen *_SECTIONS tuples, so quoting it here would be a behaviour
// change rather than a hardening; the names are matched literally because that
// is what those tuples make them.
func sectionBody(body []byte, section string) (string, bool) {
	re := regexp.MustCompile(`(?s)## ` + regexp.QuoteMeta(section) + `\n(.*?)(\n## |\z)`)
	m := re.FindSubmatch(body)
	if m == nil {
		return "", false
	}
	return string(m[1]), true
}

// commentStrip is `re.sub(r"<!--.*?-->", "", content, flags=re.DOTALL)`, which
// is what makes a section holding only its scaffolded HTML hint count as empty.
var commentStrip = regexp.MustCompile(`(?s)<!--.*?-->`)

// sectionContent is the stripped body of one section: comments removed, then
// Python's str.strip().
func sectionContent(body []byte, section string) (content string, found bool) {
	raw, ok := sectionBody(body, section)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(commentStrip.ReplaceAllString(raw, "")), true
}

// ---------------------------------------------------------------- str.format

// formatField matches one replacement field of the subset the two scaffolding
// templates use: `{name}` and `{name[index]}`.
var formatField = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)(?:\[([0-9]+)\])?$`)

// formatTemplate is Python's str.format(**kwargs) over the subset the discovery
// and PRD templates use: `{{` and `}}` escapes, `{name}`, and `{name[index]}`
// into a list.
//
// The subset is the point. Python's mini-language also has `!r` conversions,
// `:spec` formats, attribute access and nested fields; none appears in either
// built-in template, and an OVERRIDE that used one would be silently
// mis-rendered by a partial implementation. So anything outside the subset is
// refused by name (exit 4) rather than approximated — Python raises there too,
// for the values these call sites pass.
//
// Both refusals map to R-0.7a(e): Python's KeyError/IndexError is a traceback
// and exit 1, this is a diagnostic naming the offending field.
func formatTemplate(tmpl, label string, args map[string]any) (string, error) {
	var b strings.Builder
	for i := 0; i < len(tmpl); {
		c := tmpl[i]
		if c == '}' {
			if i+1 < len(tmpl) && tmpl[i+1] == '}' {
				b.WriteByte('}')
				i += 2
				continue
			}
			return "", model.Errorf(model.ExitArtifact,
				"%s: single '}' encountered in format string", label)
		}
		if c != '{' {
			b.WriteByte(c)
			i++
			continue
		}
		if i+1 < len(tmpl) && tmpl[i+1] == '{' {
			b.WriteByte('{')
			i += 2
			continue
		}
		end := strings.IndexByte(tmpl[i:], '}')
		if end < 0 {
			return "", model.Errorf(model.ExitArtifact,
				"%s: expected '}' before end of format string", label)
		}
		field := tmpl[i+1 : i+end]
		text, err := formatOne(field, label, args)
		if err != nil {
			return "", err
		}
		b.WriteString(text)
		i += end + 1
	}
	return b.String(), nil
}

func formatOne(field, label string, args map[string]any) (string, error) {
	m := formatField.FindStringSubmatch(field)
	if m == nil {
		return "", model.Errorf(model.ExitArtifact,
			"%s: unsupported replacement field '{%s}' — this CLI substitutes "+
				"'{name}' and '{name[index]}' only", label, field)
	}
	v, ok := args[m[1]]
	if !ok {
		return "", model.Errorf(model.ExitArtifact,
			"%s: template refers to unknown field '%s'", label, m[1])
	}
	if m[2] == "" {
		s, ok := v.(string)
		if !ok {
			return "", model.Errorf(model.ExitArtifact,
				"%s: field '%s' is a list; index it as '{%s[0]}'", label, m[1], m[1])
		}
		return s, nil
	}
	list, ok := v.([]string)
	if !ok {
		return "", model.Errorf(model.ExitArtifact,
			"%s: field '%s' is not indexable", label, m[1])
	}
	n, err := strconv.Atoi(m[2])
	if err != nil || n >= len(list) {
		return "", model.Errorf(model.ExitArtifact,
			"%s: index %s out of range for field '%s'", label, m[2], m[1])
	}
	return list[n], nil
}

// wrote turns a write failure into an artifact fault. Python's write_text() is
// unguarded, so a full disk is a traceback there; here it is exit 4 with the
// path named (R-0.7a(e)).
func wrote(path string, err error) error {
	if err == nil {
		return nil
	}
	return model.Errorf(model.ExitArtifact, "cannot write %s: %v", path, err)
}
