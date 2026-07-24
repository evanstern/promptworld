// Package lint holds standing project-wide gates that run as part of the
// normal `go test ./...` invocation, since this project has no CI and no
// Makefile.
package lint

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skipDir reports whether a directory's contents are never part of the
// project's own formatted Go sources: the gitignored worktree pool, hidden
// dirs (VCS metadata, editor/tool state), and conventional Go testdata dirs
// (which may deliberately hold malformed fixtures).
func skipDir(name string) bool {
	if name == "testdata" || name == ".worktrees" {
		return true
	}
	return strings.HasPrefix(name, ".")
}

// moduleRoot walks up from the current package directory to find the
// nearest ancestor containing go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found walking up from %s", dir)
		}
		dir = parent
	}
}

// TestGofmt (TASK-83): standing gate against gofmt drift. This test walks
// every .go file from the module root and compares it against go/format's
// canonical output, failing (and naming every offender) if any file
// differs. It uses the standard library formatter rather than shelling out
// to the gofmt binary so the gate has no external dependency and works the
// same in any environment that can `go test`.
func TestGofmt(t *testing.T) {
	root := moduleRoot(t)

	var unformatted []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && skipDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		formatted, err := format.Source(src)
		if err != nil {
			// A file that fails to parse is a build error elsewhere, not a
			// formatting concern for this gate.
			return nil
		}
		if !bytes.Equal(src, formatted) {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			unformatted = append(unformatted, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}

	if len(unformatted) > 0 {
		t.Errorf("gofmt drift found — run `gofmt -w` on:\n  %s", strings.Join(unformatted, "\n  "))
	}
}
