package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/tui"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// The recovery menu: what `tui` shows when it starts somewhere that is not a
// workspace root.
//
// It exists because the previous behaviour — return RequireRoot's error and
// exit 3 — was measured to be wrong in the most likely case. Standing in
// examples/banking/bank/workspaces/, a directory holding TWO workspace roots,
// the TUI refused and printed the root-resolution order. For an audience defined
// as blocked on terminal fluency, that is the wrong answer to "I am one
// directory too high" (R-5.17).
//
// The non-interactive contract is untouched: cmdTUI still runs the TTY gate
// first and still exits 7 with no filesystem change (R-5.3). Only the root
// check moved, and it moved into the UI rather than being deleted.

// nearbyRoots returns workspace roots at or immediately below dir, sorted.
//
// Depth is one level, deliberately. A recursive scan of an arbitrary directory
// is slow in the places it would matter (a home directory, /) and surprising
// everywhere: an entry appearing from four levels down is not something the
// reader can predict from where they are standing. One level covers the case
// this was built for and stays explainable.
func nearbyRoots(dir string) []string {
	var found []string
	if workspace.New(dir).IsRoot() {
		found = append(found, dir)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		// Unreadable directory: no roots, no error. The menu still offers to
		// create one, which is the useful thing to do from here anyway.
		return found
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		child := filepath.Join(dir, e.Name())
		if workspace.New(child).IsRoot() {
			found = append(found, child)
		}
	}
	sort.Strings(found)
	return found
}

// recoveryScreens is the catalog shown outside a workspace.
//
// selected is written when the reader picks a nearby root; cmdTUI reads it after
// tui.Run returns and restarts against it. The value travels through this
// closure rather than through internal/tui because that package is deliberately
// ignorant of workspaces — it only needs to know the run ended (tui.ErrHandOff).
func recoveryScreens(dir string, selected *string) []tui.Screen {
	var screens []tui.Screen

	if roots := nearbyRoots(dir); len(roots) > 0 {
		labels := make([]string, len(roots))
		for i, r := range roots {
			if r == dir {
				labels[i] = ". (here)"
			} else {
				labels[i] = filepath.Base(r)
			}
		}
		byLabel := map[string]string{}
		for i, l := range labels {
			byLabel[l] = roots[i]
		}
		screens = append(screens, tui.Screen{
			Title:   fmt.Sprintf("open a workspace found nearby (%d)", len(roots)),
			Prompt:  "workspace",
			Choices: labels,
			Run: func(label string) (string, error) {
				*selected = byLabel[label]
				return "", tui.ErrHandOff
			},
		})
	}

	screens = append(screens,
		tui.Screen{
			Title: "create a workspace here",
			Form:  initForm(dir),
		},
		tui.Screen{
			Title: "what is a workspace root?",
			Run:   func(string) (string, error) { return rootHelp(dir), nil },
		},
	)
	return screens
}

// initForm collects `init`'s three arguments and previews the invocation before
// anything is written, through the same Form/Action path the other mutating
// screens use (R-5.5, R-5.8).
func initForm(dir string) *tui.Form {
	return &tui.Form{
		Fields: []tui.Field{
			{Label: "company", Help: "display name, e.g. Acme"},
			{Label: "team", Help: "id of the first team, e.g. core"},
			{Label: "platform", Help: "id of the first platform, e.g. web"},
		},
		Build: func(v []string) (tui.Action, error) {
			// --root is threaded so the previewed line reproduces from anywhere,
			// matching what runScreen will execute.
			return &invocation{
				ws: workspace.New(dir),
				args: &Args{
					Root: dir, Cmd: "init",
					Company: v[0], Team: v[1], Platform: v[2],
				},
			}, nil
		},
	}
}

// rootHelp is the content of the error this menu replaced, shown rather than
// fatal. A reader who did not know what a root was learned nothing from exit 3.
func rootHelp(dir string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s is not a workspace root.\n\n", dir)
	b.WriteString("A workspace root is a directory containing at least one of:\n")
	for _, name := range workspace.CanonicalRoots {
		fmt.Fprintf(&b, "  %s/\n", name)
	}
	fmt.Fprintf(&b, "  %s   (a federation manifest alone also marks a root)\n\n",
		workspace.ManifestName)
	b.WriteString("company-os finds the root in this order:\n")
	b.WriteString("  1. --root <path>\n")
	b.WriteString("  2. $COMPANY_OS_WORKSPACE_ROOT\n")
	b.WriteString("  3. the current directory\n\n")
	b.WriteString("If you are one directory above your workspace, the first menu\n")
	b.WriteString("entry lists the roots found here. If you have none yet, choose\n")
	b.WriteString("\"create a workspace here\".\n")
	return b.String()
}
