package worlds

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/world"
)

func TestIsPathArg(t *testing.T) {
	cases := []struct {
		arg  string
		want bool
	}{
		{"aria", false},
		{"harbor-2", false},
		{"./aria", true},
		{"../aria", true},
		{"~/aria", true},
		{"~aria", true},
		{"/abs/path", true},
		{"rel/path", true},
		{".", true},
		{"..", true},
	}
	for _, c := range cases {
		if got := IsPathArg(c.arg); got != c.want {
			t.Errorf("IsPathArg(%q) = %v, want %v", c.arg, got, c.want)
		}
	}
}

func TestResolveHomeOnly(t *testing.T) {
	setHome(t)
	home, err := WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "aria")
	makeWorld(t, dir, "aria")

	got, err := Resolve("aria")
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(dir)
	if got != abs {
		t.Errorf("Resolve = %q, want %q", got, abs)
	}
}

func TestResolveRegistryOnly(t *testing.T) {
	setHome(t)
	dir := t.TempDir()
	makeWorld(t, dir, "harbor")
	if err := Upsert("harbor", dir); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve("harbor")
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(dir)
	if got != abs {
		t.Errorf("Resolve = %q, want %q", got, abs)
	}
}

func TestResolveNotFound(t *testing.T) {
	setHome(t)
	_, err := Resolve("nowhere")
	var nf *ErrNotFound
	if !errors.As(err, &nf) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if nf.Name != "nowhere" {
		t.Errorf("ErrNotFound.Name = %q", nf.Name)
	}
}

func TestResolveAmbiguous(t *testing.T) {
	setHome(t)
	home, err := WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	homeDir := filepath.Join(home, "aria")
	makeWorld(t, homeDir, "aria")

	customDir := t.TempDir()
	makeWorld(t, filepath.Join(customDir, "elsewhere"), "aria")
	if err := Upsert("aria", filepath.Join(customDir, "elsewhere")); err != nil {
		t.Fatal(err)
	}

	_, err = Resolve("aria")
	var amb *ErrAmbiguous
	if !errors.As(err, &amb) {
		t.Fatalf("expected ErrAmbiguous, got %v", err)
	}
	if len(amb.Paths) != 2 {
		t.Errorf("expected 2 candidate paths, got %v", amb.Paths)
	}
}

func TestResolveMissingRegistryEntry(t *testing.T) {
	// T014: a name known only to the registry, whose directory has since
	// vanished, must resolve to a specific ErrMissing — never a raw
	// world.Open error, and never the generic "never heard of it" message.
	setHome(t)
	dir := t.TempDir()
	sub := filepath.Join(dir, "ghost")
	makeWorld(t, sub, "ghost")
	if err := Upsert("ghost", sub); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(sub); err != nil {
		t.Fatal(err)
	}

	_, err := Resolve("ghost")
	var missing *ErrMissing
	if !errors.As(err, &missing) {
		t.Fatalf("expected ErrMissing, got %v (%T)", err, err)
	}
	if missing.Name != "ghost" {
		t.Errorf("ErrMissing.Name = %q, want ghost", missing.Name)
	}
	abs, _ := filepath.Abs(sub)
	if missing.Path != abs {
		t.Errorf("ErrMissing.Path = %q, want %q", missing.Path, abs)
	}
	if !strings.Contains(err.Error(), "ghost") || !strings.Contains(err.Error(), "gone") {
		t.Errorf("ErrMissing message not helpful: %q", err.Error())
	}
}

