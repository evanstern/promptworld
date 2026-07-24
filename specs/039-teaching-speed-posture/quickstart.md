# Quickstart Validation: Teaching-World Speed Posture

Prerequisites: repo builds (`go build ./...`), an LLM config (`llm.json`) for the
world, optionally a real local model for the calibrated path.

## 1. Non-teaching worlds unchanged (SC-004)

```sh
go test ./...                                # includes byte-identity/regression tests
promptworld new /tmp/pw-plain --name plain --seed 1
grep teaching /tmp/pw-plain/world.json       # expect: no match (omitempty)
```

Start it with an orchestrator; `status` and `speed` replies must carry no `posture`
field and no posture warning (spec-035 uncalibrated warning still behaves as before).

## 2. Teaching default follows the profile (US1, SC-001, SC-005)

```sh
promptworld new /tmp/pw-teach --name teach --seed 1 --teaching
grep teaching /tmp/pw-teach/world.json       # "teaching": true
# Seed a calibration.json with a known planner s/pt (e.g. 17.0 → posture 16x),
# boot the daemon, and check:
#   stdout: "teaching posture: defaulting speed to 16x ..."
#   status: clock at 16x; "posture": {"rung":"16x","calibrated":true}
# Edit calibration.json to 5.0 s/pt (→ 32x), restart, expect default 32x — no
# world.json change.
```

## 3. Soft-cap override surfaces arithmetic (US2, SC-002)

```sh
promptworld speed teach 32          # above posture 16x
# expect: speed IS 32x afterwards; WARNING block contains
#   "above teaching posture 16x" + planner Route arithmetic verbatim
promptworld speed teach 8           # at/below posture
# expect: no posture warning
```

## 4. Uncalibrated teaching world prompts calibrate (US3, SC-003)

```sh
rm /tmp/pw-teach/calibration.json   # (or use a fresh teaching world)
# boot: teaching-flavored uncalibrated block prompting `promptworld calibrate teach`,
#       posture line marked provisional
# status: "posture": {..., "calibrated": false}; CLI line says provisional
# after `promptworld calibrate teach` + restart: prompt gone, rung recomputed
```

## 5. Toggle + consumer surface (US4)

```sh
promptworld teaching teach          # prints current marker
promptworld teaching plain on       # toggles; next boot applies posture
```

## 6. Doctrine guards (FR-007)

- `promptworld speed teach max` on an LLM world still errors (max-gate untouched).
- Replay check: replay the teaching world's log — byte-identical state (the boot
  default is a recorded `clock.speed_set` event).
- `go vet ./... && go test ./...` green.
