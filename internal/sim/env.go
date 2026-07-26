package sim

// EnvAt (spec 074 FR-007, research R4) exports a pure, read-only per-tile
// environmental derivation for the TUI's look-cursor TILE view: warmth and
// light truths built from the EXACT SAME mechanics decayNeeds (executor.go)
// and the gru's own light-protection check (gruProtected, gru.go) already
// key on — never a duplicated radius, never a new constant, never a reducer
// change. warmAt (terrain.go) and litAt (gru.go) are thinned to wrappers over
// the shared private cores this file's sibling functions add, so the
// exported sample and the mechanics it describes can never disagree
// (SC-006's whole point).

// EnvSample is the derived environmental answer for one tile at one tick.
// Day/night is deliberately NOT part of it — that is world state (State.Night),
// not a per-tile derivation — so a caller combines EnvSample with State.Night
// (and, for the "indoors" note, its own shelter-presence read) to get the
// plain-language levels the TILE view header shows.
type EnvSample struct {
	Warm       bool   // ≡ warmAt(s, x, y, tick)
	WarmSource string // "fire" | "shelter" | "" (when !Warm)
	Lit        bool   // ≡ litAt(s, x, y)
}

// EnvAt derives (x,y)'s environmental sample at tick. Pure and read-only:
// no reducer arm calls this, no persistence, no new tuning constant.
func EnvAt(s *State, x, y int, tick int64) EnvSample {
	warm, warmSource := warmthSource(s, x, y, tick)
	lit, _ := lightSource(s, x, y)
	return EnvSample{Warm: warm, WarmSource: warmSource, Lit: lit}
}
