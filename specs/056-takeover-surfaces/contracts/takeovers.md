# UI Contract: takeover surfaces (spec 056)

Design authority: `docs/design/tui/overlays/ceremony.md` +
`overlays/postmortem.md` (amended to shipped in the same PR).

## §1 Trigger / precedence grammar

| Event / input | State | Effect |
|---|---|---|
| `curriculum.stage_unlocked` | no takeover | ceremony opens immediately |
| `curriculum.stage_unlocked` | postmortem open | deferred (replay surfaces carry it); postmortem uninterrupted |
| `curriculum.stage_unlocked` | ceremony open | newest replaces (same-kind, non-stacking); both replayable |
| `run.ended` | any | postmortem opens, replacing an open ceremony (postmortem always wins) |
| connect | `runEnded()` true | postmortem auto-opens |
| `esc` | takeover open | dismiss one layer; ENDED posture/read-only keys unaffected |
| `p` | `runEnded()` true | reopen postmortem from anywhere; inert on live worlds |
| `q` | ceremony open | detach with the D13 "world keeps running" framing |
| `q` | postmortem open | plain ended-world quit (no keeps-running framing) |
| `?` | takeover open | takeover keeps the body slot (help yields) |

## §2 Postmortem composition (top → bottom)

1. `THE RUN HAS ENDED` title + narrated run-end line (final cause).
2. Report card — SCORED RUNS ONLY (concluded markers), absent on ambient
   worlds and whenever rubric data is unavailable (honesty rule).
3. `morgue — no-blame evidence` rows: name · day · cause · closest charter
   observation (≤ death); rotated-away observations render unknown.
4. Footer hints: `esc dismiss · q quit`.

## §3 Ceremony composition (top → bottom)

1. `<STAGE NAME> — unlocked` title (skin-resolved).
2. Narrated chapter — the D6 player-authorship voice, skin tokens.
3. Report card (the instrument, authoritative — same renderer).
4. Footer hints: `esc dismiss · q — the world keeps running`.

## §4 Shared renderer contract (D5)

`reportCardView(definition, facts, mode, width)`: identical rows at every
site; mode changes only the marker vocabulary (concluded met/missed, live
met/pending); term language matches the exercise panel's gauge vocabulary
(one vocabulary corpus-wide). Composable into spec 053's consoleCard seam;
console production wiring belongs to TASK-115.

## §5 Replay surfaces

Postmortem: `p` + auto-open-on-attach (the morgue content is its own pull
surface). Ceremony: `promptworld stages` (facts, shipped) + the `?` overlay
ceremony-replay section (stored content, never regenerated). Both surfaces
existing is an explicit AC.

## §6 Non-goals

No console card production (TASK-115); no production unlock emission
(TASK-119); no mouse targets (parity gaps recorded from birth); no new IPC
fields; no regeneration of narrative content.
