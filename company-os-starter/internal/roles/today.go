package roles

import (
	"os"
	"path/filepath"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/frontmatter"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/workspace"
	"github.com/metuur-ai/uncle-os/company-os-starter/internal/yamlio"
	"gopkg.in/yaml.v3"
)

// productRoles and engineeringRoles are the two role sets cmd_today branches on
// (`:1173`, `:1188`). `architect` is in neither, which is why it renders the
// banner and nothing else.
var (
	productRoles     = map[string]bool{"product-owner": true, "director-of-product": true}
	engineeringRoles = map[string]bool{"developer": true, "team-lead": true, "vp-engineering": true}
)

// Today is cmd_today (`bin/company-os:1168-1203`), as records.
func Today(ws *workspace.Workspace, role string) ([]model.GateResult, error) {
	sections := []model.GateResult{{
		Ordinal: 1,
		Slug:    model.SlugHeader,
		Findings: []model.Finding{{
			Severity: model.SevOK,
			Code:     model.CodeHeader,
			Fields:   model.Fields{"role": role},
		}},
	}}
	if g, ok := GlossarySection(role, len(sections)+1); ok {
		sections = append(sections, g)
	}

	if productRoles[role] {
		for _, pdir := range ws.AllPlatforms() {
			s, err := platformSection(ws, pdir, len(sections)+1)
			if err != nil {
				return nil, err
			}
			sections = append(sections, s)
		}
	}
	if engineeringRoles[role] {
		for _, tdir := range ws.AllTeams() {
			s, err := teamSection(ws, tdir, len(sections)+1)
			if err != nil {
				return nil, err
			}
			sections = append(sections, s)
		}
	}

	if guide, ok := OnboardingGuide(ws, role); ok {
		sections = append(sections, model.GateResult{
			Ordinal: len(sections) + 1,
			Slug:    model.SlugOnboarding,
			Findings: []model.Finding{{
				Severity: model.SevOK,
				Code:     model.CodeOnboarding,
				Path:     guide,
				Fields:   model.Fields{"guide": guide, "role": role},
			}},
		})
	}
	return sections, nil
}

// OnboardingGuide is onboarding_guide_for (`:1158-1165`): team scope before
// company scope (R-6.2). The returned path is workspace-relative in POSIX form,
// which is what `guide.relative_to(ws.root)` renders.
func OnboardingGuide(ws *workspace.Workspace, role string) (string, bool) {
	for _, tdir := range ws.AllTeams() {
		if g := filepath.Join(tdir, "onboarding", role+".md"); exists(g) {
			return relTo(ws.Root, g), true
		}
	}
	if g := filepath.Join(ws.Company, "onboarding", role+".md"); exists(g) {
		return relTo(ws.Root, g), true
	}
	return "", false
}

// platformSection is the `:1174-1187` loop body for one platform.
func platformSection(ws *workspace.Workspace, pdir string, ordinal int) (model.GateResult, error) {
	name := filepath.Base(pdir)
	active := filepath.Join(pdir, "change-records", "active")
	// sorted(active.iterdir()) is every entry, not only directories: a stray
	// file yields a row with no frontmatter and therefore '?' for both fields.
	prds := entries(active)

	s := model.GateResult{Ordinal: ordinal, Slug: model.SlugPlatform, Title: name}
	s.Findings = append(s.Findings, model.Finding{
		Severity: model.SevOK,
		Code:     model.CodePlatform,
		Subject:  name,
		Path:     relTo(ws.Root, pdir),
		Fields:   model.Fields{"platform": name, "activePRDs": len(prds)},
	})

	for _, d := range prds {
		prdPath := filepath.Join(active, d, "prd.md")
		meta, err := loadFrontmatter(prdPath)
		if err != nil {
			return model.GateResult{}, err
		}
		fields := model.Fields{
			"prd":      d,
			"platform": name,
			"status":   mapStr(meta, "status", "?"),
			"team":     mapStr(meta, "team", "?"),
		}
		// Not printed, but the two things a UI needs beyond the row: the PRD's
		// own title, and where to open it.
		if t := mapStr(meta, "title", ""); t != "" {
			fields["title"] = t
		}
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeActivePRD,
			Subject:  d,
			Path:     relTo(ws.Root, prdPath),
			Fields:   fields,
		})
	}

	arch := filepath.Join(pdir, "archive", "prds")
	for _, d := range entries(arch) {
		outcomePath := filepath.Join(arch, d, "outcome.md")
		if !exists(outcomePath) {
			continue
		}
		meta, err := loadFrontmatter(outcomePath)
		if err != nil {
			return model.GateResult{}, err
		}
		if mapStr(meta, "status", "") != "pending" {
			continue
		}
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeOutcomeReview,
			Subject:  d,
			Path:     relTo(ws.Root, outcomePath),
			Fields: model.Fields{
				"prd":      d,
				"platform": name,
				// `{m.get('due')}` with no default — an absent key f-strings as
				// "None", so `outcome review due None: old-1` is the line
				// Python prints, not one with a hole in it.
				"due":    mapStr(meta, "due", "None"),
				"status": "pending",
			},
		})
	}
	return s, nil
}

