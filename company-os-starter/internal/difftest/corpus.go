// Package difftest is the end-to-end characterization suite for the company-os
// binary: an enumerated corpus of CLI invocations, each run against a fresh copy
// of a fixture workspace, with per-step stdout/stderr/exit and the resulting
// file tree frozen into testdata/.
//
// It is the port of examples/differential.py, which compared the Go binary
// against the Python reference implementation. That reference was deleted at
// cutover (R-9.3), so there is no second implementation left to differ against —
// but the corpus itself was the valuable half. `examples/acceptance.sh` byte-
// freezes `validate` only; this corpus is the only end-to-end byte-level
// coverage for discover, deviation, exception, check, governance, today, graph,
// ids, skills and scratchpad.
//
// The shape therefore changed from a differential (compare A to B) to a
// characterization test (compare to frozen truth). One consequence is worth
// stating plainly: a differential proves two implementations agree, while a
// golden only proves behaviour has not changed since someone looked at it. The
// goldens are only as good as the review of the diff that last updated them.
//
// The entire declared-divergence machinery from the Python harness — the waiver
// registry, its citation grammar and its 631 records — is deliberately NOT
// ported. Every one of those waivers sanctioned a Python/Go difference. Against
// a golden there is no second implementation, so any difference is a regression
// and waiving it would defeat the point.
package difftest

// Inv is one corpus entry: a fixture plus an ordered list of argv steps.
//
// Multiple steps run in the SAME workspace copy, in order, so lifecycle
// sequences (prd new -> validate -> complete) and state-dependent commands
// (workspace sync -> status -> sync --frozen) are reachable. Every step's
// stdout/stderr/exit is recorded; the tree is recorded once at the end.
type Inv struct {
	ID      string
	Fixture string
	Steps   [][]string
	Group   string
	Note    string

	// Unstable, when non-empty, names one stream of this invocation that is not
	// deterministic even under a SINGLE binary, so freezing it would produce a
	// golden that fails at random. The stream is excluded and the exclusion is
	// written into the golden with its reason, never silently dropped. Format:
	// "<stream>: <reason>". Every use is a hole in the suite and must justify
	// itself.
	Unstable string
}

// UnstableStream is the stream name from Unstable, or "" when the whole
// invocation is deterministic.
func (i Inv) UnstableStream() string {
	for n := 0; n < len(i.Unstable); n++ {
		if i.Unstable[n] == ':' {
			return i.Unstable[:n]
		}
	}
	return ""
}

type opt func(*Inv)

func group(g string) opt    { return func(i *Inv) { i.Group = g } }
func note(n string) opt     { return func(i *Inv) { i.Note = n } }
func unstable(u string) opt { return func(i *Inv) { i.Unstable = u } }

// Fixture short names, matching the Python corpus so invocation ids are stable.
const (
	fxW  = "workspace"
	fxS  = "standalone-team"
	fxF  = "federated"
	fxBS = "banking-small"
	fxBR = "banking-rails"
	fxBF = "banking-fraud"
	fxXW = "failing-workspace"
	fxXF = "failing-federated"
	fxXN = "failing-federated-nolock"
)

// allWS and failingWS drive the per-fixture loops. Order is load-bearing: it
// fixes the order ids are generated in, which the corpus-id golden pins.
var (
	allWS     = []string{fxW, fxS, fxF, fxBS, fxBR, fxBF}
	failingWS = []string{fxXW, fxXF, fxXN}
	roles     = []string{"developer", "team-lead", "product-owner", "architect",
		"vp-engineering", "director-of-product"}
)

var corpus []Inv

func inv(id, fixture string, steps [][]string, opts ...opt) {
	e := Inv{ID: id, Fixture: fixture, Steps: steps, Group: defaultGroup(id)}
	for _, o := range opts {
		o(&e)
	}
	corpus = append(corpus, e)
}

func defaultGroup(id string) string {
	for n := 0; n < len(id); n++ {
		if id[n] == '/' {
			return id[:n]
		}
	}
	return id
}

// Corpus returns every invocation, in declaration order.
func Corpus() []Inv { return corpus }

func init() {
	buildValidate()
	buildGovernance()
	buildDiscover()
	buildPRD()
	buildCheck()
	buildDeviation()
	buildException()
	buildToday()
	buildGraph()
	buildIDs()
	buildSkills()
	buildScratchpad()
	buildInit()
	buildAdd()
	buildReality()
	buildWorkspace()
	buildUsage()
}

