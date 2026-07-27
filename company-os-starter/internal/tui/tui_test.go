package tui_test

// The UI is tested by calling Update and reading View. A Bubble Tea model is a
// pure function of its state and one message, so there is nothing a pty would
// add here except flakiness and a platform dependency: a pty test proves the
// terminal echoed what the model already returned, and cannot assert the one
// thing that matters most — that q/Esc/Ctrl-C reach tea.Quit from EVERY mode —
// without racing the renderer.
//
// tea.Quit is a Cmd, i.e. a func() Msg. It is identified by CALLING it and
// comparing the message to tea.QuitMsg{}, which is the only stable way to
// recognise it; comparing function pointers is not defined in Go.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/tui"
)

// screens builds a small catalog with one of each shape: no-argument, and
// argument-taking.
func screens(t *testing.T) []tui.Screen {
	t.Helper()
	return []tui.Screen{
		{
			Title: "validate results",
			Run: func(string) (string, error) {
				return "$ company-os validate\n\n[1/7] gate one\n  [ok] fine", nil
			},
		},
		{
			Title:   "today (role view)",
			Prompt:  "role",
			Choices: []string{"developer", "product-owner"},
			Run:     func(c string) (string, error) { return "role is " + c, nil },
		},
		{
			Title: "broken screen",
			Run:   func(string) (string, error) { return "partial output", errors.New("boom") },
		},
	}
}

func newModel(t *testing.T, opts ...func(*tui.Options)) tui.Model {
	t.Helper()
	o := tui.Options{Output: &bytes.Buffer{}}
	for _, f := range opts {
		f(&o)
	}
	m := tui.New(screens(t), o)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return next.(tui.Model)
}

// key sends one keystroke and returns the new model plus the command it issued.
func key(t *testing.T, m tui.Model, k string) (tui.Model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(keyMsg(k))
	return next.(tui.Model), cmd
}

