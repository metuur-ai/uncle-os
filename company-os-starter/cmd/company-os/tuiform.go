package main

// The two mutating forms (R-5.5): `discover new` and `prd new`.
//
// R-5.5 names these two and forbids forms for `workspace sync` and
// `scratchpad init`. That is not an arbitrary cut. These are the two commands a
// product owner AUTHORS — the ones whose arguments are a title they are still
// wording and a team they have to look up — and they are the reason the TUI has
// an audience at all. Nobody is going to scaffold infrastructure through a form
// instead of typing the command, so a form for those two would be surface with
// no reader.
//
// Everything a form collects becomes a field of *Args and nothing else (R-5.10):
// the invocation is rendered from that *Args by screenCommand, and the SAME
// *Args is what runScreen dispatches through `commands`. There is no second
// spelling of the command and no path that shells out to this binary (R-5.12).
//
// `discover validate` is deliberately absent from both this file and the
// read-only catalog: it REWRITES `status: draft` to `status: validated` in the
// brief (internal/product/discover.go), so it is a mutation wearing a read-only
// name, and wiring it anywhere that reads as browsing is the exact defect
// read-only-first exists to prevent. If it is ever offered, it belongs HERE,
// behind a preview and a confirmation, not in a browser.

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/tui"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
)

// invocation is ONE resolved command: a workspace and an *Args, and nothing
// else.
//
// This type is the whole of R-5.7's mechanism. Preview and Commit are two views
// of ONE field — Preview renders `a.args`, Commit dispatches `a.args` — so they
// cannot describe different commands for the same reason two callers of the same
// getter cannot see different values. Adding a `preview string` field here, or a
// second closure beside Commit, is the only way to break it, and
// tuiform_test.go's round-trip law fails immediately if anyone does.
type invocation struct {
	ws   *workspace.Workspace
	args *Args
}

func (i *invocation) Preview() string { return screenCommand(i.args) }

func (i *invocation) Commit() (string, error) { return runScreen(i.ws, i.args) }

// newInvocation normalizes before it wraps, so a form-built *Args is the same
// shape the parser would have produced for the previewed line (see
// normalizeArgs).
func newInvocation(ws *workspace.Workspace, args *Args) tui.Action {
	normalizeArgs(args)
	return &invocation{ws: ws, args: args}
}

// mutatingScreens is R-5.5's list. Two entries, both marked in their title,
// because a menu that does not distinguish reading from writing is a menu that
// gets someone to write by accident.
func mutatingScreens(ws *workspace.Workspace, root string) []tui.Screen {
	teams := baseNames(ws.AllTeams())
	platforms := baseNames(ws.AllPlatforms())
	components := componentIDList(componentCatalog(ws))

	return []tui.Screen{
		{
			Title: "new discovery brief (writes)",
			Form: &tui.Form{
				Fields: []tui.Field{
					{
						Label:   "team",
						Choices: teams,
						Help: "the team that owns the discovery. Briefs are " +
							"team-private: teams/<team>/product/discovery/.",
					},
					{
						Label: "title",
						Help: "free text. The brief id is derived from it: " +
							"<year>-<slugified-title>.",
					},
				},
				Build: func(v []string) (tui.Action, error) {
					return newInvocation(ws, &Args{
						Root: root, Cmd: "discover", Action: "new",
						Team: v[0], TitleArg: v[1],
					}), nil
				},
			},
		},
		{
			Title: "new PRD (writes)",
			Form: &tui.Form{
				Fields: []tui.Field{
					{
						Label:   "platform",
						Choices: platforms,
						Help: "the platform whose reality this change record " +
							"proposes to change.",
					},
					{
						Label:    "title",
						Optional: true,
						Help: "free text. Leave it out only when a discovery " +
							"brief is chosen below — the brief's title is used " +
							"instead.",
					},
					{
						Label:    "components",
						Optional: true,
						Help:     componentsHelp(components),
					},
					{
						Label:    "team",
						Choices:  teams,
						Optional: true,
						Help: "the proposing team. Required when a discovery " +
							"brief is chosen, because the brief is read from " +
							"that team's directory.",
					},
					{
						Label:    "from-discovery",
						Choices:  validatedBriefIDs(ws),
						Optional: true,
						Help: "a validated brief in the chosen team. Its " +
							"Problem signal and Success criteria are copied " +
							"into the PRD.",
					},
				},
				Build: func(v []string) (tui.Action, error) {
					return newInvocation(ws, &Args{
						Root: root, Cmd: "prd", Action: "new",
						Platform: v[0], Title: v[1], Components: v[2],
						Team: v[3], FromDiscovery: v[4],
					}), nil
				},
			},
		},
	}
}

// componentsHelp names the ids this workspace actually has, because the flag is
// a comma-separated list and a reader who guesses one gets a PRD whose targets
// do not resolve.
func componentsHelp(ids []string) string {
	if len(ids) == 0 {
		return "comma-separated component ids. This workspace has none yet."
	}
	list := strings.Join(ids, ", ")
	if len(ids) > 8 {
		list = strings.Join(ids[:8], ", ") + ", …"
	}
	return "comma-separated component ids, e.g. " + list
}

// validatedBriefIDs lists the briefs `prd new --from-discovery` will accept.
//
// Only `status: validated` ones are offered: internal/product rejects anything
// else with exit 5, so offering a draft would be offering a value the command is
// certain to refuse. Reading the status is a read — `discover validate`, the
// command that would MAKE a brief validated, writes to it and is not called from
// anywhere in the UI.
func validatedBriefIDs(ws *workspace.Workspace) []string {
	seen := map[string]bool{}
	var out []string
	for _, tdir := range ws.AllTeams() {
		dir := filepath.Join(tdir, "product", "discovery")
		for _, id := range subdirNames(dir) {
			if seen[id] {
				continue
			}
			if frontmatterField(filepath.Join(dir, id, "brief.md"), "status") == "validated" {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	sort.Strings(out)
	return out
}
