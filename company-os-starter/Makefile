# company-os — Go CLI
#
# The Go module is rooted here; acceptance.sh and the fixture workspaces live in
# ../examples/ because they predate the port. Targets that need both reach up one
# level. There is no longer a Python reference to drive: R-9.3 deleted it, and
# the differential harness that compared against it was retired with it.
#
#   make            # build
#   make dev        # rebuild + run against the example workspace
#   make check      # everything CI would run
#   make install    # copy the binary to ~/.local/bin
#
# Run `make help` for the full list.

SHELL := /bin/bash
.DEFAULT_GOAL := build

# --- paths ---------------------------------------------------------------
MODULE     := github.com/metuur-ai/uncle-os/company-os-starter
BINARY     := company-os
PKG        := ./cmd/company-os
REPO_ROOT  := $(abspath ..)
EXAMPLES   := $(REPO_ROOT)/examples
BIN_OUT    := $(CURDIR)/$(BINARY)
PREFIX     ?= $(HOME)/.local
DIST       := $(CURDIR)/dist

# The workspace most targets operate on. Override to point elsewhere:
#   make dev WS=../examples/federated
WS         ?= $(EXAMPLES)/workspace

# --- version stamping ----------------------------------------------------
# R-6.8: a copy-distributed binary must be able to say what it is.
#
# Development builds describe the tree: v0.1.1-84-gce851f6 says "84 commits past
# v0.1.1", which is what you want from `make build` on a work branch.
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)

# RELEASE_VERSION is what `make release` stamps, and it is NOT the same thing.
#
# `git describe` reports the last tag REACHED, so on an untagged tree it yields
# v0.1.1-84-gce851f6 while install.sh and every doc say v1.1.2 — artifacts that
# contradict the instructions shipped beside them. The release-tag literals are
# the declaration of intent (`make bump` writes them, `version-check` gates
# them), so they are what a release carries.
#
# Precedence: an explicit VERSION= on the command line, else the declared
# literals, else `git describe` when the sites cannot be read.
DECLARED    = $(shell scripts/bump-version.sh --show 2>/dev/null \
                | awk '/^Current: v/{print $$2}')
ifeq ($(origin VERSION),command line)
RELEASE_VERSION := $(VERSION)
else
RELEASE_VERSION  = $(or $(DECLARED),$(VERSION))
endif

# Stamped into internal/model, not main: R-3.5 puts the same identifier in every
# --json payload, and the JSON encoder is an internal package. See buildinfo.go.
LDFLAGS    := -s -w \
              -X $(MODULE)/internal/model.version=$(VERSION) \
              -X $(MODULE)/internal/model.commit=$(COMMIT)
RELEASE_LDFLAGS = -s -w \
              -X $(MODULE)/internal/model.version=$(RELEASE_VERSION) \
              -X $(MODULE)/internal/model.commit=$(COMMIT)

# R-6.2: static linking is what makes "no runtime dependency" true rather
# than aspirational.
GOFLAGS_STATIC := CGO_ENABLED=0

# R-6.10: checksums are only worth publishing if the same source produces the
# same bytes. Without -trimpath the compiler records the absolute path of the
# checkout in the pclntab, which -s -w does not strip: two clones of the same
# commit at different paths yield different artifacts and the checksum stops
# meaning "this is that commit". Measured, not assumed — see `make repro`.
# There is deliberately no build date; buildinfo.go explains why.
BUILDFLAGS := -trimpath

# --- macOS: no signing, deliberately (R-6.3) ------------------------------
# There are no sign/notarize targets and no Apple Developer ID is required.
# `com.apple.quarantine` is set by the DOWNLOADING APP (Safari, Chrome, Mail),
# never by curl/wget/tar, and Gatekeeper only adjudicates files carrying it —
# so a binary fetched by install.sh runs unsigned, with no prompt. Publish the
# install line, never a download link. Full reasoning, and what was measured
# before signing was ruled out, in
# docs/user-guide/how-to/release-and-upgrade.md.

