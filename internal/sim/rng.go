package sim

import (
	"hash/fnv"
	"math/rand/v2"
)

// rngAt returns a PCG seeded purely from (world seed, purpose, tick, index).
// There is no long-lived RNG stream: every random decision is a pure function
// of its coordinates, so replay and recovery need no RNG state at all.
func rngAt(seed uint64, purpose string, tick int64, index int) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(purpose))
	sub := h.Sum64()
	return rand.New(rand.NewPCG(seed^sub, uint64(tick)*0x9e3779b97f4a7c15+uint64(index)))
}

// BundleRand is the exported, coordinate-seeded [0,1) draw the bundle script
// runtime (internal/bundle world.rand) exposes to Starlark tools. It is a thin
// wrapper over the SAME rngAt determinism the sim uses everywhere else: a pure
// function of (world seed, purpose, tick, index) with no long-lived stream, so a
// script's randomness replays byte-identically from its coordinates alone. The
// caller (worldview.go) owns the "bundle:<tool>:<purpose>" namespace it passes as
// purpose, keeping this wrapper a generic seeded accessor.
func BundleRand(seed uint64, purpose string, tick int64, index int) float64 {
	return rngAt(seed, purpose, tick, index).Float64()
}
