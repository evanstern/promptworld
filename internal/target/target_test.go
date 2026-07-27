package target

import (
	"errors"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestParseForms covers every data-model.md §1 example row plus the grammar's
// whitespace and normalization rules (T004): the reserved-prefix rule is total
// and deterministic, bare names stay bare names, and each structured form
// parses to its normalized Address.
func TestParseForms(t *testing.T) {
	cases := []struct {
		in   string
		want Address
	}{
		// §1 example rows (the accepting ones).
		{"Rega", Address{Form: FormName, Class: ClassVillager, Name: "Rega"}},
		{"villager:Rega", Address{Form: FormName, Class: ClassVillager, Name: "Rega"}},
		{"villager@5,5", Address{Form: FormPoint, Class: ClassVillager, X: 5, Y: 5}},
		{"structure@12,7", Address{Form: FormPoint, Class: ClassStructure, X: 12, Y: 7}},
		{"pile@3,4", Address{Form: FormPoint, Class: ClassPile, X: 3, Y: 4}},
		{"terrain@9,2", Address{Form: FormPoint, Class: ClassTerrain, X: 9, Y: 2}},
		// Rect: corners unordered in the source, normalized min/max (US3 AC1).
		{"structure@3,9..1,5", Address{Form: FormRect, Class: ClassStructure, X: 1, Y: 5, X2: 3, Y2: 9}},
		// Line: endpoint order preserved (US3 AC2).
		{"structure@1,5->1,9", Address{Form: FormLine, Class: ClassStructure, X: 1, Y: 5, X2: 1, Y2: 9}},
		{"structure@1,9->1,5", Address{Form: FormLine, Class: ClassStructure, X: 1, Y: 9, X2: 1, Y2: 5}},
		// No reserved prefix → bare villager name, whatever the word.
		{"boulder@3,4", Address{Form: FormName, Class: ClassVillager, Name: "boulder@3,4"}},
		{"structures@3,4", Address{Form: FormName, Class: ClassVillager, Name: "structures@3,4"}},
		// Whitespace: the whole string is trimmed; spaces allowed after ',' and
		// around '..'/'->' — nowhere else.
		{"  Rega  ", Address{Form: FormName, Class: ClassVillager, Name: "Rega"}},
		{" structure@12, 7 ", Address{Form: FormPoint, Class: ClassStructure, X: 12, Y: 7}},
		{"structure@1,5 .. 3,9", Address{Form: FormRect, Class: ClassStructure, X: 1, Y: 5, X2: 3, Y2: 9}},
		{"structure@1,5 -> 1,9", Address{Form: FormLine, Class: ClassStructure, X: 1, Y: 5, X2: 1, Y2: 9}},
		{"villager: Rega", Address{Form: FormName, Class: ClassVillager, Name: "Rega"}},
		// Zero-area rect and single-point line are valid.
		{"structure@2,2..2,2", Address{Form: FormRect, Class: ClassStructure, X: 2, Y: 2, X2: 2, Y2: 2}},
		{"structure@2,2->2,2", Address{Form: FormLine, Class: ClassStructure, X: 2, Y: 2, X2: 2, Y2: 2}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

// TestParseErrors pins the parse-time taxonomy (syntax and form): a string
// carrying the reserved prefix MUST parse structured — a malformed remainder
// is ErrSyntax, never a bare-name fallback — and a diagonal line is ErrForm.
func TestParseErrors(t *testing.T) {
	cases := []struct {
		in   string
		kind ErrKind
		msg  string // substring the message must carry
	}{
		{"structure@", ErrSyntax, "not a valid address"},                      // §1 example row
		{"structure@1,5->2,9", ErrForm, "diagonal"},                           // §1 example row (v1 axis-aligned)
		{"structure@12", ErrSyntax, "not a valid address"},                    // no comma
		{"structure@a,b", ErrSyntax, "non-negative integer"},                  // non-integer
		{"structure@-1,5", ErrSyntax, "non-negative integer"},                 // negative
		{"structure@1.5,2", ErrSyntax, "non-negative integer"},                // float
		{"structure@1,2junk", ErrSyntax, "non-negative integer"},              // trailing junk
		{"structure@ 12,7", ErrSyntax, "not a valid address"},                 // space after '@'
		{"structure@12 ,7", ErrSyntax, "non-negative integer"},                // space before ','
		{"structure@1,2..3,4->5,6", ErrSyntax, "mixes"},                       // both separators
		{"structure@1,2..3,4..5,6", ErrSyntax, "more than one"},               // repeated separator
		{"villager:", ErrSyntax, "needs a name"},                              // empty typed name
		{"villager:   ", ErrSyntax, "needs a name"},                           // blank typed name
		{"structure:hut", ErrSyntax, "villager:<name>"},                       // typed-name is villager-only
		{"pile:3,4", ErrSyntax, "villager:<name>"},                            // ditto
		{"Structure@1,2", ErrSyntax, ""},                                      // class tokens are exact lowercase → bare name… see below
		{"", ErrSyntax, "empty target"},                                       // empty string
		{"terrain@1,", ErrSyntax, "missing coordinate"},                       // missing Y
		{"structure@99999999999999999999,1", ErrSyntax, "value out of range"}, // overflow
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			if tc.in == "Structure@1,2" {
				// Class tokens are exact and lowercase (spec edge cases):
				// "Structure@1,2" carries NO reserved prefix, so it is a bare
				// villager name, not an error.
				got, err := Parse(tc.in)
				if err != nil || got.Form != FormName || got.Name != "Structure@1,2" {
					t.Fatalf("Parse(%q) = %+v, %v; want bare name", tc.in, got, err)
				}
				return
			}
			_, err := Parse(tc.in)
			var te *Error
			if !errors.As(err, &te) {
				t.Fatalf("Parse(%q) err = %v, want *target.Error", tc.in, err)
			}
			if te.Kind != tc.kind {
				t.Errorf("Parse(%q) kind = %q, want %q", tc.in, te.Kind, tc.kind)
			}
			if !strings.Contains(te.Msg, tc.msg) {
				t.Errorf("Parse(%q) msg = %q, want substring %q", tc.in, te.Msg, tc.msg)
			}
		})
	}
}

// TestClassErrorKind pins the ErrClass arm's message shape. It is unreachable
// through Parse today by construction (reservedSplit and classFor read the
// same table — data-model.md §5 notes the class kind is "only reachable if the
// reserved-prefix set and class set ever diverge"), so it is pinned white-box:
// if the sets ever diverge, this is the named, distinct failure they get.
func TestClassErrorKind(t *testing.T) {
	_, err := classFor("boulders")
	var te *Error
	if !errors.As(err, &te) || te.Kind != ErrClass {
		t.Fatalf("classFor err = %v, want ErrClass", err)
	}
	if !strings.Contains(te.Msg, `unknown class "boulders"`) {
		t.Errorf("msg = %q, want unknown-class shape", te.Msg)
	}
}

// TestTiles pins the deterministic enumeration order (data-model.md §2; US3
// AC1/AC2): rect row-major (y ascending, then x ascending), line in endpoint
// order stepping ±1 along the varying axis, point single-tile.
func TestTiles(t *testing.T) {
	cases := []struct {
		in   string
		want []Tile
	}{
		{"pile@3,4", []Tile{{3, 4}}},
		// Rect normalized (1,5)-(3,9) would be 15 tiles; use a small one and
		// spell the row-major order out exactly.
		{"structure@2,1..0,0", []Tile{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {1, 1}, {2, 1}}},
		{"structure@2,2..2,2", []Tile{{2, 2}}},
		// Lines: endpoint order, both directions, both axes.
		{"structure@1,5->1,9", []Tile{{1, 5}, {1, 6}, {1, 7}, {1, 8}, {1, 9}}},
		{"structure@1,9->1,5", []Tile{{1, 9}, {1, 8}, {1, 7}, {1, 6}, {1, 5}}},
		{"structure@4,2->2,2", []Tile{{4, 2}, {3, 2}, {2, 2}}},
		{"structure@2,2->2,2", []Tile{{2, 2}}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			a, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			got := a.Tiles()
			if len(got) != len(tc.want) {
				t.Fatalf("Tiles() = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("Tiles() = %v, want %v", got, tc.want)
				}
			}
		})
	}

	// A name address has no tiles.
	a, err := Parse("Rega")
	if err != nil {
		t.Fatal(err)
	}
	if a.Tiles() != nil {
		t.Errorf("name form Tiles() = %v, want nil", a.Tiles())
	}
}

// TestLeafImports is SC-004's structural guarantee: the package imports the
// standard library ONLY, so internal/tool (TASK-157's designation consumer —
// a leaf that may import neither sim nor bundle) can import it without a cycle
// or a new dependency edge. Stdlib import paths never contain a dot in their
// first element; any dotted path (module or vendored) fails here.
func TestLeafImports(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, imp := range f.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			first := path
			if i := strings.IndexByte(path, '/'); i >= 0 {
				first = path[:i]
			}
			if strings.Contains(first, ".") {
				t.Errorf("%s imports %q — internal/target must be stdlib-only (the TASK-157 leaf-safety seam)", name, path)
			}
		}
	}
}