# R-6.2: release targets. Windows is deliberately absent (R-6.11) — Go
# cross-compiles to it for free, but nothing here is tested there, so nothing
# is claimed.
PLATFORMS  := darwin/arm64 darwin/amd64 linux/amd64

.PHONY: help build install uninstall run dev test test-v cover fmt fmt-check \
        vet lint version-check version-show bump bump-build check acceptance \
        golden clean dist release tidy deps-check arch-test repro

# --- help ----------------------------------------------------------------
help: ## Show this help
	@echo "company-os Go CLI — $(VERSION) ($(COMMIT))"
	@echo
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
	@echo
	@echo "Variables: WS=<workspace> PREFIX=<dir> VERSION=<v> BUMP=<major|minor|patch|vX.Y.Z>"

# --- build ---------------------------------------------------------------
build: ## Build the binary into ./company-os
	$(GOFLAGS_STATIC) go build $(BUILDFLAGS) -ldflags '$(LDFLAGS)' -o $(BIN_OUT) $(PKG)
	@echo "built $(BIN_OUT) ($(VERSION))"

# Upgrade-in-place is the common case (R-6.9), so the copy lands on a sibling
# path and is renamed over the old binary. rename(2) is atomic — no window
# where a half-written binary is on PATH — and it unlinks rather than truncates
# the old inode, which is what an in-place `cp` cannot do on Linux while the
# old binary is still executing (ETXTBSY). Darwin permits the in-place write;
# Linux is the platform this protects, and it is a release target.
install: build ## Install to $(PREFIX)/bin (default ~/.local/bin)
	@mkdir -p $(PREFIX)/bin
	@cp $(BIN_OUT) $(PREFIX)/bin/.$(BINARY).new
	@chmod 0755 $(PREFIX)/bin/.$(BINARY).new
	@mv -f $(PREFIX)/bin/.$(BINARY).new $(PREFIX)/bin/$(BINARY)
	@echo "installed -> $(PREFIX)/bin/$(BINARY) ($(VERSION))"
	@command -v $(BINARY) >/dev/null 2>&1 \
	  || echo "note: $(PREFIX)/bin is not on PATH"

uninstall: ## Remove the installed binary
	@rm -f $(PREFIX)/bin/$(BINARY)
	@echo "removed $(PREFIX)/bin/$(BINARY)"

tidy: ## go mod tidy (never `go mod vendor` — see go.work)
	go mod tidy

# --- run / dev -----------------------------------------------------------
run: build ## Build, then run: make run ARGS="validate"
	@$(BIN_OUT) --root $(WS) $(ARGS)

dev: build ## Build and validate the example workspace
	@$(BIN_OUT) --root $(WS) validate; \
	  echo "exit: $$?"

# --- test ----------------------------------------------------------------
test: ## Run the Go test suite
	go test ./...

test-v: ## Run the Go test suite, verbose
	go test -v ./...

cover: ## Test with coverage, open the HTML report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

arch-test: ## Assert no os.Exit / stdout writes below cmd/ (R-2.10)
	go test -run TestArchitecture ./cmd/... -v

# --- quality -------------------------------------------------------------
fmt: ## Format
	gofmt -w -s .

fmt-check: ## Fail if anything is unformatted
	@out=$$(gofmt -l -s .); \
	  if [ -n "$$out" ]; then echo "unformatted:"; echo "$$out"; exit 1; fi
	@echo "gofmt: clean"

vet: ## go vet
	go vet ./...

lint: fmt-check vet ## fmt-check + vet

# The release tag is stamped into the binary from git, but it is also written by
# hand into the CI-pinning examples in docs/ and into install.sh's usage comment.
# Those copies drift silently — install.sh sat at v0.4.0 while docs said v1.0.0.
version-check: ## Fail if the hardcoded release-tag literals disagree
	@scripts/bump-version.sh --check

version-show: ## Show the current release tag and every site that carries it
	@scripts/bump-version.sh --show