// --- validate (byte-frozen by acceptance.sh; here for exit code + tree) ------

func buildValidate() {
	for _, fx := range allWS {
		inv("validate/"+fx, fx, [][]string{{"validate"}})
	}
	inv("validate/not-a-root", "empty", [][]string{{"validate"}}, note("exit 3"))
	inv("validate/after-graph-build", fxW, [][]string{{"graph", "build"}, {"validate"}})

	// R-0.9 failure-path fixtures: every gate's [FAIL]/[warn] rendering.
	for _, fx := range failingWS {
		inv("validate/"+fx, fx, [][]string{{"validate"}})
		inv("graph/build-"+fx, fx, [][]string{{"graph", "build"}}, group("graph"))
		inv("today/default-"+fx, fx, [][]string{{"today"}}, group("today"))
		inv("ids/list-"+fx, fx, [][]string{{"ids", "list"}}, group("ids"))
		inv("skills/list-"+fx, fx, [][]string{{"skills", "list"}}, group("skills"))
		inv("workspace/status-"+fx, fx, [][]string{{"workspace", "status"}}, group("workspace"))
	}
	inv("governance/resolve-failing-workspace", fxXW,
		[][]string{{"governance", "resolve", "--team", "ghost"}})
	inv("governance/explain-failing-workspace", fxXW,
		[][]string{{"governance", "explain", "orphan"}})
	inv("check/ready-failing-workspace", fxXW,
		[][]string{{"check", "ready", "--team", "ghost", "--components", "orphan"}})
	inv("workspace/frozen-failing-federated", fxXF,
		[][]string{{"workspace", "sync", "--frozen"}}, group("workspace"))
	inv("workspace/frozen-failing-nolock", fxXN,
		[][]string{{"workspace", "sync", "--frozen"}}, group("workspace"))
}

// --- governance -------------------------------------------------------------

func buildGovernance() {
	inv("governance/resolve-workspace", fxW,
		[][]string{{"governance", "resolve", "--team", "customer-engagement"}})
	inv("governance/resolve-federated", fxF,
		[][]string{{"governance", "resolve", "--team", "customer-engagement"}})
	inv("governance/resolve-standalone", fxS,
		[][]string{{"governance", "resolve", "--team", "solo"}})
	inv("governance/resolve-banking", fxBS,
		[][]string{{"governance", "resolve", "--team", "core"}})
	inv("governance/resolve-fraud", fxBF,
		[][]string{{"governance", "resolve", "--team", "fraud-detection"}})
	inv("governance/resolve-rails", fxBR,
		[][]string{{"governance", "resolve", "--team", "payments-rails"}})
	inv("governance/resolve-twice", fxW, [][]string{
		{"governance", "resolve", "--team", "customer-engagement"},
		{"governance", "resolve", "--team", "customer-engagement"},
	}, note("idempotence of generated/effective-governance.yaml"))
	inv("governance/resolve-unknown-team", fxW,
		[][]string{{"governance", "resolve", "--team", "nope"}})
	inv("governance/resolve-no-team", fxW, [][]string{{"governance", "resolve"}})
	inv("governance/explain-known", fxW,
		[][]string{{"governance", "explain", "customer-notification-service"}})
	inv("governance/explain-banking", fxBS,
		[][]string{{"governance", "explain", "banking-app"}})
	inv("governance/explain-unknown-with-suggestions", fxW,
		[][]string{{"governance", "explain", "customer-notification-servic"}},
		note("close-match suggestion path (GPF-R-2.3)"))
	inv("governance/explain-unknown-no-suggestions", fxW,
		[][]string{{"governance", "explain", "zzzzzzzz"}})
	inv("governance/explain-standalone", fxS,
		[][]string{{"governance", "explain", "anything"}})
	inv("governance/explain-no-component", fxW, [][]string{{"governance", "explain"}})
	inv("governance/explain-after-resolve", fxS, [][]string{
		{"governance", "resolve", "--team", "solo"},
		{"governance", "explain", "anything"},
	})
}

// --- discover ---------------------------------------------------------------

