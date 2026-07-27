// Package tui is the interactive terminal UI (Unit 5).
//
// It is the ONE package in the module permitted to emit ANSI escape sequences.
// cmd/company-os/ansi_test.go walks every other Go file in the tree and fails on
// an escape byte in a string literal or an import of a styling library; this
// directory is its only exemption, and the five golden files depend on that
// exemption staying exactly this narrow (R-3.10).
//
// The package is deliberately ignorant of Company OS. It knows nothing about
// workspaces, gates, findings, or renderers — it takes a []Screen of already-
// resolved titles and closures and drives them through Bubble Tea's
// model-view-update loop. That is what keeps R-5.13 true by construction: a
// screen's body is whatever internal/render wrote for it, because this package
// has no way to compose a sentence of its own. cmd/company-os/tui.go builds the
// catalog; see there for what the ten screens actually run.
//
// Everything below is a plain Model/Update/View triple, so the whole UI is
// testable by calling Update with a synthetic message and reading View — no pty,
// no terminal, no timing. tui_test.go does exactly that.
package tui

import (
	"errors"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen is one entry in the catalog: a title, an optional list of values to
// pick from, and either the closure that produces the body or — for a mutating
// screen — the Form that collects arguments, previews the invocation, and runs
// it only after confirmation (see form.go).
//
// Run returns TEXT, not records. The catalog has already passed the record set
// through the same internal/render function the flag CLI uses, so this package
// cannot render a gate differently from `company-os validate` even by accident
// (R-5.13).
//
// Run may return a body AND an error: several commands print records and then
// refuse, and dropping the body on the error path would lose the half of the
// output that explains the refusal — the same reason main.go renders before it
// checks err.
type Screen struct {
	// Title is what the menu lists, e.g. "validate results".
	Title string
	// Prompt labels the picker, e.g. "role" or "component".
	Prompt string
	// Choices are the values Run accepts. A nil slice means the screen takes no
	// argument and Run is called with "".
	Choices []string
	// Run produces the body. The argument is the picked choice, or "".
	// It is nil when Form is set.
	Run func(choice string) (string, error)
	// Form, when non-nil, makes this a MUTATING screen (R-5.5): opening it
	// collects the form's fields instead of running anything, and the write
	// happens only after the previewed invocation is confirmed. Choices and Run
	// are unused in that case.
	Form *Form
}

// Options configure a run. Input and Output are supplied by cmd/company-os
// rather than reached for here: R-2.10 forbids os.Stdout below cmd/, and the
// same seam is what lets a test drive the program with pipes.
type Options struct {
	Input  io.Reader
	Output io.Writer
	// NoColor suppresses every styling attribute (R-5.15). The caller decides,
	// because reading the environment is a decision about the process, not
	// about the widget.
	NoColor bool
}

// Run starts the UI and blocks until the user exits. It returns only a terminal
// I/O failure; a screen whose command errored is reported inside the UI, which
// is where the reader can act on it.
func Run(screens []Screen, opts Options) error {
	p := tea.NewProgram(New(screens, opts),
		tea.WithInput(opts.Input),
		tea.WithOutput(opts.Output),
		// The alternate screen is what makes "leaves no partial write" (R-5.14)
		// visibly true: on exit the terminal is restored to the scrollback the
		// user had before, with no UI residue to clean up.
		tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// Mode is which of the three views is on screen. There are only three, and the
// transitions are a line each, so a state machine type would be more machinery
// than the thing it models.
//
// It is exported, with an accessor, so a test can assert WHERE the UI is without
// driving a terminal — which is what makes "q, Esc and Ctrl-C exit from EVERY
// screen" (R-5.14) checkable rather than merely claimed.
type Mode int

const (
	// ModeMenu is the list of screens.
	ModeMenu Mode = iota
	// ModePick is a screen's argument list.
	ModePick
	// ModeBody is a screen's rendered output.
	ModeBody
	// ModeForm collects a mutating screen's arguments. Nothing is written here.
	ModeForm
	// ModeConfirm shows the derived invocation and waits (R-5.6, R-5.8). This is
	// the last mode in which the workspace is still untouched.
	ModeConfirm
)

// String names the mode for a test failure message.
func (m Mode) String() string {
	switch m {
	case ModeMenu:
		return "menu"
	case ModePick:
		return "picker"
	case ModeBody:
		return "body"
	case ModeForm:
		return "form"
	case ModeConfirm:
		return "confirm"
	}
	return "unknown"
}

// Model is the root model. One model holds the whole UI: R-5.4's ten screens are
// ten catalog entries, not ten models, because every one of them is a title, an
// optional list, and a block of text in a viewport.
type Model struct {
	screens []Screen

	mode   Mode
	menu   int // cursor in the screen list
	pick   int // cursor in the choice list
	active int // index into screens, valid in ModePick and ModeBody

	vp     viewport.Model
	body   string // the unwrapped body, re-wrapped on every resize
	title  string
	choice string // the value this screen was opened with, "" when it takes none
	fail   string // an error from Run, shown above the body

	// the mutating half (form.go). action is the ONE Action that was built from
	// values: it is what ModeConfirm previews and what commit runs, so the two
	// cannot be different invocations (R-5.7).
	form   *Form
	values []string
	field  int
	action Action

	width, height int
	sty           styles
	quitting      bool
	// handOff records that a screen returned ErrHandOff, so Update quits
	// without rendering. Separate from quitting because the two answer
	// different questions: quitting is "the user asked to leave", handOff is
	// "a screen ended this run on the caller's behalf".
	handOff bool
}

// defaultWidth and defaultHeight apply until the first WindowSizeMsg arrives.
// Bubble Tea sends one immediately on a real terminal, but View can be called
// before it lands (and always is, in a test), and a zero width would divide the
// layout by nothing.
const (
	defaultWidth  = 80
	defaultHeight = 24
	// narrowWidth is the threshold R-5.15 names. Below it the layout drops its
	// gutter and the footer shortens; above it neither changes.
	narrowWidth = 80
	// chromeHeight is the header and footer the viewport does not get: two
	// header lines, a blank, and one footer line.
	chromeHeight = 4
)

// New builds the root model.
func New(screens []Screen, opts Options) Model {
	m := Model{
		screens: screens,
		width:   defaultWidth,
		height:  defaultHeight,
		sty:     newStyles(opts.Output, opts.NoColor),
	}
	m.vp = viewport.New(m.width, m.height-chromeHeight)
	return m
}

// Mode reports which view is on screen.
func (m Model) Mode() Mode { return m.mode }

// Init implements tea.Model. There is nothing to load: every screen's Run is
// called on demand, so start-up does no filesystem work at all.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.relayout()
		return m, nil

	case tea.KeyMsg:
		// R-5.14, before anything else and from every mode. The audience for
		// this UI is defined by terminal unfamiliarity, so the exit has to be
		// the one behaviour that never depends on where you are. That is also
		// why Esc does not mean "back": a key that sometimes exits and
		// sometimes does not is the trap R-5.14 exists to prevent. Going back
		// is backspace or left, and the footer says so on every screen.
		//
		// `q` has ONE carve-out and it is stated rather than hidden: while a
		// free-text field has focus it is a character, because a form whose
		// title cannot contain the letter q is not a form. Esc and Ctrl-C stay
		// unconditional there, so every mode including a text field still has
		// two ways out, the footer names them, and nothing has been written at
		// that point in any case (R-5.8) — so no exit from a form can leave a
		// partial write.
		switch msg.String() {
		case "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			if !m.typing() {
				m.quitting = true
				return m, tea.Quit
			}
		}
		return m.key(msg)
	}

	if m.mode == ModeBody {
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

// key handles everything that is not an exit.
func (m Model) key(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeMenu:
		switch msg.String() {
		case "up", "k":
			m.menu = clamp(m.menu-1, 0, len(m.screens)-1)
		case "down", "j":
			m.menu = clamp(m.menu+1, 0, len(m.screens)-1)
		case "home", "g":
			m.menu = 0
		case "end", "G":
			m.menu = len(m.screens) - 1
		case "enter", " ", "right", "l":
			opened := m.open(m.menu)
			if opened.handOff {
				return opened, tea.Quit
			}
			return opened, nil
		}
		return m, nil

	case ModePick:
		choices := m.screens[m.active].Choices
		switch msg.String() {
		case "up", "k":
			m.pick = clamp(m.pick-1, 0, len(choices)-1)
		case "down", "j":
			m.pick = clamp(m.pick+1, 0, len(choices)-1)
		case "home", "g":
			m.pick = 0
		case "end", "G":
			m.pick = len(choices) - 1
		case "backspace", "left", "h":
			m.mode = ModeMenu
		case "enter", " ", "right", "l":
			if len(choices) > 0 {
				m.load(m.active, choices[m.pick])
				if m.handOff {
					return m, tea.Quit
				}
			}
		}
		return m, nil

	case ModeForm:
		return m.formKey(msg)

	case ModeConfirm:
		return m.confirmKey(msg)

	default: // ModeBody
		switch msg.String() {
		case "backspace", "left", "h":
			if len(m.screens[m.active].Choices) > 0 {
				m.mode = ModePick
			} else {
				m.mode = ModeMenu
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(tea.Msg(msg))
		return m, cmd
	}
}

// open enters a screen: straight to its body when it takes no argument, to the
// picker when it does.
func (m Model) open(i int) Model {
	if i < 0 || i >= len(m.screens) {
		return m
	}
	m.active = i
	if m.screens[i].Form != nil {
		m.openForm(i)
		return m
	}
	if len(m.screens[i].Choices) > 0 {
		m.pick = 0
		m.mode = ModePick
		return m
	}
	m.load(i, "")
	return m
}

// ErrHandOff, returned by a Screen's Run, ends the program cleanly instead of
// rendering a body or an error.
//
// It is an exit path, not a mode: this package still knows nothing about what
// the caller wants to do next. The one use is the recovery menu's
// "open a workspace found nearby" — the caller records the chosen root in its
// own closure and restarts the TUI against it. Showing "switching…" in a
// viewport the caller is about to replace would be a frame of noise, and
// leaving the program running after the catalog is obsolete would be worse.
var ErrHandOff = errors.New("tui: screen handed control back to the caller")

// load runs a screen and puts the result in the viewport.
//
// The pointer receiver is deliberate: load mutates the viewport, and a value
// receiver would drop the scroll reset on a copy.
func (m *Model) load(i int, choice string) {
	s := m.screens[i]
	m.title = s.Title
	m.choice = choice
	m.fail = ""
	body, err := s.Run(choice)
	if errors.Is(err, ErrHandOff) {
		m.handOff = true
		m.quitting = true
		return
	}
	if err != nil {
		m.fail = err.Error()
	}
	if strings.TrimSpace(body) == "" && m.fail == "" {
		body = "(nothing to show)"
	}
	m.body = body
	m.mode = ModeBody
	m.relayout()
	m.vp.GotoTop()
}

// relayout re-sizes the viewport and re-wraps the body. Wrapping happens here
// rather than in View because it is the expensive half and the body only changes
// when the screen or the width does — which is exactly when this is called
// (R-5.15: survives resize).
func (m *Model) relayout() {
	w, h := m.width, m.height
	if w < 1 {
		w = defaultWidth
	}
	if h < chromeHeight+1 {
		h = chromeHeight + 1
	}
	m.vp.Width = w
	m.vp.Height = h - chromeHeight
	m.vp.SetContent(wrap(m.body, w))
}

// View implements tea.Model.
func (m Model) View() string {
	if m.quitting {
		// An empty final frame: with the alternate screen this is what leaves
		// the terminal as it was found (R-5.14).
		return ""
	}
	var b strings.Builder
	b.WriteString(m.header())
	b.WriteString("\n\n")
	switch m.mode {
	case ModeMenu:
		b.WriteString(m.list(titles(m.screens), m.menu))
	case ModePick:
		b.WriteString(m.list(m.screens[m.active].Choices, m.pick))
	case ModeForm:
		if m.fail != "" {
			b.WriteString(m.sty.fail.Render(wrap(m.fail, m.width)))
			b.WriteString("\n\n")
		}
		b.WriteString(m.formView())
	case ModeConfirm:
		b.WriteString(m.confirmView())
	default:
		if m.fail != "" {
			b.WriteString(m.sty.fail.Render(wrap("error: "+m.fail, m.width)))
			b.WriteString("\n\n")
		}
		b.WriteString(m.vp.View())
	}
	b.WriteString("\n")
	b.WriteString(m.footer())
	return b.String()
}

func (m Model) header() string {
	switch m.mode {
	case ModeMenu:
		return m.sty.title.Render("company-os") + "\n" +
			m.sty.dim.Render(m.truncate("read-only workspace views"))
	case ModePick:
		s := m.screens[m.active]
		return m.sty.title.Render(m.truncate(s.Title)) + "\n" +
			m.sty.dim.Render(m.truncate("choose a "+orDefault(s.Prompt, "value")))
	case ModeForm:
		return m.sty.title.Render(m.truncate(m.title)) + "\n" +
			m.sty.dim.Render(m.truncate(
				"fill in the arguments — nothing is written yet"))
	case ModeConfirm:
		return m.sty.title.Render(m.truncate(m.title)) + "\n" +
			m.sty.dim.Render(m.truncate("review before it runs"))
	default:
		// The flag-complete invocation is NOT repeated here: the body already
		// opens with it, derived from the *Args that ran (cmd/company-os/tui.go
		// § screenCommand), and a second copy composed from a different value
		// is the drift R-5.7 exists to forbid. The subtitle carries what the
		// body cannot — which argument this screen was opened with.
		sub := "read-only view"
		if m.screens[m.active].Form != nil {
			// A body reached through a form is the result of a WRITE, and
			// labelling it "read-only view" would be the UI's only false
			// statement — made at the one moment the reader most needs to know
			// what just happened.
			sub = "this ran and changed the workspace"
		}
		if m.choice != "" {
			sub = orDefault(m.screens[m.active].Prompt, "argument") + ": " + m.choice
		}
		return m.sty.title.Render(m.truncate(m.title)) + "\n" +
			m.sty.dim.Render(m.truncate(sub))
	}
}

func (m Model) footer() string {
	var keys string
	switch {
	// The typing case comes first, and names esc/ctrl-c rather than q, because
	// in a text field q is a character — telling the reader otherwise is the
	// only way this carve-out could trap them.
	case m.typing() && m.narrow():
		keys = "type  enter=next  esc=quit"
	case m.typing():
		keys = "type   enter next   left back   esc / ctrl-c quit"
	case m.mode == ModeForm && m.narrow():
		keys = "l/r=value  enter=next  q=quit"
	case m.mode == ModeForm:
		keys = "up/down field   left/right value   enter next   q / esc / ctrl-c quit"
	case m.mode == ModeConfirm && m.narrow():
		keys = "y=run  n=back  q=cancel"
	case m.mode == ModeConfirm:
		keys = "y run   n back   q / esc / ctrl-c cancel without writing"
	case m.narrow() && m.mode == ModeMenu:
		keys = "up/down  enter  q=quit"
	case m.narrow():
		keys = "up/down  bksp=back  q=quit"
	case m.mode == ModeMenu:
		keys = "up/down move   enter open   q / esc / ctrl-c quit"
	case m.mode == ModePick:
		keys = "up/down move   enter open   backspace back   q / esc / ctrl-c quit"
	default:
		keys = "up/down scroll   backspace back   q / esc / ctrl-c quit"
	}
	return m.sty.dim.Render(m.truncate(keys))
}

// list renders a cursor list. The "> " marker carries the selection on its own,
// so the list stays readable with NO_COLOR set and on a terminal with no styling
// at all — the highlight is decoration, never the only signal (R-5.15).
func (m Model) list(items []string, cursor int) string {
	if len(items) == 0 {
		return m.sty.dim.Render("(nothing to browse here)")
	}
	// The list scrolls by window rather than in a viewport: it is one line per
	// item with a cursor, and a viewport would need its own scroll-follow logic
	// to keep the cursor visible.
	rows := m.height - chromeHeight
	if rows < 1 {
		rows = 1
	}
	first := 0
	if cursor >= rows {
		first = cursor - rows + 1
	}
	last := first + rows
	if last > len(items) {
		last = len(items)
	}
	var b strings.Builder
	for i := first; i < last; i++ {
		text := m.truncate(items[i])
		if i == cursor {
			b.WriteString(m.sty.sel.Render("> " + text))
		} else {
			b.WriteString("  " + text)
		}
		if i < last-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// truncate clips one line to the terminal width so a long title cannot wrap the
// header and push the layout down a row (R-5.15).
func (m Model) truncate(s string) string {
	w := m.width
	if w < 1 {
		w = defaultWidth
	}
	// Two columns are reserved for the list's "> " marker.
	w -= 2
	if w < 4 {
		w = 4
	}
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	return string(r[:w-1]) + "…"
}

func (m Model) narrow() bool { return m.width < narrowWidth }

// styles are the only place an escape sequence originates.
type styles struct {
	title lipgloss.Style
	dim   lipgloss.Style
	sel   lipgloss.Style
	fail  lipgloss.Style
}

// newStyles builds the palette against a renderer bound to the program's own
// output, never lipgloss's package-global default. The global one resolves its
// colour profile from os.Stdout on first use — which is the wrong stream when
// output is a pipe, and reaching for os.Stdout from an internal package is what
// R-2.10 forbids in the first place.
//
// With NO_COLOR the styles carry no attributes at all, not merely no colour.
// Dropping bold and faint as well is what turns "honours NO_COLOR" into a
// checkable property — tui_test.go asserts that not one byte of any frame is an
// escape — instead of an approximation nobody can measure. Nothing is lost: the
// selection is carried by a "> " marker, so the styling was never the signal.
func newStyles(w io.Writer, noColor bool) styles {
	if w == nil || noColor {
		var plain lipgloss.Style
		return styles{title: plain, dim: plain, sel: plain, fail: plain}
	}
	r := lipgloss.NewRenderer(w)
	return styles{
		title: r.NewStyle().Bold(true),
		dim:   r.NewStyle().Faint(true),
		sel:   r.NewStyle().Bold(true),
		fail:  r.NewStyle().Bold(true),
	}
}

func titles(screens []Screen) []string {
	out := make([]string, 0, len(screens))
	for _, s := range screens {
		out = append(out, s.Title)
	}
	return out
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		return lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// wrap hard-wraps text to width, breaking on a space where one is available.
//
// The viewport does not wrap — it emits each line as it is — so a gate finding
// longer than the terminal would be folded by the terminal itself, at the wrong
// place and outside the viewport's line accounting, which scrambles scrolling.
// Wrapping here is what makes 60 columns legible rather than garbled (R-5.15).
func wrap(text string, width int) string {
	if width < 1 {
		width = 1
	}
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, wrapLine(line, width)...)
	}
	return strings.Join(out, "\n")
}

// wrapLine wraps one line, carrying its leading indent onto every continuation.
// The gate renderers indent continuation detail under a finding; a continuation
// flush against the left margin reads as a new record, which is a wrong claim
// about the report rather than an ugly one.
func wrapLine(line string, width int) []string {
	if len([]rune(line)) <= width {
		return []string{line}
	}
	indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
	if len([]rune(indent)) >= width {
		indent = ""
	}
	var out []string
	rest, first := line, true
	for {
		prefix := ""
		limit := width
		if !first {
			prefix = indent
			limit = width - len([]rune(indent))
		}
		r := []rune(rest)
		if len(r) <= limit {
			out = append(out, prefix+rest)
			return out
		}
		cut := limit
		for i := limit; i > limit/2; i-- {
			if r[i] == ' ' {
				cut = i
				break
			}
		}
		out = append(out, strings.TrimRight(prefix+string(r[:cut]), " "))
		rest = strings.TrimLeft(string(r[cut:]), " ")
		first = false
		if rest == "" {
			return out
		}
	}
}