# `bump` rewrites the hardcoded literals; it does NOT tag and does NOT change
# what the binary reports. VERSION comes from `git describe` at build time, so
# a freshly bumped tree still stamps the OLD tag until `git tag` is run. That is
# why `bump-build` passes the new value through explicitly — it is the only way
# to get a binary carrying a version that has no tag yet, which is exactly what
# you want when testing a release before cutting it.
#
#   make bump                # patch: v1.1.0 -> v1.1.1
#   make bump BUMP=minor     # v1.1.0 -> v1.2.0
#   make bump BUMP=v2.0.0    # explicit
BUMP ?= patch

bump: ## Bump the release-tag literals (BUMP=major|minor|patch|vX.Y.Z, default patch)
	@scripts/bump-version.sh $(BUMP)
	@scripts/bump-version.sh --check

# Bump then build the matrix, WITHOUT tagging or pushing — `make dist`, not
# `make release`. This is the rehearsal: it proves the new version cross-compiles
# and stamps correctly before anything public happens. Publishing is `make
# release`, which needs the bump committed first (a tag must point at a commit
# that contains it).
bump-build: bump ## Bump, then cross-compile into ./dist at the new version (no tag, no push)
	@$(MAKE) --no-print-directory dist || exit 1
	@new=$$(scripts/bump-version.sh --show | awk '/^Current: /{print $$2}'); \
	  echo; \
	  echo "next: verify, commit, publish"; \
	  echo "  ./dist/$(BINARY)-$$(go env GOOS)-$$(go env GOARCH) --version"; \
	  echo "  git commit -am 'release $$new' && git push"; \
	  echo "  make release        # gate, rebuild, tag, push"

# --- behaviour suite -----------------------------------------------------
# acceptance.sh freezes `validate` output against golden snapshots, on the
# passing path and the failure path, and proves the generated artifacts are
# idempotent. Re-baselining is a review step, not a formality: a golden accepted
# without reading the diff protects nothing.
#
# There was briefly a second suite here, `make differential`, over a
# 288-invocation corpus in internal/difftest. It was removed — see task 6.10 in
# docs/tasks/go-cli-tui-port.md for the measurement that decided it and for what
# is consequently NOT covered.

acceptance: ## Golden-snapshot suite over `validate` (examples/acceptance.sh)
	@$(EXAMPLES)/acceptance.sh

golden: ## Re-baseline the acceptance snapshots — review the diff before committing
	@$(EXAMPLES)/acceptance.sh --update

check: lint version-check test acceptance ## Everything CI runs

# --- release -------------------------------------------------------------
# Artifact names carry no version: `/releases/latest/download/<name>` needs a
# fixed name, and install.sh is built on that URL. The version lives in the
# release tag and is stamped into the binary — `company-os --version` reports it.
#
# The stamp is RELEASE_VERSION (the declared literals), not VERSION (`git
# describe`). See the version-stamping block: a release must not tell a user
# v0.1.1-84-gce851f6 while the install.sh beside it says v1.1.2.
dist: ## Cross-compile release artifacts + checksums into ./dist (no tag, no push)
	@if [ -z "$(RELEASE_VERSION)" ]; then \
	  echo "dist: cannot determine a version to stamp" >&2; exit 1; fi
	@scripts/bump-version.sh --check >/dev/null 2>&1 \
	  || { echo "dist: release-tag literals disagree — run 'make version-show'" >&2; \
	       scripts/bump-version.sh --check; exit 1; }
	@rm -rf $(DIST) && mkdir -p $(DIST)
	@for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; \
	  out=$(DIST)/$(BINARY)-$${os}-$${arch}; \
	  echo "  $$os/$$arch"; \
	  $(GOFLAGS_STATIC) GOOS=$$os GOARCH=$$arch \
	    go build $(BUILDFLAGS) -ldflags '$(RELEASE_LDFLAGS)' -o $$out $(PKG) || exit 1; \
	done
	@cp install.sh $(DIST)/install.sh
	@cd $(DIST) && shasum -a 256 * > SHA256SUMS
	@echo
	@echo "artifacts in $(DIST) ($(RELEASE_VERSION)):"
	@ls -1 $(DIST)
	@echo
	@echo "NOT SIGNED, NOT NOTARIZED — deliberately (R-6.3). Ship install.sh as the"
	@echo "install path: com.apple.quarantine is set by the DOWNLOADING APP (Safari,"
	@echo "Chrome, Mail), not by curl/wget/tar, so a curl-fetched binary runs with no"
	@echo "Gatekeeper prompt. A browser download of these same bytes HANGS silently."
	@echo "Publish the install line, not a download link. See"
	@echo "docs/user-guide/how-to/release-and-upgrade.md."