// TestParseLocusForms covers the bare-locus entry point (spec 084 FR-003,
// research R4): the same grammar, normalization, and enumeration as the
// class-prefixed forms, with an empty designation-neutral Class. Each case
// mirrors a Parse row above so the two entries cannot drift.
func TestParseLocusForms(t *testing.T) {
	cases := []struct {
		in   string
		want Address
	}{
		{"4,5", Address{Form: FormPoint, X: 4, Y: 5}},
		{"0,0", Address{Form: FormPoint, X: 0, Y: 0}},
		// Rect: corners unordered in the source, normalized min/max.
		{"1,1..8,8", Address{Form: FormRect, X: 1, Y: 1, X2: 8, Y2: 8}},
		{"3,9..1,5", Address{Form: FormRect, X: 1, Y: 5, X2: 3, Y2: 9}},
		// Line: endpoint (author) order preserved.
		{"2,2->2,9", Address{Form: FormLine, X: 2, Y: 2, X2: 2, Y2: 9}},
		{"2,9->2,2", Address{Form: FormLine, X: 2, Y: 9, X2: 2, Y2: 2}},
		// Whitespace: whole string trimmed; spaces after ',' and around
		// '..'/'->' — the parseLocus rules verbatim.
		{"  4,5  ", Address{Form: FormPoint, X: 4, Y: 5}},
		{"1,1 .. 8,8", Address{Form: FormRect, X: 1, Y: 1, X2: 8, Y2: 8}},
		{"2,2 -> 2,9", Address{Form: FormLine, X: 2, Y: 2, X2: 2, Y2: 9}},
		{"4, 5", Address{Form: FormPoint, X: 4, Y: 5}},
		// Zero-area rect and single-point line are valid.
		{"2,2..2,2", Address{Form: FormRect, X: 2, Y: 2, X2: 2, Y2: 2}},
		{"2,2->2,2", Address{Form: FormLine, X: 2, Y: 2, X2: 2, Y2: 2}},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseLocus(tc.in)
			if err != nil {
				t.Fatalf("ParseLocus(%q): %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("ParseLocus(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

// TestParseLocusErrors pins the bare-locus taxonomy: malformed loci are
// ErrSyntax; a diagonal line is ErrForm — the same kinds parseLocus mints for
// the class-prefixed consumers (one taxonomy, one home).
func TestParseLocusErrors(t *testing.T) {
	cases := []struct {
		in   string
		kind ErrKind
	}{
		{"", ErrSyntax},
		{"   ", ErrSyntax},
		{"4", ErrSyntax},
		{"4,", ErrSyntax},
		{"a,b", ErrSyntax},
		{"-1,5", ErrSyntax},
		{"4.5,5", ErrSyntax},
		{"1,1..2,2..3,3", ErrSyntax},
		{"1,1..2,2->3,3", ErrSyntax},
		{"1,1->3,3", ErrForm}, // diagonal — axis-aligned only in v1
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			_, err := ParseLocus(tc.in)
			var te *Error
			if !errors.As(err, &te) {
				t.Fatalf("ParseLocus(%q) err = %v, want *Error", tc.in, err)
			}
			if te.Kind != tc.kind {
				t.Errorf("ParseLocus(%q) kind = %q, want %q", tc.in, te.Kind, tc.kind)
			}
		})
	}
}

// TestParseLocusTilesMatchParse pins the one-parser law's enumeration half:
// a bare locus and its class-prefixed twin enumerate IDENTICAL tiles — the
// bundle compiler and the designation tools can never disagree on a locus.
func TestParseLocusTilesMatchParse(t *testing.T) {
	pairs := [][2]string{
		{"4,5", "structure@4,5"},
		{"1,1..3,2", "structure@1,1..3,2"},
		{"2,9->2,2", "structure@2,9->2,2"},
		{"5,1->9,1", "structure@5,1->9,1"},
	}
	for _, p := range pairs {
		bare, err := ParseLocus(p[0])
		if err != nil {
			t.Fatalf("ParseLocus(%q): %v", p[0], err)
		}
		classed, err := Parse(p[1])
		if err != nil {
			t.Fatalf("Parse(%q): %v", p[1], err)
		}
		bt, ct := bare.Tiles(), classed.Tiles()
		if len(bt) != len(ct) {
			t.Fatalf("%q vs %q: %d tiles vs %d", p[0], p[1], len(bt), len(ct))
		}
		for i := range bt {
			if bt[i] != ct[i] {
				t.Errorf("%q tile %d = %v, want %v", p[0], i, bt[i], ct[i])
			}
		}
	}
}
