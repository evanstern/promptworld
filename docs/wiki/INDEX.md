# promptworld grounding wiki

Code-grounded corpus for the promptworld daemon substrate. Every note is pinned to the
commit it was verified against; a change to any file in a note's `sources:` invalidates
that note (re-pin with `/grounding-wiki:wiki-update`).

## Orientation

- [[overview]] — the system's shape: always-on daemon, attachable clients, event-sourced world
- [[design-grounding]] — the TASK-1 grounded assumptions the code implements
- [[project-rename]] — the 2026-07-22 rename script-world → promptworld: repo, module, binary, env var, data dir

## Time & simulation

- [[game-clock]] — 1 tick = 1 game second; speeds 1x–max; epoch day 1 06:00
- [[sim-loop]] — single-goroutine fixed-timestep loop; command intents; auto-slow
- [[sim-state-reducer]] — State + Apply: the single mutation path, live and replay
- [[deterministic-rng]] — per-decision PCG from (seed, purpose, tick, index); no RNG state
- [[executor]] — agent bodies: needs, intents, death, terrain overlays, walls/axes/paths
- [[gru]] — the nocturnal predator: wounds not kills, light/shelter safety, rumor fuel
- [[reflex-policy]] — survival decision ladder + deterministic BFS pathfinding
- [[mental-maps]] — per-agent private spatial knowledge: explored terrain + place-facts with provenance/freshness, gating target resolution, grown by a perception sweep, rendered in the prompt
- [[event-types]] — the event taxonomy and payload shapes

## Persistence

- [[world-save-directory]] — one dir = one run; manifest, layout, separability
- [[world-migration]] — v1→v2 snapshot-cut migration: keep the people, reset the land
- [[worldmap-generation]] — seeded terrain (water/woods/forage/rock/dens); regenerated, never stored
- [[event-log]] — append-only SQLite events table; seq contiguity; source of truth
- [[snapshots]] — hash-verified recovery accelerator; cadence and fallback chain
- [[world-tuning]] — spec-048 world tuning manifest (tuning.json): five promoted doctrine dials, clamp-validated at boot, event-logged as sim.tuning_applied, replay file-independent

## Interface

- [[ipc-protocol]] — JSON-lines over UDS: requests, responses, pushes, status shape
- [[ipc-server]] — sessions, gapless subscribe-replay, overflow drop, long-path sockets
- [[ipc-client]] — dial, request correlation, push demux
- [[tui-client]] — Bubble Tea four-pane client over a live log-shipped replica
- [[cli-promptworld]] — the single binary's subcommands and exit discipline
- [[instance-manager]] — machine-wide ps, worlds home, name-or-path addressing, advisory registry

## Inference & minds

- [[llm-orchestrator]] — two-tier model traffic: routing, spend ceiling, degraded mode
- [[llm-provider-health]] — operator-facing dead-tier/tool-silence conditions: boot+periodic model-existence preflight, worker-side tool-silence detector, daemon/status/TUI surfacing
- [[cognition]] — the cognition horizon: Fibonacci-point decision registry, seconds-per-point calibration, deterministic LLM-vs-reflex routing by staleness budget, and the adaptive-throttle debt/governor feedback controller
- [[tool-registry]] — one registry for every agent capability: derived vocabulary/validation/durations/rosters, boot-time coverage gates
- [[tool-loop]] — the bounded agent tool-use loop driver: submit/dispatch/feed-back, one-landed-action cardinality, shared by the villager planner and the guardian's turn
- [[bundle-tools]] — drop-in persona/tool bundles (spec 036): manifest/Starlark tools compiled to whitelisted effect batches, boot-frozen, sandboxed, replay-deterministic
- [[agent-mind]] — personas, souls, memory window, and the planner driver
- [[memory-retrieval]] — spec-042 embedding relevance: vectors recorded at emission, situation vectors, the three-term selector, shadow/on gating, divergence evidence
- [[decision-context]] — spec-043 per-turn context grounding: the fixed-order block inventory (self-history, need trajectories, plan echo, budgeted memories/journal), the drop priority under the size budget, and the deliberate absences
- [[agent-journal]] — the agent-authored journal: per-villager markdown notebook, one rune budget gated in the reducer, four roster tools
- [[social-fabric]] — relationships, rumors, debts, secrets, conversations
- [[nightly-consolidation]] — sleep-triggered soul digestion behind the persona firewall
- [[chronicle]] — the narrated story feed: cloud narrator, snapshot-carried catch-up ring
- [[morgue]] — the run's legacy document: per-death epitaphs + run-end summary from a deterministic genesis replay fold, narrated epilogues blockquoted, charter/orders evidence alignment, regenerable morgue.md
- [[guardian]] — the guardian (spec 052/TASK-121 renamed from "Metatron", every serialized string frozen byte-identical): console + system-authored turns, omen/vision influence, standing-order agency, clock-control meta tools, charges, the editable charter
- [[guardian-miracles]] — charge-priced world edits ("workings" to the player, spec 052): time snap, item grant, entity move/remove, gratis doctrine, shift-semantics re-base
- [[guardian-orders]] — the standing-orders subsystem: event-sourced watches compiled to free structural predicates, live matching, system-authored triggered turns, fuzzy confirm, daytime-omen deferral
- [[skin]] — the spec-052 runtime skin substrate: one token lookup (world skin.json → compiled default table → the token path) supplies every fiction display string (guardian name/epithet/vocabulary, curriculum stage identities); the event log and every serialized identifier stay skin-free
- [[governance]] — norms and votes: the daily meeting under an event-sourced convention, relationship-driven law, the village charter
- [[curriculum-ladder]] — the spec-046 four-stage teaching ladder: immutable world stage with informed override, the guardian's tool ceiling + stage-1 instruction lock (default/tutor presets), curriculum.* unlock events with auditable evidence, the per-user unlocks record, seeded exercises

## Lifecycle & quality

- [[daemon-lifecycle]] — recovery, pidfile, meta validation, signals, shutdown
- [[testing-strategy]] — determinism harness, integration, binary-level e2e scenarios
