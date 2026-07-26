package frontmatter

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"unicode/utf8"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// ErrInvalidUTF8 is the Go stand-in for the UnicodeDecodeError that
// Path.read_text() raises inside the Python frontmatter() before its regex ever
// runs. Python never returns a value in that case; neither do we.
//
// It carries exit code 4: an undecodable artifact is a malformed artifact
// (R-4.5), and R-0.7a(e) sanctions replacing Python's traceback-and-exit-1 with
// a diagnostic. errors.Is still matches it — the sentinel is compared by
// identity, and being a *model.Error changes only what CodeOf reads off it.
var ErrInvalidUTF8 error = &model.Error{
	Code: model.ExitArtifact,
	Msg:  "frontmatter: not valid UTF-8",
}

// fence mirrors bin/company-os:76 — re.match(r"^---\n(.*?)\n---\n(.*)$", text,
// re.DOTALL) — clause for clause:
//
//   - (?s) is Python's re.DOTALL. Without it Go's `.` stops at every newline and
//     no multi-line document would ever match.
//   - \A is Python's re.match anchor. Go's regexp searches the whole input, so
//     the anchor has to be written out.
//   - \z, not $. Go's `$` means end-of-text (Python's also matches just *before*
//     a trailing newline), so the two spellings diverge in general. They agree
//     here only because the trailing (.*) is greedy and reaches end-of-text
//     first — measured: group 2 keeps its trailing "\n". \z states that intent
//     directly and stays correct if the pattern is ever touched again.
//   - (.*?) stays non-greedy: it is what makes the FIRST "\n---\n" the closing
//     fence, so a "---" line inside the body cannot capture the split.
var fence = regexp.MustCompile(`(?s)\A---\n(.*?)\n---\n(.*)\z`)

// Document is one parsed artifact.
//
// It mirrors Python's (dict, body) return, minus the YAML: HasFrontmatter false
// is Python's ({}, text) — the document had no frontmatter, and Body is the
// whole (newline-translated) text.
type Document struct {
	// YAML is the raw frontmatter block, regex group 1, with no trailing
	// newline. Nil when HasFrontmatter is false. Deliberately NOT parsed here:
	// yaml.safe_load(...) or {} lives on the other side of this seam, in
	// internal/yamlio.
	YAML []byte
	// Body is regex group 2 on a match, else the whole translated text. It
	// keeps its trailing newline, exactly as Python's does.
	Body []byte
	// HasFrontmatter reports whether the fence regex matched.
	HasFrontmatter bool
}

// Parse splits a document into its frontmatter block and body with the same
// accept/reject decision as Python's frontmatter().
//
// data is the raw file bytes. Decoding them is part of the contract, not the
// caller's job: Python's Path.read_text() decodes UTF-8 and applies universal
// newline translation BEFORE the regex sees the text, which is the only reason
// a CRLF document parses at all — the pattern itself contains no \r. Callers
// must hand over untouched bytes and let this function do both steps, or CRLF
// artifacts silently stop parsing.
func Parse(data []byte) (Document, error) {
	// Order matters: read_text() decodes first and raises before it can
	// translate anything.
	if !utf8.Valid(data) {
		return Document{}, ErrInvalidUTF8
	}
	text := translateNewlines(data)
	m := fence.FindSubmatchIndex(text)
	if m == nil {
		// Python: return {}, text — rejection is not an error, and the body is
		// the translated text, not the original bytes.
		return Document{Body: text}, nil
	}
	return Document{
		YAML:           text[m[2]:m[3]],
		Body:           text[m[4]:m[5]],
		HasFrontmatter: true,
	}, nil
}

// ParseFile is Parse over the contents of path, matching the Python signature.
func ParseFile(path string) (Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("frontmatter: %w", err)
	}
	doc, err := Parse(data)
	if err != nil {
		return Document{}, fmt.Errorf("%s: %w", path, err)
	}
	return doc, nil
}

// translateNewlines reproduces Python's universal newline mode (the newline=None
// default of open()): "\r\n" and a lone "\r" both become "\n".
func translateNewlines(b []byte) []byte {
	if bytes.IndexByte(b, '\r') < 0 {
		return b
	}
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] != '\r' {
			out = append(out, b[i])
			continue
		}
		out = append(out, '\n')
		if i+1 < len(b) && b[i+1] == '\n' {
			i++
		}
	}
	return out
}
