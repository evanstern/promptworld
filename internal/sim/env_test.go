package sim

// Spec 074 (look-cursor TILE view), FR-007/SC-006: EnvAt must never disagree
// with the mechanics it decomposes — every case below cross-checks EnvAt's
// Warm/Lit fields against warmAt/litAt directly, and the source-attribution
// cases pin the fire-before-shelter scan order warmAt has always used
// (research R4).

import "testing"

func TestEnvAtAgreesWithWarmAtAndLitAt(t *testing.T) {
	m := testMap(1)
	cases := []struct {
		name       string
		structures []Structure
		x, y       int
		tick       int64
	}{
		{"empty tile", nil, 5, 5, 100},
		{"lit fire in radius", []Structure{{Kind: "fire", X: 5, Y: 5, FuelUntil: 1000}}, 6, 5, 100},
		{"lit fire out of warmth radius but in light radius", []Structure{{Kind: "fire", X: 5, Y: 5, FuelUntil: 1000}}, 8, 5, 100},
		{"cold fire (fuel expired)", []Structure{{Kind: "fire", X: 5, Y: 5, FuelUntil: 50}}, 5, 5, 100},
		{"dying fire still lit", []Structure{{Kind: "fire", X: 5, Y: 5, FuelUntil: 150}}, 5, 5, 100},
		{"shelter tile", []Structure{{Kind: "shelter", X: 5, Y: 5}}, 5, 5, 100},
		{"shelter adjacent (no warmth off-tile)", []Structure{{Kind: "shelter", X: 5, Y: 5}}, 6, 5, 100},
		{"outside every radius", []Structure{{Kind: "fire", X: 0, Y: 0, FuelUntil: 1000}}, 20, 20, 100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := NewState(1, m)
			s.Structures = c.structures
			sample := EnvAt(s, c.x, c.y, c.tick)
			if sample.Warm != warmAt(s, c.x, c.y, c.tick) {
				t.Errorf("EnvAt.Warm=%v, warmAt=%v", sample.Warm, warmAt(s, c.x, c.y, c.tick))
			}
			if sample.Lit != litAt(s, c.x, c.y) {
				t.Errorf("EnvAt.Lit=%v, litAt=%v", sample.Lit, litAt(s, c.x, c.y))
			}
			if !sample.Warm && sample.WarmSource != "" {
				t.Errorf("WarmSource=%q with Warm=false, want \"\"", sample.WarmSource)
			}
		})
	}
}

// TestEnvAtWarmSourceAttribution pins the fire-before-shelter scan order
// warmAt has always used (research R4): a tile satisfying both attributes to
// whichever structure appears first in s.Structures.
func TestEnvAtWarmSourceAttribution(t *testing.T) {
	m := testMap(1)

	t.Run("fire only", func(t *testing.T) {
		s := NewState(1, m)
		s.Structures = []Structure{{Kind: "fire", X: 5, Y: 5, FuelUntil: 1000}}
		sample := EnvAt(s, 5, 5, 100)
		if !sample.Warm || sample.WarmSource != "fire" {
			t.Fatalf("got Warm=%v Source=%q, want warm/fire", sample.Warm, sample.WarmSource)
		}
	})

	t.Run("shelter only", func(t *testing.T) {
		s := NewState(1, m)
		s.Structures = []Structure{{Kind: "shelter", X: 5, Y: 5}}
		sample := EnvAt(s, 5, 5, 100)
		if !sample.Warm || sample.WarmSource != "shelter" {
			t.Fatalf("got Warm=%v Source=%q, want warm/shelter", sample.Warm, sample.WarmSource)
		}
	})

	t.Run("fire scanned before shelter on the same tile", func(t *testing.T) {
		s := NewState(1, m)
		s.Structures = []Structure{
			{Kind: "fire", X: 5, Y: 5, FuelUntil: 1000},
			{Kind: "shelter", X: 5, Y: 5},
		}
		sample := EnvAt(s, 5, 5, 100)
		if sample.WarmSource != "fire" {
			t.Fatalf("fire precedes shelter in scan order: got source %q, want fire", sample.WarmSource)
		}
	})

	t.Run("shelter scanned before fire on the same tile", func(t *testing.T) {
		s := NewState(1, m)
		s.Structures = []Structure{
			{Kind: "shelter", X: 5, Y: 5},
			{Kind: "fire", X: 5, Y: 5, FuelUntil: 1000},
		}
		sample := EnvAt(s, 5, 5, 100)
		if sample.WarmSource != "shelter" {
			t.Fatalf("shelter precedes fire in scan order: got source %q, want shelter", sample.WarmSource)
		}
	})
}

// TestEnvAtLitWithoutWarmth: the light radius is strictly wider than the
// warmth radius (gruLightRadius > fireWarmRadius) — a tile can be lit
// without being warm, and EnvAt must show exactly that combination.
func TestEnvAtLitWithoutWarmth(t *testing.T) {
	m := testMap(1)
	s := NewState(1, m)
	s.Structures = []Structure{{Kind: "fire", X: 5, Y: 5, FuelUntil: 1000}}
	// fireWarmRadius=2, gruLightRadius=3 — distance 3 is lit but not warm.
	sample := EnvAt(s, 8, 5, 1000)
	if sample.Warm {
		t.Errorf("distance 3 from a fire should not be warm (fireWarmRadius=2)")
	}
	if !sample.Lit {
		t.Errorf("distance 3 from a fire should be lit (gruLightRadius=3)")
	}
}
