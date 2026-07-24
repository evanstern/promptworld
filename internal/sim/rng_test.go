package sim

import "testing"

// TestBundleRand covers the exported seeded accessor the bundle script runtime
// draws world.rand from (spec 036 US3): the value is in [0,1), a pure function of
// (seed, purpose, tick, index) — identical coordinates give identical draws — and
// distinct coordinates decorrelate, so replay reproduces a script's randomness from
// its coordinates alone (the determinism half of FR-011/SC-003).
func TestBundleRand(t *testing.T) {
	const seed = 42

	// Range: every draw is in [0,1).
	for i := 0; i < 1000; i++ {
		v := BundleRand(seed, "bundle:demo:pick", 100, i)
		if v < 0 || v >= 1 {
			t.Fatalf("BundleRand index %d = %v, out of [0,1)", i, v)
		}
	}

	// Determinism: identical coordinates ⇒ identical value.
	a := BundleRand(seed, "bundle:demo:pick", 100, 0)
	b := BundleRand(seed, "bundle:demo:pick", 100, 0)
	if a != b {
		t.Errorf("same coordinates gave different draws: %v vs %v", a, b)
	}

	// Decorrelation: changing any coordinate changes the stream (a purpose, tick, or
	// index collision would let two draws alias — the namespace exists to prevent it).
	if BundleRand(seed, "bundle:demo:pick", 100, 0) == BundleRand(seed, "bundle:demo:other", 100, 0) {
		t.Error("distinct purposes produced the same draw")
	}
	if BundleRand(seed, "bundle:demo:pick", 100, 0) == BundleRand(seed, "bundle:demo:pick", 101, 0) {
		t.Error("distinct ticks produced the same draw")
	}
	if BundleRand(seed, "bundle:demo:pick", 100, 0) == BundleRand(seed, "bundle:demo:pick", 100, 1) {
		t.Error("distinct indices produced the same draw")
	}
	if BundleRand(seed, "bundle:demo:pick", 100, 0) == BundleRand(seed+1, "bundle:demo:pick", 100, 0) {
		t.Error("distinct seeds produced the same draw")
	}
}
