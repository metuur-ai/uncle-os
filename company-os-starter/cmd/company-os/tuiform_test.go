package main

// R-5.6..R-5.12: the derived preview and the two mutating forms.
//
// The load-bearing test in this file is TestPreviewRoundTripsThroughTheParser.
// It does not check that the preview and the execution AGREE on the cases
// somebody thought to write down — a test like that passes on the day it is
// written and says nothing on the day a flag is added. It checks a law:
//
//	parse(shellSplit(screenCommand(a))) == a
//
// for every *Args the spec table can describe. Preview is therefore a right
// inverse of the parser, over the whole domain. A preview that drops a flag,
// mis-quotes a title, reorders a positional past an argparse `--`, or is
// hand-written for one screen breaks the equality immediately, because the only
// way to satisfy it is to render exactly what the parser will read back.
//
// The corpus is generated FROM commandSpecs rather than listed, so a flag added
// tomorrow is covered tomorrow without anyone remembering to come here.

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/tui"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// hostileValues are what a free-text slot can really hold. A discovery title is
// typed by a human, so every one of these is reachable: the shell metacharacters
// are the reason the preview quotes at all, and the leading dashes are the
// reason it can move behind an argparse `--`.
var hostileValues = []string{
	"plain",
	"two words",
	"it's mine",
	`say "hi"`,
	`mixed "and" it's`,
	"-x",
	"--force",
	"--team=other",
	"$(rm -rf /)",
	"`whoami`",
	"a\\b",
	"tab\there",
	"new\nline",
	"ünïcodé — em dash",
	"*",
	"~/relative",
	"",
	"  leading and trailing  ",
	"--",
}

// argsFor builds one *Args directly from a spec, filling every declared slot.
// Building it from the spec rather than from an argv is deliberate: an argv
// containing a leading-dash positional cannot be written naively, and the law
// under test is about *Args values, not about how they were obtained.
func argsFor(spec cmdSpec, text string) *Args {
	a := &Args{Cmd: spec.name}
	for _, f := range spec.flags {
		if f.str != nil && f.def != "" {
			*f.str(a) = f.def
		}
	}
	for _, p := range spec.pos {
		if len(p.choices) > 0 {
			*p.dest(a) = p.choices[0]
			continue
		}
		*p.dest(a) = text
	}
	for _, f := range spec.flags {
		switch {
		case f.boolean != nil:
			*f.boolean(a) = true
		case len(f.choices) > 0:
			*f.str(a) = f.choices[len(f.choices)-1]
		default:
			*f.str(a) = text
		}
	}
	normalizeArgs(a)
	return a
}

func TestPreviewRoundTripsThroughTheParser(t *testing.T) {
	globals := []struct {
		root string
		json bool
	}{
		{"", false},
		{"/tmp/some workspace", false},
		{"", true},
		{"./ws", true},
	}

	checked := 0
	for _, spec := range commandSpecs {
		if len(spec.pos) == 0 && len(spec.flags) == 0 {
			// `validate` and `tui`: still worth one pass, with globals.
			for _, g := range globals {
				a := &Args{Cmd: spec.name, Root: g.root, JSON: g.json}
				assertRoundTrip(t, a)
				checked++
			}
			continue
		}
		for _, text := range hostileValues {
			for _, g := range globals {
				a := argsFor(spec, text)
				a.Root, a.JSON = g.root, g.json
				assertRoundTrip(t, a)
				checked++
			}
		}
	}
	// A generator that quietly produced nothing would make this file a comment.
	if checked < len(commandSpecs)*len(globals) {
		t.Fatalf("only %d invocations were checked; the corpus is broken", checked)
	}
}

