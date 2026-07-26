// Package target implements the target-addressing grammar of spec 082
// (specs/082-target-addressing/data-model.md — the NORMATIVE definition).
// It is the ONE parser both consumers bind to: the bundle effect compiler
// (internal/bundle, this spec) and the future designation tools (TASK-157,
// internal/tool), so the two can never drift on what an address means.
//
// The package is deliberately a LEAF: it imports the standard library only,
// so internal/tool — which by law imports neither sim nor bundle — can adopt
// it without a cycle or a new dependency edge. target_test.go pins that
// import surface (the TASK-157 seam's structural guarantee, SC-004).
//
// Parsing is state-free: same string ⇒ same Address, always. Binding an
// Address to world entities (resolution) is each consumer's job against the
// state it was handed; the error kinds ErrBounds/ErrUnresolved exist here so
// consumers share one taxonomy, but Parse itself only ever emits
// ErrSyntax/ErrClass/ErrForm.
package target

import (
	"fmt"
	"strconv"
	"strings"
)

// Form is the shape of a parsed address.
type Form string

const (
	FormName  Form = "name"  // a villager by name (bare or "villager:<name>")
	FormPoint Form = "point" // <class>@X,Y — one tile
	FormRect  Form = "rect"  // <class>@X1,Y1..X2,Y2 — inclusive rectangle
	FormLine  Form = "line"  // <class>@X1,Y1->X2,Y2 — inclusive axis-aligned line
)

// Class is the addressed entity class. The set mirrors the miracle reducer's
// move/remove vocabulary (sim.EntityMovedPayload/EntityRemovedPayload Class).
type Class string

const (
	ClassVillager  Class = "villager"
	ClassStructure Class = "structure"
	ClassPile      Class = "pile"
	ClassTerrain   Class = "terrain"
)

// ErrKind is the taxonomy class of a target failure (data-model.md §5). Parse
// emits ErrSyntax/ErrClass/ErrForm; ErrBounds and ErrUnresolved are minted by
// consumers during resolution, named here so the taxonomy has one home.
type ErrKind string

const (
	ErrSyntax     ErrKind = "syntax"     // reserved prefix present but the address is malformed
	ErrClass      ErrKind = "class"      // <word>[@:] where word is not a known class
	ErrForm       ErrKind = "form"       // well-formed, but the form is not admitted here
	ErrBounds     ErrKind = "bounds"     // coordinates outside the world map (consumer-side)
	ErrUnresolved ErrKind = "unresolved" // in-bounds, allowed form, nothing binds (consumer-side)
)

// Error is a taxonomy-tagged target failure. Consumers branch on Kind (never
// on message text) to wrap it in their own error surface.
type Error struct {
	Kind ErrKind
	Msg  string
}

func (e *Error) Error() string { return e.Msg }

// errf mints a taxonomy-tagged error.
func errf(kind ErrKind, format string, args ...any) *Error {
	return &Error{Kind: kind, Msg: fmt.Sprintf(format, args...)}
}

// Address is the parsed form of a target string (data-model.md §2).
//
//   - FormName: Class is always ClassVillager; Name holds the trimmed name
//     (case preserved — resolution matches case-insensitively).
//   - FormPoint: (X,Y) is the tile.
//   - FormRect: (X,Y) is the componentwise-min corner and (X2,Y2) the
//     componentwise-max corner (normalized at parse; zero-area is valid).
//   - FormLine: (X,Y)→(X2,Y2) in AUTHOR order (direction is intent, e.g. wall
//     build order); v1 requires the line be axis-aligned.
type Address struct {
	Form   Form
	Class  Class
	Name   string
	X, Y   int
	X2, Y2 int
}

// Tile is one map coordinate, the unit of enumeration.
type Tile struct{ X, Y int }

// classes is the class vocabulary — ALSO the reserved-prefix word set. Kept as
// one table so the two cannot diverge silently; if a future change ever splits
// them (a reserved word without a class), classFor's ErrClass arm is the
// distinct, message-quality failure the taxonomy holds ready for it.
var classes = map[string]Class{
	"villager":  ClassVillager,
	"structure": ClassStructure,
	"pile":      ClassPile,
	"terrain":   ClassTerrain,
}

