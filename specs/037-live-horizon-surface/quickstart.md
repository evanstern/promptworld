# Quickstart: validating the live cognition-horizon surface

End-to-end validation scenarios for spec 037. Prerequisites: a built
`promptworld` binary and a world with an LLM configured (`llm.json`); a
no-LLM world for the byte-identity check.

## 1. Unit/integration sweep

```sh
go build ./... && go vet ./... && go test ./...
```

Expected: green, including the new `LiveHorizon` table tests
(`internal/cognition`), counter tests (`internal/llm`), status-composition
tests (`internal/ipc`), header/pane render tests (`internal/tui`), and the CLI
status render test.

## 2. Live verdict at speed (US1)

```sh
promptworld start <llm-world>
promptworld status <llm-world>        # baseline at 1x: horizon lines, nothing suppressed
promptworld speed <llm-world> 32x     # or the ui's ] key
promptworld status <llm-world>
```

Expected at 32x (uncalibrated or slow provider): the status output includes a
horizon section naming the suppressed class(es) with the router arithmetic and
a remedy; the `set_speed` reply's spec-035 one-shot warning still appears
independently (unchanged).

In the TUI (`promptworld ui <llm-world>`):
- header gains the `[suppressed: …]` badge within ~1 s of the speed change;
- metatron pane (`3`) shows the per-class block: standing at current speed,
  remedy, `skipped N`;
- dropping back to 1x clears the badge and flips the block's standings within
  one poll.

Governed check (US1/AS3): if the governor sheds (requested 32x → effective
16x), the verdict reflects 16x — compare against the header's governed-speed
suffix.

## 3. Suppression counters (US2)

Run hot at 32x for a minute or two while villager cadences fire, then:

```sh
promptworld status <llm-world>   # counts > 0 and growing between calls
promptworld speed <llm-world> 1x
promptworld status <llm-world>   # counts retained, no longer growing
```

Counts must match the world's own record: compare against suppressed
`cog.outcome` events, e.g.

```sh
grep '"cog.outcome"' <world-dir>/events.log | grep -c '"suppressed"'
```

(count per `"class"` for the per-class comparison). Restarting the daemon
resets counts to zero.

## 4. No-LLM byte identity (US3 / SC-004)

```sh
promptworld start <no-llm-world>
promptworld status <no-llm-world>
```

Expected: output identical to a pre-037 build — no horizon section; the raw
IPC status reply carries no `horizon` key (assert in the ipc test, or
`promptworld status --json` if available).

## 5. Verdict/router agreement (SC-002)

Covered by the `internal/cognition` invariant test: `SuppressedAt` ≡ suppressed
names of `LiveHorizon` across the speed ladder × estimate grid, and
`LiveHorizon`'s `Suppressed` ≡ `!Route(...).Allow` for every watched class.
See [contracts/status-horizon.md](contracts/status-horizon.md) rule 3 and
[data-model.md](data-model.md) invariants.
