# failing-federated — the gate `[8/8]` failure fixture

**Deliberately broken; expected to exit 1.** Do not copy it as a template.

Gate `[8/8]` only exists when a `workspace.yaml` is present, so it cannot be
reached from `examples/failing-workspace/`. This fixture drives four of the five
findings `federated_slice_problems()` can produce, in a single `validate` run,
against a committed manifest/lock pair that is inconsistent on purpose. No git,
no network, and no source repo are involved — the gate compares committed slice
bytes to committed lock hashes.

| Finding | How it is planted |
|---|---|
| slice set differs between manifest and lock | `workspace.yaml` allowlists `governance/` **and** `components/`; the lock records only `governance/` |
| recorded slice file is hand-edited | the lock's hash for `platforms/sliced-alpha/governance/requirements.yaml` is all zeroes, so it can never match the committed bytes |
| recorded slice file is missing | the lock records `platforms/sliced-alpha/components/svc-sliced.yaml`, which is not committed |
| manifest repo has no lock entry | `never-synced` appears in `workspace.yaml` and nowhere in `workspace.lock.yaml` |

The fifth finding — manifest present but lock missing or malformed — makes
`federated_slice_problems()` return early, so it can never co-occur with these
four. It has its own fixture: `examples/failing-federated-nolock/`.

Gates 1–7 are near-silent here by design. That is itself part of the snapshot:
it freezes how an empty gate renders (a header, a blank line, no findings) and
confirms all seven headers renumber to `/8` once a manifest exists.
