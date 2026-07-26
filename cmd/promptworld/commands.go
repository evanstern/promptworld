package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/daemon"
	"github.com/evanstern/promptworld/internal/guardian"
	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tui"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

func dirArg(fs *flag.FlagSet, args []string) (string, error) {
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if fs.NArg() < 1 {
		return "", fmt.Errorf("missing world directory argument")
	}
	return fs.Arg(0), nil
}

// parseDirFlags handles both "cmd <dir> --flag" and "cmd --flag <dir>".
func parseDirFlags(fs *flag.FlagSet, args []string) (string, error) {
	var dir string
	var rest []string
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		dir, rest = args[0], args[1:]
	} else {
		rest = args
	}
	if err := fs.Parse(rest); err != nil {
		return "", err
	}
	if dir == "" {
		if fs.NArg() < 1 {
			return "", fmt.Errorf("missing world directory argument")
		}
		dir = fs.Arg(0)
	}
	return dir, nil
}

// resolveWorld turns a per-world command's positional argument — a name or
// a path (FR-006) — into a directory. Path-shaped arguments bypass name
// resolution entirely and are returned verbatim, today's exact behavior
// (FR-012); bare names resolve via worlds.Resolve (FR-007/FR-011). Every
// per-world command except `new` (whose argument creates rather than
// resolves) routes through this.
func resolveWorld(arg string) (string, error) {
	if worlds.IsPathArg(arg) {
		return arg, nil
	}
	return worlds.Resolve(arg)
}

// worldArg is dirArg's name-or-path counterpart (see resolveWorld).
func worldArg(fs *flag.FlagSet, args []string) (string, error) {
	arg, err := dirArg(fs, args)
	if err != nil {
		return "", err
	}
	return resolveWorld(arg)
}

// parseWorldFlags is parseDirFlags's name-or-path counterpart (see
// resolveWorld).
func parseWorldFlags(fs *flag.FlagSet, args []string) (string, error) {
	arg, err := parseDirFlags(fs, args)
	if err != nil {
		return "", err
	}
	return resolveWorld(arg)
}

