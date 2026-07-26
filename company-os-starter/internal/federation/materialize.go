package federation

// Read-only materialization and content hashing (bin/company-os:2331-2472).
//
// Three of the package's four orderings live in this file. See doc.go.

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
)

// Slice trees are derived content, like generated/. Freezing them makes a
// hand-edit take deliberate effort, and gate 8 turns it into a [FAIL] rather
// than into silent drift.
const (
	modeFileReadonly = 0o444
	modeDirReadonly  = 0o555
	modeFileWritable = 0o644
	modeDirWritable  = 0o755
)

// rglob is `sorted(root.rglob("*"))` (bin/company-os:2355, :2422): every
// descendant of root, as a POSIX path relative to root, in CPython PurePath
// order.
//
// The sort is yamlio.PathLess, not sort.Strings. Component-wise ordering puts
// `sdd/adr/a.md` before `sdd/adr-x.md`; a byte sort puts it after, because '/'
// (0x2F) sorts above '-' and '.'. The two disagree on any tree with a directory
// whose name is a prefix of a sibling file's — which is exactly what an `adr/`
// directory next to an `adr.md` index is.
//
// Sorting the relative paths is equivalent to sorting the absolute ones Python
// compares: every entry shares the same root prefix, so the differing component
// is the same one either way.
//
// A missing root yields nothing, matching rglob on a path that does not exist.
// Symlinked directories are not descended (filepath.WalkDir does not follow
// them), and hashTree skips symlinks outright.
func rglob(root string) []string {
	var out []string
	_ = filepath.WalkDir(root, func(p string, _ os.DirEntry, err error) error {
		if err != nil {
			if p == root {
				return nil
			}
			return nil
		}
		if p == root {
			return nil
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return nil
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	yamlio.SortPaths(out)
	return out
}

// chmodWritable is _chmod_writable (bin/company-os:2331-2335): restore write so
// a previously frozen tree can be removed or overwritten. Errors are swallowed,
// as Python's bare `except OSError: pass` swallows them.
func chmodWritable(p string) {
	fi, err := os.Lstat(p)
	if err != nil {
		return
	}
	mode := os.FileMode(modeFileWritable)
	if fi.IsDir() {
		mode = modeDirWritable
	}
	_ = os.Chmod(p, mode)
}

// forceRemove is _force_remove (bin/company-os:2338-2348): remove a path even
// when a prior materialization left it read-only.
//
// The chmod pass has to cover the WHOLE subtree before the removal starts:
// unlinking an entry needs write permission on its parent directory, which
// 0555 denies, and no amount of chmod-ing the entry itself supplies it.
func forceRemove(p string) {
	fi, err := os.Lstat(p)
	if err != nil {
		return
	}
	if fi.IsDir() {
		// Reverse order, children before parents — see makeReadonly.
		rels := rglob(p)
		for i := len(rels) - 1; i >= 0; i-- {
			chmodWritable(path.Join(p, rels[i]))
		}
		chmodWritable(p)
		_ = os.RemoveAll(p)
		return
	}
	chmodWritable(p)
	_ = os.Remove(p)
}

// makeReadonly is _make_readonly (bin/company-os:2351-2360): files 0444, dirs
// 0555.
//
// TRAP. Python walks `sorted(root.rglob("*"), reverse=True)` — children strictly
// before their parents. filepath.WalkDir is pre-order, so streaming it would
// freeze a directory before descending into it. The walk is therefore collected
// and reversed explicitly.
//
// `root` itself is NOT chmod'd: rglob("*") yields descendants only, so the slice
// target keeps whatever mode MkdirAll gave it. That is observable — the
// differential harness compares mode bits on every path — so do not "fix" it.
func makeReadonly(root string) {
	rels := rglob(root)
	for i := len(rels) - 1; i >= 0; i-- {
		p := path.Join(root, rels[i])
		fi, err := os.Lstat(p)
		if err != nil {
			continue
		}
		mode := os.FileMode(modeFileReadonly)
		if fi.IsDir() {
			mode = modeDirReadonly
		}
		_ = os.Chmod(p, mode)
	}
}

// materializeSlice is materialize_slice (bin/company-os:2363-2396): copy ONLY
// allowlisted paths out of the cache working tree into target, then freeze them.
// A path absent upstream is skipped — a platform repo may have no skills/.
func materializeSlice(cache, target string, paths []string) error {
	// A prior materialization left this tree 0444/0555. Restore write across the
	// WHOLE target before touching it: removal and mkdir both need write on the
	// PARENT of each entry, which chmod-ing only the path being removed does not
	// supply. This bites exactly when an allowlist entry is nested (docs/sdd),
	// which is the norm for knowledge slices; every pre-existing fixture used
	// depth-1 entries (governance/) and never hit it.
	if exists(target) {
		chmodWritable(target)
		for _, rel := range rglob(target) {
			chmodWritable(path.Join(target, rel))
		}
	}
	for _, p := range paths {
		if rel := strings.Trim(p, "/"); rel != "" {
			forceRemove(path.Join(target, rel))
		}
	}
	for _, p := range paths {
		rel := strings.Trim(p, "/")
		if rel == "" {
			continue
		}
		src, dst := path.Join(cache, rel), path.Join(target, rel)
		fi, err := os.Stat(src)
		if err != nil {
			continue
		}
		if fi.IsDir() {
			if err := copyTree(src, dst); err != nil {
				return err
			}
		} else if fi.Mode().IsRegular() {
			if err := mkdirAll(filepath.Dir(dst)); err != nil {
				return err
			}
			if err := copyFile(src, dst); err != nil {
				return err
			}
		}
	}
	if exists(target) {
		makeReadonly(target)
	}
	return nil
}

// copyTree is shutil.copytree(dirs_exist_ok=True). Existing destinations are
// merged rather than refused: a paths list naming a file before its containing
// directory would otherwise fail, and hand-written multi-slice manifests make
// that ordering likely.
//
// File modes are deliberately not preserved. copy2 preserves them in Python,
// but makeReadonly overwrites every mode under the target immediately after, and
// nothing hashes mtime — so the only observable outcome is 0444/0555 either way.
func copyTree(src, dst string) error {
	if err := mkdirAll(dst); err != nil {
		return err
	}
	for _, rel := range rglob(src) {
		sp, dp := path.Join(src, rel), path.Join(dst, rel)
		fi, err := os.Lstat(sp)
		if err != nil {
			continue
		}
		switch {
		case fi.IsDir():
			if err := mkdirAll(dp); err != nil {
				return err
			}
		case fi.Mode()&os.ModeSymlink != 0:
			// copytree(symlinks=False) copies what the link points AT.
			if target, err := os.Stat(sp); err == nil && target.IsDir() {
				continue
			}
			if err := copyFile(sp, dp); err != nil {
				return err
			}
		default:
			if err := copyFile(sp, dp); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return model.Errorf(model.ExitExternalTool, "cannot read %s: %v", src, err)
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		return model.Errorf(model.ExitExternalTool, "cannot write %s: %v", dst, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return model.Errorf(model.ExitExternalTool, "cannot copy %s: %v", src, err)
	}
	return out.Close()
}

// sha256File is _sha256_file (bin/company-os:2399-2404).
func sha256File(p string) (string, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", model.Errorf(model.ExitExternalTool, "cannot read %s: %v", p, err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", model.Errorf(model.ExitExternalTool, "cannot read %s: %v", p, err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// hashTree is hash_tree (bin/company-os:2407-2427): {ws-relative posix path:
// sha256} for the ALLOWLISTED content of one materialized slice.
//
// Scoped to `paths` rather than walking the whole target: a derived artifact
// that appeared under the target (a generated/ aggregate, a CLAUDE.md node)
// would otherwise be absorbed into the lock and frozen 0444 by the next sync,
// after which graph build could not rewrite it.
//
// The result is an OrderedMap because its INSERTION order is what the lock
// emits. DefaultSlicePaths lists `governance/` and then
// `governance/requirements.yaml`, so that file is set twice — and OrderedMap.Set,
// like a Python dict assignment, keeps the first position.
func hashTree(ws *workspace.Workspace, target string, paths []string) (*yamlio.OrderedMap, error) {
	files := yamlio.NewOrderedMap()
	for _, p := range paths {
		rel := strings.Trim(p, "/")
		if rel == "" {
			continue
		}
		src := path.Join(target, rel)
		var candidates []string
		if isDir(src) {
			for _, r := range rglob(src) {
				candidates = append(candidates, path.Join(src, r))
			}
		} else {
			candidates = []string{src}
		}
		for _, f := range candidates {
			fi, err := os.Lstat(f)
			if err != nil || !fi.Mode().IsRegular() {
				continue
			}
			sum, err := sha256File(f)
			if err != nil {
				return nil, err
			}
			wsRel, err := filepath.Rel(ws.Root, f)
			if err != nil {
				return nil, model.Errorf(model.ExitExternalTool,
					"%s is outside the workspace", f)
			}
			files.SetString(filepath.ToSlash(wsRel), sum)
		}
	}
	return files, nil
}

// AggregateHash is aggregate_hash (bin/company-os:2430-2437): a deterministic
// digest over a {path: sha256} map.
//
// TRAP. It iterates `sorted(files)` — a plain STRING sort — over the very keys
// the lock emits in walk order. Two orders, same data, same file. Unifying them
// changes either sliceHash or the lock's byte layout.
//
// Computed ONCE over the union of a repo's slices. Chaining per-slice digests
// would make the result depend on manifest slice order, so reordering `slices:`
// with no content change would flip sliceHash and make --frozen die blaming the
// object cache.
func AggregateHash(files *yamlio.OrderedMap) string {
	agg := sha256.New()
	for _, rel := range files.SortedKeys() {
		v, _ := files.Get(rel)
		agg.Write([]byte(rel))
		agg.Write([]byte{0})
		agg.Write([]byte(v.Value))
		agg.Write([]byte{'\n'})
	}
	return hex.EncodeToString(agg.Sum(nil))
}

// materializeAll is _materialize_all (bin/company-os:2455-2472): check out ONE
// cone covering every slice's allowlist, then materialize each slice into its
// own target. Returns the UNION {relpath: sha256} map.
//
// The union checkout is mandatory: checkoutSlice uses `sparse-checkout set`, not
// `add`, so a per-slice checkout loop would have each call clobber the previous
// cone and leave every later slice empty.
func materializeAll(ws *workspace.Workspace, cache, name string, slices []Slice, sha string) (*yamlio.OrderedMap, error) {
	var union []string
	for _, s := range slices {
		for _, p := range s.Paths {
			if !containsString(union, p) {
				union = append(union, p)
			}
		}
	}
	if err := checkoutSlice(cache, sha, sparseDirs(union)); err != nil {
		return nil, err
	}
	files := yamlio.NewOrderedMap()
	for _, s := range slices {
		target, err := SliceTargetRoot(ws, name, s.LocalDirectory)
		if err != nil {
			return nil, err
		}
		if err := materializeSlice(cache, target, s.Paths); err != nil {
			return nil, err
		}
		part, err := hashTree(ws, target, s.Paths)
		if err != nil {
			return nil, err
		}
		files.Update(part)
	}
	return files, nil
}

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func mkdirAll(dir string) error {
	if err := os.MkdirAll(dir, 0o777); err != nil {
		return model.Errorf(model.ExitExternalTool, "cannot create %s: %v", dir, err)
	}
	return nil
}
