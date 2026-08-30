# Browse and scaffold in the TUI

`company-os tui` opens an interactive terminal UI over the workspace you are
standing in. It does two things: it shows you the read-only views without your
having to remember which command produces them, and it fills in the arguments
for the commands that write.

Since 2026-08-29 the second half covers a whole unit of work. You can take a
change from a discovery brief to a completed PRD without typing a command —
which is what the rest of this page is mostly about.

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
component that does not exist. The picker is built when you open the screen, so
it also contains anything you created earlier in the same session.

Some screens hand off rather than render in place: the TUI exits and the command
runs, so you get the real output rather than a reproduction of it.

## The forms that write

Nine screens, each labelled `(writes)` in the menu. They fall into two groups,
and the menu lists them in this order for a reason — read top to bottom, the
first group is the method.

**One change, start to finish:**

| Form | Runs |
| --- | --- |
| new discovery brief `(writes)` | `company-os discover new` |
| validate discovery brief `(writes)` | `company-os discover validate` |
| new PRD `(writes)` | `company-os prd new` |
| validate PRD `(writes)` | `company-os prd validate` |
| new reality doc `(writes)` | `company-os reality new` |
| complete PRD `(writes)` | `company-os prd complete` |

**Growing the federation:**

| Form | Runs |
| --- | --- |
| add team `(writes)` | `company-os add team` |
| add platform `(writes)` | `company-os add platform` |
| add component `(writes)` | `company-os add component` |

`validate PRD` writes nothing at all — it reads the record and reports. It is
labelled `(writes)` and sits with the forms anyway, because the read-only
screens are listings that dispatch no commands whatsoever, and that is a
structural guarantee worth more than one accurate title. Over-warning is the
safe direction.

Every one of them ends the same way: the TUI shows you the complete,
flag-for-flag `company-os` command it is about to run, and does nothing until
you confirm. The preview is derived from the same argument structure that gets
executed, not written out by hand for each screen — so it cannot drift from what
actually runs. Copy it and you have the command to put in a script or a runbook.

`add` is three separate screens rather than one with a kind picker, because only
`add component` takes `--platform`. One combined form would offer that field to
all three kinds and let two of them fail at the end, which is the mistake a form
is supposed to prevent.

### The whole loop, in one sitting

Open the menu once and work down it. Every picker is built when you open its
screen, so each step offers what the step before it created:

1. **new discovery brief** — pick your team, type a title. The brief id is
   derived from the title.
2. **validate discovery brief** — your new brief is in the picker. Validating
   sets `status: validated`, which is what makes it selectable in the next step.
3. **new PRD** — choose the platform and your components, leave the title blank,
   and pick the brief under `from-discovery`. Its Problem signal and Success
   criteria are copied into the PRD.
4. **validate PRD** — reports what the record is still missing.
5. **new reality doc** — only components without one are offered, since
   `reality new` refuses to overwrite.
6. **complete PRD** — archives the record and schedules a 90-day outcome review.

Step 6 will refuse the first time, and that is correct rather than a bug: a new
PRD carries an unchecked governance checklist, and no screen offers to tick it
off. Ticking it off means writing the evidence into `prd.md` in your editor. The
refusal names each reason and prints the command that fixes it.

This mattered enough to be worth a specific fix. Until 2026-08-29 four of these
pickers were built when the TUI *started* rather than when the screen opened, so
a brief created in step 1 was not offered in step 2, and a brief validated in
step 2 was not offered in step 3. Nothing errored — you simply were not shown
your own work, and quitting and relaunching between every step was the only way
through.

The advisor is still a different matter. It is computed once when the TUI opens,
so its suggestions reflect the workspace as it was at launch. Relaunch to
refresh them.

## What the TUI will not do

- **No `workspace sync`, no `scratchpad init`.** Both need values that cannot be
  derived from the workspace — a repository URL and a commit pin, a path outside
  the tree. A form that offers a plausible default for those writes a wrong one,
  which is worse than not offering the form.
- **No `--force` on `complete PRD`, and there will not be one.** `--force`
  overrides the check that a change is not done until reality is updated. A gate
  you can wave through from a menu is not a gate, and the person likeliest to
  reach for it from a menu is the one who least knows what it protects. Use it
  from a terminal, deliberately, or not at all.
- **No writing from a browsing screen.** `discover validate` rewrites the brief
  it is given, so it lives with the forms and never in the discovery browser.
  Nothing in the read-only half of the menu runs a command.
- **No editing.** It scaffolds artifacts and shows you state. Filling in a
  discovery brief or a PRD — and ticking off a governance checklist — is work
  for your editor.
- **No `--json`.** The TUI is for people. Agents and scripts use the commands
  directly, where the structured envelope and the differentiated exit codes are.

## See also

- [`company-os` CLI reference](../reference/company-os-cli.md) — every
  subcommand, including the ones the TUI fronts
- [Take a change from discovery to done](take-a-change-from-discovery-to-done.md)
  — the same lifecycle at the command line, with what each step checks
- [Grow a workspace](grow-a-workspace.md) — what `add` creates, in detail