// cmdNew implements `promptworld new` per contracts/cli.md (research.md D5):
// a bare-word argument is name-form — create <worlds-home>/<name> (or --at
// DIR exactly), manifest name = the argument. A path-shaped argument
// (worlds.IsPathArg) is legacy path-form, byte-compatible with today:
// create at that path, name from --name or the basename (FR-012).
func cmdNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	name := fs.String("name", "", "path-form only: world name (default: directory basename)")
	at := fs.String("at", "", "name-form only: create at this exact path instead of the default worlds home")
	seed := fs.Uint64("seed", 0, "world seed (default: random)")
	teaching := fs.Bool("teaching", false, "mark this a teaching world (spec 039): the daemon defaults its speed to the highest planner-safe rung")
	stageFlag := fs.String("stage", "", "curriculum-ladder stage to create at (spec 046: stage-1..stage-4); default: stage-1 for a new player, else your highest earned stage")
	override := fs.Bool("override", false, "create at an unearned stage anyway (spec 046) — the world's config records the override honestly")
	charterPresetFlag := fs.String("charter-preset", "", "charter preset to seed (default|tutor); stage-1 worlds seed the tutor preset unless this opts out to default")
	scenarioFlag := fs.String("scenario", "", "create a scenario world running this seeded exercise (spec 054): stamps the exercise's stage and seed and arms its incident schedule at boot")
	arg, err := parseDirFlags(fs, args)
	if err != nil {
		return err
	}

	// Scenario resolution (spec 054 US3, R5): a scenario id resolves against
	// the compiled exercise catalog — an unknown id refuses listing it (the
	// stage-gate refusal voice). The scenario IMPLIES its stage and pins its
	// authored seed, so an explicit --stage/--seed may only agree; the
	// earned-stage gate below then applies to the implied stage unchanged
	// (a scenario never bypasses the earn gate).
	var scenarioDef sim.ExerciseDefinition
	if *scenarioFlag != "" {
		def, ok := sim.ExerciseByID(*scenarioFlag)
		if !ok {
			ids := make([]string, 0, len(sim.ScenarioExercises))
			for _, d := range sim.ScenarioExercises {
				ids = append(ids, fmt.Sprintf("%s (%s — %s)", d.ID, d.Stage, d.Concept))
			}
			return fmt.Errorf("new: --scenario %q is not a known exercise — the catalog:\n  %s",
				*scenarioFlag, strings.Join(ids, "\n  "))
		}
		scenarioDef = def
		if *stageFlag != "" && *stageFlag != def.Stage {
			return fmt.Errorf("new: --stage %s conflicts with --scenario %s (the exercise is played at %s; drop --stage)",
				*stageFlag, def.ID, def.Stage)
		}
		*stageFlag = def.Stage
		if *seed != 0 && *seed != def.Seed {
			return fmt.Errorf("new: --seed %d conflicts with --scenario %s (the exercise pins seed %d; drop --seed)",
				*seed, def.ID, def.Seed)
		}
		*seed = def.Seed
	}

	// Curriculum-ladder stage resolution (spec 046 US1, T009, R9): validate
	// an explicit --stage, else default to stage-1 for a new player or the
	// player's highest earned stage otherwise; an unearned stage refuses
	// with an informed message naming the skipped concepts (skin names)
	// unless --override; --charter-preset defaults to tutor at stage-1
	// (opt-out to "default") and to "default" everywhere else.
	switch *stageFlag {
	case "", world.Stage1, world.Stage2, world.Stage3, world.Stage4:
	default:
		return fmt.Errorf("new: --stage %q is not a valid stage (want %s, %s, %s, or %s)",
			*stageFlag, world.Stage1, world.Stage2, world.Stage3, world.Stage4)
	}
	unlocks := worlds.LoadUnlocks()
	stage := *stageFlag
	if stage == "" {
		stage = highestEarnedStage(unlocks)
	}
	if !unlocks.StageEarned(stage) && !*override {
		var skipped []string
		for _, id := range stageOrder {
			if !unlocks.StageEarned(id) {
				skipped = append(skipped, skin.StageName(id))
			}
			if id == stage {
				break
			}
		}
		return fmt.Errorf("new: %s (%s) is not yet earned — creating here would skip %s; pass --override to proceed anyway (the world's config records the override honestly)",
			skin.StageName(stage), stage, strings.Join(skipped, ", "))
	}
	charterPreset := *charterPresetFlag
	if !world.ValidCharterPreset(charterPreset) {
		return fmt.Errorf("new: --charter-preset %q is not valid (want %q, %q, or omitted)",
			charterPreset, world.CharterPresetDefault, world.CharterPresetTutor)
	}
	if charterPreset == "" && stage == world.Stage1 {
		// R6/spec: stage-1 worlds seed the tutor preset by default; an
		// explicit --charter-preset default (opt-out) is left untouched by
		// this branch and stamped verbatim below.
		charterPreset = world.CharterPresetTutor
	}

	nameForm := !worlds.IsPathArg(arg)
	var dir, worldName string
	if nameForm {
		if *name != "" {
			return fmt.Errorf("new %q: --name is not valid with a bare name argument (the argument is already the name)", arg)
		}
		if err := worlds.ValidateName(arg); err != nil {
			return err
		}
		worldName = arg
		if *at != "" {
			dir = *at
		} else {
			home, err := worlds.WorldsHome()
			if err != nil {
				return err
			}
			dir = filepath.Join(home, arg)
		}
	} else {
		if *at != "" {
			return fmt.Errorf("new %q: --at is only valid with a bare name, not a path — the path itself is already the location", arg)
		}
		dir = arg
		worldName = *name
		if worldName == "" {
			// Backward compatible: the auto-derived basename was never
			// validated before this feature and stays that way (FR-012).
			worldName = filepath.Base(filepath.Clean(dir))
		} else if err := worlds.ValidateName(worldName); err != nil {
			// An explicit --name IS validated (contracts/cli.md D5).
			return err
		}
	}

	if *seed == 0 {
		var b [8]byte
		if _, err := rand.Read(b[:]); err != nil {
			return err
		}
		*seed = binary.LittleEndian.Uint64(b[:]) >> 12 // keep it comfortably printable
	}
	w, err := world.Create(dir, worldName, *seed)
	if err != nil {
		return err
	}
	// Teaching marker from birth (spec 039 US1): stamp the manifest so the
	// daemon reads the posture at its first boot. Set-after-create keeps
	// world.Create's signature untouched for its other callers.
	if *teaching {
		if err := world.SetTeaching(dir, true); err != nil {
			return err
		}
		w.Manifest.Teaching = true
	}
	// Curriculum-ladder stage marker from birth (spec 046 FR-002/FR-003,
	// T009): set-after-create, the SetTeaching pattern — write-once, no
	// toggle command. Stage is always stamped (even "" i.e. no --stage was
	// ever passed AND the resolved default happened to be stage-1 for a new
	// player) so `w.Manifest` in memory matches what Open would report.
	if err := world.SetStage(dir, stage, *override, charterPreset); err != nil {
		return err
	}
	w.Manifest.Stage, w.Manifest.StageOverridden, w.Manifest.CharterPreset = stage, *override, charterPreset
	// Scenario block from birth (spec 054 US3): set-after-create like the
	// stage above — write-once; the daemon arms the machinery from it at
	// every boot.
	if scenarioDef.ID != "" {
		if err := world.SetScenario(dir, scenarioDef.ID); err != nil {
			return err
		}
		w.Manifest.Scenario = &world.ScenarioConfig{Exercise: scenarioDef.ID}
	}
	st, err := store.Open(w.DBPath())
	if err != nil {
		return err
	}
	defer st.Close()
	payload, err := json.Marshal(sim.WorldCreatedPayload{Name: worldName, Seed: *seed})
	if err != nil {
		return err
	}
	genesis := []store.Event{{Tick: 0, Type: "world.created", Payload: payload}}
	// Genesis tuning pin (spec 057 / TASK-108 US2): one sim.tuning_applied event
	// carrying the full current default dial set, so this world's effective
	// doctrine is fixed in its own log at birth. Future changes to any default*
	// constant then reach only pre-057/un-pinned worlds, never rewriting a
	// pinned world's replay. Ordered right after world.created (both tick 0; the
	// reducer arms are order-independent — world.created is a no-op on state and
	// the tuning arm is a pure set of s.Tuning). migrate does NOT back-fill this.
	genesis = append(genesis, sim.GenesisTuningEvent(0))
	secretEvents, err := persona.SecretEvents()
	if err != nil {
		return err
	}
	genesis = append(genesis, secretEvents...)
	if err := st.AppendEvents(genesis); err != nil {
		return err
	}
	if err := llm.WriteDefault(w.LLMConfigPath()); err != nil {
		return err
	}
	if err := persona.Genesis(dir, charterPreset); err != nil {
		return err
	}

	// A name-form world at a custom --at location is outside the worlds
	// home, so it needs a registry pointer to be name-addressable later
	// (D1/D6) — the default name-form location is inside the home and is
	// scan-owned, no registry entry wanted. Advisory: never fatal.
	if nameForm && *at != "" {
		if err := worlds.Upsert(worldName, dir); err != nil {
			fmt.Printf("warning: could not register %q in the known-worlds registry (advisory, continuing): %v\n", worldName, err)
		}
	}

	startHint := dir
	if nameForm {
		startHint = worldName
	}
	// The fresh-world local model (spec 034 US3/T015): named from
	// llm.DefaultConfig() rather than hardcoded, so this line can never drift
	// from what WriteDefault actually wrote above (contracts/fresh-world-defaults.md).
	localModel := llm.DefaultConfig().Providers["local"].Model
	fmt.Printf("created world %q in %s (seed %d)\nllm config: %s (edit providers/routes/budget; delete the file to disable LLM traffic)\nstart it with: promptworld start %s\nlocal model: %s — pull it first if you haven't: ollama pull %s\n",
		worldName, dir, *seed, w.LLMConfigPath(), startHint, localModel, localModel)
	if line := stageStatusLine(stage, *override); line != "" {
		fmt.Println(line)
	}
	if scenarioDef.ID != "" {
		fmt.Printf("scenario: %s — %s\n  the schedule arms at boot; attach and press 6 for the exercise panel\n",
			scenarioDef.ID, scenarioDef.Concept)
	}
	return nil
}

