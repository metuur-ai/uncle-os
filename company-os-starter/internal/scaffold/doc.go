// Package scaffold creates new workspace objects from embedded templates.
//
// Responsibility: `init`, `add platform|team|component`, `reality new`, and
// `scratchpad init`. Refusing to overwrite an existing artifact is a single
// sentinel error here, mapped to exit code 8 at the dispatch boundary rather
// than decided at each of the five Python sites.
//
// Template resolution (ResolveTemplate, template.go) landed with task 1.8 and is
// shared with internal/product, which scaffolds discovery briefs and PRDs from
// the same chain. The commands themselves are not implemented yet.
package scaffold
