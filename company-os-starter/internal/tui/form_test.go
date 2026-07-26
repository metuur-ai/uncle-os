package tui_test

// The mutating half of the UI (R-5.5..R-5.12), driven through Update and read
// through View, like the rest of the package.
//
// The test that matters most here is TestTheConfirmedActionIsTheOneThatWasShown.
// The round-trip law in cmd/company-os proves that the previewed TEXT is the
// text the parser reads back; this one proves the other half — that the VALUE
// whose Preview() reached the screen is the same value whose Commit() ran. Both
// together are what make "preview and execution cannot diverge" a property of
// the code rather than a claim about it.

import (
	"bytes"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/tui"
)

// spyAction records what happened to it. Each one carries a distinct preview
// string, which is how a test can tell WHICH action reached the screen and which
// one ran.
type spyAction struct {
	preview  string
	previews int
	commits  int
	body     string
	err      error
}

func (s *spyAction) Preview() string {
	s.previews++
	return s.preview
}

func (s *spyAction) Commit() (string, error) {
	s.commits++
	return s.body, s.err
}

// formCatalog is one picker field and one free-text field — the two shapes a
// form can hold — plus a builder that hands back a fresh spy per submission.
func formCatalog(built *[]*spyAction) []tui.Screen {
	return []tui.Screen{
		{
			Title: "new discovery brief (writes)",
			Form: &tui.Form{
				Fields: []tui.Field{
					{Label: "team", Choices: []string{"core", "payments"},
						Help: "the owning team"},
					{Label: "title", Help: "free text"},
				},
				Build: func(v []string) (tui.Action, error) {
					s := &spyAction{
						preview: "PREVIEW#" + string(rune('A'+len(*built))) +
							" team=" + v[0] + " title=" + v[1],
						body: "created something",
					}
					*built = append(*built, s)
					return s, nil
				},
			},
		},
	}
}

func formModel(t *testing.T, built *[]*spyAction) tui.Model {
	t.Helper()
	m := tui.New(formCatalog(built), tui.Options{Output: &bytes.Buffer{}, NoColor: true})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m, _ = key(t, next.(tui.Model), "enter")
	if m.Mode() != tui.ModeForm {
		t.Fatalf("the screen did not open a form, mode %v", m.Mode())
	}
	return m
}

// type types a string into the focused free-text field.
func typeText(t *testing.T, m tui.Model, s string) tui.Model {
	t.Helper()
	for _, r := range s {
		if r == ' ' {
			m, _ = key(t, m, " ")
			continue
		}
		m, _ = key(t, m, string(r))
	}
	return m
}

// reachConfirm fills the form and submits.
func reachConfirm(t *testing.T, m tui.Model, title string) tui.Model {
	t.Helper()
	m, _ = key(t, m, "down")
	m = typeText(t, m, title)
	m, _ = key(t, m, "enter")
	if m.Mode() != tui.ModeConfirm {
		t.Fatalf("the form did not reach the confirmation, mode %v:\n%s", m.Mode(), m.View())
	}
	return m
}

// TestNothingRunsBeforeConfirmation is R-5.8 at this package's seam: reaching
// the preview builds the action but never calls Commit.
func TestNothingRunsBeforeConfirmation(t *testing.T) {
	var built []*spyAction
	m := reachConfirm(t, formModel(t, &built), "a title")
	if len(built) != 1 {
		t.Fatalf("%d actions built, want 1", len(built))
	}
	if built[0].commits != 0 {
		t.Errorf("the action ran before any confirmation (R-5.8)")
	}
	if !strings.Contains(m.View(), built[0].preview) {
		t.Errorf("the confirmation does not show the invocation (R-5.6):\n%s", m.View())
	}
	// The string on screen came OUT of the action, rather than being kept
	// beside it: a frame that never asked the action what it was is a frame
	// showing a copy, and a copy is what goes stale (R-5.7).
	if built[0].previews == 0 {
		t.Error("the confirmation rendered without asking the action for its preview")
	}
	m, _ = key(t, m, "y")
	if built[0].commits != 1 {
		t.Errorf("confirming did not run the action")
	}
	if m.Mode() != tui.ModeBody {
		t.Errorf("after running, mode is %v, want body", m.Mode())
	}
	if !strings.Contains(m.View(), "created something") {
		t.Errorf("the result is not shown:\n%s", m.View())
	}
}