// resolveWorldForMigrate resolves a migrate argument to a directory. Unlike
// resolveWorld, it must reach v1 worlds — which this v2 build cannot
// world.Open, so worlds.Resolve (whose name lookup gates on openability) is
// blind to them. Path arguments pass through verbatim; bare names resolve
// against the worlds home then the known-worlds registry by manifest presence
// alone, never the version gate.
func resolveWorldForMigrate(arg string) (string, error) {
	if worlds.IsPathArg(arg) {
		return arg, nil
	}
	home, err := worlds.WorldsHome()
	if err != nil {
		return "", err
	}
	if cand := filepath.Join(home, arg); hasManifest(cand) {
		return cand, nil
	}
	if reg, err := worlds.LoadRegistry(); err == nil {
		if p, ok := reg.Worlds[arg]; ok && hasManifest(p) {
			return p, nil
		}
	}
	return "", fmt.Errorf("no world named %q (searched %s and the known-worlds list) — try `promptworld ps --all`", arg, home)
}

func hasManifest(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, world.ManifestName))
	return err == nil
}

// cmdMigrate implements `promptworld migrate <world>` (spec 012 US6, spec 013):
// the offline snapshot-cut migration that upgrades an older world (v1 or v2) to
// the current format — a v1 world chains 1→2→3 in one run. It resolves the
// world, then hands the whole archive/transform/rewrite ceremony to
// world.Migrate, and prints a human summary of what carried across the break.
func cmdMigrate(args []string) error {
	fs := flag.NewFlagSet("migrate", flag.ContinueOnError)
	arg, err := dirArg(fs, args)
	if err != nil {
		return err
	}
	dir, err := resolveWorldForMigrate(arg)
	if err != nil {
		return err
	}
	res, err := world.Migrate(dir)
	if err != nil {
		return err
	}
	if res.ManifestOnly {
		// v4→v5 (spec 068): the format bump only gates old software away from
		// new-terrain worlds — this world's log, state, and terrain carry
		// over untouched (terrain_gen stays absent ⇒ legacy terrain).
		fmt.Printf("migrated %q (seed %d) to format v%d\n  manifest-only upgrade — the event log and terrain carry over unchanged\nstart it with: promptworld start %s\n",
			res.Name, res.Seed, world.FormatVersion, arg)
		return nil
	}
	fmt.Printf("migrated %q (seed %d) to format v%d\n  %d villagers carried across the break at tick %d (%s)\n  %d source events archived in %s\nstart it with: promptworld start %s\n",
		res.Name, res.Seed, world.FormatVersion, res.AgentsCarried, res.Tick, clock.Format(res.Tick),
		res.SourceEvents, res.ArchivePath, arg)
	return nil
}