func TestResolveUnknownNameStaysNotFound(t *testing.T) {
	// A name the registry has never seen at all must still be the generic
	// ErrNotFound, not ErrMissing (that's reserved for names the manager
	// actually recognized once).
	setHome(t)
	_, err := Resolve("never-heard-of-it")
	var missing *ErrMissing
	if errors.As(err, &missing) {
		t.Fatalf("unknown name must not produce ErrMissing, got %v", err)
	}
	var nf *ErrNotFound
	if !errors.As(err, &nf) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// makeUnopenableWorld writes a manifest with an unsupported format_version
// directly (bypassing world.Create/migrate) so world.Open fails on it while
// the manifest file still exists on disk — the "world but unopenable"
// fixture the live bug needs covered (stale binary vs migrated world).
func makeUnopenableWorld(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"stale","seed":1,"format_version":999999,"tick_game_seconds":1}`
	if err := os.WriteFile(filepath.Join(dir, world.ManifestName), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveHomeCandidateUnopenableSurfacesOpenError(t *testing.T) {
	// Live bug: a home-directory world with an unsupported format_version
	// must resolve to the real world.Open error (migrate-hint text) — never
	// silently fall through to ErrNotFound, which is what a stale binary
	// reported against a migrated world.
	setHome(t)
	home, err := WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "myworld-01")
	makeUnopenableWorld(t, dir)

	_, err = Resolve("myworld-01")
	var uo *ErrUnopenable
	if !errors.As(err, &uo) {
		t.Fatalf("expected ErrUnopenable, got %v (%T)", err, err)
	}
	if !strings.Contains(err.Error(), "format_version") || !strings.Contains(err.Error(), "migrate") {
		t.Errorf("ErrUnopenable message lost the world.Open migrate hint: %q", err.Error())
	}
	var nf *ErrNotFound
	if errors.As(err, &nf) {
		t.Fatalf("must not fall through to ErrNotFound: %v", err)
	}
}

func TestResolveRegistryCandidateUnopenableSurfacesOpenError(t *testing.T) {
	// Same live bug, registry-only path: a name known only via the registry
	// whose target directory exists but fails to open must surface the
	// world.Open error, not ErrMissing or ErrNotFound. Written directly via
	// writeRegistry (not Upsert): Upsert's own pruneRegistry would drop an
	// unopenable entry before Resolve ever saw it, which is exactly the
	// registry-side counterpart to the bug this test guards against.
	setHome(t)
	dir := t.TempDir()
	sub := filepath.Join(dir, "harbor")
	makeUnopenableWorld(t, sub)
	regPath, err := RegistryPath()
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(sub)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeRegistry(regPath, &Registry{Worlds: map[string]string{"harbor": abs}}); err != nil {
		t.Fatal(err)
	}

	_, err = Resolve("harbor")
	var uo *ErrUnopenable
	if !errors.As(err, &uo) {
		t.Fatalf("expected ErrUnopenable, got %v (%T)", err, err)
	}
	if !strings.Contains(err.Error(), "format_version") {
		t.Errorf("ErrUnopenable message lost the world.Open error: %q", err.Error())
	}
}

func TestResolveHomeUnopenableTakesPriorityOverHealthyRegistry(t *testing.T) {
	// Home-first (data-model.md "Name resolution"): if the home candidate
	// exists as a world but fails to open, that error must surface even
	// when a different, healthy registry candidate for the same name
	// exists — not silently fall through to the registry candidate.
	setHome(t)
	home, err := WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	homeDir := filepath.Join(home, "aria")
	makeUnopenableWorld(t, homeDir)

	regDir := t.TempDir()
	regWorldDir := filepath.Join(regDir, "elsewhere")
	makeWorld(t, regWorldDir, "aria")
	if err := Upsert("aria", regWorldDir); err != nil {
		t.Fatal(err)
	}

	_, err = Resolve("aria")
	var uo *ErrUnopenable
	if !errors.As(err, &uo) {
		t.Fatalf("expected ErrUnopenable, got %v (%T)", err, err)
	}
	abs, _ := filepath.Abs(homeDir)
	if uo.Path != abs {
		t.Errorf("ErrUnopenable.Path = %q, want home candidate %q", uo.Path, abs)
	}
}

func TestResolveSamePathIsNotAmbiguous(t *testing.T) {
	// A registry entry that happens to point at the same worlds-home dir
	// (e.g. a stale self-entry) must not be flagged ambiguous.
	setHome(t)
	home, err := WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "aria")
	makeWorld(t, dir, "aria")
	// Force a registry entry pointing at the same home dir without going
	// through Upsert's home-pruning (Upsert would prune it away, which is
	// exactly the case Resolve must also tolerate if it ever appears).
	regPath, err := RegistryPath()
	if err != nil {
		t.Fatal(err)
	}
	abs, _ := filepath.Abs(dir)
	if err := writeRegistry(regPath, &Registry{Worlds: map[string]string{"aria": abs}}); err != nil {
		t.Fatal(err)
	}

	got, err := Resolve("aria")
	if err != nil {
		t.Fatalf("expected no error for same-path collision, got %v", err)
	}
	if got != abs {
		t.Errorf("Resolve = %q, want %q", got, abs)
	}
}
