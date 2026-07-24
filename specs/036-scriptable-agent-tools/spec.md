# Feature Specification: Scriptable Agent Tools — Pluggable Bundle-Defined Tools

**Feature Branch**: `036-scriptable-agent-tools`

**Created**: 2026-07-24

**Status**: Draft

**Input**: User description: "Scriptable agent tools: pluggable script-defined tools over an engine API (TASK-85). Make agent/angel tools scriptable and pluggable instead of hard-coded. Core reframe: a tool script is a pure function (args + read-only world view -> event batch + narration); the engine validates the batch against the approved event vocabulary and lands it through the existing door — scripts cannot mutate state directly or invent event types. v1 scope: instantaneous angel/expressive tools are scriptable; tick-simulated villager world verbs remain native. Persona bundles install by dropping a folder into the world dir with boot-time validation. Determinism preserved. Phased: manifest-only declarative tools first, script runtime second, widened event vocabulary third. Dogfood: at least one existing angel tool re-expressed as a loadable bundle."

## Clarifications

### Session 2026-07-24

- Q: When one tool inside a persona bundle fails validation, does the whole persona reject or just that tool? → A: Per-tool rejection — the invalid tool is skipped with a loud boot error; the persona's charter, capabilities, and remaining valid tools still load. Structural failures (invalid charter or capabilities file) reject the whole persona.
- Q: Can a bundle shadow a built-in tool of the same name? → A: No — built-ins always win; a colliding bundle tool is skipped with a boot warning. Bundle-vs-bundle collisions are also resolved deterministically (first loaded wins, later skipped with a warning). The dogfood equivalence test runs the bundle twin in a world/persona where the built-in is not granted.
- Q: How much world state can a script read? → A: Invoker-scoped — the read-only world view exposes exactly what the invoking agent's perspective legitimately exposes (plus the invocation arguments), not omniscient world state. Scripted tools inherit the invoker's epistemic limits and cannot leak knowledge the invoker does not have.
- Q: Does a failed invocation consume charge? → A: No — failures are free. Charge is deducted only when a batch validates and lands; validation rejections, script errors, and budget exhaustion leave the charge balance untouched. (Accepted "for now" — revisit if free retries prove abusable; step/memory caps bound the abuse surface.)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Declarative tool bundle: drop a folder, get a new tool (Priority: P1)