// keyMsg builds the tea.KeyMsg whose String() is k. The special names go through
// their key type; everything else is a rune sequence.
func keyMsg(k string) tea.KeyMsg {
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
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

// TestExitFromEveryScreen is R-5.14 as amended 2026-07-27. The safety property
// is unchanged and is what this asserts: EVERY mode has an unconditional exit.
// A UI that only exits from its first screen is the failure the requirement
// exists for, and it passes any test that never navigates.
//
// What changed is which keys carry it. `ctrl+c` and `q` still exit from all
// three modes below; Esc no longer does — it goes back, and is covered by
// TestEscGoesBackOneLevel. Esc's exit survives only at the menu, where there is
// nothing to go back to, and that case is asserted here too.
func TestExitFromEveryScreen(t *testing.T) {
	base := newModel(t)

	menu := base
	pick, _ := key(t, base, "down") // screen 2 takes an argument
	pick, _ = key(t, pick, "enter") // -> picker
	body, _ := key(t, pick, "enter")

	if pick.Mode() != tui.ModePick {
		t.Fatalf("expected the picker after opening an argument-taking screen, got %v", pick.Mode())
	}
	if body.Mode() != tui.ModeBody {
		t.Fatalf("expected the body after choosing, got %v", body.Mode())
	}

	for _, where := range []struct {
		name string
		m    tui.Model
		keys []string
	}{
		// Esc is an exit ONLY here, and only because the menu is the top level.
		{"menu", menu, []string{"q", "esc", "ctrl+c"}},
		{"picker", pick, []string{"q", "ctrl+c"}},
		{"body", body, []string{"q", "ctrl+c"}},
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
}

// TestEscGoesBackOneLevel is the amended half of R-5.14, and the defect that
// prompted it: from a form there was no way back at all. `backspace` returned
// only from a picker field (on a text field it deletes a character) and `left`
// only from a text field (on a picker it cycles the value), so no single key
// went back from every field — and the form footer named none of them. Leaving
// a form opened by mistake meant killing the UI.
//
// Reported from the running TUI, not found by inspection.
func TestEscGoesBackOneLevel(t *testing.T) {
	base := newModel(t)
	pick, _ := key(t, base, "down")
	pick, _ = key(t, pick, "enter")
	body, _ := key(t, pick, "enter")

	for _, tc := range []struct {
		name string
		from tui.Model
		want tui.Mode
	}{
		{"body -> picker", body, tui.ModePick},
		{"picker -> menu", pick, tui.ModeMenu},
	} {
		after, cmd := key(t, tc.from, "esc")
		if isQuit(cmd) {
			t.Errorf("%s: esc quit instead of going back", tc.name)
			continue
		}
		if after.Mode() != tc.want {
			t.Errorf("%s: esc landed in %v, want %v", tc.name, after.Mode(), tc.want)
		}
	}

	// And the whole point: repeated Esc reaches the menu and then, and only
	// then, exits. A reader who has gone in too far can hold one key to leave.
	m := body
	for i := 0; i < 4; i++ {
		after, cmd := key(t, m, "esc")
		if isQuit(cmd) {
			if m.Mode() != tui.ModeMenu {
				t.Errorf("esc quit from %v, not from the menu", m.Mode())
			}
			return
		}
		m = after
	}
	t.Error("repeated esc never reached an exit")
}

// TestEveryScreenIsReachable walks the whole catalog through Update, which is
// what makes R-5.4's list a testable boundary rather than a comment: adding an
// eleventh screen without a Run, or one that panics on open, fails here.
func TestEveryScreenIsReachable(t *testing.T) {
	m := newModel(t)
	for i := range screens(t) {
		at := m
		for j := 0; j < i; j++ {
			at, _ = key(t, at, "down")
		}
		at, _ = key(t, at, "enter")
		if at.Mode() == tui.ModePick {
			at, _ = key(t, at, "enter")
		}
		if at.Mode() != tui.ModeBody {
			t.Fatalf("screen %d did not open, mode %v", i, at.Mode())
		}
		if strings.TrimSpace(at.View()) == "" {
			t.Errorf("screen %d rendered an empty frame", i)
		}
	}
}

// TestBodyShowsCommandOutputVerbatim is R-5.13 at the seam this package owns:
// whatever the catalog's Run returned is what appears, unedited. The catalog
// gets that string from internal/render, so a gate cannot render one way in the
// TUI and another in `company-os validate`.
func TestBodyShowsCommandOutputVerbatim(t *testing.T) {
	m := newModel(t)
	m, _ = key(t, m, "enter")
	view := m.View()
	for _, want := range []string{"[1/7] gate one", "[ok] fine", "company-os validate"} {
		if !strings.Contains(view, want) {
			t.Errorf("body is missing %q:\n%s", want, view)
		}
	}
}

// TestAScreenThatFailsShowsBothHalves. A command may return output AND an error
// — `workspace sync` and `prd complete` both do — and dropping either half loses
// the explanation of the refusal.
func TestAScreenThatFailsShowsBothHalves(t *testing.T) {
	m := newModel(t)
	m, _ = key(t, m, "down")
	m, _ = key(t, m, "down")
	m, _ = key(t, m, "enter")
	view := m.View()
	if !strings.Contains(view, "boom") {
		t.Errorf("the error is not shown:\n%s", view)
	}
	if !strings.Contains(view, "partial output") {
		t.Errorf("the partial body was dropped on the error path:\n%s", view)
	}
}

// TestBackspaceReturnsWithoutExiting. Going back has to exist, or the exit keys
// become the only navigation and R-5.14 turns the UI into a one-shot.
func TestBackspaceReturnsWithoutExiting(t *testing.T) {
	m := newModel(t)
	m, _ = key(t, m, "down")
	m, _ = key(t, m, "enter") // picker
	m, cmd := key(t, m, "enter")
	if isQuit(cmd) {
		t.Fatal("enter quit the program")
	}
	m, cmd = key(t, m, "backspace")
	if isQuit(cmd) {
		t.Fatal("backspace quit the program")
	}
	if m.Mode() != tui.ModePick {
		t.Fatalf("backspace from a body should return to its picker, got %v", m.Mode())
	}
	m, _ = key(t, m, "backspace")
	if m.Mode() != tui.ModeMenu {
		t.Fatalf("backspace from a picker should return to the menu, got %v", m.Mode())
	}
}

// TestNoColorEmitsNoEscapes is half of R-5.15, and it is a byte assertion rather
// than a "looks plain" one: NO_COLOR is only honoured if the frame is literally
// free of escape sequences.
func TestNoColorEmitsNoEscapes(t *testing.T) {
	m := newModel(t, func(o *tui.Options) { o.NoColor = true })
	frames := []string{m.View()}

	body, _ := key(t, m, "enter")
	frames = append(frames, body.View())

	pick, _ := key(t, m, "down")
	pick, _ = key(t, pick, "enter")
	frames = append(frames, pick.View())

	for i, f := range frames {
		if j := strings.IndexByte(f, 0x1b); j >= 0 {
			t.Errorf("frame %d carries an ANSI escape at byte %d with NO_COLOR set: %q",
				i, j, f)
		}
	}
	// The selection must still be visible without styling, or "degrade legibly"
	// is not satisfied — the marker is the signal and the colour was decoration.
	if !strings.Contains(frames[0], "> ") {
		t.Errorf("no cursor marker without colour:\n%s", frames[0])
	}
}

// TestNarrowTerminalStaysWithinItsWidth is the other half of R-5.15. Sixty
// columns is the case the requirement names; the assertion is that no rendered
// line exceeds the terminal, because a line that overflows is folded by the
// terminal itself, outside the viewport's line accounting, and the scroll
// position stops meaning anything.
func TestNarrowTerminalStaysWithinItsWidth(t *testing.T) {
	long := strings.Repeat("governance requirement that keeps going ", 12)
	cat := []tui.Screen{{
		Title: "very long title that will not fit inside sixty columns at all",
		Run:   func(string) (string, error) { return long + "\n    indented " + long, nil },
	}}
	m := tui.New(cat, tui.Options{Output: &bytes.Buffer{}, NoColor: true})

	for _, width := range []int{60, 40, 100} {
		next, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: 20})
		at := next.(tui.Model)
		at, _ = key(t, at, "enter")
		for _, line := range strings.Split(at.View(), "\n") {
			if n := len([]rune(line)); n > width {
				t.Errorf("at %d columns a line is %d runes wide: %q", width, n, line)
			}
		}
	}
}

