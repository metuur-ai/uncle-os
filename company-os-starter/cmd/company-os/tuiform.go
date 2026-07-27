package main

// The mutating forms (R-5.5): `discover new`, `prd new`, and the three `add`
// kinds — team, platform, component.
//
// R-5.5 still forbids forms for `workspace sync` and `scratchpad init`, and
// that cut still holds: both need values no reader has in their head at the
// menu (a repo URL and a commit pin; a path outside the workspace).
//
// The first two are the commands a product owner AUTHORS — the ones whose
// arguments are a title they are still wording and a team they have to look up.
// This file used to argue from that to a closed set of two: "nobody is going to
// scaffold infrastructure through a form instead of typing the command, so a
// form for those two would be surface with no reader." That was a prediction
// about readers, and it was wrong — the three `add` forms exist because someone
// asked for them (Amendment 4, 2026-07-27). The prediction is recorded rather
// than deleted, because the next person to argue a form has no audience should
// know this file has been wrong about that before.
//
// `add` is one command with a `kind` positional, but it is THREE screens, not
// one screen with a kind picker: only `component` takes `--platform`, and a
// single form would have to offer that field to all three and let two of them
// fail at commit. A form whose fields are wrong for the chosen value is the
// kind of surface a menu is supposed to remove.
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
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/scaffold"
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

// mutatingScreens is R-5.5's list. Every entry is marked in its title, because
// a menu that does not distinguish reading from writing is a menu that gets
// someone to write by accident.
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
		{
			Title: "add team (writes)",
			Form: &tui.Form{
				Fields: []tui.Field{idField("team")},
				Build: func(v []string) (tui.Action, error) {
					return addInvocation(ws, root, "team", v[0], "")
				},
			},
		},
		{
			Title: "add platform (writes)",
			Form: &tui.Form{
				Fields: []tui.Field{idField("platform")},
				Build: func(v []string) (tui.Action, error) {
					return addInvocation(ws, root, "platform", v[0], "")
				},
			},
		},
		{
			Title: "add component (writes)",
			// Resolved at open time, not here: a reader who adds a platform and
			// then a component to it in one sitting must see the platform they
			// just created. Everything else in this file is fixed at catalog
			// build, which is why this is the only FormFn.
			FormFn: func() *tui.Form {
				current := baseNames(ws.AllPlatforms())
				return &tui.Form{
					Fields: []tui.Field{
						{
							Label:   "platform",
							Choices: current,
							Help:    platformFieldHelp(current),
						},
						idField("component"),
					},
					Build: func(v []string) (tui.Action, error) {
						return addInvocation(ws, root, "component", v[1], v[0])
					},
				}
			},
		},
	}
}

// idField is the new-id field shared by the three `add` screens.
//
// Labelled "id" rather than "name" to match what the parser calls the positional
// ("id of the new platform/team/component", args.go), and the Help discloses the
// slugging, the same way `discover new`'s title field discloses how a brief id is
// derived. Without it, "My Team" silently becoming "my-team" reads as the CLI
// ignoring what was typed.
func idField(kind string) tui.Field {
	return tui.Field{
		Label: "id",
		Help: "id of the new " + kind + ". Lowercased, and every run of other " +
			"characters becomes one dash: \"My " + kind + "\" creates \"my-" +
			kind + "\".",
	}
}

// platformFieldHelp names the target platform's role, and says plainly when there
// is nothing to pick.
//
// The empty case is neither hypothetical nor cosmetic: a required field with no
// Choices is indistinguishable from a text box — internal/tui gives it no
// left/right hint and lets keystrokes through — and a workspace with no platforms
// is exactly where a reader reaches for `add component` first. Typing one anyway
// is refused by ws.PlatformDir with exit 3 rather than scaffolding into nowhere,
// but the field should say so before they type.
func platformFieldHelp(platforms []string) string {
	if len(platforms) == 0 {
		return "the platform that will own the component. This workspace has " +
			"none yet — add a platform first."
	}
	return "the platform that will own the component. Its descriptor is written " +
		"to platforms/<platform>/components/."
}

// addInvocation builds one `company-os add <kind> <id> [--platform p]`.
//
// The empty-slug refusal lives here because Build is the seam designed for it:
// the form's own required check only rejects whitespace, so "###" would pass it,
// slug to "", and reach scaffold.Add with an empty id — which resolves
// teams/<id>/team.yaml to teams/team.yaml and scatters a team's files into the
// layer root. Refusing in Build keeps the reader in the form with the reason, and
// writes nothing.
func addInvocation(ws *workspace.Workspace, root, kind, id, platform string) (tui.Action, error) {
	if scaffold.Slugify(id) == "" {
		return nil, fmt.Errorf("%q has no letters or digits — the id would be empty", id)
	}
	return newInvocation(ws, &Args{
		Root: root, Cmd: "add", Kind: kind, Name: id, Platform: platform,
	}), nil
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