// TestTheConfirmedActionIsTheOneThatWasShown is the identity half of R-5.7.
//
// The form is submitted, sent back, edited, and submitted again, so TWO actions
// exist and their previews differ. The assertion is not that the strings match —
// it is that the object whose preview is on screen is the object that runs, and
// that the other one is never touched. A design that rebuilt the action at
// confirmation time, or kept a preview string beside a separately built command,
// fails here.
func TestTheConfirmedActionIsTheOneThatWasShown(t *testing.T) {
	var built []*spyAction
	m := reachConfirm(t, formModel(t, &built), "first")

	m, _ = key(t, m, "n") // back to the form
	if m.Mode() != tui.ModeForm {
		t.Fatalf("n did not return to the form, mode %v", m.Mode())
	}
	m = typeText(t, m, " revised")
	m, _ = key(t, m, "enter")
	if m.Mode() != tui.ModeConfirm {
		t.Fatalf("resubmitting did not reach the confirmation, mode %v", m.Mode())
	}
	if len(built) != 2 {
		t.Fatalf("%d actions built, want 2", len(built))
	}
	if built[0].preview == built[1].preview {
		t.Fatal("the two submissions produced the same preview; the test proves nothing")
	}

	shown := m.View()
	if !strings.Contains(shown, built[1].preview) {
		t.Fatalf("the second preview is not on screen:\n%s", shown)
	}
	if strings.Contains(shown, built[0].preview) {
		t.Fatalf("the stale preview is still on screen:\n%s", shown)
	}

	m, _ = key(t, m, "y")
	if built[0].commits != 0 {
		t.Error("the action that was NOT previewed ran (R-5.7)")
	}
	if built[1].commits != 1 {
		t.Errorf("the previewed action ran %d times, want 1", built[1].commits)
	}
}

// TestCancellingRunsNothing is R-5.9's first half at this seam. Every way out of
// the confirmation that is not `y` must leave Commit uncalled.
func TestCancellingRunsNothing(t *testing.T) {
	for _, k := range []string{"q", "esc", "ctrl+c", "n", "backspace", "left"} {
		var built []*spyAction
		m := reachConfirm(t, formModel(t, &built), "a title")
		m, _ = key(t, m, k)
		if built[0].commits != 0 {
			t.Errorf("%q ran the action instead of cancelling it (R-5.9)", k)
		}
		if k == "q" || k == "esc" || k == "ctrl+c" {
			if m.View() != "" {
				t.Errorf("%q left a frame behind: %q", k, m.View())
			}
		} else if m.Mode() != tui.ModeForm {
			t.Errorf("%q from the confirmation went to %v, want the form", k, m.Mode())
		}
	}
}

// TestExitKeysFromFormAndConfirmation extends R-5.14 to the two new modes, and
// pins the ONE carve-out: `q` is a character while a free-text field has focus,
// and esc / ctrl-c still are not.
func TestExitKeysFromFormAndConfirmation(t *testing.T) {
	var built []*spyAction
	picker := formModel(t, &built)           // field 0: the team picker
	text, _ := key(t, picker, "down")        // field 1: free text
	confirm := reachConfirm(t, picker, "ok") // the confirmation

	for _, where := range []struct {
		name string
		m    tui.Model
		keys []string
	}{
		{"form/picker", picker, []string{"q", "esc", "ctrl+c"}},
		{"form/text", text, []string{"esc", "ctrl+c"}},
		{"confirm", confirm, []string{"q", "esc", "ctrl+c"}},
	} {
		for _, k := range where.keys {
			after, cmd := key(t, where.m, k)
			if !isQuit(cmd) {
				t.Errorf("%s: %q did not quit (R-5.14)", where.name, k)
			}
			if got := after.View(); got != "" {
				t.Errorf("%s: %q left a frame behind: %q", where.name, k, got)
			}
		}
	}

	// The carve-out, asserted rather than assumed: in a text field q types.
	typed, cmd := key(t, text, "q")
	if isQuit(cmd) {
		t.Fatal("q quit from a text field, so a title cannot contain the letter q")
	}
	if !strings.Contains(typed.View(), "q") {
		t.Errorf("q was neither an exit nor a character:\n%s", typed.View())
	}
	// And the footer must say which keys DO leave, or the carve-out is a trap.
	view := text.View()
	if !strings.Contains(view, "esc") || !strings.Contains(view, "ctrl-c") {
		t.Errorf("the text field's footer does not name a way out:\n%s", view)
	}
}

