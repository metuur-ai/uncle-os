# Example: `federated/` — a real multi-repo workspace fixture (Phase 4, Option B)

**What it demonstrates:** the federation machinery with **real, verifiable
hashes**. The `communications` platform is declared in `workspace.yaml` as a
read-only governance slice pulled from a separate repo, pinned by commit
(`714635868b00…`). `workspace.lock.yaml` records the resolved commit, a slice
hash, and per-file sha256 hashes. The materialized slice is committed alongside
the lock, so `company-os validate` passes gate **[8/8]** on a fresh checkout
with **no network and no source repo** — it compares the committed slice bytes
to the committed lock hashes.

**Key files:**
- `workspace.yaml` — manifest: repo `platform-communications`, `root:
  platforms/communications`, pinned `paths:` (governance/, components/,
  reality/, skills/, templates/).
- `workspace.lock.yaml` — written by `workspace sync`; never hand-edit.
- Everything else mirrors `../workspace/` (same team, ontology, baseline) so
  you can diff the two examples to see exactly what federation adds.

**Try it:**
```bash
export PATH="$PWD/../../bin:$PATH"
company-os validate            # includes the federation drift gate [8/8]
company-os workspace status    # per-repo pin / lock / slice drift
company-os workspace sync --frozen   # lock-only, refuses network
```
Contrast with `../banking/bank/workspaces/`, where pins are illustrative and
no lock is committed (hashes there cannot be faked).