// assertRoundTrip is the law itself.
func assertRoundTrip(t *testing.T, want *Args) {
	t.Helper()
	preview := screenCommand(want)
	if preview == "" {
		t.Fatalf("no preview for %+v", want)
	}
	tokens := shellSplit(preview)
	if len(tokens) == 0 || tokens[0] != "company-os" {
		t.Fatalf("preview %q does not start with the program name", preview)
	}
	got, err := parse(tokens[1:])
	if err != nil {
		t.Errorf("the previewed command does not parse.\n  preview: %s\n  argv:    %q\n  error:   %v",
			preview, tokens[1:], err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("preview and execution describe different invocations (R-5.7).\n"+
			"  preview:  %s\n  executed: %+v\n  re-parsed:%+v", preview, want, got)
	}
}

// TestPreviewRoundTripsFromRealArgv is the same law approached from the other
// side, so the generator above is not the only thing keeping it true: these
// start life as argv the way a user types it, go through the real parser, and
// must survive the trip back out and in again.
func TestPreviewRoundTripsFromRealArgv(t *testing.T) {
	for _, argv := range [][]string{
		{"validate"},
		{"--json", "validate"},
		{"--root", "/tmp/ws", "validate"},
		{"today"},
		{"today", "--role", "product-owner"},
		{"skills", "list"},
		{"ids", "list", "--prefix", "component://", "--team", "core"},
		{"governance", "explain", "svc"},
		{"governance", "resolve", "--team", "core"},
		{"discover", "new", "--team", "core", "Reduce onboarding drift"},
		{"discover", "validate", "--team", "core", "2026-reduce-onboarding-drift"},
		{"prd", "new", "--platform", "plat", "--title", "Queue alerts",
			"--components", "svc-a,svc-b", "--team", "core"},
		{"prd", "new", "--platform", "plat", "--from-discovery", "2026-x", "--team", "core"},
		{"prd", "complete", "--platform", "plat", "2026-queue-alerts", "--force"},
		{"check", "ready", "--team", "core", "--components", "svc"},
		{"deviation", "declare", "req://x#R1", "--team", "core", "--rationale", "we can't yet"},
		{"exception", "request", "req://x#R1", "--team", "core", "--component", "svc",
			"--expires", "2035-01-01", "--reason", "vendor"},
		{"add", "component", "svc-a", "--platform", "plat"},
		{"reality", "new", "svc-a", "--platform", "plat"},
		{"init", "--company", "Acme Inc.", "--team", "core", "--platform", "plat"},
		{"workspace", "sync", "--frozen", "--only", "docs"},
		{"scratchpad", "init", "--repo", "/tmp/a repo"},
		{"graph", "build"},
	} {
		a, err := parse(argv)
		if err != nil {
			t.Fatalf("fixture argv %q does not parse: %v", argv, err)
		}
		assertRoundTrip(t, a)
	}
}

// TestPreviewIsNotCached asserts there is no second copy of the command line to
// go stale: changing the executed *Args changes the preview, field by field,
// across every slot the spec declares. A screen that had memoised its label, or
// spelled one out beside the args, fails here on the first field it forgot.
func TestPreviewIsNotCached(t *testing.T) {
	covered := 0
	for _, spec := range commandSpecs {
		if !hasFreeText(spec) {
			// `today`, `graph`, `skills`: every slot is a fixed choice set, so
			// there is no free text to vary. Their derivation is covered by the
			// round-trip law instead.
			continue
		}
		covered++
		base := argsFor(spec, "before")
		want := screenCommand(base)
		mutated := argsFor(spec, "after")
		if got := screenCommand(mutated); got == want {
			t.Errorf("%s: changing every free-text argument did not change the "+
				"preview — it is not derived from the executed args (R-5.7): %q",
				spec.name, got)
		}
	}
	if covered == 0 {
		t.Fatal("no command has a free-text argument; the check is vacuous")
	}
}

// hasFreeText reports whether a spec has any slot a user types rather than picks.
func hasFreeText(spec cmdSpec) bool {
	for _, p := range spec.pos {
		if len(p.choices) == 0 {
			return true
		}
	}
	for _, f := range spec.flags {
		if f.boolean == nil && len(f.choices) == 0 {
			return true
		}
	}
	return false
}

// TestCommittedOutputCarriesThePreviewedCommand: what actually ran announces
// itself with the SAME derivation the preview used. The header is not compared
// to a literal — it is compared to Preview(), which is the point: one function,
// one value, two readers.
func TestCommittedOutputCarriesThePreviewedCommand(t *testing.T) {
	root := tuiWorkspace(t)
	ws := workspace.New(root)
	inv := newInvocation(ws, &Args{Cmd: "skills", Action: "list"})

	body, err := inv.Commit()
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if !strings.HasPrefix(body, "$ "+inv.Preview()+"\n") {
		t.Errorf("the executed output does not open with the previewed command.\n"+
			"  preview: %s\n  body:    %.120q", inv.Preview(), body)
	}
}

// ------------------------------------------------------------- the two forms

// TestMutatingScreensAreTheTwoR55Names. R-5.5 names `discover new` and `prd new`
// and forbids forms for `workspace sync` and `scratchpad init`. Both halves are
// asserted, because the forbidden half is the one that gets added by someone
// being helpful.
func TestMutatingScreensAreTheTwoR55Names(t *testing.T) {
	ws := workspace.New(tuiWorkspace(t))
	got := mutatingScreens(ws, "")
	if len(got) != 2 {
		t.Fatalf("%d mutating screens, R-5.5 ships two", len(got))
	}
	for i, want := range []string{
		"new discovery brief (writes)", "new PRD (writes)"} {
		if got[i].Title != want {
			t.Errorf("mutating screen %d is %q, want %q", i, got[i].Title, want)
		}
		if got[i].Form == nil {
			t.Fatalf("screen %q has no form", got[i].Title)
		}
	}

	// No screen in the whole catalog may reach a forbidden command, and no
	// browsing screen may reach `discover validate`, which rewrites the brief it
	// is asked about.
	for _, s := range screensFor(ws, "") {
		if s.Form == nil {
			continue
		}
		action, err := s.Form.Build(sampleValues(s.Form))
		if err != nil {
			t.Fatalf("%s: build: %v", s.Title, err)
		}
		for _, forbidden := range []string{
			"workspace sync", "scratchpad init", "discover validate"} {
			if strings.Contains(action.Preview(), forbidden) {
				t.Errorf("%s previews %q — R-5.5 forbids a form for it", s.Title, forbidden)
			}
		}
	}
}

// TestEveryFormFieldHasAFlagEquivalent is R-5.10, checked rather than asserted:
// each field is filled with a value unique to it, and the previewed invocation
// is parsed back. A field whose value does not survive that trip is a value the
// TUI can collect and the CLI cannot reproduce — which is the whole property
// R-5.10 exists for.
func TestEveryFormFieldHasAFlagEquivalent(t *testing.T) {
	ws := workspace.New(tuiWorkspace(t))
	for _, s := range screensFor(ws, "") {
		if s.Form == nil {
			continue
		}
		values := sampleValues(s.Form)
		action, err := s.Form.Build(values)
		if err != nil {
			t.Fatalf("%s: build: %v", s.Title, err)
		}
		preview := action.Preview()
		tokens := shellSplit(preview)
		back, err := parse(tokens[1:])
		if err != nil {
			t.Fatalf("%s: the previewed command does not parse: %s (%v)", s.Title, preview, err)
		}
		for i, f := range s.Form.Fields {
			if values[i] == "" {
				continue
			}
			if !strings.Contains(preview, values[i]) {
				t.Errorf("%s: field %q does not appear in the invocation (R-5.10): %s",
					s.Title, f.Label, preview)
			}
			if !argsCarry(back, values[i]) {
				t.Errorf("%s: field %q survives the preview but not the parse "+
					"(R-5.10): %s", s.Title, f.Label, preview)
			}
		}
	}
}

// TestFormPickersOfferOnlyValuesThatExist. A picker that offers a team the
// workspace does not have offers a command that is certain to fail, which is
// worse than a text box.
func TestFormPickersOfferOnlyValuesThatExist(t *testing.T) {
	root := tuiWorkspace(t)
	ws := workspace.New(root)
	teams := map[string]bool{}
	for _, n := range baseNames(ws.AllTeams()) {
		teams[n] = true
	}
	platforms := map[string]bool{}
	for _, n := range baseNames(ws.AllPlatforms()) {
		platforms[n] = true
	}
	for _, s := range mutatingScreens(ws, "") {
		for _, f := range s.Form.Fields {
			for _, c := range f.Choices {
				switch f.Label {
				case "team":
					if !teams[c] {
						t.Errorf("%s offers team %q, which does not exist", s.Title, c)
					}
				case "platform":
					if !platforms[c] {
						t.Errorf("%s offers platform %q, which does not exist", s.Title, c)
					}
				}
			}
		}
	}
}

// ---------------------------------------------- R-5.8 / R-5.9 on a real tree

// TestCancelledFormLeavesTheWorkspaceExactlyAsItWas is R-5.8 and the first half
// of R-5.9, proven on a REAL workspace through the REAL catalog: the form is
// filled in completely, the preview is reached, and then each of the four ways
// out is taken. The tree is hashed around every one of them.
//
// This is the assertion that a form cannot be written to disk "just to be
// helpful" — a draft file, a recent-values cache, a scaffolded directory — since
// any of those would move the digest.
func TestCancelledFormLeavesTheWorkspaceExactlyAsItWas(t *testing.T) {
	root := tuiWorkspace(t)
	ws := workspace.New(root)

	for _, cancel := range []string{"q", "esc", "ctrl+c", "n"} {
		before := treeDigest(t, root)
		m := formModel(t, ws, "new discovery brief (writes)")
		m = fillForm(t, m, []string{"", "Reduce onboarding drift"})
		if m.Mode() != tui.ModeConfirm {
			t.Fatalf("%s: expected the confirmation, got %v", cancel, m.Mode())
		}
		// The preview must be on screen BEFORE anything can run (R-5.6).
		if !strings.Contains(m.View(), "company-os discover new") {
			t.Errorf("%s: the confirmation does not show the invocation:\n%s", cancel, m.View())
		}
		if after := treeDigest(t, root); after != before {
			t.Fatalf("%s: reaching the preview already changed the workspace (R-5.8):\n%s",
				cancel, diffTrees(t, before, after))
		}
		m, _ = tuiKey(t, m, cancel)
		if after := treeDigest(t, root); after != before {
			t.Errorf("%s cancelled the form but changed the workspace (R-5.9):\n%s",
				cancel, diffTrees(t, before, after))
		}
	}
}

// TestConfirmingRunsTheCommandThroughTheSameCodePath is R-5.11 and R-5.12: the
// confirmed form produces exactly what the flag CLI produces from the previewed
// argv, because it IS the flag CLI's code path — the same `commands` entry, the
// same renderer, in-process.
//
// The comparison is against a second workspace driven by run() with the argv the
// preview printed, so what is compared is the whole created tree, not a message.
func TestConfirmingRunsTheCommandThroughTheSameCodePath(t *testing.T) {
	viaTUI := tuiWorkspace(t)
	viaCLI := tuiWorkspace(t)

	m := formModel(t, workspace.New(viaTUI), "new discovery brief (writes)")
	m = fillForm(t, m, []string{"", "Reduce onboarding drift"})
	preview := previewLine(t, m)
	m, _ = tuiKey(t, m, "y")
	if m.Mode() != tui.ModeBody {
		t.Fatalf("confirming did not run the command, mode %v", m.Mode())
	}

	argv := shellSplit(preview)[1:]
	var out, errOut strings.Builder
	if code := run(append([]string{"--root", viaCLI}, argv...), &out, &errOut); code != 0 {
		t.Fatalf("the previewed command failed under the flag CLI (%d): %s", code, errOut.String())
	}

	if a, b := treeDigest(t, viaTUI), treeDigest(t, viaCLI); a != b {
		t.Errorf("the TUI and the previewed command produced different workspaces "+
			"(R-5.12)\n--- tui ---\n%s\n--- cli ---\n%s", a, b)
	}
	if !strings.Contains(m.View(), "brief.md") {
		t.Errorf("the result screen does not name what was created:\n%s", m.View())
	}
}

// TestAFormWritesNothingUntilConfirmed walks every field of every form and
// hashes the tree after every keystroke. R-5.8 is a statement about the whole
// interaction, not only about its last step.
func TestAFormWritesNothingUntilConfirmed(t *testing.T) {
	root := tuiWorkspace(t)
	ws := workspace.New(root)
	before := treeDigest(t, root)

	for _, s := range screensFor(ws, "") {
		if s.Form == nil {
			continue
		}
		m := formModel(t, ws, s.Title)
		for _, k := range []string{
			"down", "right", "right", "up", "left", "a", "b", " ", "c",
			"backspace", "tab", "enter", "enter", "enter", "enter", "enter",
		} {
			m, _ = tuiKey(t, m, k)
			if m.Mode() == tui.ModeBody {
				t.Fatalf("%s: %q ran the command without a confirmation (R-5.8)", s.Title, k)
			}
			if after := treeDigest(t, root); after != before {
				t.Fatalf("%s: %q changed the workspace before confirmation (R-5.8):\n%s",
					s.Title, k, diffTrees(t, before, after))
			}
		}
	}
}

// ------------------------------------------------------------------- helpers

// sampleValues fills a form the way a user would: the first offered choice for a
// picker, and a value unique to the field for free text.
func sampleValues(f *tui.Form) []string {
	out := make([]string, len(f.Fields))
	for i, fld := range f.Fields {
		if len(fld.Choices) > 0 {
			out[i] = fld.Choices[0]
			continue
		}
		out[i] = "field" + strings.ToUpper(fld.Label[:1]) + fld.Label[1:] + "Value"
	}
	return out
}

// argsCarry reports whether any string field of an *Args holds v. It is how
// "the value survived the parse" is asserted without naming which field each
// form chose — the mapping is the form's business; the survival is R-5.10's.
func argsCarry(a *Args, v string) bool {
	rv := reflect.ValueOf(*a)
	for i := 0; i < rv.NumField(); i++ {
		if f := rv.Field(i); f.Kind() == reflect.String && f.String() == v {
			return true
		}
	}
	return false
}

// formModel opens one named screen's form on a fresh model.
func formModel(t *testing.T, ws *workspace.Workspace, title string) tui.Model {
	t.Helper()
	screens := screensFor(ws, "")
	at := -1
	for i, s := range screens {
		if s.Title == title {
			at = i
			break
		}
	}
	if at < 0 {
		t.Fatalf("no screen titled %q", title)
	}
	m := tui.New(screens, tui.Options{Output: &strings.Builder{}, NoColor: true})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = next.(tui.Model)
	for i := 0; i < at; i++ {
		m, _ = tuiKey(t, m, "down")
	}
	m, _ = tuiKey(t, m, "enter")
	if m.Mode() != tui.ModeForm {
		t.Fatalf("%q did not open a form, mode %v", title, m.Mode())
	}
	return m
}

// fillForm types one value per field, leaving pickers on their default when the
// corresponding entry is empty, then submits.
func fillForm(t *testing.T, m tui.Model, values []string) tui.Model {
	t.Helper()
	for i, v := range values {
		if i > 0 {
			m, _ = tuiKey(t, m, "down")
		}
		for _, r := range v {
			m, _ = tuiKey(t, m, string(r))
		}
	}
	m, _ = tuiKey(t, m, "enter")
	return m
}

// previewLine pulls the invocation out of the confirmation frame. It reads the
// rendered UI on purpose: R-5.6 is about what the reader is shown before the
// write, so the string under test has to be the one on screen.
func previewLine(t *testing.T, m tui.Model) string {
	t.Helper()
	for _, line := range strings.Split(m.View(), "\n") {
		if trimmed := strings.TrimSpace(line); strings.HasPrefix(trimmed, "$ company-os") {
			return strings.TrimPrefix(trimmed, "$ ")
		}
	}
	t.Fatalf("no previewed command in the confirmation frame:\n%s", m.View())
	return ""
}

func tuiKey(t *testing.T, m tui.Model, k string) (tui.Model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(tuiKeyMsg(k))
	return next.(tui.Model), cmd
}

func tuiKeyMsg(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "backspace":
		return tea.KeyMsg{Type: tea.KeyBackspace}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
}

// shellSplit is the inverse of shellQuote, and it is TEST-ONLY on purpose: the
// binary never needs to read a command line back, and a splitter in production
// would be a second, unexercised parser.
//
// It implements the sliver of POSIX word splitting that shellQuote can emit —
// blanks separate words, single quotes suppress everything, and a backslash
// outside quotes escapes one character, which is how the standard `'\”`
// sequence for an embedded quote is spelled.
func shellSplit(s string) []string {
	var out []string
	var cur strings.Builder
	inWord, inQuote := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQuote:
			if c == '\'' {
				inQuote = false
				continue
			}
			cur.WriteByte(c)
		case c == '\'':
			inQuote, inWord = true, true
		case c == '\\':
			if i+1 < len(s) {
				i++
				cur.WriteByte(s[i])
				inWord = true
			}
		case c == ' ' || c == '\t' || c == '\n':
			if inWord {
				out = append(out, cur.String())
				cur.Reset()
				inWord = false
			}
		default:
			cur.WriteByte(c)
			inWord = true
		}
	}
	if inWord {
		out = append(out, cur.String())
	}
	return out
}

// TestShellSplitInvertsShellQuote guards the guard: the round-trip law is only
// worth anything if the splitter it runs through is a real inverse of the
// quoter, so that is asserted directly, on the same hostile corpus.
func TestShellSplitInvertsShellQuote(t *testing.T) {
	for _, v := range hostileValues {
		got := shellSplit(shellQuote(v))
		if v == "" {
			if len(got) != 1 || got[0] != "" {
				t.Errorf("quoting the empty string does not survive splitting: %q", got)
			}
			continue
		}
		if len(got) != 1 || got[0] != v {
			t.Errorf("shellSplit(shellQuote(%q)) = %q, want one token %q",
				v, got, v)
		}
	}
}
