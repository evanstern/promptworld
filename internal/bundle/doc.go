// Package bundle loads pluggable, world-scoped tool bundles
// (spec 036-scriptable-agent-tools): manifest-only or Starlark-scripted tool
// folders dropped into a world's bundles/ directory, synthesized into
// tool.Tool entries the metatron turn assembly merges alongside the
// compile-time roster.
//
// Invariants:
//
//   - Effects, never raw events: a bundle tool (declarative or scripted)
//     expresses its result as an effect batch over an audited vocabulary
//     (move_entity, remove_entity, grant_item, snap_time, narrate). The
//     effect→store.Event compiler in this package is the sole factory that
//     turns an effect into a store.Event; scripts can never construct one
//     directly.
//   - Boot-frozen: bundles are discovered, validated, and compiled into an
//     immutable BundleSet once at daemon boot (FR-008). There is no hot
//     reload — a change to bundles/ takes effect on the next boot.
//   - Whitelist-subset: every event type a bundle declares must be a subset
//     of the sim package's InjectSocial whitelist (injectSocialWhitelist);
//     anything wider fails boot validation rather than landing at runtime.
//   - Pure, sandboxed, step-capped scripts: a scripted tool's apply() runs in
//     a starlark.Thread with no wall clock, no I/O, no ambient randomness,
//     recursion off, and a hard step cap (manifest-configurable, capped at
//     1,000,000) — determinism and replay byte-identity depend on it.
//
// go.starlark.net is pinned here as a direct dependency ahead of the
// executor (script.go, spec 036 Phase 5/T023); the blank import below keeps
// it live in go.mod until that file lands.
package bundle

import (
	_ "go.starlark.net/starlark"
)