func buildDiscover() {
	inv("discover/new", fxW,
		[][]string{{"discover", "new", "--team", "customer-engagement", "Quiet hours v2"}})
	inv("discover/new-twice-conflict", fxW, [][]string{
		{"discover", "new", "--team", "customer-engagement", "Quiet hours v2"},
		{"discover", "new", "--team", "customer-engagement", "Quiet hours v2"},
	}, note("second must refuse (exit 8)"))
	inv("discover/new-then-validate", fxW, [][]string{
		{"discover", "new", "--team", "customer-engagement", "Quiet hours v2"},
		{"discover", "validate", "--team", "customer-engagement", "2026-quiet-hours-v2"},
	}, note("fresh brief is a stub -> validate should fail"))
	inv("discover/new-standalone", fxS,
		[][]string{{"discover", "new", "--team", "solo", "Solo idea"}})
	inv("discover/new-banking", fxBS,
		[][]string{{"discover", "new", "--team", "core", "Statements v2"}})
	inv("discover/new-unknown-team", fxW,
		[][]string{{"discover", "new", "--team", "nope", "X"}})
	inv("discover/new-slugify", fxW,
		[][]string{{"discover", "new", "--team", "customer-engagement", "  Weird // Title!! 2026  "}})
	inv("discover/new-no-title", fxW,
		[][]string{{"discover", "new", "--team", "customer-engagement"}})
	inv("discover/validate-passing", fxW,
		[][]string{{"discover", "validate", "--team", "customer-engagement", "2026-per-channel-quiet-hours"}})
	inv("discover/validate-federated", fxF,
		[][]string{{"discover", "validate", "--team", "customer-engagement", "2026-per-channel-quiet-hours"}})
	inv("discover/validate-banking", fxBS,
		[][]string{{"discover", "validate", "--team", "core", "2026-instant-statements"}})
	inv("discover/validate-fraud", fxBF,
		[][]string{{"discover", "validate", "--team", "fraud-detection", "2026-alert-triage-queues"}})
	inv("discover/validate-missing", fxW,
		[][]string{{"discover", "validate", "--team", "customer-engagement", "2026-nope"}})
	inv("discover/validate-unknown-team", fxW,
		[][]string{{"discover", "validate", "--team", "nope", "x"}})
	inv("discover/missing-team-flag", fxW,
		[][]string{{"discover", "new", "Title"}}, note("parser usage error"))
	inv("discover/bad-action", fxW,
		[][]string{{"discover", "frobnicate", "--team", "customer-engagement", "x"}})
}

// --- prd (lifecycle is invariant #4) ----------------------------------------

var prdNewW = []string{"prd", "new", "--team", "customer-engagement", "--platform", "communications",
	"--components", "customer-notification-service",
	"--from-discovery", "2026-per-channel-quiet-hours"}

func buildPRD() {
	inv("prd/new-from-discovery", fxW, [][]string{prdNewW})
	inv("prd/new-then-validate", fxW, [][]string{prdNewW,
		{"prd", "validate", "--platform", "communications", "2026-per-channel-quiet-hours"}})
	inv("prd/full-lifecycle", fxW, [][]string{prdNewW,
		{"prd", "validate", "--platform", "communications", "2026-per-channel-quiet-hours"},
		{"prd", "complete", "--platform", "communications", "2026-per-channel-quiet-hours"},
	}, note("done-check should refuse: stale reality/ (invariant #4)"))
	inv("prd/full-lifecycle-force", fxW, [][]string{prdNewW,
		{"prd", "complete", "--platform", "communications", "2026-per-channel-quiet-hours", "--force"},
	}, note("archive + outcome.md + log.md written"))
	inv("prd/new-twice-conflict", fxW, [][]string{prdNewW, prdNewW})
	inv("prd/new-with-title", fxW, [][]string{{"prd", "new", "--team", "customer-engagement",
		"--platform", "communications", "--components", "customer-notification-service",
		"--title", "Ad hoc change"}})
	inv("prd/new-no-title-no-discovery", fxW, [][]string{{"prd", "new", "--team", "customer-engagement",
		"--platform", "communications", "--components", "customer-notification-service"}})
	inv("prd/new-missing-discovery", fxW, [][]string{{"prd", "new", "--team", "customer-engagement",
		"--platform", "communications", "--components", "customer-notification-service",
		"--from-discovery", "2026-nope"}})
	inv("prd/new-draft-discovery", fxW, [][]string{
		{"discover", "new", "--team", "customer-engagement", "Draft idea"},
		{"prd", "new", "--team", "customer-engagement", "--platform", "communications",
			"--components", "customer-notification-service", "--from-discovery", "2026-draft-idea"},
	}, note("brief is draft, not validated -> exit 5"))
	inv("prd/new-unknown-platform", fxW, [][]string{{"prd", "new", "--team", "customer-engagement",
		"--platform", "nope", "--components", "x", "--title", "T"}})
	inv("prd/new-missing-platform-flag", fxW,
		[][]string{{"prd", "new", "--team", "customer-engagement", "--title", "T"}})
	inv("prd/validate-missing", fxW,
		[][]string{{"prd", "validate", "--platform", "communications", "2026-nope"}})
	inv("prd/complete-missing", fxW,
		[][]string{{"prd", "complete", "--platform", "communications", "2026-nope"}})
	inv("prd/validate-banking-active", fxBS,
		[][]string{{"prd", "validate", "--platform", "product", "2026-instant-statements"}})
	inv("prd/complete-banking-active", fxBS,
		[][]string{{"prd", "complete", "--platform", "product", "2026-instant-statements"}})
	inv("prd/complete-banking-force", fxBS,
		[][]string{{"prd", "complete", "--platform", "product", "2026-instant-statements", "--force"}})
	inv("prd/new-multi-component", fxW, [][]string{{"prd", "new", "--team", "customer-engagement",
		"--platform", "communications",
		"--components", "customer-notification-service,ghost-component",
		"--title", "Multi component"}})
}