// reservedSplit reports whether s carries the reserved prefix
// ^(villager|structure|pile|terrain)[@:] and, if so, splits it into the class
// word, the separator byte, and the remainder. Any other string — including an
// unknown word before '@'/':' — is a bare villager name by the grammar.
func reservedSplit(s string) (word string, sep byte, rest string, ok bool) {
	i := strings.IndexAny(s, "@:")
	if i < 0 {
		return "", 0, "", false
	}
	if _, known := classes[s[:i]]; !known {
		return "", 0, "", false
	}
	return s[:i], s[i], s[i+1:], true
}

// classFor resolves a reserved-prefix word to its Class. Unreachable through
// Parse today (reservedSplit checks the same table), kept distinct so a future
// divergence of the prefix set and class set fails with a named class error
// rather than a generic syntax one (data-model.md §5).
func classFor(word string) (Class, error) {
	c, ok := classes[word]
	if !ok {
		return "", errf(ErrClass, "unknown class %q", word)
	}
	return c, nil
}

// syntaxHint is the one-line grammar reminder every syntax error carries.
const syntaxHint = "want class@X,Y, class@X1,Y1..X2,Y2, or class@X1,Y1->X2,Y2"

// Parse parses one target string per the spec-082 grammar (data-model.md §1–2):
// the whole string is trimmed; a string carrying the reserved prefix
// ^(villager|structure|pile|terrain)[@:] MUST parse as a structured address —
// if malformed that is an ErrSyntax, never a fallback to bare-name — and any
// other string is a bare villager name (v1 compat, byte-identical behavior for
// every string the v1 compiler accepted).
func Parse(s string) (Address, error) {
	t := strings.TrimSpace(s)
	word, sep, rest, reserved := reservedSplit(t)
	if !reserved {
		if t == "" {
			return Address{}, errf(ErrSyntax, "empty target")
		}
		return Address{Form: FormName, Class: ClassVillager, Name: t}, nil
	}
	class, err := classFor(word)
	if err != nil {
		return Address{}, err
	}
	if sep == ':' {
		// typed-name — the grammar admits it for villager only.
		if class != ClassVillager {
			return Address{}, errf(ErrSyntax, "%q is not a valid address (only \"villager:<name>\" takes a name; %s)", t, syntaxHint)
		}
		name := strings.TrimSpace(rest)
		if name == "" {
			return Address{}, errf(ErrSyntax, "%q is not a valid address (villager: needs a name)", t)
		}
		return Address{Form: FormName, Class: ClassVillager, Name: name}, nil
	}
	return parseLocus(t, class, rest)
}

// parseLocus parses the tile-address remainder after "<class>@": a point, an
// inclusive rect ("..", corners normalized min/max), or an inclusive
// axis-aligned line ("->", endpoint order preserved). Spaces are permitted
// after ',' and around '..'/'->' — nowhere else.
func parseLocus(whole string, class Class, rest string) (Address, error) {
	ri := strings.Index(rest, "..")
	li := strings.Index(rest, "->")
	switch {
	case ri >= 0 && li >= 0:
		return Address{}, errf(ErrSyntax, "%q is not a valid address (mixes '..' and '->'; %s)", whole, syntaxHint)
	case ri >= 0:
		if strings.Count(rest, "..") > 1 {
			return Address{}, errf(ErrSyntax, "%q is not a valid address (more than one '..'; %s)", whole, syntaxHint)
		}
		x1, y1, err := parsePoint(whole, strings.TrimRight(rest[:ri], " "))
		if err != nil {
			return Address{}, err
		}
		x2, y2, err := parsePoint(whole, strings.TrimLeft(rest[ri+2:], " "))
		if err != nil {
			return Address{}, err
		}
		// Normalize: (X,Y) = componentwise min, (X2,Y2) = componentwise max.
		return Address{Form: FormRect, Class: class,
			X: min(x1, x2), Y: min(y1, y2), X2: max(x1, x2), Y2: max(y1, y2)}, nil
	case li >= 0:
		if strings.Count(rest, "->") > 1 {
			return Address{}, errf(ErrSyntax, "%q is not a valid address (more than one '->'; %s)", whole, syntaxHint)
		}
		x1, y1, err := parsePoint(whole, strings.TrimRight(rest[:li], " "))
		if err != nil {
			return Address{}, err
		}
		x2, y2, err := parsePoint(whole, strings.TrimLeft(rest[li+2:], " "))
		if err != nil {
			return Address{}, err
		}
		if x1 != x2 && y1 != y2 {
			return Address{}, errf(ErrForm, "%q is a diagonal line — lines are axis-aligned (horizontal or vertical) in v1", whole)
		}
		// Endpoint ORDER IS PRESERVED — direction is author intent.
		return Address{Form: FormLine, Class: class, X: x1, Y: y1, X2: x2, Y2: y2}, nil
	default:
		x, y, err := parsePoint(whole, rest)
		if err != nil {
			return Address{}, err
		}
		return Address{Form: FormPoint, Class: class, X: x, Y: y}, nil
	}
}

