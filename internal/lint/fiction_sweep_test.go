package lint

// The fiction-denylist sweep (spec 052 T008, SC-001/SC-002): no angel-fiction
// vocabulary ships on any user-facing surface of the default experience — and
// no NEW bare fiction literal compiles without failing here. Two sweeps:
//
//  1. Go string literals (production files): every literal in the tree is
//     scanned for the denylist; occurrences are allowed ONLY in the frozen
//     serialized forms (spec 052 ruling 2 / research R4 — event types, wire
//     names, JSON tags, tool ids, paths, correlation prefixes) or in the
//     explicitly allowlisted legacy constants (history, annotated at their
//     definition sites).
//  2. Docs (README + the design corpus prose): the same denylist, with
//     backtick/code contexts allowed (frozen identifiers are quoted as code)
//     and named historical passages allowlisted.
//
// Rendered-output sweeps (composed prompts, TUI views) live beside their
// packages: internal/metatron/fiction_prompt_test.go and
// internal/tui/fiction_render_test.go — unexported surfaces are only
// reachable there.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// denylistRe matches the fiction vocabulary spec 052 SC-001 bans from the
// default experience: Metatron, angel, miracle (as display text), and the
// divine/heaven/scripture register. Vision/omen are default-skin-retained
// folk vocabulary (spec assumption 1) and deliberately absent.
var denylistRe = regexp.MustCompile(`(?i)\b(metatron|angels?|miracles?|divine|heavens?|scriptures?)\b`)

// frozenGoForms are the serialized spellings a "metatron"/"miracle" hit may
// legitimately take inside a Go string literal (ruling 2; annotated at their
// definition sites). Each is checked against the exact matched region plus
// its immediate neighbors in the literal.
func allowedGoOccurrence(lit, word string, start, end int) bool {
	lower := strings.ToLower(word)
	before := byte(0)
	if start > 0 {
		before = lit[start-1]
	}
	after := byte(0)
	if end < len(lit) {
		after = lit[end]
	}
	switch lower {
	case "metatron":
		// Frozen forms: "metatron.*" event types, "metatron_*" wire/JSON/kind
		// names, "*-metatron-*" correlation prefixes, "metatron/..." paths,
		// and the exact bare identifier "metatron" (the grammar family key,
		// the llm.json kind, the cognition class, the on-disk dir name, and
		// the hidden CLI compat alias).
		if word != "metatron" {
			return false // capital-M Metatron is display text, never frozen
		}
		switch {
		case after == '.', after == '_', after == '/':
			return true
		case before == '-' && after == '-':
			return true
		case lit == "metatron":
			return true
		}
		return false
	case "miracle", "miracles":
		// Frozen forms: the work_miracle tool id, the miracle_kinds manifest
		// key, a frozen JSON struct tag, and the exact "miracle" literal (the
		// IPC command name; the CLI compat alias too).
		switch {
		case strings.HasPrefix(lit[start:], "miracle_kinds"):
			return true
		case start >= len("work_") && lit[start-len("work_"):start] == "work_":
			return true
		case start >= len(`json:"`) && lit[start-len(`json:"`):start] == `json:"`:
			return true
		case lit == "miracle":
			return true
		}
		return false
	}
	return false
}

// allowlistedDecls are declaration names whose string values are exempt —
// each is annotated history at its definition site.
var allowlistedDecls = map[string]string{
	"LegacyDefaultCharter": "internal/persona/charter.go — the pre-052 seed kept only for default-charter recognition (SC-003)",
}

func TestFictionDenylistGoSources(t *testing.T) {
	root := repoRoot(t)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".worktrees", "node_modules", "backlog", "topics", "specs", "docs", "research":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return nil // a broken file fails the build, not this sweep
		}
		var owner []string // decl-name stack for the allowlist check
		ast.Inspect(f, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.ValueSpec:
				var names []string
				for _, id := range node.Names {
					names = append(names, id.Name)
				}
				owner = names
			case *ast.BasicLit:
				if node.Kind != token.STRING {
					return true
				}
				s, uerr := strconv.Unquote(node.Value)
				if uerr != nil {
					return true
				}
				for _, name := range owner {
					if _, ok := allowlistedDecls[name]; ok {
						return true
					}
				}
				for _, m := range denylistRe.FindAllStringIndex(s, -1) {
					word := s[m[0]:m[1]]
					if allowedGoOccurrence(s, word, m[0], m[1]) {
						continue
					}
					t.Errorf("%s:%d: fiction literal %q in %q — route it through the skin lookup or freeze-annotate it (spec 052)",
						relPath(root, path), fset.Position(node.Pos()).Line, word, truncate(s, 90))
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// docAllowlisted reports whether a docs-file denylist hit is acceptable: a
// code context (inside backticks or a fenced block — frozen identifiers are
// quoted as code) or a named historical passage.
func docAllowlisted(file, line string, inFence bool) bool {
	if inFence {
		// Fenced blocks are mockups/code: frozen identifiers and recorded
		// examples may appear; new fiction strings there are a review
		// concern, not this sweep's (mockups use {{skin.*}} tokens).
		return true
	}
	// Inline code spans: every hit must sit inside `...`.
	stripped := regexp.MustCompile("`[^`]*`").ReplaceAllString(line, "")
	return !denylistRe.MatchString(stripped)
}

// docHistoryAllowlist names doc files (relative to repo root) whose listed
// sections quote history verbatim — the SC-001 "history files" allowance.
var docHistoryAllowlist = map[string]string{
	"docs/design/tui/INDEX.md": "## History", // TASK-34 decision record, preserved verbatim by spec 047
}

func TestFictionDenylistDocs(t *testing.T) {
	root := repoRoot(t)
	targets := []string{filepath.Join(root, "README.md")}
	filepath.WalkDir(filepath.Join(root, "docs", "design", "tui"), func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".md") {
			targets = append(targets, path)
		}
		return nil
	})
	for _, path := range targets {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rel := relPath(root, path)
		historyMarker := docHistoryAllowlist[rel]
		inFence, inHistory := false, false
		for i, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inFence = !inFence
				continue
			}
			if historyMarker != "" && strings.HasPrefix(line, historyMarker) {
				inHistory = true
			}
			if inHistory {
				continue
			}
			if !denylistRe.MatchString(line) {
				continue
			}
			if docAllowlisted(rel, line, inFence) {
				continue
			}
			t.Errorf("%s:%d: fiction vocabulary in doc prose: %q", rel, i+1, truncate(strings.TrimSpace(line), 110))
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(wd, "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root not found from %s: %v", wd, err)
	}
	return root
}

func relPath(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
