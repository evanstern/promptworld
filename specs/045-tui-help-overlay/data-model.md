# Data Model: `?` help overlay

**Spec**: [spec.md](./spec.md) | **Plan**: [plan.md](./plan.md) | **Date**: 2026-07-25

Client-side view state and static content only — nothing here touches the daemon, the
event log, or `sim.State`.

## Model state (internal/tui/tui.go)

| Field | Type | Semantics |
|---|---|---|
| `helpOpen` | `bool` | Overlay visible; while true the overlay owns key dispatch (head of chain, after ctrl+c, only when `!mbFocused` for the open trigger) |
| `helpSection` | small enum/int | Current section: mode-keys / walkthrough / lessons reference |
| `helpTier` | `bool` or int | Basic vs advanced page of the mode-keys section |
| `helpMode` | captured mode key | The mode help was opened *from* (frozen at open so content can't drift while open — spec edge case) |
| `helpScroll` | `int` | Pager offset; increment-unbounded/floor-0/clamp-at-render (chronicleDetailPane idiom) |

All reset on dismissal; dismissal restores nothing else because nothing else changed
(FR-006 is satisfied by construction — the overlay never mutates non-help state).

## Content model (internal/tui/help.go — static data)

| Entity | Shape | Notes |
|---|---|---|
| **helpModeKey** | `{key, action string, tier}` | One binding row; per-mode tables for: global/home, minibuffer (reachable as content from other modes, FR-001), inspect, villagers roster, villagers detail, solo/narrow. Source of truth: docs/design/tui/patterns/keymap.md; basic tier ≈ footer-hinted keys |
| **glyphEntry** | `{glyph, name, meaning string}` | THE shared table: extracted from renderMapGrid's legend string (views.go:615-617); consumed by both the renderer's legend line and the overlay's glyph page — single source (FR-005) |
| **headerAnatomy** | ordered `{element, meaning, whenShown}` rows | Every headerView element + every conditional badge: running/PAUSED (and ENDED when spec 044 lands), governed-speed suffix, [degraded], [llm: …], [suppressed: …], disconnected form |
| **dockTabEntry** | `{key, name, purpose}` | From paneNames/dockTabKey (chronicle/metatron/villagers) |
| **helpLesson** | `{id, title, body string}` | The pull-reference seam: an empty table today + contract comment; future first-occurrence-lesson feature appends entries only (contracts/help-content.md, SC-006) |

## Derivations & invariants

- **Keymap sweep (SC-003)**: a test walks every mode's overlay table and asserts each
  listed key is handled by that mode's handler (and, inverse, that every handler-listed
  key appears in exactly one tier). Mechanical anti-drift gate.
- **Legend identity (FR-005)**: `renderMapGrid`'s legend line is rendered *from* the
  glyph table; the overlay's glyph page renders the same table long-form. One edit
  point.
- **No-LLM identity (SC-004)**: content tables are consts; overlay renders with nil
  `status`/`replica`.
- **Exact-height**: `helpPanelView` output passes `clipContent` + styleBox sizing;
  render sweep tests include help states at all sizes.
