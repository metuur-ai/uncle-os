# Browse and scaffold in the TUI

`company-os tui` opens an interactive terminal UI over the workspace you are
standing in. It does two things: it shows you the read-only views without your
having to remember which command produces them, and it fills in the arguments
for the handful of commands that scaffold new artifacts.

It is not a second implementation of the CLI. Every screen runs the same code
the equivalent command runs, and every form previews the exact `company-os`
invocation it is about to execute and waits for you to approve it. Nothing is
written before you confirm.

```bash
cd examples/workspace
company-os tui
```

Two things make `tui` behave unlike the other subcommands.

**It needs a real terminal.** Both stdin and stdout must be a TTY. Pipe it,
redirect it, or run it in CI and it refuses with exit `7` and a message naming
which half of your command line to change — plus what to run instead, since
every other subcommand works without a terminal and `validate` gives the same
findings as text.

**It does not require a workspace root.** Every command except `init`,
`scratchpad init`, and this one fails immediately outside a workspace. The TUI
opens anyway, on a recovery menu with a workspace picker; choose a root there
and the catalog rebuilds against it. So `company-os tui` is a reasonable thing
to type when you are not sure where you are.

## Keys

| Key | Does |
| --- | --- |
| `↑` `↓`, or `k` `j` | move the cursor |
| `g` / `Home`, `G` / `End` | jump to first / last entry |
| `Enter` | open the highlighted entry, or advance a form |
| `Esc` | go back one level; quits only when already at the top |
| `Backspace` | go back, when you are choosing from a list |
| `q` | quit, except while typing into a field |
| `Ctrl-C` | quit, always |

`Esc` used to quit from anywhere. It now steps back one level, so a wrong turn
into a browser costs one keystroke instead of the whole session. At the top
level it still quits, because a key that does nothing is worse than a key with
one meaning too many. `Ctrl-C` is the unconditional exit from every mode, and
the footer always names a way out.

`q` is deliberately inert while you are typing — otherwise the letter `q` could
not be typed into a title.

## The read-only screens

Ten of them. Each is a view you could get from a command, listed so you do not
have to recall which:

| Screen | Equivalent command |
| --- | --- |
| workspace overview | — |
| today (role view) | `company-os today --role <role>` |
| validate results | `company-os validate` |
| component browser | — |
| PRD browser | — |
| discovery browser | — |
| governance explain | `company-os governance explain <component>` |
| skills list | `company-os skills list` |
| ids list | `company-os ids list` |
| workspace status | `company-os workspace status` |

Screens that need an argument — a role, a component id — ask for it with a
picker built from what the workspace actually contains, so you cannot choose a
component that does not exist.

Some screens hand off rather than render in place: the TUI exits and the command
runs, so you get the real output rather than a reproduction of it.

## The forms that write

Five screens scaffold artifacts. Each is labelled `(writes)` in the menu:

| Form | Runs |
| --- | --- |
| new discovery brief `(writes)` | `company-os discover new` |
| new PRD `(writes)` | `company-os prd new` |
| add team `(writes)` | `company-os add team` |
| add platform `(writes)` | `company-os add platform` |
| add component `(writes)` | `company-os add component` |

Every one of them ends the same way: the TUI shows you the complete,
flag-for-flag `company-os` command it is about to run, and does nothing until
you confirm. The preview is derived from the same argument structure that gets
executed, not written out by hand for each screen — so it cannot drift from what
actually runs. Copy it and you have the command to put in a script or a runbook.

`add` is three separate screens rather than one with a kind picker, because only
`add component` takes `--platform`. One combined form would offer that field to
all three kinds and let two of them fail at the end, which is the mistake a form
is supposed to prevent.

### Creating a platform and then a component in it

This sequence works, and it is worth knowing why it is called out. The
`add component` form's platform picker is built when you open that screen, not
when the TUI starts. So a platform you created a moment ago in the same session
is already offerable:

1. Open **add platform `(writes)`**, name it, confirm.
2. Open **add component `(writes)`** — the new platform is in the list.

Without this, the picker would show you everything except the platform you just
made, and quitting and relaunching would be the only way forward.

The advisor is a different matter. It is computed once when the TUI opens, so
its suggestions reflect the workspace as it was at launch. Relaunch to refresh
them.

## What the TUI will not do

- **No `workspace sync`, no `scratchpad init`.** Both need values that cannot be
  derived from the workspace — a repository URL and a commit pin, a path outside
  the tree. A form that offers a plausible default for those writes a wrong one,
  which is worse than not offering the form.
- **No editing.** It scaffolds artifacts and shows you state. Filling in a
  discovery brief or a PRD is work for your editor.
- **No `--json`.** The TUI is for people. Agents and scripts use the commands
  directly, where the structured envelope and the differentiated exit codes are.

## See also

- [`company-os` CLI reference](../reference/company-os-cli.md) — every
  subcommand, including the ones the TUI fronts
- [Take a change from discovery to done](take-a-change-from-discovery-to-done.md)
  — the lifecycle the two writing forms start
- [Grow a workspace](grow-a-workspace.md) — what `add` creates, in detail