// TestRequiredFieldsBlockThePreview. A form that reaches the confirmation with a
// required value missing shows a command line that cannot run, which is worse
// than no preview at all.
func TestRequiredFieldsBlockThePreview(t *testing.T) {
	var built []*spyAction
	m := formModel(t, &built)
	m, _ = key(t, m, "down") // the title, left empty
	m, _ = key(t, m, "enter")
	if m.Mode() != tui.ModeForm {
		t.Fatalf("an empty required field reached %v", m.Mode())
	}
	if len(built) != 0 {
		t.Error("an action was built for an incomplete form")
	}
	if !strings.Contains(m.View(), "title is required") {
		t.Errorf("the form does not name the missing field:\n%s", m.View())
	}
}

// TestBuildFailureStaysInTheForm. Build is the only place the catalog can refuse
// a set of answers; refusing must not strand the reader on a blank screen.
func TestBuildFailureStaysInTheForm(t *testing.T) {
	cat := []tui.Screen{{
		Title: "refuses",
		Form: &tui.Form{
			Fields: []tui.Field{{Label: "x"}},
			Build: func([]string) (tui.Action, error) {
				return nil, errors.New("that combination is not allowed")
			},
		},
	}}
	m := tui.New(cat, tui.Options{Output: &bytes.Buffer{}, NoColor: true})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	at, _ := key(t, next.(tui.Model), "enter")
	at = typeText(t, at, "v")
	at, _ = key(t, at, "enter")
	if at.Mode() != tui.ModeForm {
		t.Fatalf("a refused build left the form, mode %v", at.Mode())
	}
	if !strings.Contains(at.View(), "not allowed") {
		t.Errorf("the refusal is not shown:\n%s", at.View())
	}
}

// TestPickerCyclesAndOnlyOffersItsChoices. A picker must never hold a value the
// CLI would reject, so no keystroke may put text into one.
func TestPickerCyclesAndOnlyOffersItsChoices(t *testing.T) {
	var built []*spyAction
	m := formModel(t, &built)
	m, _ = key(t, m, "z") // must not be typed into the picker
	m, _ = key(t, m, "right")
	m = reachConfirm(t, m, "t")
	if got := built[0].preview; !strings.Contains(got, "team=payments") {
		t.Errorf("right did not cycle the picker: %q", got)
	}
	if strings.Contains(built[0].preview, "z") {
		t.Errorf("a keystroke was typed into a picker: %q", built[0].preview)
	}

	// An optional picker must be able to hold nothing.
	cat := []tui.Screen{{
		Title: "opt",
		Form: &tui.Form{
			Fields: []tui.Field{{Label: "maybe", Choices: []string{"a", "b"}, Optional: true}},
			Build: func(v []string) (tui.Action, error) {
				return &spyAction{preview: "value=[" + v[0] + "]"}, nil
			},
		},
	}}
	om := tui.New(cat, tui.Options{Output: &bytes.Buffer{}, NoColor: true})
	next, _ := om.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	at, _ := key(t, next.(tui.Model), "enter")
	at, _ = key(t, at, "enter")
	if !strings.Contains(at.View(), "value=[]") {
		t.Errorf("an optional picker did not start unset:\n%s", at.View())
	}
}

// TestFormsDegradeLegibly extends R-5.15 to the new modes: no escape byte with
// NO_COLOR, and no line wider than the terminal at sixty columns.
func TestFormsDegradeLegibly(t *testing.T) {
	for _, width := range []int{60, 40, 100} {
		var built []*spyAction
		m := tui.New(formCatalog(&built),
			tui.Options{Output: &bytes.Buffer{}, NoColor: true})
		next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: 20})
		at, _ := key(t, next.(tui.Model), "enter")
		frames := []string{at.View()}
		at = reachConfirm(t, at,
			strings.Repeat("a very long discovery title that will not fit ", 3))
		frames = append(frames, at.View())
		at, _ = key(t, at, "y")
		frames = append(frames, at.View())

		for i, f := range frames {
			if j := strings.IndexByte(f, 0x1b); j >= 0 {
				t.Errorf("width %d frame %d carries an escape at byte %d", width, i, j)
			}
			for _, line := range strings.Split(f, "\n") {
				if n := len([]rune(line)); n > width {
					t.Errorf("at %d columns a line is %d runes wide: %q", width, n, line)
				}
			}
		}
	}
}

