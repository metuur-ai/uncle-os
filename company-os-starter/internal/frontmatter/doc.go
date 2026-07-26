// Package frontmatter splits the `^---\n...\n---\n` block that every artifact
// carries from the body that follows it.
//
// Responsibility: the exact regex semantics of bin/company-os:76, including the
// universal-newline decode that Python's Path.read_text() performs before the
// regex ever runs, Go's need for (?s), and the difference between Go's and
// Python's `$` anchor.
//
// Scope, deliberately: this package stops at the fence. It hands back the
// frontmatter block as raw bytes and does not call a YAML parser — Python's
// `yaml.safe_load(...) or {}`, its exception behavior, and preserving unknown
// keys and key order on write all belong to internal/yamlio.
//
// Measured Python truth, corpus, and the surprises (`---\n---\n` is rejected;
// CRLF is accepted): .devlocal/go-port/frontmatter-truth.md
package frontmatter