func cmdLLM(args []string) error {
	fs := flag.NewFlagSet("llm", flag.ContinueOnError)
	system := fs.String("system", "", "system prompt")
	maxTokens := fs.Int64("max-tokens", 0, "max output tokens (matters most for priced providers)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 3 {
		return fmt.Errorf("usage: promptworld llm <world> <kind> <prompt...>")
	}
	dir, err := resolveWorld(fs.Arg(0))
	if err != nil {
		return err
	}
	kind := fs.Arg(1)
	prompt := strings.Join(fs.Args()[2:], " ")
	w, err := world.Open(dir)
	if err != nil {
		return err
	}
	c, err := ipc.Dial(w.SockPath())
	if err != nil {
		return err
	}
	defer c.Close()
	data, err := c.Call("llm_call", ipc.LLMCallArgs{
		Kind: kind, System: *system, Prompt: prompt, MaxTokens: *maxTokens,
	})
	if err != nil {
		return err
	}
	var resp llm.Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	fmt.Print(formatLLMOneShot(resp))
	return nil
}

// formatLLMOneShot renders a one-shot llm_call response for the CLI (spec
// 024 US6/T018): the header names the serving PROVIDER (Response.Provider,
// always set — FR-011), not the legacy tier; any candidates the chain-walk
// passed over before landing here (Response.Skipped, contracts/status.md)
// get their own line with each mechanical reason, so a fallback is visible
// without reaching for `promptworld status`.
func formatLLMOneShot(resp llm.Response) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s · %s · %d in / %d out tokens · $%.4f · %dms]\n",
		resp.Provider, resp.Model, resp.InputTokens, resp.OutputTokens, resp.CostUSD, resp.Millis)
	if len(resp.Skipped) > 0 {
		skipped := make([]string, len(resp.Skipped))
		for i, sk := range resp.Skipped {
			skipped[i] = fmt.Sprintf("%s (%s)", sk.Provider, sk.Reason)
		}
		fmt.Fprintf(&b, "skipped: %s\n", strings.Join(skipped, ", "))
	}
	b.WriteString(resp.Text)
	b.WriteByte('\n')
	return b.String()
}

// cmdGuardian is the console one-shot (TASK-12; canonical name `guardian`
// since spec 052 FR-008, with `guardian` as a hidden compat alias): with a
// message, one mediated turn; without, the model-free status peek.
func cmdGuardian(args []string) error {
	fs := flag.NewFlagSet("guardian", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: promptworld guardian <world> [message...]")
	}
	dir, err := resolveWorld(fs.Arg(0))
	if err != nil {
		return err
	}
	w, err := world.Open(dir)
	if err != nil {
		return err
	}
	c, err := ipc.Dial(w.SockPath())
	if err != nil {
		return err
	}
	defer c.Close()

	if fs.NArg() == 1 {
		st, err := c.GuardianStatus()
		if err != nil {
			return err
		}
		charter := "custom charter in effect"
		if st.CharterDefault {
			charter = "default charter"
		}
		fmt.Printf("charges %s (%d/%d) · %s · charter.md at %s\n",
			chargeGlyphs(st.Charges), st.Charges, sim.GuardianChargeCap, charter, w.CharterPath())
		if len(st.Orders) > 0 {
			fmt.Printf("\n--- standing orders ---\n")
			for _, o := range st.Orders {
				fmt.Printf("%s\n", orderStatusLine(o))
			}
		}
		if strings.TrimSpace(st.SoulTail) != "" {
			fmt.Printf("\n--- recent notes ---\n%s\n", strings.TrimSpace(st.SoulTail))
		}
		return nil
	}

	r, err := c.GuardianChat(strings.Join(fs.Args()[1:], " "))
	if err != nil {
		return err
	}
	for _, m := range r.Moments {
		fmt.Printf("! %s\n", m)
	}
	fmt.Printf("\n%s\n", r.Reply)
	if r.Nudge != nil {
		fmt.Printf("\n⚡ %s → %s: %q\n", r.Nudge.Form, strings.Join(r.Nudge.Targets, ", "), r.Nudge.Text)
	}
	if r.Order != nil {
		fmt.Printf("\n👁 watch set (%s): %q\n", r.Order.ID, r.Order.Condition)
	}
	for _, id := range r.Cancelled {
		fmt.Printf("\n👁 watch released: %s\n", id)
	}
	if r.Clock != "" {
		fmt.Printf("\n⏲ %s\n", r.Clock)
	}
	fmt.Printf("\n[charges %s %d/%d]\n", chargeGlyphs(r.Charges), r.Charges, sim.GuardianChargeCap)
	return nil
}

// orderStatusLine renders one standing order for the CLI status peek (spec 029
// T023): id, a fuzzy marker, origin, remaining game-day, and status, followed
// by the condition text — the same fields the console/TUI surfaces show.
func orderStatusLine(o guardian.OrderStatus) string {
	fuzzy := ""
	if o.Fuzzy {
		fuzzy = " (fuzzy)"
	}
	return fmt.Sprintf("👁 %s%s [%s · day %d · %s]: %q", o.ID, fuzzy, o.Origin, o.ExpiresDay, o.Status, o.Condition)
}

func chargeGlyphs(n int) string {
	return strings.Repeat("⚡", n) + strings.Repeat("·", sim.GuardianChargeCap-n)
}

func cmdDaemon(args []string) error {
	fs := flag.NewFlagSet("daemon", flag.ContinueOnError)
	dir, err := worldArg(fs, args)
	if err != nil {
		return err
	}
	return daemon.Run(dir)
}