// parsePoint parses "X,Y" with X/Y non-negative decimal integers. A space is
// permitted after the comma only; whole is the full address, for error text.
func parsePoint(whole, s string) (x, y int, err error) {
	c := strings.IndexByte(s, ',')
	if c < 0 {
		return 0, 0, errf(ErrSyntax, "%q is not a valid address (%s)", whole, syntaxHint)
	}
	x, err = parseCoord(whole, s[:c])
	if err != nil {
		return 0, 0, err
	}
	y, err = parseCoord(whole, strings.TrimLeft(s[c+1:], " "))
	if err != nil {
		return 0, 0, err
	}
	return x, y, nil
}

// parseCoord parses one coordinate: decimal digits only (non-negative — no
// sign, no float, no spaces, no trailing junk).
func parseCoord(whole, s string) (int, error) {
	if s == "" {
		return 0, errf(ErrSyntax, "%q is not a valid address (missing coordinate; %s)", whole, syntaxHint)
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, errf(ErrSyntax, "%q is not a valid address (coordinate %q is not a non-negative integer; %s)", whole, s, syntaxHint)
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, errf(ErrSyntax, "%q is not a valid address (coordinate %q: %v)", whole, s, err)
	}
	return n, nil
}

// Tiles enumerates the address's tiles deterministically (data-model.md §2) —
// a pure function of the Address, no state, no map, so any two consumers
// enumerate identically:
//
//   - point → the single tile;
//   - rect  → row-major: y from Y to Y2 ascending, inner x from X to X2
//     ascending (corners were normalized at parse);
//   - line  → from (X,Y) to (X2,Y2) inclusive, stepping ±1 along the single
//     varying axis, in endpoint (author) order.
//
// FormName has no tiles (resolution is by name); it returns nil.
func (a Address) Tiles() []Tile {
	switch a.Form {
	case FormPoint:
		return []Tile{{a.X, a.Y}}
	case FormRect:
		out := make([]Tile, 0, (a.Y2-a.Y+1)*(a.X2-a.X+1))
		for y := a.Y; y <= a.Y2; y++ {
			for x := a.X; x <= a.X2; x++ {
				out = append(out, Tile{x, y})
			}
		}
		return out
	case FormLine:
		if a.X == a.X2 && a.Y == a.Y2 {
			return []Tile{{a.X, a.Y}}
		}
		if a.X == a.X2 { // vertical
			out := make([]Tile, 0, abs(a.Y2-a.Y)+1)
			for y := a.Y; ; y += sign(a.Y2 - a.Y) {
				out = append(out, Tile{a.X, y})
				if y == a.Y2 {
					break
				}
			}
			return out
		}
		// horizontal (axis alignment was enforced at parse)
		out := make([]Tile, 0, abs(a.X2-a.X)+1)
		for x := a.X; ; x += sign(a.X2 - a.X) {
			out = append(out, Tile{x, a.Y})
			if x == a.X2 {
				break
			}
		}
		return out
	}
	return nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	if n < 0 {
		return -1
	}
	return 1
}
