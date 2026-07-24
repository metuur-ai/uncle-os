# Example: `banking/` — multi-repo & multi-user pattern catalog

**What it demonstrates:** the nine collaboration patterns (P1-P9) from the
research doc `.devlocal/research/2026-07-22-multi-repo-multi-user-use-cases.md`,
staged as two org profiles built from the **same substrate**:

- `small-company/` — ~10 people, one monorepo, no federation. See its
  `EXAMPLE_README.md`.
- `bank/` — ~250-person bank: simulated source repos (`bank/repos/`),
  federated team workspace roots (`bank/workspaces/`), and a proposed
  cross-platform initiative artifact (`bank/initiatives/`). See its
  `EXAMPLE_README.md`.

**Start at `README.md`** in this directory — it maps every pattern P1-P9 to the
exact files that demonstrate it, and lists what is real vs illustrative
(placeholder commit pins, no fake lock hashes, sync slices not materialized).

Unlike `../workspace/` and `../federated/`, this example is documentation-
oriented: it is not wired into `acceptance.sh` and is not meant to pass
`company-os validate` as-is (the federated workspaces would first need a real
`workspace sync` against real repos).