func cmdStart(args []string) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	dir, err := worldArg(fs, args)
	if err != nil {
		return err
	}
	w, err := world.Open(dir)
	if err != nil {
		return err
	}
	if running, pid := daemon.IsRunning(dir); running {
		return fmt.Errorf("daemon already running (pid %d)", pid)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	logf, err := os.OpenFile(w.LogPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer logf.Close()
	cmd := exec.Command(exe, "daemon", dir)
	cmd.Stdout, cmd.Stderr = logf, logf
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true} // detach from our session
	if err := cmd.Start(); err != nil {
		return err
	}
	// The child is re-parented on our exit; never wait on it.

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if c, err := ipc.Dial(w.SockPath()); err == nil {
			sd, err := c.Status("status", nil)
			c.Close()
			if err == nil {
				fmt.Printf("daemon started (pid %d): %s\n", sd.Daemon.Pid, clockLine(sd))
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("daemon did not answer within 5s — check %s", w.LogPath())
}

// requireWorldDir confirms dir at least looks like a world directory (a
// world.json manifest is present) without validating its content — the
// version-agnostic half of world.Open's own check. Daemon-lifecycle
// commands (stop/status, TASK-147) must reach a running daemon regardless
// of format_version, but a directory that was never a world at all (a typo,
// a stray path) still needs to fail clearly rather than silently reporting
// "daemon not running".
func requireWorldDir(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, world.ManifestName)); err != nil {
		return fmt.Errorf("not a world directory (missing %s): %w", world.ManifestName, err)
	}
	return nil
}

func cmdStop(args []string) error {
	fs := flag.NewFlagSet("stop", flag.ContinueOnError)
	dir, err := worldArg(fs, args)
	if err != nil {
		return err
	}
	if err := requireWorldDir(dir); err != nil {
		return err
	}
	// Deliberately no world.Open here (TASK-147): stop's job is reaching the
	// daemon PROCESS, which must work even for a format_version this build
	// can no longer Open — otherwise a running old-version daemon can never
	// be stopped by a newer binary, and migrate (which refuses a running
	// world) can't help either. daemon.IsRunning and the socket/pid paths
	// below are all pure-path / pidfile checks, not validating opens.
	running, pid := daemon.IsRunning(dir)
	if !running {
		fmt.Println("daemon not running")
		return nil // idempotent
	}
	if c, err := ipc.Dial(world.SockPathIn(dir)); err == nil {
		c.Call("shutdown", nil)
		c.Close()
	} else {
		// Socket dead but pid alive: fall back to SIGTERM (same graceful path).
		syscall.Kill(pid, syscall.SIGTERM)
	}
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if running, _ := daemon.IsRunning(dir); !running {
			fmt.Println("daemon stopped")
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("daemon (pid %d) did not stop within 30s", pid)
}

func cmdStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	dir, err := parseWorldFlags(fs, args)
	if err != nil {
		return err
	}
	if err := requireWorldDir(dir); err != nil {
		return err
	}

	// Reaching a live daemon (TASK-147) is version-agnostic, same reasoning
	// as cmdStop: the socket path is a pure join, and a running daemon
	// answers regardless of what format_version it booted with.
	if c, err := ipc.Dial(world.SockPathIn(dir)); err == nil {
		defer c.Close()
		sd, err := c.Status("status", nil)
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(sd)
		}
		fmt.Print(renderStatusHuman(sd))
		return nil
	}

	// No live daemon answered. The rest of this — manifest fields, the
	// offline store snapshot — is genuine world content, so it keeps the
	// version gate (TASK-147's carve-out: only the lifecycle check above
	// bypasses it).
	w, err := world.Open(dir)
	if err != nil {
		var vmErr *world.ErrFormatVersionMismatch
		if errors.As(err, &vmErr) {
			// requireWorldDir above already confirmed a manifest is present,
			// so this is specifically "this build can't read that
			// format_version" rather than a bogus path or corrupt content.
			// Daemon liveness is still checkable version-agnostically (the
			// dial above already ruled out a live daemon) — report that
			// instead of the migrate-hint error, which would misleadingly
			// read as "this world is broken" (TASK-147).
			if *asJSON {
				return printJSON(map[string]any{"daemon": map[string]any{"running": false}})
			}
			fmt.Println("daemon not running")
			return nil
		}
		return err
	}

	// Offline: last-known state from the store, read-only (shared with
	// `ps --all`'s stopped rows — specs/008-instance-manager D7).
	tick, paused, speed, lastSeq, ended, endedDay, err := worlds.OfflineSnapshot(w)
	if err != nil {
		return err
	}
	if *asJSON {
		clockMap := map[string]any{
			"tick": tick, "game_time": clock.Format(tick),
			"paused": paused, "speed": speed,
		}
		// Run-end posture (spec 044 FR-004): present only when ended, so a
		// living world's offline JSON is unchanged — mirroring the live
		// ClockStatus omitempty fields.
		if ended {
			clockMap["ended"] = true
			clockMap["ended_day"] = endedDay
		}
		worldMap := map[string]any{"name": w.Manifest.Name, "seed": w.Manifest.Seed, "format_version": w.Manifest.FormatVersion}
		if w.Manifest.Stage != "" {
			worldMap["stage"] = w.Manifest.Stage
			if w.Manifest.StageOverridden {
				worldMap["stage_overridden"] = true
			}
		}
		return printJSON(map[string]any{
			"world":  worldMap,
			"daemon": map[string]any{"running": false},
			"clock":  clockMap,
			"log":    map[string]any{"last_seq": lastSeq},
		})
	}
	fmt.Printf("world %q (seed %d) — daemon not running\nlast known: tick %d (%s), speed %s, paused %v\n",
		w.Manifest.Name, w.Manifest.Seed, tick, clock.Format(tick), speed, paused)
	if ended {
		fmt.Printf("run ended day %d, all villagers dead; world is an archive (read-only)\n", endedDay)
	}
	fmt.Printf("log: last seq %d\n", lastSeq)
	if line := stageStatusLine(w.Manifest.Stage, w.Manifest.StageOverridden); line != "" {
		fmt.Println(line)
	}
	return nil
}

