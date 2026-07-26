# failing-federated-nolock — the "lock is missing" gate `[8/8]` fixture

**Deliberately broken; expected to exit 1.** Do not copy it as a template.

One file: a `workspace.yaml` with no `workspace.lock.yaml` beside it.

`federated_slice_problems()` returns early when the lock is absent or malformed,
emitting exactly one finding and skipping every per-repo and per-file check.
That early return is why this shape cannot share a fixture with the four
findings in `examples/failing-federated/` — hence the separate directory for a
single message.

It also freezes the most minimal `validate` render there is: eight gate headers,
seven of them with no findings at all, and a `FAIL — 1 problem(s)` footer.

Running `company-os workspace sync` here would fix the fixture by writing the
lock. Don't.
