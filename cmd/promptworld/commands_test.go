package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/metatron"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// isolatedHome points PROMPTWORLD_HOME at a fresh temp dir for one test,
// mirroring e2e/manager_e2e_test.go's helper of the same purpose.
func isolatedHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("PROMPTWORLD_HOME", home)
	return home
}

// --- T011: cmdNew forms ---

func TestCmdNewNameFormCreatesUnderWorldsHome(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	home, err := worlds.WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "aria")
	w, err := world.Open(dir)
	if err != nil {
		t.Fatalf("expected a readable world at %s: %v", dir, err)
	}
	if w.Manifest.Name != "aria" {
		t.Errorf("manifest name = %q, want aria", w.Manifest.Name)
	}
}

// TestCmdNewTeachingMarksManifest (spec 039 US1/T006): `new --teaching` stamps
// the teaching marker from birth; without the flag the marker is absent.
func TestCmdNewTeachingMarksManifest(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"class", "--seed", "1", "--teaching"}); err != nil {
		t.Fatal(err)
	}
	home, err := worlds.WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	w, err := world.Open(filepath.Join(home, "class"))
	if err != nil {
		t.Fatalf("Open teaching world: %v", err)
	}
	if !w.Manifest.Teaching {
		t.Errorf("--teaching did not set the manifest marker")
	}

	if err := cmdNew([]string{"plain", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	w, err = world.Open(filepath.Join(home, "plain"))
	if err != nil {
		t.Fatalf("Open plain world: %v", err)
	}
	if w.Manifest.Teaching {
		t.Errorf("world created without --teaching should not be a teaching world")
	}
}

// TestCmdNewPrintsLocalModelPullGuidance (spec 034 US3/T015): `promptworld new`
// prints a closing line naming the fresh-world default's local model and a
// copy-pasteable `ollama pull` command, derived from llm.DefaultConfig() so it
// can never drift from the model WriteDefault actually wrote.
func TestCmdNewPrintsLocalModelPullGuidance(t *testing.T) {
	isolatedHome(t)
	out := captureStdout(t, func() {
		if err := cmdNew([]string{"aria", "--seed", "1"}); err != nil {
			t.Fatal(err)
		}
	})
	model := llm.DefaultConfig().Providers["local"].Model
	wantLine := fmt.Sprintf("local model: %s — pull it first if you haven't: ollama pull %s", model, model)
	if !strings.Contains(out, wantLine) {
		t.Errorf("cmdNew stdout missing pull guidance line:\n got: %q\nwant substring: %q", out, wantLine)
	}
}

func TestCmdNewNameFormRefusesDuplicateUntouched(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	home, err := worlds.WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, "aria")
	before, err := world.Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := cmdNew([]string{"aria", "--seed", "2"}); err == nil {
		t.Fatal("expected the duplicate `new aria` to be refused")
	}

	after, err := world.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if after.Manifest.Seed != before.Manifest.Seed {
		t.Errorf("existing world was touched: seed %d -> %d", before.Manifest.Seed, after.Manifest.Seed)
	}
}

func TestCmdNewNameFormRejectsNameFlag(t *testing.T) {
	isolatedHome(t)
	err := cmdNew([]string{"aria", "--name", "somethingelse", "--seed", "1"})
	if err == nil {
		t.Fatal("expected --name to be rejected in name-form")
	}
}