func cmdTimeCtl(cmd string, args []string) error {
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	dir, err := worldArg(fs, args)
	if err != nil {
		return err
	}
	w, err := world.Open(dir)
	if err != nil {
		return err
	}
	c, err := ipc.Dial(w.SockPath())
	if err != nil {
		return err
	}
	defer c.Close()
	sd, err := c.Status(cmd, nil)
	if err != nil {
		return err
	}
	fmt.Println(clockLine(sd))
	return nil
}

func cmdSpeed(args []string) error {
	fs := flag.NewFlagSet("speed", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		return fmt.Errorf("usage: promptworld speed <world> <1x|4x|8x|16x|32x|max>")
	}
	val := fs.Arg(1)
	if _, err := clock.ParseSpeed(val); err != nil {
		return err
	}
	dir, err := resolveWorld(fs.Arg(0))
	if err != nil {
		return err
	}
	w, err := world.Open(dir)
	if err != nil {
		return err
	}
	c, err := ipc.Dial(w.SockPath())
	if err != nil {
		return err
	}
	defer c.Close()
	sd, err := c.Status("set_speed", ipc.SetSpeedArgs{Speed: val})
	if err != nil {
		return err
	}
	fmt.Println(setSpeedLine(sd))
	return nil
}

// cmdTeaching prints or toggles a world's teaching-posture marker (spec 039
// US4, contracts/posture.md §5). With no on|off argument it reports the current
// marker; with one it rewrites the manifest offline (world.SetTeaching) — the
// daemon reads the new value at its next boot, so the command says so. Exit
// non-zero only on IO/parse errors, never on state.
func cmdTeaching(args []string) error {
	fs := flag.NewFlagSet("teaching", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: promptworld teaching <world> [on|off]")
	}
	dir, err := resolveWorld(fs.Arg(0))
	if err != nil {
		return err
	}
	if fs.NArg() < 2 {
		w, err := world.Open(dir)
		if err != nil {
			return err
		}
		fmt.Printf("teaching: %s\n", onOff(w.Manifest.Teaching))
		return nil
	}
	var on bool
	switch fs.Arg(1) {
	case "on":
		on = true
	case "off":
		on = false
	default:
		return fmt.Errorf("teaching: want on or off, got %q", fs.Arg(1))
	}
	if err := world.SetTeaching(dir, on); err != nil {
		return err
	}
	fmt.Printf("teaching %s (applies at next daemon start)\n", onOff(on))
	return nil
}

func onOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

func cmdUI(args []string) error {
	fs := flag.NewFlagSet("ui", flag.ContinueOnError)
	dir, err := worldArg(fs, args)
	if err != nil {
		return err
	}
	w, err := world.Open(dir)
	if err != nil {
		return err
	}
	// WithMouseCellMotion (spec 049, research R2) enables click reporting
	// app-wide; only the chronicle inspect-list line click binds to it
	// (internal/tui handleMouse) — every other mouse event is unhandled and
	// falls through inert, so no existing keyboard behavior changes (FR-005).
	m, err := tea.NewProgram(tui.New(w), tea.WithAltScreen(), tea.WithMouseCellMotion()).Run()
	if err != nil {
		return err
	}
	// An unrecoverable protocol failure (e.g. reply over the cap, TASK-19)
	// quits the TUI; surface it as a real error and a non-zero exit.
	if fm, ok := m.(tui.Model); ok && fm.FatalErr() != "" {
		return fmt.Errorf("%s", fm.FatalErr())
	}
	return nil
}

