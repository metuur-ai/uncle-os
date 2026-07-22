---
id: skill://governance/requesting-an-exception
version: '1.0'
authority: canonical
appliesTo: ['company://all-teams']
inputs: [a mandatory rule that a specific component cannot satisfy]
outputs: [an approved entry in governance/exceptions.yaml with an expiry]
tags: [authority/canonical, kind/skill, process/exception]
---

# Requesting an Exception to a Mandatory Rule

Deviations are for `default` rules. Exceptions are for `mandatory` rules,
and they are loud, temporary, and owned.

1. (mandatory) Draft it:
   `company-os exception request "<rule-id>" --team <team> \
      --component <id> --expires <YYYY-MM-DD> --reason "<why>"`
2. (mandatory) Fill `compensatingControls` — what mitigates the risk while
   the exception exists. "None" is not acceptable for mandatory rules.
3. (mandatory) Get sign-off from the rule owner (the platform or Company OS
   that defines the rule) and record it in `approvedBy`.
4. (mandatory) An exception without `expires` is invalid, and an expired one
   fails `company-os validate` in CI. Plan the remediation before the expiry,
   not after the pipeline turns red.
5. (guidance) If the same exception keeps being renewed, the rule may be
   wrong — raise it with the platform architect as a requirement change.