// --- check ------------------------------------------------------------------

func buildCheck() {
	inv("check/ready-workspace", fxW, [][]string{{"check", "ready", "--team", "customer-engagement",
		"--components", "customer-notification-service"}})
	inv("check/done-workspace", fxW, [][]string{{"check", "done", "--team", "customer-engagement",
		"--components", "customer-notification-service"}})
	inv("check/ready-federated", fxF, [][]string{{"check", "ready", "--team", "customer-engagement",
		"--components", "customer-notification-service"}})
	inv("check/ready-banking", fxBS,
		[][]string{{"check", "ready", "--team", "core", "--components", "banking-app"}})
	inv("check/done-banking", fxBS,
		[][]string{{"check", "done", "--team", "core", "--components", "banking-app"}})
	inv("check/ready-standalone", fxS,
		[][]string{{"check", "ready", "--team", "solo", "--components", "none"}})
	inv("check/done-standalone", fxS,
		[][]string{{"check", "done", "--team", "solo", "--components", "none"}})
	inv("check/ready-multi", fxW, [][]string{{"check", "ready", "--team", "customer-engagement",
		"--components", "customer-notification-service,ghost,another"}})
	inv("check/ready-unknown-team", fxW,
		[][]string{{"check", "ready", "--team", "nope", "--components", "x"}})
	inv("check/ready-unknown-component", fxW, [][]string{{"check", "ready",
		"--team", "customer-engagement", "--components", "does-not-exist"}})
	inv("check/missing-components-flag", fxW,
		[][]string{{"check", "ready", "--team", "customer-engagement"}})
	inv("check/bad-kind", fxW,
		[][]string{{"check", "sideways", "--team", "customer-engagement", "--components", "x"}})
	inv("check/ready-fraud", fxBF, [][]string{{"check", "ready", "--team", "fraud-detection",
		"--components", "fraud-scoring-engine"}})
}

// --- deviation --------------------------------------------------------------