func cmdAttach(args []string) error {
	fs := flag.NewFlagSet("attach", flag.ContinueOnError)
	dir, err := worldArg(fs, args)
	if err != nil {
		return err
	}
	w, err := world.Open(dir)
	if err != nil {
		return err
	}
	c, err := ipc.Dial(w.SockPath())
	if err != nil {
		return err
	}
	defer c.Close()

	sd, err := c.Status("status", nil)
	if err != nil {
		return err
	}
	fmt.Printf("attached to %q — %s\ncommands: pause | resume | speed <v> | status | quit\n", sd.World.Name, clockLine(sd))
	if err := c.Subscribe(nil); err != nil {
		return err
	}

	go func() {
		for p := range c.Pushes() {
			switch p.Push {
			case "event":
				fmt.Println(eventLine(*p.Event))
			case "dropped":
				fmt.Printf("-- stream overflowed at seq %d; re-syncing --\n", p.LastSeq)
				since := p.LastSeq
				if err := c.Subscribe(&since); err != nil {
					return
				}
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "quit", "exit", "q":
			fmt.Println("detached (the world keeps running)")
			return nil
		case "pause", "resume", "status":
			if sd, err := c.Status(fields[0], nil); err != nil {
				fmt.Printf("error: %v\n", err)
			} else {
				fmt.Println(clockLine(sd))
			}
		case "speed":
			if len(fields) < 2 {
				fmt.Println("usage: speed <1x|4x|8x|16x|32x|max>")
				continue
			}
			if sd, err := c.Status("set_speed", ipc.SetSpeedArgs{Speed: fields[1]}); err != nil {
				fmt.Printf("error: %v\n", err)
			} else {
				fmt.Println(setSpeedLine(sd))
			}
		default:
			fmt.Printf("unknown command %q\n", fields[0])
		}
	}
	return scanner.Err()
}

func cmdTail(args []string) error {
	fs := flag.NewFlagSet("tail", flag.ContinueOnError)
	since := fs.Int64("since", -1, "start after this seq (default: last 20 events)")
	follow := fs.Bool("follow", false, "keep following live events (requires a running daemon)")
	dir, err := parseWorldFlags(fs, args)
	if err != nil {
		return err
	}
	w, err := world.Open(dir)
	if err != nil {
		return err
	}

	// History always comes read-only from the store, daemon or not.
	st, err := store.Open(w.DBPath())
	if err != nil {
		return err
	}
	from := *since
	if from < 0 {
		from = st.LastSeq() - 20
		if from < 0 {
			from = 0
		}
	}
	events, err := st.EventsSince(from, 0)
	if err != nil {
		st.Close()
		return err
	}
	last := from
	for _, e := range events {
		fmt.Println(eventLine(e))
		last = e.Seq
	}
	st.Close()

	if !*follow {
		return nil
	}
	c, err := ipc.Dial(w.SockPath())
	if err != nil {
		return fmt.Errorf("--follow needs a running daemon: %w", err)
	}
	defer c.Close()
	if err := c.Subscribe(&last); err != nil {
		return err
	}
	for p := range c.Pushes() {
		switch p.Push {
		case "event":
			fmt.Println(eventLine(*p.Event))
		case "dropped":
			since := p.LastSeq
			if err := c.Subscribe(&since); err != nil {
				return err
			}
		}
	}
	return nil
}

// renderStatusHuman renders `promptworld status`'s online human-readable
// output: the world/daemon summary line, the clock line, one WARNING line
// per provider carrying an active health condition (contracts/
// provider-conditions.md "Human surfaces", spec 034 T009), and the log tail.
// Factored out of cmdStatus so the WARNING-block logic is unit-testable
// without a live daemon (T011). A healthy world — no LLM status, or an LLM
// status whose providers are all condition-free — renders byte-identical to
// pre-034 output; the loop below is a no-op in that case.
func renderStatusHuman(sd *ipc.StatusData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "world %q (seed %d) — daemon running (pid %d, up %ds, %d subscriber(s))\n%s\n",
		sd.World.Name, sd.World.Seed, sd.Daemon.Pid, sd.Daemon.UptimeSeconds, sd.Daemon.Subscribers,
		clockLine(sd))
	for _, line := range llmConditionWarnings(sd) {
		b.WriteString(line + "\n")
	}
	// Per-provider calibration state (spec 035 FR-004/US3): visible for the
	// whole life of the daemon, not just the boot line that scrolled away.
	if sd.LLM != nil {
		for _, p := range sd.LLM.Providers {
			b.WriteString(providerCalibrationLine(p) + "\n")
		}
	}
	for _, line := range horizonStatusLines(sd) {
		b.WriteString(line + "\n")
	}
	if line := postureStatusLine(sd); line != "" {
		b.WriteString(line + "\n")
	}
	if line := stageStatusLine(sd.World.Stage, sd.World.StageOverridden); line != "" {
		b.WriteString(line + "\n")
	}
	if line := scenarioStatusLine(sd); line != "" {
		b.WriteString(line + "\n")
	}
	fmt.Fprintf(&b, "log: last seq %d\n", sd.Log.LastSeq)
	return b.String()
}

// scenarioStatusLine renders the scenario line for `promptworld status`
// (spec 054 FR-007, D1): the exercise id and its model-free outcome. Empty
// for ambient worlds and old daemons — the wire fields are absent, so their
// output is unchanged (the stageStatusLine precedent).
func scenarioStatusLine(sd *ipc.StatusData) string {
	if sd.World.ScenarioExercise == "" {
		return ""
	}
	outcome := sd.World.ScenarioOutcome
	if outcome == "failed" {
		outcome = "failed (run ended)"
	}
	return fmt.Sprintf("exercise: %s — %s", sd.World.ScenarioExercise, strings.ReplaceAll(outcome, "_", " "))
}

// stageStatusLine renders the curriculum-ladder stage line for `promptworld
// status` (spec 046 FR-002/FR-009, R10): the active skin's display identity
// for the world's stage, plus the override marker when creation skipped
// ahead. Empty for a pre-046/pre-ladder world (Stage absent) — the wire field
// is absent, so their output is unchanged.
func stageStatusLine(stage string, overridden bool) string {
	if stage == "" {
		return ""
	}
	line := fmt.Sprintf("stage: %s (%s)", skin.StageName(stage), stage)
	if overridden {
		line += " [overridden]"
	}
	return line
}

