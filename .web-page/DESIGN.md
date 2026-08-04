# Uncle OS — Design System

Light editorial documentation system for an **interactive tutorial site**. The
home tab is a product page; the other nine tabs teach by doing.

All tokens live in `src/index.css` under `@theme`. Shared components live in
`src/components/ui/`. This file is the contract — if a rule here conflicts with
what a component currently does, the component is wrong.

---

## 1. The one hard rule

**Never use a raw Tailwind palette utility.** No `bg-slate-50`, `text-indigo-600`,
`border-gray-200`, `bg-white`, `text-black`. Use the semantic names:

| Instead of | Use |
|---|---|
| `bg-slate-50` / page bg | `bg-canvas` |
| `bg-white` (card) | `bg-surface` |
| `bg-slate-100` (well) | `bg-surface-sunken` |
| `text-slate-900` | `text-fg` |
| `text-slate-600` | `text-fg-muted` |
| `text-slate-500` | `text-fg-subtle` |
| `border-slate-200` | `border-border` |
| `bg-indigo-600` | `bg-accent` |
| `text-indigo-600` | `text-accent-text` |
| `bg-indigo-50` | `bg-accent-soft` |
| `border-indigo-200` | `border-accent-border` |
| `bg-cyan-600` (Team OS) | `bg-scope` |
| green / amber / red | `success` / `warn` / `danger` families |

Each status family has four members: `--color-X`, `--color-X-text`,
`--color-X-soft`, `--color-X-border`. Use `X-text` when the colour *is* the text
(it is contrast-checked against white/soft), `X` when it is a solid fill behind
`text-fg-inverse`.

The single deliberate exception is the terminal surface, which uses the `term-*`
tokens (`bg-term-bg`, `text-term-fg`, `text-term-success`, …). Terminals are dark
because terminals are dark. Nothing else on the site is.

---

## 2. Type

Root is **16px**. Body is `text-base` (16px / 1.625). There is no "easy read"
toggle — the base scale is already the accessible scale.

| Token | Size | Use |
|---|---|---|
| `text-5xl` / `text-4xl` | 60 / 48 | Hero headline only |
| `text-3xl` | 36 | Page (tab) title, one per view |
| `text-2xl` | 28 | Section heading |
| `text-xl` | 22 | Subsection / card title |
| `text-lg` | 18 | Lead paragraph under a title |
| `text-base` | 16 | Body — the default |
| `text-sm` | 14 | Secondary text, dense tables, captions |
| `text-xs` | 12 | Labels, badges, tab bar |
| `text-2xs` | 11 | Metadata, `kbd`, micro-labels. **Never body copy.** |

Weights: 700 hero, 600 headings, 500 labels/UI, 400 body. Nothing else.

**Measure.** Any paragraph gets `.measure` (68ch). Narrow columns use
`.measure-tight` (56ch). Never let prose span a full-width container.

**Numbers.** Add `.tabular` to anything where digits change in place — gate
numbers, counts, exit codes, timers, table numerics.

`font-mono` is for commands, paths, IDs, file names, code, and keyboard keys.
Not for headings, not for decoration.

---

## 3. Layout

```
Header (sticky, z-50)
Navbar (sticky under header, z-40)
main  →  <PageShell>  →  <PageHeader> + <Section>…
Footer
```

- Page container: `mx-auto w-full max-w-6xl px-4 sm:px-6 lg:px-8`.
  The old `max-w-[95%]` produced ~180-character lines. Gone.
- Wide views that genuinely need it (Reference matrix, Architecture tree) may
  use `max-w-7xl`. Never wider.
- Vertical rhythm between sections: `space-y-12` (desktop) / `space-y-10` (mobile).
  Inside a section: `space-y-4`. Inside a card: `space-y-3`.
- Spacing is 4/8-based: `1, 1.5, 2, 3, 4, 6, 8, 12, 16`. No arbitrary `p-[13px]`.
- Breakpoints: 375 / `sm:`640 / `md:`768 / `lg:`1024 / `xl:`1280. Design mobile
  first. Verify no horizontal scroll at 375px.

---

## 4. Radius & elevation

Radius: `rounded-md` (inputs, badges, small buttons) · `rounded-lg` (buttons,
cards) · `rounded-xl` (panels, modals) · `rounded-2xl` (hero surfaces).
Do not mix `rounded-3xl`/`rounded-full` except on avatars, pills and dots.

Elevation, three steps only:
- `shadow-xs` — resting input / subtle separation
- `shadow-sm` — cards
- `shadow-md` — dropdowns, popovers, sticky bars
- `shadow-lg` — modals only

Most cards should use **border + `bg-surface`** with no shadow. Borders carry
hierarchy in an editorial system; shadows are the exception.