func buildDeviation() {
	devDefault := []string{"deviation", "declare",
		"platform-standard://communications/prd-structure", "--team", "customer-engagement"}

	inv("deviation/declare-default-rule", fxW, [][]string{devDefault})
	inv("deviation/declare-then-resolve", fxW, [][]string{devDefault,
		{"governance", "resolve", "--team", "customer-engagement"},
	}, note("deviation must appear in deviationsApplied"))
	inv("deviation/declare-mandatory-then-resolve", fxW, [][]string{
		{"deviation", "declare", "platform-standard://communications/delivery-reliability",
			"--team", "customer-engagement"},
		{"governance", "resolve", "--team", "customer-engagement"},
	}, note("mandatory rule -> deviationRejected (invariant #1)"))
	inv("deviation/declare-with-rationale", fxW, [][]string{{"deviation", "declare",
		"platform-standard://communications/prd-structure", "--team", "customer-engagement",
		"--rationale", "we ship a different shape"}})
	inv("deviation/declare-twice", fxW, [][]string{devDefault, devDefault,
		{"governance", "resolve", "--team", "customer-engagement"}})
	inv("deviation/declare-standalone", fxS,
		[][]string{{"deviation", "declare", "company-control://change-log", "--team", "solo"}})
	inv("deviation/declare-banking", fxBS, [][]string{
		{"deviation", "declare", "company-control://change-log", "--team", "core"},
		{"governance", "resolve", "--team", "core"},
	})
	inv("deviation/declare-unknown-team", fxW,
		[][]string{{"deviation", "declare", "x", "--team", "nope"}})
	inv("deviation/declare-then-validate", fxW, [][]string{devDefault,
		{"governance", "resolve", "--team", "customer-engagement"},
		{"validate"},
	}, note("reviewDate is TODAY+180 -> gate [2/8] must stay green"))
	inv("deviation/bad-action", fxW,
		[][]string{{"deviation", "revoke", "x", "--team", "customer-engagement"}})
}

// --- exception --------------------------------------------------------------

func buildException() {
	inv("exception/request-future", fxW, [][]string{{"exception", "request",
		"platform-standard://communications/delivery-reliability",
		"--team", "customer-engagement", "--component", "customer-notification-service",
		"--expires", "2035-01-01"}})
	inv("exception/request-then-validate", fxW, [][]string{{"exception", "request",
		"platform-standard://communications/delivery-reliability",
		"--team", "customer-engagement", "--component", "customer-notification-service",
		"--expires", "2035-01-01"}, {"validate"}})
	inv("exception/request-expired-then-validate", fxW, [][]string{{"exception", "request",
		"platform-standard://communications/delivery-reliability",
		"--team", "customer-engagement", "--component", "customer-notification-service",
		"--expires", "2020-01-01"}, {"validate"},
	}, note("past expiry -> gate [2/8] must FAIL"))
	inv("exception/request-with-reason", fxW, [][]string{{"exception", "request",
		"platform-standard://communications/message-schema",
		"--team", "customer-engagement", "--component", "customer-notification-service",
		"--expires", "2035-06-30", "--reason", "legacy consumer"}})
	inv("exception/request-standalone", fxS, [][]string{{"exception", "request",
		"company-control://security-service-baseline",
		"--team", "solo", "--component", "none", "--expires", "2035-01-01"}})
	inv("exception/request-banking", fxBS, [][]string{{"exception", "request",
		"platform-standard://product/release-safety",
		"--team", "core", "--component", "banking-app", "--expires", "2035-01-01"}, {"validate"}})
	inv("exception/request-unknown-team", fxW, [][]string{{"exception", "request", "x",
		"--team", "nope", "--component", "c", "--expires", "2035-01-01"}})
	inv("exception/missing-expires", fxW,
		[][]string{{"exception", "request", "x", "--team", "customer-engagement", "--component", "c"}})
	inv("exception/missing-component", fxW, [][]string{{"exception", "request", "x",
		"--team", "customer-engagement", "--expires", "2035-01-01"}})
	inv("exception/garbage-expires", fxW, [][]string{{"exception", "request",
		"platform-standard://communications/message-schema",
		"--team", "customer-engagement", "--component", "customer-notification-service",
		"--expires", "not-a-date"}, {"validate"}})
}

// --- today ------------------------------------------------------------------

func buildToday() {
	for _, fx := range allWS {
		inv("today/default-"+fx, fx, [][]string{{"today"}})
	}
	for _, role := range roles {
		inv("today/role-"+role, fxW, [][]string{{"today", "--role", role}})
	}
	inv("today/po-banking", fxBS, [][]string{{"today", "--role", "product-owner"}},
		note("active PRD + archived outcome review"))
	inv("today/dev-banking", fxBS, [][]string{{"today", "--role", "developer"}})
	inv("today/dev-standalone-no-generated", fxS, [][]string{{"today", "--role", "developer"}},
		note("warn: no effective-governance.yaml"))
	inv("today/after-resolve", fxS, [][]string{
		{"governance", "resolve", "--team", "solo"}, {"today", "--role", "developer"}})
	inv("today/bad-role", fxW, [][]string{{"today", "--role", "wizard"}})
	inv("today/not-a-root", "empty", [][]string{{"today"}})
}