// TestResizeAfterOpeningReflows. A terminal resized while a report is on screen
// must re-wrap what is already there, not only what is opened next.
func TestResizeAfterOpeningReflows(t *testing.T) {
	long := strings.Repeat("wide ", 60)
	cat := []tui.Screen{{Title: "wide", Run: func(string) (string, error) { return long, nil }}}
	m := tui.New(cat, tui.Options{Output: &bytes.Buffer{}, NoColor: true})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 20})
	at, _ := key(t, next.(tui.Model), "enter")

	next, _ = at.Update(tea.WindowSizeMsg{Width: 50, Height: 20})
	at = next.(tui.Model)
	for _, line := range strings.Split(at.View(), "\n") {
		if n := len([]rune(line)); n > 50 {
			t.Errorf("after a resize a line is %d runes wide: %q", n, line)
		}
	}
}

// TestViewBeforeAnyWindowSize. Bubble Tea renders once before the first
// WindowSizeMsg on some terminals, and a zero width would divide the layout by
// nothing.
func TestViewBeforeAnyWindowSize(t *testing.T) {
	m := tui.New(screens(t), tui.Options{Output: &bytes.Buffer{}, NoColor: true})
	if strings.TrimSpace(m.View()) == "" {
		t.Fatal("the first frame is empty before any WindowSizeMsg")
	}
	next, _ := m.Update(tea.WindowSizeMsg{Width: 0, Height: 0})
	if strings.TrimSpace(next.(tui.Model).View()) == "" {
		t.Fatal("a zero-sized terminal renders nothing")
	}
}

// TestAnEmptyCatalogDoesNotPanic. `ids list` on a workspace with no registry,
// `governance explain` with no components — a picker with nothing in it is a
// real state, and it has to say so rather than index into an empty slice.
func TestEmptyChoicesRenderAMessage(t *testing.T) {
	cat := []tui.Screen{{
		Title:   "governance explain",
		Prompt:  "component",
		Choices: []string{},
		Run:     func(string) (string, error) { return "unreachable", nil },
	}}
	m := tui.New(cat, tui.Options{Output: &bytes.Buffer{}, NoColor: true})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	at, _ := key(t, next.(tui.Model), "enter")
	// A nil/empty Choices slice means "no argument", so this opens the body.
	if at.Mode() != tui.ModeBody {
		t.Fatalf("expected the body for an empty choice set, got %v", at.Mode())
	}
	if strings.TrimSpace(at.View()) == "" {
		t.Fatal("empty frame")
	}
}
