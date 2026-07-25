# Quickstart: validating the tuning manifest

Prerequisites: a built `promptworld` binary and a scratch world (do not use
world-01 for validation runs).

## 1. Absent file — behavior unchanged (SC-001)

```sh
go test ./...                      # full suite green, unchanged
promptworld new scratch-tune       # no tuning.json written
promptworld daemon scratch-tune    # boots silently w.r.t. tuning
```

Expected: no tuning warnings, no `sim.tuning_applied` in the log
(inspect the store / TUI event stream), behavior identical to pre-048.

## 2. Tune a dial (SC-002)

```sh
cat > <world-dir>/tuning.json <<'EOF'
{ "fire_burn_per_wood": 28800 }
EOF
# restart daemon
```

Expected: boot log reports the applied set; log gains exactly one
`sim.tuning_applied` with the full five-field payload
(`fire_burn_per_wood: 28800`, others default). A built/refueled fire's
`FuelUntil` reflects 28800 per wood. Restarting again with the same file
appends **no** second event.

## 3. Clamp + reject (SC-004)

```sh
echo '{ "fire_burn_per_wood": 999999 }' > <world-dir>/tuning.json   # clamps
echo '{ "fire_burn_per_woods": 1 }'     > <world-dir>/tuning.json   # unknown key → boot fails
echo '{ "fire_burn_per_wood": "hot" }'  > <world-dir>/tuning.json   # wrong type → boot fails
```

Expected: first form boots with a clamp warning naming field/raw/clamped;
the other two refuse to boot with the file path and the specific problem.

## 4. Replay determinism (SC-003)

Covered by tests (the decisive check): a sim test drives a world past
fire/refuel/gru/encounter activity under a tuned set, snapshots the state
hash, then replays the log into a fresh state and asserts hash equality —
with no tuning file in reach of the replay path. The daemon-level
`embed_replay_test.go` pattern is the model. Manual spot-check: stop the
daemon, delete `tuning.json`, restart — recovered state must still carry the
tuned values (they come from the log), and the boot must seed a new event
returning the world to defaults only if the operator intended that
(absent file = defaults = a *different* effective set → one new event; this
is correct and visible in the log).

## 5. Per-dial proof (SC-005)

`go test ./internal/sim -run Tuning` (and the mind cadence/cooldown tests):
one test per dial asserting the tuned value — not the default constant —
drives the behavior (refuel trigger window, fuel deadline, emergence roll,
planner stagger/bucket cadence, encounter gate).

## 6. Docs (FR-008)

`docs/design/control-surface-and-calibration.md` §6 names `tuning.json`, the
five promoted dials, and the event discipline as shipped mechanism.