func TestCmdNewNameFormWithAtCreatesExactPathAndRegisters(t *testing.T) {
	isolatedHome(t)
	target := filepath.Join(t.TempDir(), "exact-spot")
	if err := cmdNew([]string{"custom", "--at", target, "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	w, err := world.Open(target)
	if err != nil {
		t.Fatalf("expected a world at the exact --at path: %v", err)
	}
	if w.Manifest.Name != "custom" {
		t.Errorf("manifest name = %q, want custom", w.Manifest.Name)
	}
	// --at is outside the worlds home, so it must be registry-addressable
	// by name afterward (D1/D6).
	resolved, err := worlds.Resolve("custom")
	if err != nil {
		t.Fatalf("expected `custom` to resolve via the registry: %v", err)
	}
	abs, _ := filepath.Abs(target)
	if resolved != abs {
		t.Errorf("resolved = %q, want %q", resolved, abs)
	}
}

func TestCmdNewPathFormUnchanged(t *testing.T) {
	isolatedHome(t)
	dir := filepath.Join(t.TempDir(), "w")
	if err := cmdNew([]string{dir, "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	w, err := world.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if w.Manifest.Name != "w" {
		t.Errorf("manifest name = %q, want basename \"w\"", w.Manifest.Name)
	}
	// A path-form world must NOT be registered — it isn't addressable by
	// name until a daemon boots for it (T007's job, not `new`'s).
	if _, err := worlds.Resolve("w"); err == nil {
		t.Error("expected a path-form world to not be registry-resolvable before any daemon boot")
	}
}

func TestCmdNewPathFormRejectsAt(t *testing.T) {
	isolatedHome(t)
	dir := filepath.Join(t.TempDir(), "w")
	err := cmdNew([]string{dir, "--at", "/tmp/should-not-be-used", "--seed", "1"})
	if err == nil {
		t.Fatal("expected --at to be rejected for a path-shaped argument")
	}
}

func TestCmdNewPathFormValidatesExplicitName(t *testing.T) {
	isolatedHome(t)
	dir := filepath.Join(t.TempDir(), "w")
	err := cmdNew([]string{dir, "--name", "-badname", "--seed", "1"})
	if err == nil {
		t.Fatal("expected an explicit flag-like --name to be rejected (contracts/cli.md D5)")
	}
}

func TestCmdNewPathFormDefaultBasenameStaysUnvalidated(t *testing.T) {
	// Backward compatibility (FR-012): the auto-derived basename was never
	// validated before this feature, and a dotted directory name (itself
	// only reachable via explicit path syntax, never as a bare name) must
	// keep working exactly as it did.
	isolatedHome(t)
	dir := filepath.Join(t.TempDir(), "my.world")
	if err := cmdNew([]string{dir, "--seed", "1"}); err != nil {
		t.Fatalf("expected the legacy dotted-basename default to still work: %v", err)
	}
}

// --- T012: name-or-path resolution plumbing ---

func TestResolveWorldPassesPathsThroughVerbatim(t *testing.T) {
	isolatedHome(t)
	for _, p := range []string{"./aria", "../aria", "~/aria", "/abs/aria", "rel/aria"} {
		got, err := resolveWorld(p)
		if err != nil {
			t.Fatalf("resolveWorld(%q) unexpected error: %v", p, err)
		}
		if got != p {
			t.Errorf("resolveWorld(%q) = %q, want verbatim passthrough", p, got)
		}
	}
}

func TestResolveWorldResolvesBareNameViaWorldsHome(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	dir, err := resolveWorld("aria")
	if err != nil {
		t.Fatal(err)
	}
	home, err := worlds.WorldsHome()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "aria"); dir != want {
		t.Errorf("resolveWorld(aria) = %q, want %q", dir, want)
	}
}

func TestResolveWorldUnknownNameErrors(t *testing.T) {
	isolatedHome(t)
	_, err := resolveWorld("never-created")
	if err == nil {
		t.Fatal("expected an error for an unresolvable bare name")
	}
	var nf *worlds.ErrNotFound
	if !errors.As(err, &nf) {
		t.Errorf("expected worlds.ErrNotFound, got %v (%T)", err, err)
	}
}

func TestWorldArgResolvesNameAtTheCallSite(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"harbor", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	dir, err := worldArg(fs, []string{"harbor"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.Open(dir); err != nil {
		t.Fatalf("worldArg did not resolve to a real world: %v", err)
	}
}

func TestParseWorldFlagsResolvesNameAtTheCallSite(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"harbor", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "")
	dir, err := parseWorldFlags(fs, []string{"harbor", "--json"})
	if err != nil {
		t.Fatal(err)
	}
	if !*asJSON {
		t.Error("expected --json to still parse alongside a name argument")
	}
	if _, err := world.Open(dir); err != nil {
		t.Fatalf("parseWorldFlags did not resolve to a real world: %v", err)
	}
}

func TestWorldArgPathStillBypassesResolution(t *testing.T) {
	// A path to a directory that doesn't even exist yet must pass through
	// unresolved (today's exact behavior) rather than erroring inside
	// resolution — whatever downstream code (world.Open etc.) reports the
	// error, not worlds.Resolve.
	isolatedHome(t)
	fs := flag.NewFlagSet("t", flag.ContinueOnError)
	dir, err := worldArg(fs, []string{"./does-not-exist-anywhere"})
	if err != nil {
		t.Fatalf("path-shaped args must bypass resolution, got error: %v", err)
	}
	if dir != "./does-not-exist-anywhere" {
		t.Errorf("dir = %q, want verbatim path", dir)
	}
}

func TestCmdStatusAcceptsWorldByName(t *testing.T) {
	// End-to-end at the cmd-function layer (no subprocess): `status` on a
	// stopped, name-created world resolves and reports offline state.
	isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	// cmdStatus prints to stdout; we only care that it resolves and
	// succeeds rather than erroring on "no world directory".
	if err := cmdStatus([]string{"aria"}); err != nil {
		t.Fatalf("cmdStatus by name failed: %v", err)
	}
}

func TestCmdStopIdempotentByName(t *testing.T) {
	isolatedHome(t)
	if err := cmdNew([]string{"aria", "--seed", "1"}); err != nil {
		t.Fatal(err)
	}
	if err := cmdStop([]string{"aria"}); err != nil {
		t.Fatalf("stop on a never-started name-created world must be idempotent, got: %v", err)
	}
}

func TestCmdStopUnknownNameFailsClearly(t *testing.T) {
	isolatedHome(t)
	err := cmdStop([]string{"nowhere"})
	if err == nil {
		t.Fatal("expected an error for an unknown world name")
	}
	var nf *worlds.ErrNotFound
	if !errors.As(err, &nf) {
		t.Errorf("expected worlds.ErrNotFound, got %v (%T)", err, err)
	}
}

// --- T018: formatLLMOneShot (spec 024 US6) ---

// TestFormatLLMOneShotNamesServingProvider: the header names the serving
// provider, not the legacy tier — a head dispatch carries no Skipped line.
func TestFormatLLMOneShotNamesServingProvider(t *testing.T) {
	resp := llm.Response{
		Provider: "cogito", Model: "cogito:3b",
		InputTokens: 12, OutputTokens: 34, CostUSD: 0, Millis: 812,
		Text: "hi there",
	}
	out := formatLLMOneShot(resp)
	if !strings.HasPrefix(out, "[cogito · cogito:3b · 12 in / 34 out tokens · $0.0000 · 812ms]\n") {
		t.Errorf("header = %q", out)
	}
	if strings.Contains(out, "skipped:") {
		t.Errorf("a head dispatch must not print a skipped line: %q", out)
	}
	if !strings.HasSuffix(out, "hi there\n") {
		t.Errorf("body missing: %q", out)
	}
}

// TestFormatLLMOneShotPrintsSkippedReasons: a fallback response's Skipped
// candidates each print with their mechanical reason (contracts/status.md).
func TestFormatLLMOneShotPrintsSkippedReasons(t *testing.T) {
	resp := llm.Response{
		Provider: "gemma", Model: "gemma4:12b-mlx",
		Skipped: []llm.RouteSkip{
			{Provider: "cogito", Reason: llm.SkipCircuitOpen},
			{Provider: "bogus", Reason: llm.SkipQueueFull},
		},
		Text: "hi there",
	}
	out := formatLLMOneShot(resp)
	if !strings.Contains(out, "skipped: cogito (circuit-open), bogus (queue-full)\n") {
		t.Errorf("skipped line missing/wrong: %q", out)
	}
}

// --- T023: orderStatusLine (spec 029 polish, `promptworld metatron` status peek) ---

// TestOrderStatusLineFields: id, fuzzy marker, origin, expiry day, status, and
// condition all appear; a structural (non-fuzzy) order carries no marker.
func TestOrderStatusLineFields(t *testing.T) {
	structural := metatron.OrderStatus{
		ID: "ord-120-1", Condition: "Rowan falls asleep", Origin: "player", ExpiresDay: 6, Status: "active",
	}
	line := orderStatusLine(structural)
	for _, want := range []string{"ord-120-1", "player", "day 6", "active", "Rowan falls asleep"} {
		if !strings.Contains(line, want) {
			t.Errorf("structural order line missing %q: %q", want, line)
		}
	}
	if strings.Contains(line, "fuzzy") {
		t.Errorf("a structural order should carry no fuzzy marker: %q", line)
	}

	fuzzy := metatron.OrderStatus{
		ID: "ord-130-1", Condition: "Rowan seems heartbroken", Origin: "system", Fuzzy: true, ExpiresDay: 7, Status: "active",
	}
	if line := orderStatusLine(fuzzy); !strings.Contains(line, "fuzzy") {
		t.Errorf("fuzzy order line missing fuzzy marker: %q", line)
	}
}

// --- T009/T011: promptworld status WARNING block (spec 034,
// contracts/provider-conditions.md "Human surfaces") ---

// TestRenderStatusHumanWarningBlock is the golden test for cmdStatus's
// online human render: a no-LLM status renders byte-identical to pre-034
// output, and a healthy LLM status renders no WARNING line at all (the
// offline-world fallback path never even reaches this function); one and
// two affected providers each print one WARNING line, in wire order, in
// the exact contract format. Any LLM status additionally renders one
// calibration-state row per provider after the WARNING block — spec 035
// US3/FR-004, a conscious retune of 034's healthy-world byte-identity to
// the wire surface only (spec 035 SC-002).
func TestRenderStatusHumanWarningBlock(t *testing.T) {
	base := ipc.StatusData{
		World:  ipc.WorldStatus{Name: "aria", Seed: 42},
		Clock:  ipc.ClockStatus{Tick: 100, GameTime: "Day 1, 06:00", Speed: "16x", EffectiveRate: 16.0},
		Daemon: ipc.DaemonStatus{Pid: 123, UptimeSeconds: 45, Subscribers: 1},
		Log:    ipc.LogStatus{LastSeq: 7},
	}

	tests := []struct {
		name      string
		llm       *llm.Status
		wantLines []string // exact WARNING lines expected, in order
	}{
		{
			name: "no LLM status",
			llm:  nil,
		},
		{
			name: "healthy providers only",
			llm: &llm.Status{Providers: []llm.ProviderStatus{
				{Name: "local", Model: "cogito:3b", Up: true},
			}},
		},
		{
			name: "one affected provider",
			llm: &llm.Status{Providers: []llm.ProviderStatus{
				{
					Name: "local", Model: "cogito:3b", Endpoint: "http://localhost:11434/v1", Up: false,
					Condition:       "model-missing",
					ConditionDetail: `model "cogito:3b" not served by http://localhost:11434/v1`,
					ConditionRemedy: "ollama pull cogito:3b",
				},
			}},
			wantLines: []string{
				`WARNING llm provider "local": model "cogito:3b" not served by http://localhost:11434/v1 — ollama pull cogito:3b`,
			},
		},
		{
			name: "two affected providers",
			llm: &llm.Status{Providers: []llm.ProviderStatus{
				{
					Name: "local", Model: "cogito:3b", Up: false,
					Condition: "model-missing", ConditionDetail: "model missing detail", ConditionRemedy: "ollama pull cogito:3b",
				},
				{
					Name: "cloud", Model: "claude-opus-4-8", Up: false,
					Condition: "endpoint-unreachable", ConditionDetail: "endpoint unreachable detail", ConditionRemedy: "start the model server",
				},
			}},
			wantLines: []string{
				`WARNING llm provider "local": model missing detail — ollama pull cogito:3b`,
				`WARNING llm provider "cloud": endpoint unreachable detail — start the model server`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := base
			sd.LLM = tt.llm
			got := renderStatusHuman(&sd)

			want := fmt.Sprintf("world %q (seed %d) — daemon running (pid %d, up %ds, %d subscriber(s))\n%s\n",
				sd.World.Name, sd.World.Seed, sd.Daemon.Pid, sd.Daemon.UptimeSeconds, sd.Daemon.Subscribers, clockLine(&sd))
			for _, l := range tt.wantLines {
				want += l + "\n"
			}
			// Calibration-state rows (spec 035 US3): one per provider,
			// after the WARNING block, before the log line.
			if tt.llm != nil {
				for _, p := range tt.llm.Providers {
					want += providerCalibrationLine(p) + "\n"
				}
			}
			want += fmt.Sprintf("log: last seq %d\n", sd.Log.LastSeq)

			if got != want {
				t.Errorf("renderStatusHuman() =\n%q\nwant\n%q", got, want)
			}
			if len(tt.wantLines) == 0 && strings.Contains(got, "WARNING") {
				t.Errorf("healthy/no-LLM world should render no WARNING line: %q", got)
			}
		})
	}
}

// --- spec 037 US3 (T010): CLI horizon section ---

// TestHorizonStatusLinesRenders: one line per watched class — the standing at
// the current effective speed, the calibrate-vs-slow-down remedy split by the
// calibrated flag, and the "skipped N" count. A thinking class with a zero
// count carries no count.
func TestHorizonStatusLinesRenders(t *testing.T) {
	sd := &ipc.StatusData{
		Clock: ipc.ClockStatus{Speed: "32x"},
		Horizon: []ipc.HorizonClass{
			{Class: "planner", Suppressed: true, Calibrated: false, SuppressedCount: 214},
			{Class: "conversation", Suppressed: true, Calibrated: true, SuppressedCount: 3},
			{Class: "meeting", Suppressed: false, Calibrated: true, SuppressedCount: 0},
		},
	}
	lines := horizonStatusLines(sd)
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %v", len(lines), lines)
	}
	if lines[0] != "horizon: planner suppressed at 32x — calibrate or slow down (skipped 214)" {
		t.Errorf("uncalibrated suppressed line = %q", lines[0])
	}
	if lines[1] != "horizon: conversation suppressed at 32x — slow down (skipped 3)" {
		t.Errorf("calibrated suppressed line = %q", lines[1])
	}
	if lines[2] != "horizon: meeting thinking at 32x" {
		t.Errorf("thinking line = %q", lines[2])
	}
}

// TestHorizonStatusLinesAbsent: a world with no horizon (no-LLM) renders no
// horizon section — the CLI output is unchanged from pre-037 (FR-008 / US3
// AC2).
func TestHorizonStatusLinesAbsent(t *testing.T) {
	if lines := horizonStatusLines(&ipc.StatusData{}); lines != nil {
		t.Errorf("no-horizon world should render no horizon lines, got %v", lines)
	}
}

// TestRenderStatusHumanWithHorizon: an LLM world's full render carries the
// horizon section after the calibration rows and before the log line; a
// no-horizon world's render contains no "horizon:" line at all.
func TestRenderStatusHumanWithHorizon(t *testing.T) {
	base := ipc.StatusData{
		World:  ipc.WorldStatus{Name: "aria", Seed: 42},
		Clock:  ipc.ClockStatus{Tick: 100, GameTime: "Day 1, 06:00", Speed: "32x", EffectiveRate: 32.0},
		Daemon: ipc.DaemonStatus{Pid: 123, UptimeSeconds: 45, Subscribers: 1},
		Log:    ipc.LogStatus{LastSeq: 7},
	}
	withHorizon := base
	withHorizon.LLM = &llm.Status{Providers: []llm.ProviderStatus{{Name: "local", Model: "cogito:3b", Up: true}}}
	withHorizon.Horizon = []ipc.HorizonClass{{Class: "planner", Suppressed: true, SuppressedCount: 5, Verdict: "x"}}
	got := renderStatusHuman(&withHorizon)
	if !strings.Contains(got, "horizon: planner suppressed at 32x — calibrate or slow down (skipped 5)") {
		t.Errorf("render missing horizon line: %q", got)
	}
	// The horizon section sits before the log line.
	if strings.Index(got, "horizon:") > strings.Index(got, "log: last seq") {
		t.Errorf("horizon section should precede the log line: %q", got)
	}

	if got := renderStatusHuman(&base); strings.Contains(got, "horizon:") {
		t.Errorf("no-horizon world rendered a horizon line: %q", got)
	}
}

// --- T009/T013: calibration-UX rendering (spec 035) ---

// TestSetSpeedLineWithWarning: a set_speed reply carrying a warning renders
// it on its own line, visually distinct from the clock line.
func TestSetSpeedLineWithWarning(t *testing.T) {
	sd := &ipc.StatusData{Warning: "uncalibrated world at 32x: planner, conversation suppressed at current estimates — run `promptworld calibrate demo`"}
	sd.Clock.Speed = "32x"
	sd.Clock.GameTime = "day 1, 06:00"
	out := setSpeedLine(sd)
	if !strings.Contains(out, "speed 32x") {
		t.Errorf("clock line missing from output: %q", out)
	}
	if !strings.Contains(out, "WARNING: uncalibrated world at 32x") {
		t.Errorf("warning missing/misrendered: %q", out)
	}
}

// TestSetSpeedLineWithoutWarning: no warning field means no WARNING line —
// a calibrated (or no-LLM) world's rendering is unchanged.
func TestSetSpeedLineWithoutWarning(t *testing.T) {
	sd := &ipc.StatusData{}
	sd.Clock.Speed = "4x"
	sd.Clock.GameTime = "day 1, 06:00"
	out := setSpeedLine(sd)
	if strings.Contains(out, "WARNING") {
		t.Errorf("no-warning reply must not render a WARNING line: %q", out)
	}
}

// TestProviderCalibrationLineBootstrap: an empty CalibratedAt renders the
// explicit "uncalibrated (bootstrap)" marker (FR-004).
func TestProviderCalibrationLineBootstrap(t *testing.T) {
	line := providerCalibrationLine(llm.ProviderStatus{Name: "local", Model: "gemma3:12b"})
	if !strings.Contains(line, "local") || !strings.Contains(line, "uncalibrated (bootstrap)") {
		t.Errorf("bootstrap provider line = %q", line)
	}
}

// TestProviderCalibrationLineCalibrated: a present CalibratedAt renders the
// profile's timestamp, not the bootstrap marker.
func TestProviderCalibrationLineCalibrated(t *testing.T) {
	line := providerCalibrationLine(llm.ProviderStatus{Name: "cloud", Model: "claude-opus-4-8", CalibratedAt: "2026-07-20T21:40:00Z"})
	if !strings.Contains(line, "2026-07-20T21:40:00Z") {
		t.Errorf("calibrated provider line missing timestamp: %q", line)
	}
	if strings.Contains(line, "bootstrap") {
		t.Errorf("calibrated provider line must not show the bootstrap marker: %q", line)
	}
}