// TestFormValuesDoNotLeakAcrossModels. Update takes the model by value, so two
// models made from the same parent share the answer slice's backing array unless
// it is copied on write. Without the copy, editing a form in one branch silently
// edits it in a model someone else still holds — including the earlier model the
// exit-key tests keep around.
func TestFormValuesDoNotLeakAcrossModels(t *testing.T) {
	var built []*spyAction
	base := formModel(t, &built)
	base, _ = key(t, base, "down")

	branchA := typeText(t, base, "alpha")
	branchB := typeText(t, base, "beta")

	if !strings.Contains(branchA.View(), "alpha") {
		t.Errorf("branch A lost its own value:\n%s", branchA.View())
	}
	if strings.Contains(branchA.View(), "beta") {
		t.Errorf("branch B's edit leaked into branch A:\n%s", branchA.View())
	}
	if strings.Contains(branchB.View(), "alpha") {
		t.Errorf("branch A's edit leaked into branch B:\n%s", branchB.View())
	}
}

// TestBackspaceEditsTextAndLeavesPickers. Backspace is the one key whose meaning
// depends on the field, so both meanings are pinned.
func TestBackspaceEditsTextAndLeavesPickers(t *testing.T) {
	var built []*spyAction
	m := formModel(t, &built)
	// On the picker, backspace goes back to the menu.
	back, _ := key(t, m, "backspace")
	if back.Mode() != tui.ModeMenu {
		t.Errorf("backspace on a picker went to %v, want the menu", back.Mode())
	}
	// On a text field it deletes a rune.
	at, _ := key(t, m, "down")
	at = typeText(t, at, "abc")
	at, _ = key(t, at, "backspace")
	if !strings.Contains(at.View(), "ab") || strings.Contains(at.View(), "abc") {
		t.Errorf("backspace did not delete one character:\n%s", at.View())
	}
	// Deleting past the start is not an error.
	for i := 0; i < 5; i++ {
		at, _ = key(t, at, "backspace")
	}
	if at.Mode() != tui.ModeForm {
		t.Errorf("over-deleting left the form, mode %v", at.Mode())
	}
}

// TestThisPackageCannotComposeAnInvocation is the structural half of R-5.7, in
// the same spirit as cmd/company-os/ansi_test.go: it reads this package's own
// source with go/ast and fails on any string literal that looks like a
// company-os command line.
//
// The point is not that today's code happens not to contain one. It is that a
// preview written HERE could never be derived from the args that execute — this
// package has no access to them — so the only preview it could produce is a
// hand-written one, which is exactly what R-5.7 forbids. Making that
// unrepresentable is cheaper than reviewing for it.
//
// The bare program name is allowed: it is the menu's own title. What is not
// allowed is the program name followed by anything, because that is a sentence
// about a command rather than a label.
func TestThisPackageCannotComposeAnInvocation(t *testing.T) {
	const prog = "company-os"
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("reading the package directory: %v", err)
	}
	scanned := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, perr := parser.ParseFile(fset, name, nil, 0)
		if perr != nil {
			t.Fatalf("parsing %s: %v", name, perr)
		}
		scanned++
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			v, uerr := strconv.Unquote(lit.Value)
			if uerr != nil {
				v = lit.Value
			}
			if at := strings.Index(v, prog); at >= 0 && len(v) > at+len(prog) {
				t.Errorf("%s: string literal %s composes a command line — R-5.7 "+
					"requires the preview to be DERIVED by cmd/company-os from "+
					"the args it executes, and nothing here can see those",
					fset.Position(lit.Pos()), lit.Value)
			}
			return true
		})
	}
	if scanned == 0 {
		t.Fatal("scanned no source files; the guard is broken, not clean")
	}
}
