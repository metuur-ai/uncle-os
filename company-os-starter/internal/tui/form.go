package tui

// Forms and the mandatory preview (R-5.5..R-5.12).
//
// This file adds the two modes a mutating screen needs — collect, then confirm —
// without giving the package any more knowledge of Company OS than it had
// before. It still cannot compose an invocation: the sentence a form shows
// before it writes anything comes back from Action.Preview(), built by
// cmd/company-os from the same *Args it will execute. The AST guard in
// form_test.go asserts that no string literal in this package even LOOKS like a
// command line, which is what keeps R-5.7 structural rather than aspirational.

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Action is ONE resolved invocation, and the reason preview and execution cannot
// diverge (R-5.7).
//
// The model builds exactly one of these from the collected answers, stores that
// value, renders Preview() from it, and — only after the user confirms — calls
// Commit() on THE SAME VALUE. There is no second argument, no second closure,
// and no path through Update that could preview one action and run another: the
// only Action the model has ever seen is the one in m.action.
//
// Commit returns TEXT, for the same reason Screen.Run does: the catalog has
// already passed the records through the renderer the flag CLI uses, so this
// package cannot format a result differently from `company-os` (R-5.13).
type Action interface {
	// Preview is the exact, flag-complete invocation Commit will perform
	// (R-5.6). It must be derived from the same values Commit dispatches on.
	Preview() string
	// Commit performs it. Nothing in this package calls it before the user has
	// confirmed the string Preview returned (R-5.8).
	Commit() (string, error)
}

// Field is one answer a form collects.
//
// Every field corresponds to a command-line flag or positional (R-5.10); the
// mapping lives in the catalog, because that is the only place that knows the
// parser's spec. What this package guarantees is narrower and sufficient: a
// field's value is a plain string, and the whole set of them is what Build sees.
type Field struct {
	// Label names the value, e.g. "team". It is also what the invocation calls
	// it, so a reader can match the form row to the previewed flag.
	Label string
	// Help is one line under the field list explaining the focused field.
	Help string
	// Choices, when non-empty, makes the field a picker rather than a text
	// entry: left/right cycle it and no keystroke can produce a value the CLI
	// would reject.
	Choices []string
	// Default is the value the field starts with. Empty means the first choice
	// for a required picker, and an empty string for anything else.
	Default string
	// Optional allows the field to stay empty. An optional picker gains an
	// unset entry ahead of its choices.
	Optional bool
}

// Form is a mutating screen's body: the values to collect and the one function
// that turns them into an Action.
//
// Build is deliberately the ONLY hook. A separate Preview closure alongside a
// separate Commit closure would put two independently written functions behind
// one screen, which is exactly the drift R-5.7 forbids — the second one would be
// correct on the day it was written and wrong on the day the first one changed.
type Form struct {
	Fields []Field
	// Build resolves the answers into one Action. It MUST NOT touch the
	// filesystem: it runs the moment the user asks to see the preview, which is
	// before any confirmation has been given (R-5.8).
	Build func(values []string) (Action, error)
}

// unsetLabel is how an optional picker renders "no value". It is a UI label for
// the empty string, not a value: the empty string is what reaches Build, and an
// empty value is what the catalog omits from the invocation.
const unsetLabel = "(none)"

// openForm enters a screen's form with its defaults filled in.
//
// The form arrives ALREADY RESOLVED (see Screen.ResolveForm): a screen may
// supply it lazily, and resolving it here instead would call that closure again
// on a path that must not re-read the workspace mid-edit.
func (m *Model) openForm(i int, f *Form) {
	s := m.screens[i]
	m.active = i
	m.title = s.Title
	m.choice = ""
	m.fail = ""
	m.action = nil
	m.form = f
	m.field = 0
	m.values = make([]string, len(f.Fields))
	for j, fld := range f.Fields {
		m.values[j] = initialValue(fld)
	}
	m.mode = ModeForm
	if len(f.Fields) == 0 {
		// A form with nothing to collect is an OFFER: a fixed invocation the
		// reader either confirms or declines. Go straight to the preview so the
		// confirmation is the whole interaction, rather than showing an empty
		// field list that cannot be advanced (m.field would index nothing).
		// preview() still calls Build, so R-5.6/R-5.7 hold identically — the
		// previewed line is derived, not written for this screen.
		*m = m.preview()
	}
}

func initialValue(f Field) string {
	if f.Default != "" {
		return f.Default
	}
	if !f.Optional && len(f.Choices) > 0 {
		return f.Choices[0]
	}
	return ""
}

// typing reports whether the focused field takes free text. It is the one place
// the plain `q` exit is suspended — see Update for why that carve-out is this
// narrow and no narrower.
func (m Model) typing() bool {
	return m.mode == ModeForm && m.form != nil &&
		m.field >= 0 && m.field < len(m.form.Fields) &&
		len(m.form.Fields[m.field].Choices) == 0
}

// setValue writes one answer, copying the slice first.
//
// The copy is not defensive style, it is required: Update takes Model BY VALUE
// and every keystroke produces a new one, so two models made from the same
// parent share this slice's backing array. Without the copy, editing a field in
// one branch would silently edit it in a model a caller still holds — including
// the earlier model a test keeps to prove the exit keys work from every mode.
func (m *Model) setValue(i int, v string) {
	vals := make([]string, len(m.values))
	copy(vals, m.values)
	vals[i] = v
	m.values = vals
}