---

## 5. Interaction

- Every interactive element: `cursor-pointer`, a visible hover state, and a
  transition of `150ms`–`200ms`. Use `transition-colors` where only colour moves.
- **Press feedback must not shift layout.** Use colour/opacity, or `active:scale-[0.98]`
  on self-contained cards. Never animate `width`, `height`, `top`, `left`, or margin.
- Touch targets ≥ 44×44px. If the visual is smaller, pad it or add hit area.
  Icon-only buttons: `p-2` minimum plus `aria-label`.
- Focus rings come free from `:focus-visible` in base CSS. Do not add
  `focus:outline-none` without replacing the ring.
- Disabled: `disabled:opacity-50 disabled:cursor-not-allowed` **and** the real
  `disabled` attribute. Never a div that looks dead but still fires.

Animation budget: 1–2 animated elements per view. Enter 180–220ms, exit ~120ms.
Use `.animate-fade-rise`, `.animate-scale-in`, `.stagger`. Reduced-motion is
handled globally — do not re-implement it.

---

## 6. Accessibility floor

Non-negotiable, checked before any view is considered done:

- Body text ≥ 4.5:1, large text and UI glyphs ≥ 3:1. The tokens satisfy this;
  breaking it means you used a raw palette colour.
- Colour never carries meaning alone. A pass/fail state needs an icon or a word,
  not just green/red. Every status badge in this app ships with an icon.
- Headings run `h1 → h2 → h3` with no skipped levels. One `h1` per view
  (`PageHeader` renders it).
- Icon-only buttons need `aria-label`. Decorative icons need `aria-hidden="true"`.
- Tab bars: `role="tablist"` / `role="tab"` / `aria-selected` / `aria-controls`.
  Arrow-key navigation between tabs.
- Modals: focus trapped, `Esc` closes, focus returns to the trigger, backdrop
  uses `bg-overlay`, `role="dialog"` + `aria-modal="true"` + `aria-labelledby`.
- Async status (`command ran`, `answer correct`) announces via `aria-live="polite"`.
- Form inputs get a real visible `<label>`. Placeholder is not a label.
- No emoji as icons. `lucide-react` only, `strokeWidth` left at default.

---

## 7. Tutorial view contract

Every non-home tab is a lesson and renders through the same shell:

```tsx
<PageShell>
  <PageHeader
    eyebrow="Lesson 3 of 9"
    title="CLI Terminal Explorer"
    lead="Run validate, graph build and prd new against a simulated workspace."
    icon={Terminal}
  />
  <Section title="…" description="…"> … </Section>
  <LessonFooter onNext={…} nextLabel="Governance Tiers" />
</PageShell>
```

Rules for lessons:
- State the objective before the interaction. The learner should know what they
  are about to do and why.
- Interactive widgets get an explicit affordance — a button labelled with a verb,
  never "click anywhere to continue".
- Show result state clearly: loading (skeleton if > 300ms), success, error with a
  recovery path. Never a dead end.
- Long lessons use `<StepList>` with visible progress and a current-step marker.
- Every lesson ends with `<LessonFooter>` pointing at the next tab, so the site
  reads as a sequence rather than ten disconnected panels.

---

## 8. Component inventory (`src/components/ui/`)

| Component | Purpose |
|---|---|
| `PageShell` | Width container + vertical rhythm for a whole view |
| `PageHeader` | `h1`, eyebrow, lead, icon. One per view |
| `Section` | `h2` + optional description + content well |
| `Card` | Bordered surface. Optional `interactive` for clickable cards |
| `Callout` | `info` / `success` / `warn` / `danger` / `accent`. Icon + title + body |
| `Badge` | Small status/label pill. Icon optional, tone-driven |
| `Button` | `primary` / `secondary` / `ghost` / `danger`; `sm`/`md`/`lg`; `loading` |
| `CodeBlock` | Mono block with language label + copy button |
| `InlineCode` | Mono inline token |
| `Terminal` | Dark terminal surface with prompt/output lines |
| `StepList` / `Step` | Numbered lesson steps with completion state |
| `ProgressBar` | Labelled progress with `role="progressbar"` |
| `Tabs` | Accessible in-view tab group (tablist semantics + arrow keys) |
| `Prose` | Applies `.measure` and long-form spacing to text blocks |
| `EmptyState` | Icon + message + action for no-data panes |
| `LessonFooter` | "Next lesson" navigation |

Import from `../ui` (barrel). If you need a new primitive, add it there rather
than inventing a local one-off.

---

## 9. Content

`src/data/*.ts` is **frozen**. Copy is not edited, reworded, shortened or
invented. Layout and presentation change; words do not. If a design needs copy
that does not exist in the data files, restructure to use what is there.
