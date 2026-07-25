# Research: Grounded feedback layer (spec 063)

## R1 — Explain: composition and gating

**Decision**: `explain` is a registry tool (internal/tool) with
expressive-class effect and EMPTY events (the pause/start meta-tool
precedent — no whitelist entry, nothing lands). Its handler composes fact
sheets tool-side from: `tool.LoopRosterMetatron()`/the registry's declared
schemas + `MetatronToolGuidance` derivations, the miracle kind/cost table,
charge doctrine constants (cap, genesis, regen cadence — the spec-050
export), decision classes (the TASK-63 vocabulary), and the map-glyph
legend vocabulary. Scoping: the sheet reflects the world's EFFECTIVE grant
(post `applyStageCeiling` + manifest intersection) and says so when a
cataloged tool is ungranted. Unknown topic → the explainable-topic catalog
as the result (repairable, the rejected_gate voice without being a gate).

**Rationale**: "described ≡ declared by construction" is already the
registry's doctrine (wiki metatron note); explain extends the same property
to the player-facing answer path.

## R2 — Tutor-lane mechanics (zero-cost by construction)

**Decision**: read-only tools (a new registry attribute or the
expressive+empty-events class plus a `ReadOnly` marker) do NOT consume the
turn's one mediated act: the loop driver treats them like converse-adjacent
work — the turn may call explain multiple rounds and still land one act or
none. No charge path exists (charges are spent by landers; explain has no
lander). Faith: nothing to exclude yet (TASK-118 unshipped) — the
invariant is recorded as a sweep-test obligation for when the field exists.
Rubric hygiene: extend the exercise catalog sweep to assert no rubric term
matches tutor-lane telemetry types/prefixes.

**Rationale**: the task's contract line ("explaining is speech, not an
act"); structural exclusion beats prose exclusion (spec 021's lesson).

## R3 — Tutor guide seam

**Decision**: `persona.TutorGuide` compiled constant (TutorCharter
sibling, same ≤4,000-char cap discipline); composed by the guardian's turn
assembly ONLY when the world's charter preset is `tutor` — inserted in the
editable zone after the charter (and after skin voice/bundle SOULs), before
skills/frame. Non-tutor worlds: byte-identical composition (tested). The
guide's content instructs: orient first, answer mechanics via explain,
point at ?/keys for UI questions, never invent numbers.

**Rationale**: standing resolution 2 — the stage-3 skill lock stays
untouched; compiled-in text is tamper-proof like the preset charters.

## R4 — Cheap chain routing + budget

**Decision**: new route kind (the `metatron_watch` precedent):
`KindReportCard` in internal/llm, default-routed to the cheap/local chain,
configurable per llm.json like every kind; per-call token budget of the
watch-confirm class; producer-side debounce (R5) bounds call volume. No
LLM configured / route failure / budget exhausted → skip the note,
deterministic parts only, one dim log line, no error theater.

## R5 — Card production, storage, and doors

**Decision**: a stopping-point consumer in internal/guardian (the digest
worker's notify-consumer pattern): triggers on `run.ended`,
`curriculum.exercise_passed` (and scenario fail = run end), and
pause-episode starts (`clock.paused` debounced — at most one card per
pause episode, and only when new guardian activity exists since the last
card). Grading inputs: the recorded `cog.tool_call` verdict trail (TASK-63
events), rubric evidence, the charter-revision timeline
(`charter_observed` fingerprints) + effective charter text, and R1's fact
sheets. Storage: pre-end cards land as a whitelisted prose event
`guardian.report_card{fingerprint, note, citations[]}` through
InjectSocial (reduced to latest-card state + log history); the RUN-END
card rides the existing run-end epilogue path (`morgue.epilogue` agent −1
carries the note beside the narrator's epilogue — the postmortem already
renders that section) so the ended-world door narrowing needs no new
entry. Client: console card seam composes checklist (127's renderer) +
the stored note; unseen-badge between stopping points.

**Rationale**: every door/pattern is a shipped precedent; "stored, never
re-graded" falls out of event-sourcing the note.

**Alternative considered**: status-carried card — rejected: not durable,
invisible to replay/postmortem, and would re-grade on every poll.

## R6 — The `?` guardian section (D9)

**Decision**: a new help section in internal/tui/help.go rendering from
status only: stage identity (skin StageName), the effective granted-verb
summary (the existing `Status.GrantedTools` rendering vocabulary), and one
compiled example ask per verb (skin-token-resolved nouns). Byte-identical
per identical status (the spec-045 no-LLM floor invariant, extended to
status-derived sections per the ceremony-replay precedent in
overlays/help.md). Content contract amendment recorded on the page.

## R7 — Skin tokens

New tokens expected: `skin.guardian.example_ask.<verb>` family (or one
composed template per verb), card labels (`report card`, attribution
header), guide framing strings. All per the contract §4: default table +
doc twin + completeness test in the introducing commit.

## R8 — Design pages in scope

overlays/help.md (D9 section), pages/guardian-console.md (card production
real), patterns/skin-tokens.md (token twin), plus keymap.md only if any
binding changes (none expected — the card has no keys; badge reuses the
existing pattern). Re-pins throughout; wiki/tool-registry + player docs in
the post-merge re-ground.
