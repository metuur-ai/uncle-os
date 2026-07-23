# What `workspace sync` would materialize here

This sample commits only the team's own writable content. Running
`company-os workspace sync` against real repos would additionally materialize
read-only slices from `../../repos/`:

    company-os/            <- bank/company-os.git        (pin: v2026.2)
    company-ontology/      <- bank/company-ontology.git  (pin: v2026.2)
    platforms/fraud/       <- bank/platform-fraud.git
    platforms/payments/    <- bank/platform-payments.git
    platforms/cards/       <- bank/platform-cards.git
    platforms/identity/    <- bank/platform-identity.git

and write `workspace.lock.yaml` with resolved commits + per-file slice hashes
(validate gate [8/8] then enforces byte-for-byte drift). `workspace status`
reports pin/lock/slice drift per repo. Lock hashes cannot be faked here, so no
illustrative lock file is committed.
