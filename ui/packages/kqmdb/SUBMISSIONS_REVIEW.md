# KQM submission and review interface check

Date: 22 September 2026.

## Scope and coverage

Scope: the public submission form, saved status page, reviewer sign-in, queue,
review details, matching-team comparisons, import form, and decision dialogs.
The check includes their loading, empty, error, pending, approved, and rejected
states. Existing result charts received a functional check; their internal UI
was not part of this interface review.

Stack: React 19, Wouter, Tailwind 4, and the existing Gauge primitives. The pages
use the site's KQM logo, dark purple colors, `g-*` design tokens, Hanken Grotesk
body font, Space Grotesk headings, and JetBrains Mono for identifiers.

Project documents read: `CLAUDE.md`, `CONTRIBUTING.md`,
`docs/adr/0001-design-system-migration.md`, and this package's `README.md`.
No nested `AGENTS.md` file applied to these files. The user-supplied working
instructions also applied.

The review used `better-interface` and all six of its domain skills. The tag
helper page at `https://taghelper.simpact.app/id/LNJ6bpFTFDCw` supplied the flow
reference. The KQM pages use direct review actions in place of copied commands.

| Domain | Evidence inspected | Result |
| --- | --- | --- |
| Accessibility | Labels, field errors, live status, native controls, dialog name and focus, Escape and focus return, skip link, heading order, language control, 44 px actions | Clear |
| Layout | Form and review columns; queue, receipt, and matching-team sections; 320 px width and 200% CSS zoom; dialog scroll and wrapping | Clear |
| Writing | Form hints, saved-result limitation, public name notice, error recovery, empty queue, decision effects, required review notes, import scope | Clear |
| Typography | Heading scale, body and field sizes, line height, DPS values, dates, long identifiers, full descriptions and queue truncation | Clear |
| Colors | Computed foreground/background pairs, KQM tokens, focus outline, success/warning/error text and labels | Clear |
| UI | Primary and destructive actions, native disclosures, spacing, panel radius, feedback, pending actions, reduced-motion and forced-color CSS | Clear |

## Findings

No actionable interface findings.

The review found and fixed a dialog focus-return error and stale comparison
text after a replacement. It also added a name to the language control and
page titles for submission and review pages.

## Verification

Passed:

- `pnpm --filter @gcsim/db test:worker`: 17 tests passed. Checks cover saved
  results, Basic Auth, same-origin writes, retries, approval, rejection,
  replacement, conflicting decisions, import, source-update protection, and
  input limits. These tests use SQLite transactions and an in-memory R2 store.
- `pnpm --filter @gcsim/db typecheck`: passed.
- `pnpm --filter @gcsim/db build`: passed. The existing large-bundle warning
  remains; no new dependency was added.
- `pnpm --filter @gcsim/e2e exec playwright test --config playwright.db.config.ts`:
  the six existing database/viewer cases passed. The first run found the dialog
  focus error and an error-retry fixture that recovered too early under React
  Strict Mode. Both were corrected.
- `pnpm --filter @gcsim/e2e exec playwright test --config playwright.db.config.ts submissions.spec.ts`:
  all six new flow tests passed after those corrections. They cover submission
  errors and receipt, login and approval, rejection note, replacement, queue
  retry and empty state, 320 px width, and 200% CSS zoom.
- `pnpm --filter @gcsim/e2e exec playwright test --config playwright.db.config.ts submissions.spec.ts --grep replacement`:
  passed after the final comparison-state fix. The old comparison is removed
  and the original viewer link is retained.
- Targeted Biome checks and `git diff --check`: passed.
- In-app Browser with the local Worker: submitted the provided Neuvillette
  result, opened the saved receipt, signed in, approved it, and opened its
  published viewer. All four characters, weapons, artifacts, configuration,
  and result charts loaded.
- In-app Browser with the local Worker: imported `LNJ6bpFTFDCw`, confirmed the
  original author and description, inspected the four builds and three matching
  teams, then rejected the local copy with a note. The status and note updated,
  and focus moved to the decision heading. The source entry was not changed.
- In-app Browser screenshots: inspected the form, sign-in, review details,
  comparisons, and decision dialog with the KQM design.
- Live Cloudflare deployment `002f9ecd-f0a1-447e-a855-f157707d97ec`:
  the public form and reviewer sign-in loaded in the in-app Browser. An empty
  public submission showed its field error and moved focus to the link field.
  HTTP checks returned 401 for an unauthenticated review request and 200 for
  authenticated sign-in and queue reads. The private responses had `no-store`
  and no public CORS header. The public database retained 5,996 records.
- In-app Browser at live `/db/wGtDfgdt9n8G`: all four character builds loaded,
  DPS was 66,610, and 15 chart graphics were present. No console error appeared.
- The live D1 console confirmed migration `0002_submissions.sql`, the new table
  and source column, and the unchanged public record count. No production test
  simulation was published.

Measured text contrast from rendered colors:

| Element | Foreground / background | Contrast |
| --- | --- | --- |
| Summary details | `#d1c6d8` / `#2d282f` | 8.78:1 |
| Pending status | `#e7be80` / `#423745` | 6.48:1 |
| Primary action | `#232024` / `#dab2f9` | 8.99:1 |
| Rejection outline text | `#ff97ac` / `#423745` | 5.51:1 |
| Destructive action | `#232024` / `#ff97ac` | 7.88:1 |
| Dialog field text | `#faf7fc` / `#423745` | 10.61:1 |

The rendered field focus outline is 2 px with a 3 px offset. Its color
`#dab2f9` has 6.29:1 contrast against the field background `#423745`.
Status is also given in text; color alone does not convey the decision.

Not verified:

- A real screen reader session, native browser zoom, and mobile Safari.
- A complete audit of the existing shared chart and character-detail widgets.
- Manual forced-colors and reduced-motion device checks; the CSS rules were
  inspected.
- Translation of the new English workflow into the site's other languages.

## Verdict

Approve: no HIGH findings remain in the stated scope.
