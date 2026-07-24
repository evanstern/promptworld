package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/evanstern/promptworld/internal/tool"
)

// Bundle discovery and the boot-frozen BundleSet (contracts/boot-validation.md).
// Discovery is deterministic — direct child dirs in ascending bytewise name
// order, dotfiles skipped, unknown files ignored — so the same world dir always
// yields the same BundleSet and the same BootReport ordering (the replay and
// SC-005 stories depend on it). Collision resolution lives here (C1 built-in
// wins, C2 first-loaded wins); the structural ladder (B1–B4, T1–T7) lives in
// validate.go.

// GrantDoc is a bundle's optional capabilities.json — the same shape as the
// world-level manifest (internal/metatron manifestDoc): a persona narrows the
// world grant, never widens it (Phase 6 applies the intersection).
type GrantDoc struct {
	Tools        []string `json:"tools"`
	MiracleKinds []string `json:"miracle_kinds"`
}

// BundleTool is a validated tool: its synthesized tool.Tool (roster surface)
// plus the manifest and parsed effect templates the handler factory (Phase 3)
// needs. Templates is nil for script-mode tools; Script is the compiled,
// compile-once Starlark program (spec 036 US3, T024) and is nil for declarative
// tools. Exactly one of Templates/Script is populated, matching the manifest.
type BundleTool struct {
	Name      string
	Tool      tool.Tool
	Manifest  *Manifest
	Templates []effectTemplate
	Script    *scriptProgram
	Dir       string
}

// Bundle is a validated bundle: its identity, optional SOUL fragment and grant
// narrowing, and its collision-resolved tools in load order.
type Bundle struct {
	Name  string
	Soul  string
	Grant *GrantDoc
	Tools []BundleTool
}

// BootIssue is one entry in the BootReport: a skip or rejection, naming the
// bundle (and tool, when tool-scoped), the offending file (relative to the
// bundles root), the ladder rule, its severity, and a human sentence that names
// the specific problem and offending value (SC-005).
type BootIssue struct {
	Bundle   string
	Tool     string
	File     string
	Rule     string
	Severity string // "error" | "warning"
	Message  string
}

// BundleSet is the boot-frozen aggregate the daemon holds and the metatron turn
// assembly reads. It is immutable after Discover returns.
type BundleSet struct {
	bundles []Bundle
	report  []BootIssue
}

// Discover walks <worldDir>/bundles/, validates every bundle, resolves name
// collisions, and returns the frozen BundleSet. An absent bundles/ dir yields
// an empty set with no error (bundles are additive); only a genuine I/O failure
// reading the root is returned as an error — every validation problem is a
// BootReport entry, never a hard failure (a bad bundle must not brick boot).
func Discover(worldDir string) (*BundleSet, error) {
	root := filepath.Join(worldDir, "bundles")
	names, err := childDirs(root)
	if err != nil {
		return nil, fmt.Errorf("read bundles dir %q: %w", root, err)
	}
	bs := &BundleSet{}
	seen := make(map[string]string) // tool name -> first bundle that loaded it (C2)
	for _, name := range names {
		b, issues := validateBundle(root, filepath.Join(root, name))
		bs.report = append(bs.report, issues...)
		if b == nil {
			continue // bundle rejected (B1–B4); issues already say why
		}
		kept := make([]BundleTool, 0, len(b.Tools))
		for _, bt := range b.Tools {
			rel := relTo(root, filepath.Join(bt.Dir, "tool.json"))
			if _, builtin := tool.Lookup(bt.Name); builtin {
				bs.report = append(bs.report, BootIssue{
					Bundle: b.Name, Tool: bt.Name, File: rel, Rule: "C1", Severity: "warning",
					Message: fmt.Sprintf("tool %q collides with a built-in tool of the same name — the built-in wins, this bundle tool is skipped", bt.Name),
				})
				continue
			}
			if first, ok := seen[bt.Name]; ok {
				bs.report = append(bs.report, BootIssue{
					Bundle: b.Name, Tool: bt.Name, File: rel, Rule: "C2", Severity: "warning",
					Message: fmt.Sprintf("tool %q was already provided by bundle %q — first-loaded wins, this one is skipped", bt.Name, first),
				})
				continue
			}
			seen[bt.Name] = b.Name
			kept = append(kept, bt)
		}
		b.Tools = kept
		bs.bundles = append(bs.bundles, *b)
	}
	return bs, nil
}

// Bundles returns the validated bundles in load order.
func (bs *BundleSet) Bundles() []Bundle { return bs.bundles }

// BootReport returns every skip/rejection in deterministic discovery order.
func (bs *BundleSet) BootReport() []BootIssue { return bs.report }

// Roster returns every synthesized tool.Tool in deterministic order (bundle
// load order × within-bundle tool order) — the surface the metatron turn
// assembly merges alongside the granted built-in roster (Phase 3).
func (bs *BundleSet) Roster() []tool.Tool {
	var out []tool.Tool
	for _, b := range bs.bundles {
		for _, t := range b.Tools {
			out = append(out, t.Tool)
		}
	}
	return out
}

// SoulFragments returns each bundle's SOUL.md content, in load order, skipping
// bundles without one (spec 036 US4, T029). Each fragment is already B2-capped
// (≤4000 chars) at load time; the metatron turn assembly appends these verbatim
// after the charter section of the system prompt.
func (bs *BundleSet) SoulFragments() []string {
	var out []string
	for _, b := range bs.bundles {
		if b.Soul != "" {
			out = append(out, b.Soul)
		}
	}
	return out
}

// childDirs returns the direct child directory names of dir in ascending
// bytewise order, skipping dotfiles and non-dir entries. A missing dir is the
// common, unremarkable case: (nil, nil). Any other read error is returned.
func childDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue // non-dir entries (docs, licenses) ignored
		}
		n := e.Name()
		if strings.HasPrefix(n, ".") {
			continue // dotfile dirs skipped
		}
		names = append(names, n)
	}
	sort.Strings(names) // explicit ascending bytewise — the deterministic order
	return names, nil
}

// relTo renders p relative to root for a BootReport file field, falling back to
// the absolute path if it cannot be made relative.
func relTo(root, p string) string {
	if r, err := filepath.Rel(root, p); err == nil {
		return r
	}
	return p
}