// teamSection is the `:1189-1198` loop body for one team.
func teamSection(ws *workspace.Workspace, tdir string, ordinal int) (model.GateResult, error) {
	name := filepath.Base(tdir)
	effPath := filepath.Join(tdir, "generated", "effective-governance.yaml")
	eff, err := loadYAMLFile(effPath)
	if err != nil {
		return model.GateResult{}, err
	}

	s := model.GateResult{Ordinal: ordinal, Slug: model.SlugTeam, Title: name}
	// `if not eff` is truthiness, so an empty or null document warns exactly as
	// an absent one does.
	if eff == nil || eff.IsFalsy() {
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevWarn,
			Code:     model.CodeGovernanceMissing,
			Subject:  name,
			Path:     relTo(ws.Root, effPath),
			Fields:   model.Fields{"team": name},
		})
		return s, nil
	}

	root := yamlio.Deref(eff.Root())
	s.Findings = append(s.Findings, model.Finding{
		Severity: model.SevOK,
		Code:     model.CodeTeam,
		Subject:  name,
		Path:     relTo(ws.Root, effPath),
		Fields: model.Fields{
			"team": name,
			// str(eff.get('generatedAt')) — an absent key renders as "None", so
			// the default carries Python's rendering rather than "".
			"generatedAt": mapStr(root, "generatedAt", "None"),
		},
	})

	// `eff.get("components", {}).items()` — an ABSENT key yields {} and the loop
	// does not run, but a present key of any non-mapping shape (null, a list, a
	// scalar) raises AttributeError before a single component line is printed.
	// R-0.7a(j): exit 4 with a diagnostic, never exit 0 with the components
	// silently skipped.
	components := yamlio.Deref(yamlio.MapGet(root, "components"))
	if components == nil {
		return s, nil
	}
	if components.Kind != yaml.MappingNode {
		return model.GateResult{}, model.Errorf(model.ExitArtifact,
			"%s: 'components:' must be a mapping", effPath)
	}
	for _, cid := range yamlio.MapKeys(components) {
		e := yamlio.Deref(yamlio.MapGet(components, cid))
		// `e["requirements"]["platform"]` and `["company"]` are INDEXED, not
		// .get() — a missing key is a KeyError and a non-mapping `e` is a
		// TypeError, both of which stop the command.
		reqs, err := requireMapping(effPath, cid, e, "requirements")
		if err != nil {
			return model.GateResult{}, err
		}
		platform, err := requireMapping(effPath, cid, reqs, "platform")
		if err != nil {
			return model.GateResult{}, err
		}

		// sum(len(v) for v in requirements['platform'].values()), plus the
		// per-platform split the sentence throws away — the one thing a
		// component view needs that this line cannot show.
		//
		// len() is not "length of a sequence": a mapping-valued platform counts
		// its keys and a string-valued one counts its characters, so
		// `platform: {p1: abcd}` is 4 under Python. Narrowing to sequences
		// reported 0 on a path that exits 0 either way — a silently wrong
		// number, which is worse than a refusal.
		total := 0
		byPlatform := make([]string, 0, len(platform.Content)/2)
		for _, pid := range yamlio.MapKeys(platform) {
			byPlatform = append(byPlatform, pid)
			n, ok := yamlio.PyLen(yamlio.MapGet(platform, pid))
			if !ok {
				return model.GateResult{}, model.Errorf(model.ExitArtifact,
					"%s: components.%s.requirements.platform.%s has no length",
					effPath, cid, pid)
			}
			total += n
		}

		company := yamlio.MapGet(reqs, "company")
		if company == nil {
			return model.GateResult{}, model.Errorf(model.ExitArtifact,
				"%s: components.%s.requirements has no 'company' key", effPath, cid)
		}
		controls, ok := yamlio.PyLen(company)
		if !ok {
			return model.GateResult{}, model.Errorf(model.ExitArtifact,
				"%s: components.%s.requirements.company has no length", effPath, cid)
		}

		fields := model.Fields{
			"component":            cid,
			"team":                 name,
			"platformRequirements": total,
			"companyControls":      controls,
		}
		if len(byPlatform) > 0 {
			fields["platforms"] = byPlatform
		}
		s.Findings = append(s.Findings, model.Finding{
			Severity: model.SevOK,
			Code:     model.CodeComponent,
			Subject:  cid,
			Fields:   fields,
		})
	}
	return s, nil
}

