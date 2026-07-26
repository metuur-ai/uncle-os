---
id: skill://governance/requesting-an-exception
version: '1.1'
authority: canonical
appliesTo: ['company://all-teams']
inputs: [a mandatory rule that a specific component cannot satisfy]
outputs: [an approved entry in governance/exceptions.yaml with an expiry]
tags: [authority/canonical, kind/skill, process/exception]
---

# Requesting an Exception to a Mandatory Rule

Deviations are for `default` rules. Exceptions are for `mandatory` rules,
and they are loud, temporary, and owned.

**Agents: run every command with `--json` and branch on the exit code.** The
envelope carries `exitCode`, a per-finding `severity`/`code`, and a `guidance`
array holding the next command. Codes are a contract; the English in `message`
is not. Never parse prose. Full envelope:
[reference/company-os-cli.md § `--json`](../../docs/user-guide/reference/company-os-cli.md#--json).

Before drafting, confirm the rule's tier — the choice between an exception and
a deviation is not yours to guess:

```bash
company-os --json governance explain <component-id> \
  | jq -r '.sections[].findings[] | select(.code=="governance.explain-requirement")
           | [.subject, .fields.level, .fields.deviated] | @tsv'
```

`fields.level` is the tier. `mandatory` → an exception (this skill).
`default` → a deviation
(`company-os --json deviation declare "<rule-id>" --team <team>`; read the
review date from the `deviation.review-due` finding's `fields`). `guidance` →
neither is tracked. A deviation aimed at a `mandatory` rule is rejected by
`governance resolve`, so guessing wrong costs you a round trip.

1. (mandatory) Draft it:

   ```bash
   company-os --json exception request "<rule-id>" --team <team> \
      --component <id> --expires <YYYY-MM-DD> --reason "<why>"
   ```

   - `0` — the drafted entry's file and expiry are in the
     `exception.drafted` finding's `fields`. The `exception.approval-note`
     finding is the standing reminder that a drafted exception is not yet a
     valid one; carry it into whatever you report.
   - `2` — `--expires` is missing or not a date. An exception with no expiry
     is not a thing that can exist; do not retry without one.
   - `3` — the team or component does not exist.
   - `4` — `governance/exceptions.yaml` is present but malformed. Fix the YAML
     by hand; the command wrote nothing.

2. (mandatory) Fill `compensatingControls` — what mitigates the risk while
   the exception exists. "None" is not acceptable for mandatory rules.
3. (mandatory) Get sign-off from the rule owner (the platform or Company OS
   that defines the rule) and record it in `approvedBy`. An agent must never
   fill `approvedBy` on someone's behalf.
4. (mandatory) An exception without `expires` is invalid, and an expired one
   fails `company-os validate` in CI. Plan the remediation before the expiry,
   not after the pipeline turns red. To see what is about to expire:

   ```bash
   company-os --json validate \
     | jq -r '.sections[].findings[]
              | select(.code=="expiry.exception-expired"
                    or .code=="expiry.exception-no-expiry"
                    or .code=="expiry.deviation-expired")
              | [.code, .subject] | @tsv'
   ```

   `validate` exits `1` when any of these fire. `expiry.exception-valid` and
   `expiry.deviation-current` are the healthy counterparts — do not treat
   their presence as a problem.
5. (guidance) If the same exception keeps being renewed, the rule may be
   wrong — raise it with the platform architect as a requirement change.
