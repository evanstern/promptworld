# Contract: help overlay structure & content

## Layering

1. `?` opens the overlay from every mode except text-entry (minibuffer focused: `?` is
   a buffer character; minibuffer help is content reachable from other modes' overlays).
2. While open, the overlay owns the keyboard: its own navigation + dismissal only; all
   other keys are inert — no silent fallthrough to the mode beneath (FR-012).
3. Dismissal: `esc` or `?` (toggle). Exactly one layer released; the esc chain beneath
   (minibuffer → decisions → detail → solo → home) is untouched. Never a stacked second
   overlay.
4. Opening/using/dismissing sends no IPC command, emits no event, changes no world or
   layout state beyond the help fields themselves (FR-006).

## Sections & tiers

- **Section 1 — keys for the mode you were in** (frozen at open): basic tier first
  (newcomer keys ≈ the footer-hinted set), advanced tier behind one in-overlay action
  (remaining keys + layered globals). Every working binding appears in exactly one tier;
  every listed binding works (FR-003 — enforced by the keymap sweep test).
- **Section 2 — the screen**: header anatomy (every element and every conditional
  badge/posture the header can render), map glyph legend (rendered from the same glyph
  table `renderMapGrid` uses — FR-005), dock tabs by key/name/purpose.
- **Section 3 — lessons (pull reference)**: lists `helpLesson` entries for on-demand
  reading.

## The lessons seam (SC-006)

`helpLesson{id, title, body}` is the whole contract: the future first-occurrence lesson
projection registers its lessons as entries in this table (or a registry that feeds
it). Adding a lesson is a content addition; the overlay's navigation, rendering, and
tests require no structural change. Until that feature lands the table ships empty and
the section renders a short "lessons appear here as the village teaches them" line.

## Content rules

- Static, local, model-independent strings; identical bytes on no-LLM worlds (FR-008).
- Plain language, one concept per line; overlay pages scroll with the standard pager
  (overflow footer shows how) and remain usable at the client's minimum size (FR-009).
- Footer hints advertise `? help` in every mode's hint line (FR-011), and the keymap
  design doc's footer table stays in sync (updated in the same PR).