// --- graph ------------------------------------------------------------------

func buildGraph() {
	for _, fx := range allWS {
		inv("graph/build-"+fx, fx, [][]string{{"graph", "build"}})
	}
	inv("graph/build-twice", fxW, [][]string{{"graph", "build"}, {"graph", "build"}},
		note("idempotence (R-0.6)"))
	inv("graph/build-then-validate", fxBS, [][]string{{"graph", "build"}, {"validate"}})
	inv("graph/build-after-discover", fxW, [][]string{
		{"discover", "new", "--team", "customer-engagement", "Tagged brief"},
		{"graph", "build"},
	}, note("derived tags: for a brand-new artifact"))
	inv("graph/build-not-a-root", "empty", [][]string{{"graph", "build"}})
	inv("graph/bad-action", fxW, [][]string{{"graph", "rebuild"}})
}

// --- ids --------------------------------------------------------------------

func buildIDs() {
	for _, fx := range allWS {
		inv("ids/list-"+fx, fx, [][]string{{"ids", "list"}})
	}
	inv("ids/filter-team", fxW, [][]string{{"ids", "list", "--team", "customer-engagement"}})
	inv("ids/filter-platform", fxW, [][]string{{"ids", "list", "--platform", "communications"}})
	inv("ids/filter-prefix-component", fxW, [][]string{{"ids", "list", "--prefix", "component://"}})
	inv("ids/filter-prefix-team", fxW, [][]string{{"ids", "list", "--prefix", "team://"}})
	inv("ids/filter-prefix-nomatch", fxW, [][]string{{"ids", "list", "--prefix", "zzzz"}})
	inv("ids/filter-team-nomatch", fxW, [][]string{{"ids", "list", "--team", "nope"}})
	inv("ids/filter-combined", fxW, [][]string{{"ids", "list",
		"--team", "customer-engagement", "--platform", "communications"}})
	for _, role := range roles {
		inv("ids/role-"+role, fxW, [][]string{{"ids", "list", "--role", role}})
	}
	inv("ids/role-unknown", fxW, [][]string{{"ids", "list", "--role", "wizard"}})
	inv("ids/list-banking-filtered", fxBS, [][]string{{"ids", "list", "--platform", "product"}})
	inv("ids/bad-action", fxW, [][]string{{"ids", "show"}})
	inv("ids/not-a-root", "empty", [][]string{{"ids", "list"}})
}

// --- skills -----------------------------------------------------------------

func buildSkills() {
	for _, fx := range allWS {
		inv("skills/list-"+fx, fx, [][]string{{"skills", "list"}})
	}
	inv("skills/list-after-scratchpad", fxW, [][]string{
		{"scratchpad", "init", "--repo", "teams/customer-engagement"},
		{"skills", "list"},
	}, note("personal-rules layer becomes discoverable"))
	inv("skills/bad-action", fxW, [][]string{{"skills", "show"}})
	inv("skills/not-a-root", "empty", [][]string{{"skills", "list"}})
}

// --- scratchpad (exempt from require_root) ----------------------------------

func buildScratchpad() {
	inv("scratchpad/init-cwd", fxW, [][]string{{"scratchpad", "init"}})
	inv("scratchpad/init-repo-dot", fxW, [][]string{{"scratchpad", "init", "--repo", "."}})
	inv("scratchpad/init-nested-new-dir", fxW, [][]string{{"scratchpad", "init", "--repo", "sub/dir"}})
	inv("scratchpad/init-twice", fxW, [][]string{{"scratchpad", "init"}, {"scratchpad", "init"}},
		note(".gitignore must not be appended twice"))
	inv("scratchpad/init-standalone", fxS, [][]string{{"scratchpad", "init"}})
	inv("scratchpad/init-outside-workspace", "empty", [][]string{{"scratchpad", "init"}},
		note("exempt from require_root"))
	inv("scratchpad/init-team-dir", fxW,
		[][]string{{"scratchpad", "init", "--repo", "teams/customer-engagement"}})
	inv("scratchpad/bad-action", fxW, [][]string{{"scratchpad", "reset"}})
}

// --- init -------------------------------------------------------------------