// postureStatusLine renders the teaching-posture line for `promptworld status`
// (spec 039 US4, contracts/posture.md §5): the effective planner-safe rung and
// whether it is measured (calibrated) or a provisional bootstrap derivation
// (which points the operator at calibrate). Empty for non-teaching and pure-sim
// worlds — the wire field is absent, so their output is unchanged.
func postureStatusLine(sd *ipc.StatusData) string {
	if sd.Posture == nil {
		return ""
	}
	if sd.Posture.Calibrated {
		return fmt.Sprintf("teaching posture: %s (calibrated)", sd.Posture.Rung)
	}
	return fmt.Sprintf("teaching posture: %s (provisional — run `promptworld calibrate %s`)", sd.Posture.Rung, sd.World.Name)
}

// horizonStatusLines renders the live cognition-horizon section (spec 037 US3,
// FR-008): one line per watched class — its standing at the current effective
// speed, the calibrate-vs-slow-down remedy when suppressed, and the
// daemon-lifetime "skipped N" count. nil when the world carries no horizon, so
// a no-LLM world's output is unchanged (the wire field is absent).
func horizonStatusLines(sd *ipc.StatusData) []string {
	if len(sd.Horizon) == 0 {
		return nil
	}
	speed := sd.Clock.Speed
	lines := make([]string, 0, len(sd.Horizon))
	for _, e := range sd.Horizon {
		var line string
		if e.Suppressed {
			line = fmt.Sprintf("horizon: %s suppressed at %s — %s", e.Class, speed, horizonRemedy(e.Calibrated))
		} else {
			line = fmt.Sprintf("horizon: %s thinking at %s", e.Class, speed)
		}
		if e.Suppressed || e.SuppressedCount > 0 {
			line += fmt.Sprintf(" (skipped %d)", e.SuppressedCount)
		}
		lines = append(lines, line)
	}
	return lines
}

// horizonRemedy mirrors the TUI's calibrate-vs-slow-down split (spec 037
// FR-007): an uncalibrated class may still benefit from calibration; a
// calibrated one can only slow down.
func horizonRemedy(calibrated bool) string {
	if calibrated {
		return "slow down"
	}
	return "calibrate or slow down"
}

// llmConditionWarnings renders one `WARNING llm provider …` line per
// provider with an active health condition, per contracts/
// provider-conditions.md: `WARNING llm provider %q: %s — %s` (name, detail,
// remedy). nil when the world has no LLM status (offline-world fallback and
// no-orchestrator worlds both carry sd.LLM == nil) or every provider is
// healthy.
func llmConditionWarnings(sd *ipc.StatusData) []string {
	if sd.LLM == nil {
		return nil
	}
	var lines []string
	for _, p := range sd.LLM.Providers {
		if p.Condition == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("WARNING llm provider %q: %s — %s", p.Name, p.ConditionDetail, p.ConditionRemedy))
	}
	return lines
}

func clockLine(sd *ipc.StatusData) string {
	// Postmortem posture (spec 044, contracts/status.md): the run-over line
	// replaces the running/paused clock line entirely.
	if sd.Clock.Ended {
		return fmt.Sprintf("tick %d (%s) — run ended day %d, all villagers dead; world is an archive (read-only)",
			sd.Clock.Tick, sd.Clock.GameTime, sd.Clock.EndedDay)
	}
	state := "running"
	if sd.Clock.Paused {
		state = "paused"
	}
	extra := ""
	if sd.Clock.Degraded {
		extra = " [degraded]"
	}
	return fmt.Sprintf("tick %d (%s) — %s, speed %s (%.1f ticks/s effective)%s",
		sd.Clock.Tick, sd.Clock.GameTime, state, sd.Clock.Speed, sd.Clock.EffectiveRate, extra)
}

// setSpeedLine renders a set_speed reply (spec 035 FR-002): the clock line,
// plus the calibration-UX warning when the daemon sent one — additive; the
// speed change has already applied by the time this prints (the warning
// never blocks it).
func setSpeedLine(sd *ipc.StatusData) string {
	line := clockLine(sd)
	if sd.Warning != "" {
		line += "\nWARNING: " + sd.Warning
	}
	return line
}

// providerCalibrationLine renders one provider's calibration state for the
// human status rendering (spec 035 FR-004): the profile timestamp when
// seeded, or an explicit "uncalibrated (bootstrap)" marker when the wire's
// calibrated_at is absent (the bootstrap signal, FR-008).
func providerCalibrationLine(p llm.ProviderStatus) string {
	state := "uncalibrated (bootstrap)"
	if p.CalibratedAt != "" {
		state = "calibrated " + p.CalibratedAt
	}
	return fmt.Sprintf("  llm %s (%s): %s", p.Name, p.Model, state)
}

func eventLine(e store.Event) string {
	return fmt.Sprintf("#%-6d t%-8d %-14s %-18s %s", e.Seq, e.Tick, clock.Format(e.Tick), e.Type, string(e.Payload))
}

func printJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
