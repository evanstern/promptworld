package sim

// EntityLookup is the single seam every positional entity query routes
// through (spec 099 D1): "is there a pile/chest/structure at (x,y)?" and "walk
// every pile" both go through here instead of a raw scan at the call site. v1
// (linearLookup below) simply wraps the pre-existing O(n) scans — pileAt,
// chestAt, structureAt, and the Piles slice itself — with byte-identical
// semantics, including tie-break ordering (first match in slice order wins,
// same as before). A future grid/spatial index implements this interface and
// swaps in at Lookup()'s single return statement; no call site changes
// (SC-002).
type EntityLookup interface {
	// Pile returns the ground pile at (x,y), or nil when the tile holds none.
	Pile(x, y int) *Pile
	// Chest returns the chest structure at (x,y), or nil when the tile holds
	// none (or holds a non-chest structure).
	Chest(x, y int) *Structure
	// Structure reports whether a structure of the given kind stands at
	// (x,y).
	Structure(kind string, x, y int) bool
	// Piles returns every ground pile in the world, in the state's canonical
	// creation order — the whole-world walk the rot sweep (executor.go) and
	// similar sweeps use instead of ranging s.Piles directly.
	Piles() []Pile
}

// Lookup returns s's entity-lookup accessor. Call sites that need a
// positional pile/chest/structure query go through this rather than calling
// the underlying scan helpers directly (FR-001, FR-002).
func (s *State) Lookup() EntityLookup { return linearLookup{s} }

// linearLookup is the v1 EntityLookup: the existing linear scans, unchanged.
// A value type wrapping *State — cheap to construct at each call site, no
// allocation.
type linearLookup struct{ s *State }

func (l linearLookup) Pile(x, y int) *Pile                  { return l.s.pileAt(x, y) }
func (l linearLookup) Chest(x, y int) *Structure            { return l.s.chestAt(x, y) }
func (l linearLookup) Structure(kind string, x, y int) bool { return l.s.structureAt(kind, x, y) }
func (l linearLookup) Piles() []Pile                        { return l.s.Piles }
