# Quickstart — Validating Tile Registry + New Terrain Tiles (spec 068)

End-to-end validation guide. Contracts: [contracts/tile-registry.md](contracts/tile-registry.md);
shapes: [data-model.md](data-model.md).

## Prerequisites

- Go toolchain per `go.mod`; repo root; a 256-color terminal for the visual checks.

## 1. Full test suite (C1–C13 are all test-enforced)

```sh
go test ./...
```

Expected: green, including the new tests — tui byte-identity pin + token-class sweep +
registered-tile round-trip; worldmap legacy-Hash pin + gen=2 presence/determinism; world
migrate terrain-preservation; sim naming sweep.

## 2. Legacy worlds are untouched (US2-AS2 / SC-006)

```sh
# any pre-feature world save:
go run ./cmd/promptworld migrate <world-dir>     # v4 → v5, terrain_gen stays absent
go run ./cmd/promptworld <daemon/run subcommand> <world-dir>
```

Expected: migrate succeeds; the world opens; its map is visually identical (no marsh `░`,
no sand `▒` anywhere). The migrate test pins this via `Map.Hash()` equality.

## 3. New worlds carry the new vocabulary (US2-AS1/AS3)

```sh
go run ./cmd/promptworld new <fresh-dir>
# world.json now has "format_version": 5, "terrain_gen": 2
```

Attach the TUI and verify visually:

- marsh `░` clusters on wet ground near water; sand `▒` rings shorelines;
- the map legend line carries `░marsh` and `▒sand` tokens;
- the `?` overlay glyph walkthrough has a row for each with a plain-language meaning;
- at night both dim (never disappear);
- villagers walk across both; neither hosts builds.

## 4. Old-software refusal (C10 — one-time manual check)

Run any pre-feature build against the fresh v5 world:

```sh
git stash && git checkout <pre-feature-main> && go run ./cmd/promptworld <run> <fresh-dir>
```

Expected: refusal at Open — `world format_version 5 unsupported (this build supports 4)…`
— never a loaded world with different terrain. (The automated stand-in is the Open
version-mismatch test; this manual check is optional confirmation.)

## 5. Byte-identity spot check (C6)

The pinned-fixture test is authoritative, but for eyes-on confidence: render the same
pre-feature world before and after the change (same terminal size) and diff the map
region — zero differences expected.