func buildInit() {
	inv("init/full-flags", "empty",
		[][]string{{"init", "--company", "Acme", "--team", "core", "--platform", "platform-1"}})
	inv("init/then-validate", "empty", [][]string{
		{"init", "--company", "Acme", "--team", "core", "--platform", "platform-1"}, {"validate"}})
	inv("init/slugify", "empty", [][]string{{"init",
		"--company", "Acme Inc.", "--team", "Core Team!", "--platform", "My Platform!!"}})
	inv("init/non-interactive-no-flags", "empty", [][]string{{"init"}}, note("no TTY -> exit 7"))
	inv("init/non-interactive-partial-flags", "empty", [][]string{{"init", "--company", "Acme"}})
	inv("init/refuse-reinit", fxW,
		[][]string{{"init", "--company", "Acme", "--team", "core", "--platform", "p"}},
		note("already a root -> exit 8, mutating nothing"))
	inv("init/refuse-reinit-manifest-only", fxBR,
		[][]string{{"init", "--company", "Acme", "--team", "core", "--platform", "p"}},
		note("workspace.yaml alone marks a root (GPF-R-6.2)"))
	inv("init/then-add-then-validate", "empty", [][]string{
		{"init", "--company", "Acme", "--team", "core", "--platform", "platform-1"},
		{"add", "component", "widget", "--platform", "platform-1"},
		{"validate"},
	})
}

// --- add --------------------------------------------------------------------

func buildAdd() {
	inv("add/platform", fxW, [][]string{{"add", "platform", "newplat"}})
	inv("add/team", fxW, [][]string{{"add", "team", "newteam"}})
	inv("add/component", fxW, [][]string{{"add", "component", "newcomp", "--platform", "communications"}})
	inv("add/component-no-platform", fxW, [][]string{{"add", "component", "newcomp"}})
	inv("add/component-unknown-platform", fxW,
		[][]string{{"add", "component", "newcomp", "--platform", "nope"}})
	inv("add/platform-slugify", fxW, [][]string{{"add", "platform", "Weird Name!!"}})
	inv("add/platform-standalone", fxS, [][]string{{"add", "platform", "newplat"}},
		note("platforms/ does not exist yet"))
	inv("add/team-standalone", fxS, [][]string{{"add", "team", "second"}})
	inv("add/platform-then-component-then-validate", fxW, [][]string{
		{"add", "platform", "newplat"},
		{"add", "component", "newcomp", "--platform", "newplat"},
		{"validate"},
	})
	inv("add/duplicate-platform", fxW, [][]string{{"add", "platform", "communications"}},
		note("existing platform id"))
	inv("add/bad-kind", fxW, [][]string{{"add", "widget", "x"}})
	inv("add/banking", fxBS, [][]string{{"add", "team", "second"}, {"validate"}})
}

// --- reality ----------------------------------------------------------------

func buildReality() {
	inv("reality/new-for-new-component", fxW, [][]string{
		{"add", "component", "newcomp", "--platform", "communications"},
		{"reality", "new", "--platform", "communications", "newcomp"},
	})
	inv("reality/new-conflict", fxW,
		[][]string{{"reality", "new", "--platform", "communications", "customer-notification-service"}},
		note("reality doc exists -> exit 8"))
	inv("reality/new-unknown-platform", fxW, [][]string{{"reality", "new", "--platform", "nope", "x"}})
	inv("reality/new-no-platform-flag", fxW, [][]string{{"reality", "new", "x"}})
	inv("reality/new-orphan-component", fxW, [][]string{
		{"reality", "new", "--platform", "communications", "never-declared"}, {"validate"},
	}, note("reality doc for a component with no descriptor"))
	inv("reality/bad-action", fxW,
		[][]string{{"reality", "delete", "x", "--platform", "communications"}})
	inv("reality/new-banking", fxBS, [][]string{{"reality", "new", "--platform", "product", "second-app"}})
}

// --- workspace --------------------------------------------------------------

