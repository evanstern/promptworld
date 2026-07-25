# Contract: the skin system (spec 052)

**This is the published contract TASK-115 and TASK-117 consume (reorient D2).**
Once this spec's PR merges, downstream features MUST obtain every fiction
string via §2's lookup and MUST register new tokens per §4. The doc twin of
this contract lives in `docs/design/tui/patterns/skin-tokens.md` (runtime
section added by this feature).

## §1 The skin bundle (`<world>/skin.json`, optional)

```json
{
  "name": "Raven",
  "epithet": "raven",
  "tab_label": "raven",
  "voice": "You speak in riddles and barter, never commanding…",
  "strings": {
    "skin.guardian.working_noun": "trick",
    "skin.guardian.notes_label": "the raven's ledger"
  },
  "stages": {
    "stage-1": {"name": "The Whisper", "line": "…"},
    "stage-2": {"name": "The Bargain", "line": "…"}
  }
}
```

- **Identity fields** (`name`, `epithet`, `tab_label`): single-line,
  length-capped (name ≤ 40, epithet/tab_label ≤ 20 runes, no control chars);
  an invalid field falls back to the default skin's value with one notice.
- **`voice`**: the persona-voice text composed into the guardian's prompt at
  the editable-zone SOUL seam; same cap/validation as bundle SOUL fragments
  (≤ 4,000 chars). NEVER composed after the fixed frame.
- **`strings`**: token-path → value overrides; unknown token paths are
  ignored with a notice (typo never bricks). Systems/telemetry surfaces have
  no tokens, so they are unskinnable by construction (D10).
- **`stages`**: display identities per neutral stage id; missing stages fall
  back to default.
- **Loading**: boot-frozen (SetBundles/SetStage discipline). Missing file →
  default skin silently; malformed JSON → default + one notice.

## §2 Resolution

`world skin.json override → compiled default table → the token path itself`
(a rendered `skin.guardian.name` literal means a bug; the token-completeness
test fails before it ships). Identity fields are themselves tokens
(`skin.guardian.name/epithet/tab_label`) — the typed fields are convenience
accessors over the same table.

## §3 The default skin (normative identity)

| Token | Default value |
|---|---|
| `skin.guardian.name` | `Guardian` |
| `skin.guardian.epithet` | `guardian` |
| `skin.guardian.tab_label` | `guardian` |
| `skin.guardian.working_noun` | `working` (plural `workings`) |
| `skin.guardian.family_label` | `guardian` (chronicle Type-column alias for the frozen `metatron` family) |
| `skin.guardian.notes_label` | `the guardian's notes` |
| `skin.stage.stage-1.name` … `stage-4.name` | The Voice / The Written Word / The Craft / The Stewardship |
| `skin.stage.stage-N.line` | the existing one-line identities (internal/skin) |

The full table grows during implementation as the sweep converts literals;
every added token lands in this table, its doc twin, and the completeness
test in the same commit. Vision/omen display terms are default-skin-retained
vocabulary; they get tokens (`skin.guardian.vision_noun`, `.omen_noun`) so
custom skins may re-voice them, while their tool ids and recorded payloads
stay frozen.

## §4 Downstream obligations (TASK-115/117 and later)

1. No new bare fiction literal — every fiction string is a token lookup; the
   denylist sweep test enforces this repo-wide.
2. New tokens: add to the default table + doc twin + completeness test in the
   introducing PR (the design page's control-table `skin-token` column names
   it).
3. Non-fiction chrome (telemetry, key hints, structural labels) stays
   literal with `—` in the skin-token column (skin-tokens.md rule 5).

## §5 What is never skinnable

- The fixed frame (spec 021): `metatronNonNegotiables`, the initiative
  frame, registry-derived tool guidance structure — compile-time constants
  appended last on every composition path. Skins insert only at the
  editable zone.
- Mechanics: tool roster/ids, costs, charge arithmetic, reducer rules,
  stage ceilings.
- The event log (spec ruling 1): recorded types, payloads, memory text,
  correlation ids — skin-free forever; skinning is render-time only.
- Systems/telemetry tab content (D10) — no tokens exist for it.

## §6 Frozen serialized vocabulary (annotate, never rename)

Event types `metatron.*`/`curriculum.*`; IPC `metatron_chat`/
`metatron_status`; JSON keys/tags `metatron_charges`; llm.json kinds
`metatron`/`metatron_watch`; tool ids `send_vision`/`send_omen`/
`work_miracle` + kind enum; paths `charter.md`, `skills/`,
`capabilities.json`, `metatron/soul.md`, `metatron/transcript.md`;
correlation prefixes `turn-metatron-`/`watch-metatron-`; payload value
`"omen"` origin. (Normative list: research.md R4.)

## §7 Status surface additions (additive, omitempty)

`metatron_status`/world status gains resolved display facts: skin name,
epithet, tab label, family label, stage identities. Clients render from
status only; absent fields = default skin (old-daemon compatibility).
