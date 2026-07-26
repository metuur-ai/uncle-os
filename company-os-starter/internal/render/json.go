package render

// The `--json` renderer: ONE encoder over the record types, for all sixteen
// subcommands (R-3.4b).
//
// This is the whole of Unit 3's output side, and it is deliberately small. Every
// command in the CLI returns []model.GateResult; a JSON writer over that type
// therefore serves every command by construction, and R-7.7's tuple-equality
// between the text and JSON renderers is a property of the shared record set
// rather than sixteen separate agreements to keep. A per-command encoder would
// have reintroduced exactly the drift R-2.9 exists to prevent, and would have
// needed sixteen tests to say what one says here.
//
// Nothing in this file knows a command's name, a gate's meaning, or a code's
// wording. It knows three things: the shape of a Finding, that a Fields entry
// named model.FieldNext is next-command guidance (R-3.6), and that every payload
// carries a schema version and a build identifier (R-3.4, R-3.5).

import (
	"encoding/json"
	"io"

	"github.com/metuur-ai/uncle-os/company-os-starter/internal/model"
)

// SchemaVersion is the `schemaVersion` in every payload (R-3.4).
//
// It starts at 1 and is incremented when a documented field is removed or
// repurposed. Adding a field is not a break: a consumer reading `sections` is
// unaffected by a new sibling, so additions ship without a bump.
const SchemaVersion = 1

// Result is what one command run produced, as the JSON encoder needs it. The
// dispatcher fills it in; nothing below cmd/ constructs one.
type Result struct {
	// Command and Action are the subcommand and its action verb, e.g.
	// "prd" and "complete". Action is empty for the commands that have none.
	Command string
	Action  string
	// Root is the resolved workspace root. It is carried for every command, not
	// just `validate`: a machine consumer reading a payload out of a log has no
	// other way to know which workspace it describes.
	Root string
	// Sections is the record set, verbatim — the same slice the text renderer
	// was handed.
	Sections []model.GateResult
	// Err is the command's error, if any. Its message is published; its exit
	// code is ExitCode.
	Err error
	// ExitCode is the process status this run will exit with (R-3.8): the
	// payload is written even on failure, and it says what the failure was.
	ExitCode model.ExitCode
}

// JSON writes one payload. It is the only place `--json` bytes are produced.
//
// The document is indented and newline-terminated. Indentation costs nothing a
// consumer cares about and makes a failing CI log readable; the trailing newline
// makes the stream line-oriented, so `jq` and a shell `read` both work without
// special-casing the last byte.
func JSON(w io.Writer, r Result) error {
	doc := document{
		SchemaVersion: SchemaVersion,
		Build:         model.BuildInfo(),
		Command:       r.Command,
		Action:        r.Action,
		Root:          r.Root,
		ExitCode:      int(r.ExitCode),
		Sections:      make([]section, 0, len(r.Sections)),
		// Guidance is never null: a consumer testing `.guidance | length` should
		// not have to distinguish "no next step" from "field absent" (R-3.6).
		Guidance: []string{},
	}
	if r.Err != nil {
		doc.Error = r.Err.Error()
	}
	if next := model.NextCommands(r.Sections); len(next) > 0 {
		doc.Guidance = next
	}
	for _, s := range r.Sections {
		doc.Sections = append(doc.Sections, encodeSection(s))
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

// document is the payload. The field names are the R-3.4 contract and are frozen
// on first publish; `sections` in particular is NOT `gates` (R-3.4a), because
// GateResult carries the sections of `ids list`, `skills list`, `today`,
// `governance explain` and the scaffolding commands as well as validate's gates,
// and `gates` reads wrong for most of the surface.
type document struct {
	SchemaVersion int         `json:"schemaVersion"`
	Build         model.Build `json:"build"`
	Command       string      `json:"command"`
	Action        string      `json:"action,omitempty"`
	Root          string      `json:"root,omitempty"`
	ExitCode      int         `json:"exitCode"`
	Sections      []section   `json:"sections"`
	Guidance      []string    `json:"guidance"`
	Error         string      `json:"error,omitempty"`
}

// section is one model.GateResult. Findings is always an array: a gate that ran
// and found nothing is not the same fact as a gate that did not run, and
// examples/golden-validate.txt:11-12 is a real gate with no findings (R-2.1).
type section struct {
	Ordinal  int       `json:"ordinal"`
	Slug     string    `json:"slug"`
	Title    string    `json:"title,omitempty"`
	Findings []finding `json:"findings"`
}

// finding is one model.Finding.
//
// Severity is the lowercase machine name from Severity.String(), never the
// bracketed `[FAIL]` marker — that capitalization is the text renderer's and
// lives in severityMarker.
//
// Fields is model.Fields unchanged, which is what keeps counts encoded as
// numbers and ordered list values in their authored order (R-2.3). Message and
// Subject are published alongside it so that a consumer can show the human
// sentence without recomposing it, and so a reader diffing text against JSON has
// the same string on both sides.
//
// Message can be empty and is still always present. Where the sentence is worth
// having at record level the producer composes it (validate, the product and
// governance clusters, the scaffolding commands); where the whole line is a
// renderer's own formatting of Fields — `ids list`, `today`, `skills list`,
// `graph build` — there is no Message to publish and Fields carries everything.
// The key stays in the document either way so that consumers do not have to
// branch on its presence.
type finding struct {
	Severity string       `json:"severity"`
	Code     string       `json:"code"`
	Subject  string       `json:"subject,omitempty"`
	Path     string       `json:"path,omitempty"`
	Message  string       `json:"message"`
	Fields   model.Fields `json:"fields,omitempty"`
}

func encodeSection(s model.GateResult) section {
	out := section{
		Ordinal:  s.Ordinal,
		Slug:     s.Slug,
		Title:    s.Title,
		Findings: make([]finding, 0, len(s.Findings)),
	}
	for _, f := range s.Findings {
		out.Findings = append(out.Findings, finding{
			Severity: f.Severity.String(),
			Code:     f.Code,
			Subject:  f.Subject,
			Path:     f.Path,
			Message:  f.Message,
			Fields:   f.Fields,
		})
	}
	return out
}
