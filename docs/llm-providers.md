# LLM providers & routing — operator guide

How model traffic is configured in `llm.json` (in each world's save directory), as of
spec 024 (multi-provider routing, TASK-35, PR #52), spec 025 (robustness knobs,
TASK-72), spec 029 (metatron agency, TASK-27), and spec 034 (fresh-world defaults +
preflight, TASK-84). Formal shapes live in
`specs/024-provider-routing/contracts/llm-config.md`; this is the operator-facing
reference.

## Is a migration required?

**No.** The legacy two-entry shape (`"local": {...}, "cloud": {...}`) loads forever and
behaves byte-identically — it is derived internally into a two-provider registry named
`local` and `cloud` with exactly the pre-024 routes. Existing worlds boot untouched,
keep their persisted spend, and their calibration profiles keep matching by name.

Two rules:

- One file cannot mix both shapes — declaring `providers` **and** `local`/`cloud`
  together is a boot error.
- Config is read at boot only: edit `llm.json`, then restart the daemon
  (`promptworld stop <world> && promptworld start <world>`).

New worlds (`promptworld new`) are written in the v2 shape. As of spec 109 the fresh-world
default local provider is `gemma4:latest` with `tool_mode: "native"` and `parallel: 4` — a
gguf-served model, chosen because Ollama's build format (not model family or size) decides
whether JSON-Schema constraints are honored at all, and `gemma4:latest` both honors them
and tool-calls natively out of the box (see "MLX models silently ignore schema
constraints" below). `promptworld new` prints the model name and an
`ollama pull gemma4:latest` command to make first-run painless. `qwen3.6:latest` remains a
documented upgrade path for operators with the RAM to serve it (~24 GB), but a fresh world
is not written expecting one.

## The v2 shape: providers + routes

The example below is a hand-tuned, multi-provider upgrade — a live-proven division of
labor across three named providers — not what a fresh world is written with; see above
for the actual `promptworld new` default.

```json
{
  "monthly_budget_usd": 100,
  "providers": {
    "qwen":   { "transport": "openai_compat", "endpoint": "http://localhost:11434/v1",
                "model": "qwen3.6:latest", "parallel": 2, "endpoint_capacity": 4 },
    "cogito": { "transport": "openai_compat", "endpoint": "http://localhost:11434/v1",
                "model": "cogito:3b", "parallel": 4, "endpoint_capacity": 4,
                "tool_mode": "json" },
    "cloud":  { "transport": "anthropic", "model": "claude-opus-4-8",
                "input_usd_per_mtok": 5, "output_usd_per_mtok": 25,
                "api_key_env": "ANTHROPIC_API_KEY" }
  },
  "routes": {
    "planner":       ["qwen"],
    "conversation":  ["cogito", "qwen"],
    "meeting":       ["cogito", "qwen"],
    "consolidation": ["cloud"],
    "narrator":      ["cloud"],
    "drama":         ["cloud"],
    "metatron":      { "chain": ["cloud"], "no_fallback": true }
  }
}
```

This example is a division of labor: high-volume structured kinds (conversation turns,
meeting flavor) on the small parallel model, prose kinds on the quality model, the
nightly/narrative tier on cloud. `qwen3.6:latest` is gguf-served (honors schema
constraints, tool-calls natively) and is the documented upgrade for machines with the RAM
to serve a 24 GB MoE model — see "MLX models silently ignore schema constraints" below for
why an MLX-served model (e.g. `gemma4:12b-mlx`) does not belong in this slot.

### Providers

Each named entry declares one model source. Fields:

| Field | Meaning |
|-------|---------|
| `transport` | `"openai_compat"` (Ollama, LAN routers, 9router) or `"anthropic"` (official SDK) |
| `endpoint` | required for `openai_compat`; optional override for `anthropic` |
| `model` | model id at that endpoint |
| `input_usd_per_mtok` / `output_usd_per_mtok` | pricing; both 0 (or absent) = zero-priced |
| `api_key_env` | NAME of an env var holding the key (keys are never stored) |
| `api_key` | inline key — LAN-local routers only; wins over `api_key_env` |
| `parallel` | concurrent worker slots against this provider (1–16, warn-and-clamp) |
| `reasoning_effort` | hidden chain-of-thought posture; zero-priced default `"none"`, priced default omit |
| `tool_mode` | `"native"` (default) or `"json"` fallback envelope; `anthropic` transport ignores it. `"native"` is the fresh-world default because the shipped local model (`gemma4:latest`) tool-calls natively; cogito:3b needs `"json"` — measured live (TASK-52): it never function-calls natively. Inert on non-tool kinds (conversation/meeting), so set it on the entry regardless — a future chain edit then can't trip the native-mode failure |
| `endpoint_capacity` | opt-in cross-world concurrency bound for the endpoint (see leases below) |

Every knob that used to be per-tier is now per-provider. Zero-priced providers are
"local-class": never refused for budget reasons, seeded with local-class latency
bootstraps.

### MLX models silently ignore schema constraints

promptworld's JSON-reliability work (planner tool calls, conversation outcomes, the
utterance route) rests on **sampler-level constraint**: the caller sets a JSON Schema,
`internal/llm/providers.go` attaches it as an OpenAI-compat `response_format:
{type: json_schema}` envelope (or, for tool calls, native function-calling), and the
provider is expected to make invalid output unrepresentable.

**Ollama's MLX engine (`details.format: "safetensors"`) silently discards that envelope.**
Measured 2026-08-02 on the same harness across four models:

| model | `details.format` | schema-constrained JSON | native tool call |
|---|---|---|---|
| `cogito:3b` | gguf | valid | yes |
| `gemma4:latest` (8B) | gguf | valid | yes |
| `gemma4:12b-mlx` | safetensors (MLX) | **returns prose** | yes |
| `qwen3.6:latest` (36B MoE) | gguf | valid | yes |

The split tracks the **build format**, not the model family or size: `gemma4:12b-mlx`
fails head-to-head against its own gguf sibling `gemma4:latest`. All three constraint
mechanisms were tried against it and all three returned free prose instead of JSON:
OpenAI-compat `json_schema` strict, the same non-strict, and Ollama's own native
`/api/chat` `format` parameter.

**Check before you serve a model in this slot**: query the endpoint and read
`details.format`.

```sh
curl -s http://localhost:11434/api/show -d '{"model":"gemma4:12b-mlx"}' | jq '.details.format'
```

`"gguf"` is safe; `"safetensors"` (or any non-`"gguf"` value) means schema constraints
will be silently ignored by that model.

**The symptom, and why it is dangerous**: a schema-constrained call returns prose where
JSON was demanded. Nothing downstream distinguishes "the model refused to follow the
schema" from any other kind of malformed output, so it surfaces only as parse failures,
retries, and — in conversation/utterance routes — abandoned scenes. **Nothing in the
daemon log, the boot sequence, or `promptworld calibrate` reports the underlying cause**;
the provider reports success (HTTP 200, a normal-looking completion) the entire time. This
cost a 12-game-day soak (TASK-174/spec 103) before the cause was isolated to the provider,
not the code (spec 109).

### Routes: the chain IS the policy

Each call kind maps to an **ordered chain** of provider names. Membership means "meets
this kind's quality floor"; position means preference. There is no runtime scoring —
cost, latency, and quality are things **you** weigh when writing the chain.

At submission the chain is walked in order and the call goes to the first admissible
candidate. A candidate is skipped only for a mechanical, observable reason:

- `circuit-open` — its breaker is open (3 consecutive genuine failures),
- `wallet-exhausted` — the monthly ceiling is hit and this candidate is priced,
- `queue-full` / `busy` — its bounded queue is saturated.

Skips are recorded on the response (`skipped: name (reason)` in the one-shot output)
and visible in telemetry. If every candidate is inadmissible the call fails fast with
the head's reason and the sim's normal degrade paths take over (reflex layer,
conversation tolerance). Once a provider accepts a call, its failure is final — the
orchestrator never re-dispatches a failed call elsewhere.

`{"chain": ["x"], "no_fallback": true}` declares a kind that must fail rather than
substitute (chain must be a single entry).

**Continuity is automatic, not configured**: a conversation scene resolves its provider
once at scene start, and a planner/metatron tool-loop run pins its provider at run
start — no mid-scene or mid-thought model switches, including on the spec-025
in-loop retry.

### Validation

v2 configs are validated strictly at boot, and structural errors *fail the boot* with
the offender named: a route naming an undeclared provider, an accepted kind with no
route, an unknown kind key, a duplicate provider in a chain, an empty chain,
`no_fallback` with more than one entry, missing `transport`/`model`, `openai_compat`
without `endpoint`, or both shapes in one file. Tuning knobs (`parallel`, `tool_mode`,
`reasoning_effort`, `loop_max_rounds`, `max_tokens`) never fail the boot — out-of-range
values clamp with an operator warning.

### The `metatron_watch` kind (spec 029)

The angel's standing orders can be phrased fuzzily ("when Rowan seems
heartbroken…"), and confirming a fuzzy hit needs a model — but cheaply and rarely,
not through the metatron's premium conversational chain. `metatron_watch` is a
dedicated kind for exactly that: one bare yes/no call per confirm (16 tokens, no
tools, no tool loop), rate-capped per standing order so a chatty world can never
turn watching into spend. It routes and prices like any other kind — nothing about
it is a special case in the wallet or the breaker.

Default chain: `["local", "cloud"]` — cheap-first, with a reliable fallback so a
confirm still resolves when the local tier is down. Re-route it like any kind (e.g.
pin it to a small dedicated provider) by adding `"metatron_watch": [...]` under
`routes`.

**Upgrading an existing world**: a v2 `llm.json` written before this kind existed
has no route for it. Rather than failing the boot, the missing route is
backfilled from the default chain above with one boot log line naming the
backfill (`llm: route for call kind "metatron_watch" missing — backfilled from
defaults …`); add the route explicitly to silence the warning. This backfill is
scoped to kinds introduced after the v2 format shipped — a route missing for any
other kind is still a boot error, and an unknown route *key* (a typo) still fails
the boot exactly as before. Legacy (v1, two-tier) configs need no attention at
all: they resolve entirely through the same default table and pick the new kind
up for free.

### The `embedding` kind (spec 042)

Every episodic memory a villager forms can gain a recorded meaning vector, and
each villager a rolling "situation" vector — the inputs to relevance-aware memory
recall. The vectors are produced by a dedicated `embedding` kind: a local
embedding model served over the OpenAI-compatible `/embeddings` path (Ollama
serves this for embedding models).

```json
  "providers": {
    "embedder": { "transport": "openai_compat", "endpoint": "http://localhost:11434/v1",
                  "model": "all-minilm:latest" }
  },
  "routes": {
    "embedding": ["embedder"]
  }
```

Setup: `ollama pull all-minilm` (the 384-dim reference pin), add the provider +
route above, restart the daemon — the boot line reads
`daemon: embedder on (all-minilm:latest via provider "embedder")`.

Prefer the **fully tagged id** (`all-minilm:latest`, exactly as `ollama list` prints
it) over the bare alias: Ollama resolves the alias fine for the calls themselves,
but the provider-health preflight compares ids against the server's model listing
and a bare `all-minilm` raises a spurious `model-missing` warning at boot —
live-found during the spec-042 walkthrough. A successful embed call clears it
(TASK-102), so the warning is a transient boot-time blip, not a permanent one; the
tagged id avoids it entirely.

The kind is deliberately unusual in three ways:

- **Absence is the off switch.** No `embedding` route means the subsystem is OFF:
  one boot info line (`daemon: embedding off …`), a vectorless world, and **no
  backfill** — unlike `metatron_watch`, a missing route is never spliced in,
  because absence is the feature toggle (the "no llm.json → reflex-only" posture).
- **It never chats.** Embedding calls bypass the chat machinery (queues, breaker,
  budget arithmetic — embedding providers are zero-priced local by doctrine) and
  only the chain **head** serves: model identity travels with every recorded
  vector, so a chain fallback would silently mix incomparable vector spaces.
- **The `anthropic` transport is refused at boot.** The Messages API serves no
  embeddings endpoint; routing `embedding` to an anthropic provider is a config
  error naming the offender.

The provider's `model` string is the identity recorded on every vector: **treat it
as pinned per world lineage**. Changing it later is safe but neutral — vectors
from different models are never compared (old memories score neutral until the
world accretes new-model vectors); nothing is re-embedded.

**Warm-pin (Ollama)**: at driver start and hourly, the daemon pins the embedding
model resident via the Ollama-native `/api/embed` `{"keep_alive": -1}` call, so
per-memory embeds never pay a cold model load. Best-effort and Ollama-specific: a
non-Ollama `openai_compat` server 404s the pin, which is logged once at boot and
ignored — only cold-load latency depends on it, never correctness. The pin
targets the embedding model only; chat-model eviction behavior is untouched.

**Failure posture**: if the endpoint is down or the model missing, memories keep
landing — vectorless forever (no retro-fill) — and the operator hears about it
once per failure episode: a daemon-log WARNING plus a durable
`daemon.llm_warning` event. The tick path never blocks on embedding.

#### `memory_relevance` — the selection mode flag (world.json, not llm.json)

Recall behavior is gated by a three-state flag in **world.json** (deliberately not
llm.json, so deleting llm.json can never silently change memory-selection
semantics):

| value | behavior |
|-------|----------|
| absent / `""` | today's salience+recency selection — the default |
| `"shadow"` | prompts still get the legacy window **bit-identically**; every selection also computes the relevance-augmented ranking and records the divergence (`cog.memory_divergence`) |
| `"on"` | prompts consume the relevance-augmented window (query = the villager's recorded situation vector); divergence keeps recording |

An unknown value refuses to boot (a typo must never silently run as off). The
flag is read at daemon start — flipping it is an offline world.json edit.

**Shadow→on gate procedure (FR-007)**: run the world in `"shadow"` for at least
one full game day, then summarize the recorded evidence:

```sh
promptworld divergence <world>          # per-agent/per-day table
promptworld divergence <world> --json   # machine-readable
```

The summary reports mean overlap@K, the promoted-memory share (selections where
relevance surfaced a memory the legacy ranking excluded), mean rank displacement,
and vectorless counts. The flip to `"on"` is an **operator decision recorded on
the board task** — cite the numbers and the threshold call there, then edit
world.json and restart. Both outcomes (flip or no-flip) are recorded; nothing
auto-flips.

## Money: one wallet, per-provider attribution

`monthly_budget_usd` remains a single global ceiling checked before any priced call.
Every call is priced by its serving provider's rates, and the month's spend is
attributed per provider (persisted; survives restarts; per-provider rows sum to the
global total). Months from before the upgrade show their total as `(unattributed)`.

## Kind-scoped knobs stay top-level

These describe the *thought class*, not the provider, and apply unchanged whichever
chain candidate serves:

- `loop_max_rounds` — tool-loop round cap (default 8, max 16).
- `max_tokens` (spec 025) — `{"planner": n, "metatron_turn": n, "consolidation": n}`;
  absent fields keep the built-in defaults (512/1024/1024), bound 4096, warn-and-clamp.

## Sharing an endpoint: `endpoint_capacity` (opt-in)

If two providers — or two *worlds* — drive the same endpoint (e.g. one Ollama serving
both gemma and cogito, or a proving world plus your own world), declare the endpoint's
true concurrent capacity on each provider that uses it. Participating daemons then
coordinate through advisory file leases (`~/.promptworld/endpoint-leases/`, flock-based,
crash-reclaimable — a killed daemon's slots free automatically), so combined load can
no longer push calls past the 2-minute worker cap and trip breakers on both sides
(the TASK-24 failure mode). A world waiting on a saturated endpoint reports
`contended` in status instead of striking its breaker. Leave the field off and
behavior is exactly as before — the mechanism only binds worlds that opt in.

## Observability

- `promptworld status <world>` (and the TUI's LLM pane) shows a per-provider table:
  name, model, endpoint, up/down, queue depth, inflight/slots, contended, spend share.
- `promptworld llm <world> <kind> "..."` (the one-shot proof path) prints the serving
  provider and any skipped candidates with reasons.
- `promptworld calibrate <world>` iterates the declared providers (each sample pinned
  to its provider) and writes one profile entry per provider name; `--provider <name>`
  narrows to one. `--tier local|cloud` still works as a deprecated alias
  (local→zero-priced, cloud→priced).

## Behavioral edges worth knowing

- A **zero-priced cloud router** is no longer budget-refused when the ceiling is hit
  (refusal keys on pricing, not tier identity). Priced providers behave as before.
- Routing chatty kinds to a small model raises empty-utterance risk; the conversation
  tolerance machinery (TASK-42: one bad utterance + one bad outcome absorbed per
  scene) is the shipped companion. Retune chains from telemetry — quality floors are
  never auto-degraded at runtime.
- Deleting `llm.json` still disables the orchestrator entirely; the world runs with
  reflex-only minds.
