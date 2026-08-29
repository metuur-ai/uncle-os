package main

// The mutating forms (R-5.5). Two groups, and the difference matters:
//
//   - The lifecycle of one change, in order — `discover new`, `discover
//     validate`, `prd new`, `prd validate`, `reality new`, `prd complete`. A
//     reader following the menu top to bottom is following the method.
//   - Growing the federation — the three `add` kinds: team, platform, component.
//
// The lifecycle group was completed on 2026-08-29 (Amendments 5–8). Before
// that the catalog could START a unit of work and not finish one, and the gap
// was invisible rather than blocked: `prd new`'s picker lists validated briefs
// only, nothing in the UI validated one, so a reader was simply not offered the
// brief they had just written.
//
// R-5.5 still forbids forms for `workspace sync` and `scratchpad init`, and
// that cut still holds: both need values no reader has in their head at the
// menu (a repo URL and a commit pin; a path outside the workspace).
//
// `discover new` and `prd new` are the commands a product owner AUTHORS — the
// ones whose arguments are a title they are still wording and a team they have
// to look up. They shipped first, alone.
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
// `reality new` is the one form that CANNOT be split that way — its two fields
// constrain each other and a form cannot narrow one picker from another's value
// — so it enforces the same property at the other end, in Build. See
// realityInvocation, and R-5.5's combination clause.
//
// Everything a form collects becomes a field of *Args and nothing else (R-5.10):
// the invocation is rendered from that *Args by screenCommand, and the SAME
// *Args is what runScreen dispatches through `commands`. There is no second
// spelling of the command and no path that shells out to this binary (R-5.12).
//
// `discover validate` is in this file and NOT in the read-only catalog, and the
// distinction is the whole reason it took an amendment to add. It REWRITES
// `status: draft` to `status: validated` in the brief
// (internal/product/discover.go), so it is a mutation wearing a read-only name,
// and wiring it anywhere that reads as browsing is the exact defect
// read-only-first exists to prevent. This comment used to say it was absent from
// both — "if it is ever offered, it belongs HERE, behind a preview and a
// confirmation, not in a browser." Amendment 5 offered it exactly there. The
// prohibition that carried the safety is untouched and still asserted: no
// browsing screen may reach it.
//
// `prd validate` is here on a weaker claim, and it is worth being honest about
// which: unlike `discover validate`, it mutates nothing at all. It is a form
// rather than a browser because the read-only catalog dispatches no commands — a
// structural property, not a per-command judgement — and the first browsing
// screen to dispatch would be the precedent that erodes it. `(writes)` in that
// one title over-warns, which is the safe direction to be wrong in.

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
			// tui-lifecycle-completion unit 1. This is the screen the file
			// header used to argue against, and the argument still holds — it
			// is why the screen is HERE, in mutatingScreens, titled "(writes)"
			// and behind Preview/Commit, rather than in the discovery browser.
			// `discover validate` rewrites status: draft to status: validated;
			// a browsing screen must never quietly edit what is being browsed.
			//
			// Without it the UI dead-ends: a brief created one screen up cannot
			// reach the "new PRD" picker, which lists validated briefs only.
			Title: "validate discovery brief (writes)",
			Form: &tui.Form{
				Fields: []tui.Field{
					{
						Label:   "brief",
						Choices: draftBriefIDs(ws),
						Help:    briefFieldHelp(draftBriefIDs(ws)),
					},
				},
				Build: func(v []string) (tui.Action, error) {
					// Id only: the team is resolved from the brief's own
					// directory (ux-simplification 1.2, resolveTeam). Passing a
					// team here would add a second answer to a question the
					// filesystem already answers, and the preview line stays
					// short enough for a non-developer to read.
					return newInvocation(ws, &Args{
						Root: root, Cmd: "discover", Action: "validate",
						TitleArg: v[0],
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
			// tui-lifecycle-completion unit 2. Sited immediately after the screen
			// that creates what it validates, for the same reason unit 1 sits
			// after `discover new`: the two are one step of the reader's work,
			// and a checker that lives three screens from the thing it checks is
			// a checker nobody finds.
			//
			// `prd validate` reads the record and reports; it does not rewrite it
			// the way `discover validate` does. It is HERE anyway, not in the PRD
			// browser, because the browser is a listing and the rule that keeps it
			// safe is structural — browsing screens do not dispatch commands at
			// all, so there is no per-command judgement call to get wrong later.
			Title: "validate PRD (writes)",
			Form: &tui.Form{
				Fields: []tui.Field{
					{
						Label:   "prd",
						Choices: activePRDIDs(ws),
						Help:    prdFieldHelp(activePRDIDs(ws)),
					},
				},
				Build: func(v []string) (tui.Action, error) {
					// Id only, as in unit 1: `prd validate` searches every
					// platform's change-records/active/ for the id
					// (ux-simplification 1.2, resolvePlatform), so asking for the
					// platform would be asking the reader a question the tool
					// already answers from the id they just picked.
					return newInvocation(ws, &Args{
						Root: root, Cmd: "prd", Action: "validate", ID: v[0],
					}), nil
				},
			},
		},
		{
			// tui-lifecycle-completion unit 3.
			//
			// Resolved at open time for the same reason `add component` is: a
			// reader who scaffolds a component and then its reality doc in one
			// sitting must see the component they just created, and the choices
			// here also SHRINK as docs are written (R-5.26).
			Title: "new reality doc (writes)",
			FormFn: func() *tui.Form {
				rows := componentCatalog(ws)
				targets := realityTargets(rows)
				current := baseNames(ws.AllPlatforms())
				return &tui.Form{
					Fields: []tui.Field{
						{
							Label:   "platform",
							Choices: current,
							Help:    platformFieldHelp(current),
						},
						{
							Label:   "component",
							Choices: targets,
							Help:    realityFieldHelp(targets),
						},
					},
					Build: func(v []string) (tui.Action, error) {
						return realityInvocation(ws, root, rows, v[0], v[1])
					},
				}
			},
		},
		{
			// tui-lifecycle-completion unit 4 — the last step of a change, and
			// the one that decides whether the previous three amounted to
			// anything.
			//
			// There is NO force field, and there will not be one. `prd complete
			// --force` overrides the done-gate that enforces invariant 4, and a
			// gate that can be waved through from a menu by a reader who does not
			// yet know what it protects is not a gate. The flag stays where using
			// it is a deliberate act: typed, at a terminal. This is not an
			// R-5.10 gap — R-5.10 requires every value the TUI collects to have a
			// flag, not every flag to have a field.
			Title: "complete PRD (writes)",
			// Resolved at open time: this list shrinks as records are completed,
			// and a reader who completes two in one sitting must not be offered
			// the first one again (R-5.26).
			FormFn: func() *tui.Form {
				active := activePRDIDs(ws)
				return &tui.Form{
					Fields: []tui.Field{
						{
							Label:   "prd",
							Choices: active,
							Help:    completeFieldHelp(active),
						},
					},
					Build: func(v []string) (tui.Action, error) {
						return newInvocation(ws, &Args{
							Root: root, Cmd: "prd", Action: "complete", ID: v[0],
						}), nil
					},
				}
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

// completeFieldHelp names what completing does, because it is the one action in
// the catalog whose effects a reader cannot undo from the catalog: the record
// moves to archive/prds/, an outcome review is scheduled, and no screen here
// moves it back.
func completeFieldHelp(active []string) string {
	if len(active) == 0 {
		return "the PRD to complete. This workspace has no active change records."
	}
	return "the PRD to complete. It is archived, an outcome review is scheduled " +
		"for 90 days out, and the done-check refuses while any checklist item " +
		"is unchecked or a reality doc is older than the PRD."
}

// realityTargets lists the components `reality new` can still act on: the ones
// with no reality doc yet.
//
// scaffold.RealityNew refuses to overwrite an existing doc, so offering a
// component that already has one is offering a certain conflict error. Ids are
// NOT deduplicated across platforms — the same id under two platforms is two
// separate documents to write, and the platform field is what tells them apart.
func realityTargets(rows []componentRow) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rows {
		if r.reality || seen[r.id] {
			continue
		}
		seen[r.id] = true
		out = append(out, r.id)
	}
	sort.Strings(out)
	return out
}

// realityFieldHelp says plainly when there is nothing left to scaffold, which
// here is the HEALTHY state rather than the empty one — every component already
// has its reality doc — and must not read like a workspace defect.
func realityFieldHelp(targets []string) string {
	if len(targets) == 0 {
		return "the component whose current state to describe. Every component " +
			"in this workspace already has a reality doc."
	}
	return "the component whose current state to describe. Only components " +
		"without one are offered; `prd complete` refuses while this doc is " +
		"older than the PRD."
}

// realityInvocation refuses a component that belongs to a different platform.
//
// The two pickers are independent — a form has no way to narrow one field from
// another's value — so the catalog can hand Build a pair that exists nowhere.
// That pair does not fail: scaffold.RealityNew resolves the platform directory
// and writes `<chosen-platform>/reality/components/<id>.md` without ever asking
// whether the component lives there, so a mis-picked pair scaffolds a reality
// doc for one platform's component underneath another. Refusing here is
// addInvocation's seam, used for the same reason: the reader stays in the form
// with the reason, and nothing is written.
//
// An id the catalog does not know at all is passed through rather than refused.
// It cannot have come from the picker, so it was typed, and the CLI's own
// handling of an unknown component is the same whoever invoked it.
func realityInvocation(ws *workspace.Workspace, root string, rows []componentRow, platform, component string) (tui.Action, error) {
	var elsewhere []string
	for _, r := range rows {
		if r.id != component {
			continue
		}
		if r.platform == platform {
			elsewhere = nil
			break
		}
		elsewhere = append(elsewhere, r.platform)
	}
	if len(elsewhere) > 0 {
		return nil, fmt.Errorf("%q is a component of %s, not of %q",
			component, strings.Join(elsewhere, ", "), platform)
	}
	return newInvocation(ws, &Args{
		Root: root, Cmd: "reality", Action: "new",
		Platform: platform, ComponentArg: component,
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
// certain to refuse.
//
// (Historical note, now resolved: this used to add "…and `discover validate`,
// the command that would MAKE a brief validated, is not called from anywhere in
// the UI." That was the dead end — a brief created in the TUI could never reach
// this list without leaving for a terminal. The validate screen below closes it.)
func validatedBriefIDs(ws *workspace.Workspace) []string {
	return briefIDsWithStatus(ws, "validated")
}

// draftBriefIDs lists the briefs the validate screen can act on.
//
// The inverse selection matters: offering an already-validated brief would be
// offering a no-op dressed as a choice. `discover validate` on a validated brief
// succeeds and rewrites it byte-identically, so nothing breaks — it just wastes
// the only decision the screen asks for.
func draftBriefIDs(ws *workspace.Workspace) []string {
	return briefIDsWithStatus(ws, "draft")
}

// briefIDsWithStatus is the shared scan. Ids are deduplicated across teams
// because the picker offers an id and `discover validate` resolves the team from
// it (ux-simplification 1.2); two teams holding the same id is the ambiguity
// that command already reports, and it reports it better than a picker could.
func briefIDsWithStatus(ws *workspace.Workspace, status string) []string {
	seen := map[string]bool{}
	var out []string
	for _, tdir := range ws.AllTeams() {
		dir := filepath.Join(tdir, "product", "discovery")
		for _, id := range subdirNames(dir) {
			if seen[id] {
				continue
			}
			if frontmatterField(filepath.Join(dir, id, "brief.md"), "status") == status {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	sort.Strings(out)
	return out
}

// activePRDIDs lists the change records `prd validate` can act on.
//
// Active only: archived records under archive/prds/ have already been through
// `prd complete`, and offering one would offer a check on work that is finished.
// Ids are deduplicated across platforms for the same reason briefIDsWithStatus
// deduplicates across teams — the picker offers an id, resolvePlatform resolves
// the rest, and a genuinely ambiguous id is reported better by that command than
// by a list that cannot say which one it means.
func activePRDIDs(ws *workspace.Workspace) []string {
	seen := map[string]bool{}
	var out []string
	for _, pdir := range ws.AllPlatforms() {
		dir := filepath.Join(pdir, "change-records", "active")
		for _, id := range subdirNames(dir) {
			if seen[id] {
				continue
			}
			seen[id] = true
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

// prdFieldHelp says plainly when there is nothing to validate, following
// platformFieldHelp's rule: a required field with no Choices is
// indistinguishable from a text box, so the help has to carry the news.
func prdFieldHelp(active []string) string {
	if len(active) == 0 {
		return "the active PRD to check. This workspace has none — create one " +
			"with \"new PRD\" first."
	}
	return "the active PRD to check. Validating reports what the record is " +
		"still missing; it changes nothing."
}

// briefFieldHelp says plainly when there is nothing to validate, following
// platformFieldHelp's rule: a required field with no Choices is
// indistinguishable from a text box, so the help has to carry the news.
func briefFieldHelp(drafts []string) string {
	if len(drafts) == 0 {
		return "the draft brief to validate. This workspace has none — create " +
			"one with \"new discovery brief\" first."
	}
	return "the draft brief to validate. Validating fills in " +
		"status: validated, which is what makes it selectable in \"new PRD\"."
}
