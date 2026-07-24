# Quickstart Validation: Paused Authoring Chain-Completion

Prerequisites: repo root, Go 1.26.x. No daemon or live LLM needed — every
scenario below runs on scripted models / pure functions.

## 1. Full gate

```sh
go build ./... && go vet ./... && go test ./...
```

Expected: green, including the untouched pause-doctrine tests
(`TestPauseInFlightThoughtLandsAtFrozenTick`, `TestPauseStartsNoNewThoughts`,
`TestPauseConversationLandsAtFrozenTick`, `TestResumeNoBurst`) and the
byte-identical replay tests in `internal/sim`.

## 2. Feature-focused slices

```sh
# Pure paused routing (US2): allow, drift 0, arithmetic contains "paused"
go test ./internal/cognition/ -run 'Paused' -v

# Wake + bounded round + running-world unchanged (US1/US3)
go test ./internal/mind/ -run 'Pause|Nudge' -v
```

Expected assertions (names indicative; exact names in tasks.md):
- paused + nudge ⇒ exactly one planner call for the target at the frozen tick,
  `trigger_seq` = the nudge event's Seq, outcome at zero staleness
  ([contracts/recorded-events.md](contracts/recorded-events.md) C2/C3)
- second nudge, same pause ⇒ no second call (debounce bound, US1 scenario 2)
- running + nudge ⇒ zero nudge-armed calls (US3 scenario 1)
- paused world set to a suppressing speed (or uncapped) ⇒ verdict allow with
  the C1 arithmetic string; resumed ⇒ today's set-speed strings byte-identical

## 3. Replay determinism (SC-004)

```sh
go test ./internal/sim/ -run 'Replay' -v
go test ./internal/mind/ -run 'Replay' -v
```

Expected: green — paused verdicts and frozen-tick thoughts are event-log
arithmetic; any new scenario test replays a paused nudge session to a
byte-identical state hash (pattern: internal/sim/governor_replay_test.go).

## 4. Manual end-to-end (optional, needs a configured provider)

1. Start a world, attach the TUI, set speed 32x, pause.
2. Edit Metatron's charter; in Metatron chat: "nudge Aldric".
3. Observe: the vision lands, Aldric thinks once at the frozen tick under the
   new charter (decisions view shows the thought; horizon surface shows no
   planner suppression while paused).
4. Nudge Aldric again without resuming — no second thought.
5. Resume — cadence normal, no burst.