func buildWorkspace() {
	// Manifest + monorepo paths (git-free).
	inv("workspace/sync-no-manifest", fxW, [][]string{{"workspace", "sync"}})
	inv("workspace/status-no-manifest", fxW, [][]string{{"workspace", "status"}})
	inv("workspace/status-no-manifest-standalone", fxS, [][]string{{"workspace", "status"}})
	inv("workspace/status-federated", fxF, [][]string{{"workspace", "status"}},
		note("committed lock + committed slices -> clean"))
	inv("workspace/status-rails", fxBR, [][]string{{"workspace", "status"}}, note("never synced"))
	inv("workspace/status-fraud", fxBF, [][]string{{"workspace", "status"}}, note("never synced"))
	inv("workspace/frozen-federated-no-cache", fxF, [][]string{{"workspace", "sync", "--frozen"}},
		note("lock present, cache absent -> exit 6"))
	inv("workspace/frozen-no-lock", fxBR, [][]string{{"workspace", "sync", "--frozen"}})
	inv("workspace/only-nomatch", fxF, [][]string{{"workspace", "sync", "--only", "nope"}})
	inv("workspace/bad-action", fxF, [][]string{{"workspace", "pull"}})
	inv("workspace/not-a-root", "empty", [][]string{{"workspace", "status"}})

	// One `workspace status` + one `workspace sync` per bad manifest; each
	// exercises a distinct manifest-load rejection. Sorted so ids are stable.
	for _, name := range badManifestNames() {
		inv("workspace/manifest-"+name, "badmanifest-"+name,
			[][]string{{"workspace", "status"}}, group("workspace"))
		inv("workspace/manifest-"+name+"-sync", "badmanifest-"+name,
			[][]string{{"workspace", "sync"}}, group("workspace"))
	}
	inv("workspace/manifest-validate-gate", "badmanifest-empty-repos",
		[][]string{{"validate"}}, group("workspace"),
		note("a malformed manifest must also break `validate`"))

	// Real sync (git >= 2.27 required; file:// URL, no network).
	inv("workspace/sync-online", "gitfed", [][]string{{"workspace", "sync"}}, group("workspace-git"))
	inv("workspace/sync-then-status", "gitfed",
		[][]string{{"workspace", "sync"}, {"workspace", "status"}}, group("workspace-git"))
	inv("workspace/sync-then-frozen", "gitfed", [][]string{
		{"workspace", "sync"}, {"workspace", "sync", "--frozen"}, {"workspace", "status"},
	}, group("workspace-git"), note("--frozen must reproduce slice bytes from the lock"))
	inv("workspace/sync-twice", "gitfed", [][]string{{"workspace", "sync"}, {"workspace", "sync"}},
		group("workspace-git"), note("re-sync over a read-only nested slice"))
	inv("workspace/sync-only-known", "gitfed",
		[][]string{{"workspace", "sync", "--only", "testplat"}}, group("workspace-git"))
	inv("workspace/sync-only-unknown", "gitfed",
		[][]string{{"workspace", "sync", "--only", "nope"}}, group("workspace-git"))
	inv("workspace/sync-then-validate", "gitfed",
		[][]string{{"workspace", "sync"}, {"validate"}}, group("workspace-git"))
	inv("workspace/sync-then-tamper-then-validate", "gitfed",
		[][]string{{"workspace", "sync"}, {tamperStep}, {"validate"}},
		group("workspace-git"), note("gate [8/8] hash-integrity FAIL on a hand-edited slice"))
	inv("workspace/sync-bad-pin", "gitfed-badpin", [][]string{{"workspace", "sync"}},
		group("workspace-git"), note("short commit pin -> exit 4"))
	inv("workspace/sync-missing-ref", "gitfed-missingref", [][]string{{"workspace", "sync"}},
		group("workspace-git"), note("unreachable commit -> git failure, exit 6"),
		unstable("stderr: git relays the local upload-pack process's stderr and its own on "+
			"separate pipes, so the two `not our ref` lines interleave in a non-deterministic "+
			"ORDER between runs of ONE binary. Exit code and file tree are still compared; the "+
			"wrapping `error: `git ...` failed (exit 128)` line is covered deterministically by "+
			"workspace/sync-bad-pin."))
}

// --- top-level / usage ------------------------------------------------------

func buildUsage() {
	inv("usage/no-args", fxW, [][]string{{}})
	inv("usage/unknown-subcommand", fxW, [][]string{{"frobnicate"}})
	inv("usage/help", fxW, [][]string{{"--help"}})
	inv("usage/validate-help", fxW, [][]string{{"validate", "--help"}})
	inv("usage/prd-help", fxW, [][]string{{"prd", "--help"}})
	inv("usage/bad-flag", fxW, [][]string{{"validate", "--nope"}})
	inv("usage/root-flag-nonexistent-dir", fxW,
		[][]string{{"--root", "/nonexistent/xyz", "validate"}},
		note("--root is passed through verbatim by the runner"))
}
