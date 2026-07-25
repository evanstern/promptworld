package skin

import (
	"encoding/json"
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

// tokenRe matches a skin token path (the design corpus's {{skin.*}} dotted
// convention, contract §2) inside a string literal.
var tokenRe = regexp.MustCompile(`\bskin\.(guardian|stage)\.[a-z0-9_.-]+`)

// TestTokenCompleteness (spec 052 T006, US1 AS-3, SC-002): every skin token
// consumed anywhere in the repo — Go string literals, the in-repo example
// skins, and the doc twin's default table — exists in the compiled default
// table, and every default-table token resolves to a non-empty, non-path
// value. A missing token renders its own path (TestResolveOrder proves the
// visible degradation); THIS test is what fails before that ships.
func TestTokenCompleteness(t *testing.T) {
	table := DefaultTable()
	for token, v := range table {
		if v == "" || v == token {
			t.Errorf("default table token %q resolves to %q — never empty, never the path", token, v)
		}
	}

	root := repoRoot(t)

	// 1) Go source: every token literal in the tree must be a table row.
	// Deliberately excludes nothing (tests included): a test consuming a
	// token that doesn't exist is the same bug shipping.
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == ".worktrees" || name == "node_modules" || name == "backlog" || name == "topics" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return nil // a broken file fails the build, not this sweep
		}
		ast.Inspect(f, func(n ast.Node) bool {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			s, uerr := strconv.Unquote(lit.Value)
			if uerr != nil {
				return true
			}
			for _, tok := range tokenRe.FindAllString(s, -1) {
				if _, ok := table[tok]; !ok && !strings.Contains(s, "nonexistent") && !strings.Contains(s, "nope") {
					t.Errorf("%s: consumed token %q is not in the default table (add it + doc twin + this table in the same commit)",
						relPath(root, path), tok)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// 2) Example skins: every override key must name a real token.
	matches, _ := filepath.Glob(filepath.Join(root, "examples", "skins", "*.json"))
	for _, m := range matches {
		data, rerr := os.ReadFile(m)
		if rerr != nil {
			t.Fatal(rerr)
		}
		var doc struct {
			Strings map[string]string `json:"strings"`
		}
		if jerr := json.Unmarshal(data, &doc); jerr != nil {
			t.Errorf("%s: not valid JSON: %v", relPath(root, m), jerr)
			continue
		}
		for tok := range doc.Strings {
			if _, ok := table[tok]; !ok {
				t.Errorf("%s: strings override names unknown token %q", relPath(root, m), tok)
			}
		}
	}

	// 3) Doc twin (patterns/skin-tokens.md, contract §3's published table):
	// every default-table token appears in the page, and every `skin.*`
	// token the page names exists in the table — the doc and the runtime
	// can never silently drift (FR-002, T007).
	docPath := filepath.Join(root, "docs", "design", "tui", "patterns", "skin-tokens.md")
	doc, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("doc twin missing: %v", err)
	}
	docText := string(doc)
	for token := range table {
		if !strings.Contains(docText, token) && !stageTokenCoveredByRange(docText, token) {
			t.Errorf("default-table token %q is missing from the doc twin %s", token, relPath(root, docPath))
		}
	}
	for _, tok := range tokenRe.FindAllString(docText, -1) {
		if _, ok := table[tok]; !ok && !stageRangeSpelling(tok) {
			t.Errorf("doc twin names token %q that is not in the default table", tok)
		}
	}
}

// stageTokenCoveredByRange accepts the doc twin's compact stage-row spelling
// ("skin.stage.stage-1.name … skin.stage.stage-4.name" or a stage-N form)
// covering all four per-stage tokens without sixteen literal rows.
func stageTokenCoveredByRange(docText, token string) bool {
	if !strings.HasPrefix(token, "skin.stage.") {
		return false
	}
	generic := regexp.MustCompile(`skin\.stage\.stage-[N14]\.(name|line)`)
	suffix := token[strings.LastIndex(token, ".")+1:]
	for _, m := range generic.FindAllString(docText, -1) {
		if strings.HasSuffix(m, "."+suffix) {
			return true
		}
	}
	return false
}

// stageRangeSpelling reports whether a doc-twin token is the generic
// stage-N placeholder form rather than a real token.
func stageRangeSpelling(tok string) bool {
	return strings.Contains(tok, "stage-N") || regexp.MustCompile(`^skin\.stage\.stage-[14]\.(name|line)$`).MatchString(tok)
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