// formKey drives the collect step.
func (m Model) formKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	fields := m.form.Fields
	// A field-less form is only ever passed through on the way to its
	// confirmation; if a keystroke reaches here with nothing to edit, the only
	// coherent thing to do is leave. Indexing fields[0] would panic.
	if len(fields) == 0 {
		m.mode = ModeMenu
		return m, nil
	}
	f := fields[m.field]
	switch msg.String() {
	case "up":
		m.field, m.fail = clamp(m.field-1, 0, len(fields)-1), ""
		return m, nil
	case "down", "tab":
		m.field, m.fail = clamp(m.field+1, 0, len(fields)-1), ""
		return m, nil
	case "left":
		// Value cycling ONLY. On a text field there is no value to cycle and
		// left no longer navigates: since 2026-07-27 the arrows mean exactly one
		// thing in this UI, and Esc is the way back from every field.
		if len(f.Choices) > 0 {
			m.setValue(m.field, cycle(f, m.values[m.field], -1))
		}
		return m, nil
	case "right":
		if len(f.Choices) > 0 {
			m.setValue(m.field, cycle(f, m.values[m.field], +1))
		}
		return m, nil
	case "enter":
		if m.field < len(fields)-1 {
			m.field, m.fail = m.field+1, ""
			return m, nil
		}
		return m.preview(), nil
	case "backspace":
		if len(f.Choices) > 0 {
			back, _ := m.goBack()
			return back, nil
		}
		if r := []rune(m.values[m.field]); len(r) > 0 {
			m.setValue(m.field, string(r[:len(r)-1]))
		}
		return m, nil
	}

	// Anything else is text, and only for a text field: a picker never accepts a
	// keystroke, so it cannot hold a value the parser would reject.
	if len(f.Choices) > 0 {
		return m, nil
	}
	switch msg.Type {
	case tea.KeyRunes:
		m.setValue(m.field, m.values[m.field]+string(msg.Runes))
	case tea.KeySpace:
		m.setValue(m.field, m.values[m.field]+" ")
	}
	return m, nil
}

// cycle moves a picker one step, wrapping. An optional picker gains an unset
// entry ahead of its choices, so "leave it out" is reachable by the same key as
// every other value.
func cycle(f Field, current string, delta int) string {
	vals := f.Choices
	if f.Optional {
		vals = append([]string{""}, f.Choices...)
	}
	if len(vals) == 0 {
		return current
	}
	at := 0
	for i, v := range vals {
		if v == current {
			at = i
			break
		}
	}
	at = (at + delta + len(vals)) % len(vals)
	return vals[at]
}

// preview builds the Action and shows it. This is the LAST step before anything
// could be written, and nothing is written here: Build resolves arguments, and
// the write happens in commit, after the confirmation this mode exists to
// collect (R-5.8).
func (m Model) preview() Model {
	for i, f := range m.form.Fields {
		if !f.Optional && strings.TrimSpace(m.values[i]) == "" {
			m.field = i
			m.fail = f.Label + " is required"
			return m
		}
	}
	action, err := m.form.Build(m.values)
	if err != nil {
		m.fail = err.Error()
		return m
	}
	m.action = action
	m.fail = ""
	m.mode = ModeConfirm
	return m
}

// confirmKey drives the confirmation step. Only `y` and `enter` run anything;
// every other exit from this mode — n, backspace, left, and the three quit keys
// handled in Update — leaves the workspace exactly as it was (R-5.9).
func (m Model) confirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		m.commit()
		return m, nil
	case "n", "backspace":
		// One definition of back, shared with Esc — see goBack, which also owns
		// the field-less-offer case (an offer has no form to return to).
		//
		// `left` was dropped here with the other arrow aliases: there is no
		// focused field in this mode, so it had no value-cycling meaning and was
		// pure navigation.
		back, _ := m.goBack()
		return back, nil
	}
	return m, nil
}

// commit runs the action that was previewed — m.action, the same value
// Preview() was rendered from — and shows its output.
//
// R-5.9's second half is honoured by NOT pretending: once Commit begins, a
// failure part-way through is not rolled back, because only `init` is atomic
// (it stages into a temporary directory) and the rest of the system has no
// transaction to offer. The body and the error are both shown, which is what
// lets the reader see how far it got.
func (m *Model) commit() {
	body, err := m.action.Commit()
	m.fail = ""
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

// formView lists the fields with their current answers.
func (m Model) formView() string {
	var b strings.Builder
	width := 0
	for _, f := range m.form.Fields {
		if n := len([]rune(f.Label)); n > width {
			width = n
		}
	}
	for i, f := range m.form.Fields {
		value := m.values[i]
		if value == "" {
			value = unsetLabel
		}
		row := pad(f.Label, width) + "  " + value
		if len(f.Choices) > 0 {
			row += "   " + m.sty.dim.Render("(left/right)")
		}
		if i == m.field {
			b.WriteString(m.sty.sel.Render("> " + m.truncate(row)))
		} else {
			b.WriteString("  " + m.truncate(row))
		}
		b.WriteString("\n")
	}
	if help := m.form.Fields[m.field].Help; help != "" {
		b.WriteString("\n")
		b.WriteString(m.sty.dim.Render(wrap(help, m.width)))
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// confirmView shows the invocation and asks. The command line is Preview()'s
// bytes, unedited and un-reassembled — this package never composes one.
func (m Model) confirmView() string {
	var b strings.Builder
	b.WriteString(m.sty.sel.Render("run this command?"))
	b.WriteString("\n\n")
	b.WriteString(wrap("  $ "+m.action.Preview(), m.width))
	b.WriteString("\n\n")
	b.WriteString(m.sty.dim.Render(wrap(
		"nothing has been written yet. y runs it; n or esc goes back; "+
			"q or ctrl-c leaves the workspace untouched.", m.width)))
	return b.String()
}

func pad(s string, width int) string {
	if n := len([]rune(s)); n < width {
		return s + strings.Repeat(" ", width-n)
	}
	return s
}
