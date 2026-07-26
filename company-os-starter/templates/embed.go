// Package templates carries the built-in scaffolding templates into the binary.
//
// It exists for one structural reason: a //go:embed directive may only name
// files at or below its own package directory. Nothing under internal/ can
// therefore embed company-os-starter/templates/, and the alternative — copying
// the .md files down beside a deeper package — would fork the one source of
// flavor. So the directive lives here, next to the files it embeds, and the
// template files stay exactly where they are, matching the layout sketch in
// docs/lld/go-cli-tui-port.md ("templates/  //go:embed").
//
// Nothing but embedded bytes lives here. Resolution — the workspace-relative
// override chain of bin/company-os:533-548 — is internal/scaffold's business, so
// the only logic stays under the AST-enforced no-exit/no-stdout rule that
// cmd/company-os/architecture_test.go applies to internal/.
package templates

import _ "embed"

// RealityComponent is templates/reality-component.md, verbatim.
//
// This single file is the whole of R-1.11 and R-6.7: _builtin_template
// (bin/company-os:526-529) read it from TEMPLATES_DIR, a path relative to the
// script's own location, and it is the only such read anywhere in the CLI.
// Embedding it makes the "reality template not found" die at bin/company-os:2041
// unreachable, which R-0.7a(c) sanctions.
//
// The discovery-brief and PRD built-ins are not embedded: no file on disk holds
// their text. They are Python module strings today (bin/company-os:378, :470)
// and Go constants in internal/scaffold now. The sibling files in this directory
// — templates/discovery-brief.md, templates/prd.md, adr.md, deviations.yaml and
// the rest — are human reference copies with different placeholder text that the
// CLI has never read; embedding them would ship bytes no code path can reach.
//
//go:embed reality-component.md
var RealityComponent string