# `release` publishes. It runs the gate, builds the matrix, tags, and pushes the
# tag — which triggers .github/workflows/release.yml, which rebuilds from the
# tag and uploads the artifacts to GitHub. The local dist/ is the rehearsal; CI
# produces the bytes users download.
#
# Pushing a tag is public and effectively irreversible, so everything that can
# refuse does so BEFORE the push: the gate, a dirty tree, a tag that already
# exists, an unpushed HEAD. Use `make dist` to build without publishing.
#
#   make release              # gate, build, tag, push
#   make release SKIP_GATE=1  # same, without re-running make check
#   make dist                 # artifacts only
release: ## Gate, build, tag, and push — publishes via CI
	@v="$(RELEASE_VERSION)"; \
	if [ -z "$$v" ]; then echo "release: no version to stamp" >&2; exit 1; fi; \
	if git rev-parse -q --verify "refs/tags/$$v" >/dev/null 2>&1; then \
	  echo "release: tag $$v already exists." >&2; \
	  echo "  Bump first:  make bump           (or make bump BUMP=minor)" >&2; \
	  exit 1; fi; \
	if [ -n "$$(git status --porcelain)" ]; then \
	  echo "release: working tree is dirty. Commit first — a tag must point at" >&2; \
	  echo "         a commit that contains what you are shipping." >&2; \
	  git status --short >&2; exit 1; fi; \
	up=$$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null); \
	if [ -z "$$up" ]; then \
	  echo "release: branch has no upstream; push it first." >&2; exit 1; fi; \
	if [ -n "$$(git rev-list @{upstream}..HEAD 2>/dev/null)" ]; then \
	  echo "release: HEAD is ahead of $$up. Push your commits first, or CI will" >&2; \
	  echo "         build a tag whose commit the remote does not have." >&2; exit 1; fi; \
	echo "releasing $$v from $$(git rev-parse --short HEAD) on $$up"; \
	echo
	@if [ -z "$(SKIP_GATE)" ]; then \
	  echo "== gate =="; $(MAKE) --no-print-directory check || exit 1; echo; \
	else echo "== gate skipped (SKIP_GATE set) =="; echo; fi
	@echo "== build =="
	@$(MAKE) --no-print-directory dist || exit 1
	@echo
	@echo "== publish =="
	@v="$(RELEASE_VERSION)"; \
	  git tag -a "$$v" -m "release $$v" && echo "  tagged $$v" && \
	  git push origin "$$v" && echo "  pushed $$v" && \
	  echo && \
	  echo "CI is building and will attach the artifacts to the release:" && \
	  echo "  https://github.com/metuur-ai/uncle-os/actions" && \
	  echo "  https://github.com/metuur-ai/uncle-os/releases/tag/$$v"

repro: ## R-6.10: build the release matrix twice and prove the checksums match
	@tmp=$$(mktemp -d) || exit 1; \
	  echo "pass 1 (warm cache)"; \
	  $(MAKE) --no-print-directory dist >/dev/null || exit 1; \
	  cp $(DIST)/SHA256SUMS $$tmp/first; \
	  echo "pass 2 (cold cache — GOCACHE=$$tmp/cache)"; \
	  GOCACHE=$$tmp/cache $(MAKE) --no-print-directory dist >/dev/null || exit 1; \
	  if diff -q $$tmp/first $(DIST)/SHA256SUMS >/dev/null; then \
	    echo; echo "reproducible — identical across a cold-cache rebuild:"; \
	    sed 's/^/  /' $(DIST)/SHA256SUMS; rc=0; \
	  else \
	    echo; echo "NOT reproducible:"; diff -u $$tmp/first $(DIST)/SHA256SUMS; rc=1; \
	  fi; \
	  rm -rf $$tmp; exit $$rc

