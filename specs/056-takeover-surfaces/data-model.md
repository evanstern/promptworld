# Data Model: Takeover surfaces (spec 056)

No persistence, no wire changes.

## Takeover owner (client state)

| Field | Type | Meaning |
|---|---|---|
| takeover | none · ceremony · postmortem | who owns the body slot |
| ceremonyDeferred | bool | unlock arrived during a postmortem; replay surfaces carry it |
| postmortemDismissed | bool (per attach) | suppresses re-auto-open this attach; `p` overrides |

**Transitions**: stage_unlocked → ceremony (defer if postmortem);
run.ended → postmortem (replaces ceremony); esc → none; connect ∧
runEnded() ∧ !dismissed → postmortem; `p` ∧ runEnded() → postmortem.
Takeovers never stack; help overlay yields to takeovers.

## Report card (renderer input)

(exercise definition, rubric facts, mode ∈ {concluded, live}) → rows of
{term plain-language, marker, backing reference}. Marker vocabulary:
concluded = met/missed; live = met/pending. One view function; a
consoleCard wrapper for seam composition.

## Postmortem content (derived at open)

Run-end line (final cause), optional report card (scored ∧ data present),
morgue rows: {name, day, cause, closest charter observation ≤ death} from
replica facts; rotated-away observations render unknown honestly.

## Ceremony content (derived at open / replay)

{stage identity (skin-resolved), D6 authorship chapter line, proving
exercise, rubric evidence} — from the unlock event + compiled definition +
unlocks facts; stored, never regenerated.