// requireMapping is `container[key]` where the result is then subscripted or
// iterated: a missing key raises KeyError, a non-mapping container raises
// TypeError, and a non-mapping result raises on the next access. All three stop
// cmd_today before it prints the component line, so all three are refusals here
// (R-0.7a(j)).
func requireMapping(path, cid string, container *yaml.Node, key string) (*yaml.Node, error) {
	if container == nil || container.Kind != yaml.MappingNode {
		return nil, model.Errorf(model.ExitArtifact,
			"%s: components.%s is not a mapping", path, cid)
	}
	v := yamlio.Deref(yamlio.MapGet(container, key))
	if v == nil {
		return nil, model.Errorf(model.ExitArtifact,
			"%s: components.%s has no '%s' key", path, cid, key)
	}
	if v.Kind != yaml.MappingNode {
		return nil, model.Errorf(model.ExitArtifact,
			"%s: components.%s '%s' must be a mapping", path, cid, key)
	}
	return v, nil
}

// entries is sorted(dir.iterdir()) by name, or nothing when dir is absent.
// os.ReadDir already sorts by filename, which for one parent is the order
// Python's sorted() over Path objects produces.
func entries(dir string) []string {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(des))
	for _, d := range des {
		out = append(out, d.Name())
	}
	return out
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// relTo renders a path under root the way Path.relative_to does, in POSIX form
// (R-1.12).
func relTo(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

// loadFrontmatter is frontmatter(path) reduced to its dict half: a document
// without the `---\n…\n---\n` fence yields nil, which is Python's {}.
func loadFrontmatter(path string) (*yaml.Node, error) {
	if !exists(path) {
		return nil, nil
	}
	doc, err := frontmatter.ParseFile(path)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
	}
	if !doc.HasFrontmatter {
		return nil, nil
	}
	parsed, err := yamlio.LoadFrontmatter(doc.YAML)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
	}
	return yamlio.Deref(parsed.Root()), nil
}

// loadYAMLFile is load_yaml(path, None) (`bin/company-os:58-63`): an absent file
// is (nil, nil), never an error, and the caller applies `or default` through
// Document.IsFalsy rather than a nil check (R-1.7a).
func loadYAMLFile(path string) (*yamlio.Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
	}
	doc, err := yamlio.Load(data)
	if err != nil {
		return nil, model.Errorf(model.ExitArtifact, "%s: %v", path, err)
	}
	return doc, nil
}

// mapStr is `str(m.get(key, default))` for a node mapping, giving Python's
// rendering of the value — which is what makes an unquoted `due: 2026-10-16`
// print as a date rather than as a Go timestamp.
//
// A container value goes through Python's repr rather than falling back to def:
// `status: {a: 1}` prints `[{'a': 1}]` under Python, and printing `[?]` instead
// is a plausible-looking line that says something untrue about the artifact.
func mapStr(m *yaml.Node, key, def string) string {
	v := yamlio.Deref(yamlio.MapGet(m, key))
	if v == nil {
		return def
	}
	return yamlio.PyText(v)
}