A world author wants to give their angel a custom expressive tool — say, a "teleport" that relocates a villager and announces "vanished in a poof of smoke" — without touching or rebuilding the engine. They write a small bundle folder containing a tool manifest that names the tool, describes it, declares its parameters, declares which primitive effects it produces (from the engine's approved effect vocabulary), and templates the narration. They drop the folder into the world directory and restart the world. The tool appears on the angel's roster exactly like a built-in tool: the angel's language model sees its schema and description, can invoke it, the declared effects land in the world, and the narration is broadcast to nearby villagers.

**Why this priority**: This is the thinnest slice that proves the entire pipeline — load from disk → validate → register → derive the LLM-facing schema → route an invocation → land effects. Every later story builds on this pipeline. It requires no scripting runtime at all: a declarative manifest is a parameterized macro over effects the engine already knows how to apply.

**Independent Test**: Author a teleport bundle by hand, drop it into a test world, boot, and confirm (a) the tool is listed on the angel's roster with correct schema, (b) invoking it moves the target villager and broadcasts the narration, (c) the world's event log records only approved event types.

**Acceptance Scenarios**:

1. **Given** a world directory containing a valid manifest-only tool bundle, **When** the world boots, **Then** the tool is registered and appears in the angel's available tools with the manifest's name, description, and parameter schema.
2. **Given** the bundle tool is registered, **When** the angel invokes it with valid arguments, **Then** the declared effects are applied to the world through the same validated path as built-in tools, and the narration is broadcast.
3. **Given** a manifest that declares an effect type outside the engine's approved vocabulary, **When** the world boots, **Then** the bundle is rejected with an error naming the file and the offending effect type, and the world boots without it.
4. **Given** an invocation whose arguments fail the manifest's parameter schema, **When** the tool is invoked, **Then** the invocation is rejected with a descriptive error and no effects land.

---

### User Story 2 - Dogfood: an existing built-in angel tool becomes a bundle (Priority: P2)

A maintainer re-expresses one existing built-in angel tool as a loadable bundle, producing behavior indistinguishable from the built-in form: same name, same LLM-facing schema, same effects, same narration style, same charge cost against the angel's miracle economy. This proves the bundle format is expressive enough to carry real, already-shipped tools and exercises the full pipeline against a known-good baseline.

**Why this priority**: Converting a real tool is the strongest evidence the abstraction is right — it either round-trips cleanly or exposes gaps in the manifest format before third-party authors hit them. It depends on Story 1's pipeline existing.

**Independent Test**: Side-by-side comparison: invoke the built-in tool in one world, and its bundle twin in an identical world where the built-in is not granted, with the same arguments; the resulting world states, event sequences, narrations, and charge deductions must match.

**Acceptance Scenarios**:

1. **Given** a bundle re-expressing an existing built-in angel tool, **When** it is invoked with the same arguments as the built-in in an identical world (where the built-in is not granted), **Then** the resulting events, narration, and charge cost are equivalent.
2. **Given** the bundle version is installed in a world where the built-in of the same name IS granted, **When** the world boots, **Then** the built-in wins: the bundle tool is skipped with a boot warning and exactly one tool with that name is available.

---

### User Story 3 - Scripted tools with conditional logic (Priority: P3)

A world author wants a tool whose behavior depends on world state — e.g. a "cast light" tool that behaves differently at night than by day, or a blessing that only applies to villagers below a health threshold. They add a script file to the tool's bundle. The script is a pure function: it receives the invocation arguments and a read-only view of the world, and returns a batch of effects plus narration. It cannot mutate the world directly, invent effect types, perform input/output, read the wall clock, or draw randomness from outside the world's own seeded sources. The engine validates the returned batch against the approved vocabulary before landing it, exactly as in Story 1.

**Why this priority**: Conditional logic is what makes tools feel alive, but it is only safe once the validated-batch pipeline (Story 1) is proven. The scripting runtime is the largest new machinery in the feature, so it rides after the zero-runtime slices de-risk everything around it.

**Independent Test**: Author a scripted tool whose output branches on observable world state; invoke it under both branch conditions in a test world and confirm each branch produces its declared effects; confirm a script attempting a forbidden operation (I/O, clock, unseeded randomness, over-budget computation) is stopped without any state change.

**Acceptance Scenarios**:

1. **Given** a bundle with a script that branches on world state, **When** invoked under each condition, **Then** each branch's effects and narration land correctly.
2. **Given** a script that attempts to emit an effect type not declared in its manifest, **When** invoked, **Then** the whole batch is rejected, no effects land, and the failure is reported to the invoking agent.
3. **Given** a script that exceeds its execution budget (steps or memory), **When** invoked, **Then** execution is aborted, no effects land, and the failure is reported.
4. **Given** a script that does not parse or fails static checks, **When** the world boots, **Then** the bundle is rejected at boot with an error naming the file and problem.

---

### User Story 4 - Persona bundle: a character installs as one folder (Priority: P4)

A world author installs a complete persona — e.g. `gandalf/` containing a soul/charter fragment, a capabilities grant file, and a `tools/` folder of manifest+script tools — by dropping the single folder into the world directory. On boot, the world validates the whole bundle (manifest schemas, declared effects within the approved vocabulary, scripts parse, execution caps configured, capabilities well-formed) and the persona's identity, permissions, and tools all take effect together.

**Why this priority**: This is the headline authoring experience — shareable, installable characters — but it composes pieces the earlier stories already proved (tool loading + existing per-world identity/capability file mechanisms). It is packaging, not new machinery.

**Independent Test**: Drop a complete persona bundle into a fresh world; boot; confirm the persona's charter text is in effect, its capability grants apply, and every tool in its `tools/` folder is on the roster. Then corrupt one manifest and confirm boot reports the specific problem while the rest of the world remains usable.

**Acceptance Scenarios**:

1. **Given** a valid persona bundle folder in the world directory, **When** the world boots, **Then** the persona's charter fragment, capability grants, and all bundled tools are active.
2. **Given** a persona bundle with one invalid tool, **When** the world boots, **Then** that tool is rejected with a specific error naming the file and problem, and the persona's charter, capability grants, and remaining valid tools still load.
3. **Given** a persona bundle whose charter fragment or capability grants file is invalid, **When** the world boots, **Then** the entire persona bundle is rejected with a specific error (a persona must not run with broken identity or permissions).

---

### Edge Cases

- Two bundles (or a bundle and a built-in) declare the same tool name — load order and collision resolution must be deterministic and reported.
- A bundle is removed or edited after a world has already recorded events produced by its tools — replay must still work, because recorded events are self-contained data and replay never re-executes tool logic.
- A script returns an empty effect batch (narration only, or nothing) — allowed or rejected must be defined; narration-only is a legitimate expressive tool.
- A script returns effects targeting entities that no longer exist by the time the batch is validated — the batch fails validation cleanly, no partial application.
- A manifest declares zero parameters, or parameters with types the LLM-facing schema cannot represent.
- An invocation arrives while the tool's charge cost exceeds the angel's remaining charge — rejected by the existing economy rules, consistent with built-in tools.
- A bundle folder contains extra unknown files — ignored or rejected must be defined (default: ignored, so bundles can carry docs/licenses).
- Execution budget exhaustion mid-script must not leave any partially applied state (all-or-nothing landing).
- A world with bundles is opened on a build of the engine whose approved effect vocabulary no longer includes an effect a manifest declares — bundle rejected at boot with a clear versioning error.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The system MUST load tool definitions from bundle folders placed in a world's directory at world boot, without engine rebuild or code change.
- **FR-002**: A tool bundle MUST consist of a manifest (name, description, parameter schema, declared effect types, charge cost, narration template) and optionally a script; manifest-only bundles MUST be fully functional as parameterized macros over the approved effect vocabulary.
- **FR-003**: The system MUST derive the LLM-facing tool schema and prompt description from the manifest using the same derivation path as built-in tools, so loaded tools are indistinguishable to the invoking agent.
- **FR-004**: Every effect emitted by a loaded tool (declarative or scripted) MUST be validated against the engine's approved effect vocabulary and against the manifest's declared effect list before landing; batches MUST land atomically through the same validated entry point as built-in expressive tools — all-or-nothing, with no partial application.
- **FR-005**: Loaded tools MUST NOT be able to mutate world state directly, invent effect types, or bypass validation; the only way a loaded tool changes the world is by returning an effect batch for the engine to validate and apply.
- **FR-006**: Scripts MUST execute as pure functions of (invocation arguments, read-only world view) with no filesystem, network, or other I/O access, no wall-clock access, and no sources of randomness other than those seeded from the world's own recorded state. The world view MUST be invoker-scoped: it exposes only what the invoking agent's perspective legitimately exposes, so a scripted tool cannot read or leak knowledge its invoker does not have.
- **FR-007**: Script execution MUST be bounded by configurable step and memory caps; exceeding a cap MUST abort the invocation with no state change and a reported failure.
- **FR-008**: The system MUST validate bundles at boot — manifest schema correctness, declared effects being a subset of the approved vocabulary, script parse success, and execution caps configured — and reject invalid bundles with errors naming the file and the specific problem, while still booting the world with the remaining valid tools.
- **FR-009**: A persona bundle (identity/charter fragment + capability grants + tools folder) MUST install as a single dropped folder, with all its parts validated and activated together at boot.
- **FR-010**: Loaded tools MUST participate in the existing charge/miracle economy: each manifest declares a charge cost, invocations deduct it under the same rules as built-in tools, and a failed invocation MUST NOT consume charge.
- **FR-011**: Determinism MUST be preserved: replaying a world whose event log contains bundle-tool events MUST reproduce identical state hashes, and replay MUST NOT require re-executing tool logic (recorded events are self-contained data).
- **FR-012**: Tool-name collisions MUST be resolved deterministically: built-in tools always win over bundle tools of the same name, and among bundles the first-loaded definition wins (load order itself deterministic); every skipped definition MUST be reported with a boot warning.
- **FR-013**: v1 scope boundary: only instantaneous expressive/angel tools (effect-batch tools) are loadable from bundles; tick-simulated villager world verbs (e.g. hunting, foraging, building) remain native and are explicitly out of scope.
- **FR-014**: At least one existing built-in angel tool MUST be re-expressed as a loadable bundle with equivalent observable behavior (schema, effects, narration, charge cost), as ongoing proof the bundle format is sufficient.
- **FR-015**: Failures at invocation time (validation rejection, script error, budget exhaustion) MUST be reported back to the invoking agent in the same way built-in tool failures are reported, so the agent can react in-world.

### Key Entities

- **Tool Bundle**: A folder holding one tool: a manifest and optionally a script. The unit of authoring and distribution for a single tool.
- **Tool Manifest**: The declarative contract of a tool — name, description, parameters, declared effect types, charge cost, narration template. Everything the engine needs to register, schema-derive, validate, and price the tool.
- **Tool Script**: Optional pure-function logic that computes the effect batch and narration from arguments plus a read-only world view. Sandboxed, deterministic, budget-capped.
- **Approved Effect Vocabulary**: The engine's closed, audited set of effect types loaded tools may emit. This vocabulary — not an imperative API — is the extension surface; it grows only by deliberate engine-side addition.
- **Effect Batch**: The atomic unit a tool invocation produces: zero or more effects plus narration, validated and landed all-or-nothing.
- **Persona Bundle**: A folder composing an identity/charter fragment, capability grants, and a set of tool bundles into one installable character.
- **Read-only World View**: The snapshot exposed to scripts — invoker-scoped (only what the invoking agent legitimately perceives), sufficient to make decisions, incapable of causing mutation.
- **Charge Cost**: The price of an invocation against the angel's existing miracle economy, declared per tool in the manifest.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A world author can add a new expressive tool to a world by creating one folder and restarting the world — zero engine code changes, zero rebuilds — and complete the whole authoring loop (write manifest, drop folder, boot, invoke) in under 30 minutes for a manifest-only tool.
- **SC-002**: 100% of effects landed by loaded tools pass vocabulary and declaration validation; no unapproved or undeclared effect type ever reaches the world's event log, under both normal use and adversarial test scripts.
- **SC-003**: Replaying any world whose history includes bundle-tool events reproduces bit-identical state hashes, including after the bundle files are modified or deleted.
- **SC-004**: One existing built-in angel tool ships as a bundle whose observable behavior is equivalent to the built-in original in side-by-side comparison (same events, narration, charge cost for the same inputs).
- **SC-005**: Every invalid bundle is rejected at boot with an error that names the file and the specific problem; a world with a mix of valid and invalid bundles still boots and serves the valid ones.
- **SC-006**: No script invocation can run unbounded: 100% of over-budget or forbidden-operation attempts abort with zero state change.

## Assumptions

- **Boot-time loading only (v1)**: bundles are discovered and validated at world boot; hot-reloading tools into a running world is out of scope.
- **Failure policy for invalid bundles**: an invalid bundle is skipped (with a specific, loud boot error) and the world boots with the remaining valid tools; boot is not aborted. Within a persona bundle, one invalid tool rejects that tool, not the whole persona; an invalid charter or capabilities file rejects the whole persona (decided in Clarifications 2026-07-24).
- **Failed invocations don't consume charge**: an invocation that fails validation, errors, or exceeds budget leaves both world state and the angel's charge balance untouched (decided in Clarifications 2026-07-24; revisit if free retries prove abusable).
- **Replay never re-executes tools**: the event log stores self-contained effect data, so replay is independent of bundle presence or content; determinism obligations therefore attach to live execution only.
- **Narration-only tools are legitimate**: an empty effect batch with narration is a valid expressive tool result.
- **Unknown extra files in bundles are ignored**, so bundles can carry documentation.
- **Collision precedence**: built-in tools always win over bundle tools of the same name; among bundles, first-loaded wins; all collisions are reported at boot (decided in Clarifications 2026-07-24).
- **Trust model**: bundle authors are semi-trusted (the world owner chose to install the folder); the sandbox exists to guarantee determinism and resource bounds, and to keep authoring mistakes from corrupting worlds — not to defend against a hostile author with filesystem access to the world directory (who could edit world files directly anyway).
- **Scripting runtime selection** (a hermetic, deterministic, embeddable runtime is required; the concrete choice) is a design/plan-phase decision, not a spec-level one.
- **Existing mechanisms are reused**: the angel's capability-grant file format, charter/skill file loading, the tool registry/derivation path, and the miracle-charge economy already exist and are extended, not replaced.
- **Widening the approved effect vocabulary** (e.g. adding a "heal" effect) is deliberately out of this feature's v1 scope beyond what existing primitives provide; the feature ships the extension surface, and vocabulary growth happens one audited effect at a time in follow-up work.
