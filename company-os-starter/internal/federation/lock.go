package federation

// The lock: drift detection and the validate gate (bin/company-os:2475-2526).
//
// workspace.lock.yaml is machine-owned and committed. It records, per repo, the
// original pin, the commit the pin resolved to at sync time, the resolved slice
// list, an aggregate hash over the UNION of that repo's slices, and a
// {relpath: sha256} map that is the hand-edit oracle (GPF-R-7.5).

import (
	"fmt"
	"os"
	"path"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// GateSlug and GateTitle are gate 8's identity. The header text is frozen by
// examples/federated-golden-validate.txt, so it lives beside the producer.
const (
	GateSlug  = "federated-slice-integrity"
	GateTitle = "federated slice integrity (read-only derived content)"
)

// State is slice_state's three-way answer (bin/company-os:2475-2487).
type State string

const (
	// StateClean means every file the lock records is present and hashes to the
	// recorded value.
	StateClean State = "clean"
	// StateMissing means the lock records no files at all, or one of them is
	// gone from disk.
	StateMissing State = "missing"
	// StateDrifted means a recorded file is present with different bytes.
	StateDrifted State = "drifted"
)

// Lock is a loaded workspace.lock.yaml. Usable is false when the file is
// absent, empty, or does not carry a `repos:` list — the three states the
// oracle collapses into one problem.
type Lock struct {
	Data   yamlio.PyMap
	Repos  []yamlio.PyMap
	Usable bool
}

// ByName indexes the lock's repo entries. The key is repr(name) rather than
// str(name) so that a lock written with `name: 1` cannot be matched by a
// manifest saying `name: '1'`, which is what Python's dict lookup does.
func (l *Lock) ByName(name yamlio.PyValue) (yamlio.PyMap, bool) {
	want := yamlio.PyRepr(name)
	for _, r := range l.Repos {
		if n := r.Get("name"); n != nil && yamlio.PyRepr(n) == want {
			return r, true
		}
	}
	return nil, false
}

// LoadLock is `load_yaml(ws.root / LOCK_NAME, None)` plus the `repos:` shape
// test every caller applies to the result.
func LoadLock(ws *workspace.Workspace) (*Lock, error) {
	raw, err := yamlio.PyLoadFile(LockPath(ws))
	if err != nil {
		return nil, err
	}
	l := &Lock{}
	data, ok := raw.(yamlio.PyMap)
	if !ok {
		return l, nil
	}
	l.Data = data
	repos, ok := data.Get("repos").(yamlio.PySeq)
	if !ok {
		return l, nil
	}
	l.Usable = true
	for _, e := range repos {
		if m, ok := e.(yamlio.PyMap); ok {
			l.Repos = append(l.Repos, m)
		}
	}
	return l, nil
}

// LockFiles is `lr.get("files") or {}` — the recorded {relpath: sha256} map, in
// the lock's DOCUMENT order.
//
// TRAP. That order is the lock's emission order arriving back through the
// parser, which is a nested walk of manifest slice order, then paths: order,
// then sorted(rglob) — the reverse of alphabetical for every committed fixture.
// Gate 8 renders its [FAIL] lines in exactly this order and
// examples/failing-federated-golden-validate.txt freezes it. PyMap preserves it;
// a Go map would not.
func LockFiles(lr yamlio.PyMap) yamlio.PyMap {
	files, _ := lr.Get("files").(yamlio.PyMap)
	return files
}

// SliceState is slice_state (bin/company-os:2475-2487): 'clean' | 'drifted' |
// 'missing' by comparing on-disk slice bytes to the hashes in the lock.
//
// It reads ONLY the `files:` key. That contract is what keeps --only per-repo
// and gate 8 cheap, and selftest ST-071 pins it by passing an entry with no
// slices:, no url: and no pin:.
func SliceState(ws *workspace.Workspace, lr yamlio.PyMap) (State, error) {
	files := LockFiles(lr)
	if len(files) == 0 {
		return StateMissing, nil
	}
	for _, p := range files {
		abs := path.Join(ws.Root, p.K)
		fi, err := os.Stat(abs)
		if err != nil || !fi.Mode().IsRegular() {
			return StateMissing, nil
		}
		sum, err := sha256File(abs)
		if err != nil {
			return StateMissing, nil
		}
		if sum != yamlio.PyString(p.V) {
			return StateDrifted, nil
		}
	}
	return StateClean, nil
}

// Integrity is everything gate 8 needs, and nothing about how it reads.
// internal/validate wraps it in a GateResult with the ordinal it computed; this
// package does not know whether it is gate 8 of 8 or a TUI panel.
type Integrity struct {
	// Findings is in render order and already carries the SevOK line when the
	// slices are intact. Empty is not a valid state.
	Findings []model.Finding
	// Files is the number of recorded files checked, and Repos the number of
	// manifest repos — the two counts the clean line reports.
	Files int
	Repos int
}

// SliceFindings is federated_slice_problems (bin/company-os:2490-2526),
// decomposed.
//
// Python returns finished English sentences; R-2.12 requires facts. Each of the
// five problem shapes gets its own code and its own typed Fields, and Message
// below is the only thing that turns either into prose.
//
// Only called when a manifest is present, so it never runs in monorepo mode —
// the gate self-suppresses and the monorepo golden stays byte-identical.
func SliceFindings(ws *workspace.Workspace, m *Manifest) (Integrity, error) {
	out := Integrity{Repos: len(m.Repos)}
	lock, err := LoadLock(ws)
	if err != nil {
		return out, err
	}
	if !lock.Usable {
		out.Findings = append(out.Findings, finding(model.CodeSliceLockMissing,
			model.Fields{"manifest": workspace.ManifestName, "lock": workspace.LockName}))
		return out, nil
	}
	for _, repo := range m.Repos {
		nameVal := repo.Get("name")
		name := yamlio.PyString(nameVal)
		lr, ok := lock.ByName(nameVal)
		if !ok {
			out.Findings = append(out.Findings, finding(model.CodeRepoNotLocked,
				model.Fields{"repo": name, "manifest": workspace.ManifestName,
					"lock": workspace.LockName}))
			continue
		}
		// Slice-set drift: move a localDirectory (or add a slice) without
		// re-syncing and the OLD files still exist and still hash-match, so the
		// per-file loop below reports clean. Compare the sets explicitly. This is
		// the case nothing else catches.
		slices, err := RepoSlices(repo)
		if err != nil {
			return out, err
		}
		if SliceKey(slices) != SliceKeyOf(lr.Get("slices")) {
			out.Findings = append(out.Findings, finding(model.CodeSliceSetDrift,
				model.Fields{"repo": name, "manifest": workspace.ManifestName,
					"lock": workspace.LockName}))
		}
		for _, p := range LockFiles(lr) {
			out.Files++
			abs := path.Join(ws.Root, p.K)
			fields := model.Fields{"path": p.K, "repo": name,
				"lock": workspace.LockName}
			fi, statErr := os.Stat(abs)
			if statErr != nil || !fi.Mode().IsRegular() {
				f := finding(model.CodeSliceFileMissing, fields)
				f.Path = p.K
				out.Findings = append(out.Findings, f)
				continue
			}
			sum, err := sha256File(abs)
			if err != nil {
				return out, err
			}
			if sum != yamlio.PyString(p.V) {
				f := finding(model.CodeSliceHandEdited, fields)
				f.Path = p.K
				out.Findings = append(out.Findings, f)
			}
		}
	}
	if len(out.Findings) == 0 {
		fields := model.Fields{"files": out.Files, "repos": out.Repos,
			"lock": workspace.LockName}
		out.Findings = []model.Finding{{
			Severity: model.SevOK,
			Code:     model.CodeFederationSlicesMatch,
			Message:  Message(model.CodeFederationSlicesMatch, fields),
			Fields:   fields,
		}}
	}
	return out, nil
}

// Gate runs gate 8 and returns its result (bin/company-os:1095-1103). Callers
// supply the ordinal because the denominator is dynamic: the gate exists only in
// federated mode, and gates 1-7 are never renumbered.
func Gate(ws *workspace.Workspace, m *Manifest, ordinal int) (model.GateResult, error) {
	g := model.GateResult{Ordinal: ordinal, Slug: GateSlug, Title: GateTitle}
	res, err := SliceFindings(ws, m)
	if err != nil {
		return g, err
	}
	g.Findings = res.Findings
	return g, nil
}

func finding(code string, f model.Fields) model.Finding {
	return model.Finding{
		Severity: model.SevFail,
		Code:     code,
		Message:  Message(code, f),
		Fields:   f,
	}
}

// Message composes a gate-8 finding's sentence from its code and its typed
// fields, and from nothing else.
//
// This is the only function in the package that produces gate prose. Everything
// upstream of it appends facts.
func Message(code string, f model.Fields) string {
	switch code {
	case model.CodeSliceLockMissing:
		return fmt.Sprintf(
			"%s present but %s is missing or malformed — run: company-os workspace sync",
			f.Str("manifest"), f.Str("lock"))
	case model.CodeRepoNotLocked:
		return fmt.Sprintf(
			"repo '%s' in %s has no %s entry (lock does not cover the manifest) — "+
				"run: company-os workspace sync",
			f.Str("repo"), f.Str("manifest"), f.Str("lock"))
	case model.CodeSliceSetDrift:
		return fmt.Sprintf(
			"repo '%s': slice set in %s differs from %s (a target or allowlist "+
				"changed without a re-sync) — run: company-os workspace sync",
			f.Str("repo"), f.Str("manifest"), f.Str("lock"))
	case model.CodeSliceFileMissing:
		return fmt.Sprintf(
			"federated slice file missing: %s (recorded in %s) — re-sync: "+
				"company-os workspace sync", f.Str("path"), f.Str("lock"))
	case model.CodeSliceHandEdited:
		return fmt.Sprintf(
			"federated slice hand-edited: %s — content hash differs from %s; "+
				"slices are read-only derived content — re-sync: company-os workspace sync",
			f.Str("path"), f.Str("lock"))
	case model.CodeFederationSlicesMatch:
		return fmt.Sprintf(
			"federated slices match %s (%d file(s) across %d repo(s); no hand-edits)",
			f.Str("lock"), f.Int("files"), f.Int("repos"))
	}
	return ""
}