# What this can and cannot prove, stated plainly, because the previous version
# of this target inspected the *local host build* and reported "no dynamic
# dependencies" for a matrix it had never looked at.
#
#   `go version -m` reads the build metadata out of the artifact itself. It is
#   arch- and OS-independent, so it is the only check here that works on all
#   three artifacts from one machine, and it is what proves CGO_ENABLED=0 and
#   -trimpath. It does NOT prove the version stamp: cmd/go omits the -ldflags
#   setting from build info whenever -trimpath is set, because ldflags may
#   carry absolute paths. So the stamp is checked against the artifact's own
#   bytes instead, and executed outright on the one artifact this host can run.
#
#   `otool -L` parses Mach-O and therefore covers both darwin artifacts from
#   either darwin host. It cannot read ELF: run against the linux artifact it
#   says "is not an object file", which is not a pass.
#
#   `ldd` exists only on Linux. On a darwin release host the linux artifact's
#   linkage is checked structurally via file(1) and confirmed for real only on
#   task 5.3's clean Linux box.
deps-check: dist ## Prove each release artifact needs nothing at runtime (R-6.1, R-6.2, R-6.8)
	@fail=0; \
	for f in $(DIST)/$(BINARY)-*; do \
	  echo "== $$(basename $$f)"; \
	  meta=$$(go version -m $$f 2>/dev/null); \
	  echo "$$meta" | grep -qE 'CGO_ENABLED=0' \
	    && echo "   ok    CGO_ENABLED=0" \
	    || { echo "   FAIL  not built with CGO_ENABLED=0"; fail=1; }; \
	  echo "$$meta" | grep -qE '\-trimpath=true' \
	    && echo "   ok    -trimpath (R-6.10: path-independent bytes)" \
	    || { echo "   FAIL  built without -trimpath"; fail=1; }; \
	  grep -qa '$(VERSION)' $$f \
	    && echo "   ok    version string '$(VERSION)' present in the image (R-6.8)" \
	    || { echo "   FAIL  -X model.version did not reach the linker"; fail=1; }; \
	  if [ "$$(basename $$f)" = "$(BINARY)-$$(go env GOOS)-$$(go env GOARCH)" ]; then \
	    got=$$($$f --version 2>&1); \
	    case "$$got" in \
	      *"$(VERSION)"*) echo "   ok    executes and reports: $$got" ;; \
	      *) echo "   FAIL  --version says: $$got"; fail=1 ;; \
	    esac; \
	  else echo "         (not host-native — cannot exec here; --version verified on 5.3's clean boxes)"; fi; \
	  case $$f in \
	    *-darwin-*) \
	      libs=$$(otool -L $$f 2>/dev/null | tail -n +2 | awk '{print $$1}'); \
	      extra=$$(echo "$$libs" | grep -v '^/usr/lib/libSystem\.B\.dylib$$' \
	                             | grep -v '^/usr/lib/libresolv\.9\.dylib$$' | grep -v '^$$'); \
	      if [ -n "$$extra" ]; then echo "   FAIL  links non-OS libraries:"; echo "$$extra" | sed 's/^/         /'; fail=1; \
	      else echo "   ok    links only OS-provided dylibs (libSystem, libresolv)"; fi; \
	      echo "         note: Apple ships no static libSystem — a fully static Mach-O"; \
	      echo "         does not exist. Both dylibs are present on every macOS." ;; \
	    *-linux-*) \
	      file $$f | grep -q 'statically linked' \
	        && echo "   ok    ELF reports 'statically linked'" \
	        || { echo "   FAIL  ELF is not statically linked: $$(file $$f)"; fail=1; }; \
	      if command -v ldd >/dev/null 2>&1; then ldd $$f 2>&1 | sed 's/^/   ldd: /'; \
	      else echo "         (ldd unavailable on this host — confirmed on the clean Linux box, task 5.3)"; fi ;; \
	  esac; \
	done; \
	exit $$fail

clean: ## Remove build output
	@rm -rf $(BIN_OUT) $(DIST) coverage.out
	@go clean -testcache
	@echo "cleaned"
