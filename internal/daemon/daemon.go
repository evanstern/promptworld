// Package daemon wires world + store + sim loop + ipc server into the
// always-on process, and owns lifecycle: recovery, pidfile, signals,
// graceful shutdown.
package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/evanstern/promptworld/internal/bundle"
	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/cognition"
	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/metatron"
	"github.com/evanstern/promptworld/internal/mind"
	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/scribe"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// Run is the foreground daemon primitive: recover, bind, tick until
// SIGTERM/SIGINT or a shutdown command, then snapshot and exit cleanly.
func Run(dir string) error {
	startWall := time.Now()

	// Tool registry gate (spec 014, FR-003/R9): a malformed registry or roster,
	// or a world tool missing its sim resolver/duration, aborts boot before the
	// world runs — never at tick time. World-independent, so it runs first.
	if err := tool.Validate(); err != nil {
		return fmt.Errorf("tool registry invalid: %w", err)
	}
	if err := sim.ValidateToolCoverage(); err != nil {
		return fmt.Errorf("tool registry coverage: %w", err)
	}

	w, err := world.Open(dir)
	if err != nil {
		return err
	}

	// Bundle tools (spec 036 T013): discover + validate the world's bundles/
	// folder once at boot, alongside the registry validators above, and freeze it
	// into a BundleSet the metatron turn assembly reads. Every BootReport entry —
	// a skipped tool or a rejected bundle — is logged on the boot channel naming
	// its file and rule (SC-005); a bad bundle never bricks boot (bundles are
	// additive). An absent/empty bundles/ dir yields an empty set and changes
	// nothing. Only a genuine I/O failure reading the root is fatal.
	bundleSet, err := bundle.Discover(dir)
	if err != nil {
		return fmt.Errorf("bundle discovery: %w", err)
	}
	for _, iss := range bundleSet.BootReport() {
		tl := ""
		if iss.Tool != "" {
			tl = " tool=" + iss.Tool
		}
		fmt.Printf("daemon: bundle=%s%s file=%s rule=%s severity=%s: %s\n",
			iss.Bundle, tl, iss.File, iss.Rule, iss.Severity, iss.Message)
	}
	if n := len(bundleSet.Roster()); n > 0 {
		fmt.Printf("daemon: bundles on (%d tool(s) from %d bundle(s))\n", n, len(bundleSet.Bundles()))
	}

	if err := acquirePidfile(w); err != nil {
		return err
	}
	// Remove the pidfile ONLY if it is still ours: a slow shutdown can
	// overlap a successor daemon that has already claimed the file, and
	// deleting its pid would orphan it (live-found: stop then reported "not
	// running" while the successor still held the database).
	defer func() {
		if data, err := os.ReadFile(w.PidPath()); err == nil &&
			strings.TrimSpace(string(data)) == strconv.Itoa(os.Getpid()) {
			os.Remove(w.PidPath())
		}
	}()

	// Advisory instance-manager registration (specs/008-instance-manager
	// D1/D6, FR-008): only worlds living outside the worlds home need a
	// registry entry (the home is scan-owned); registering here, not from
	// the `start` client, means even a foreground `promptworld daemon <dir>`
	// run becomes visible to `ps`. Best-effort — a failure never blocks boot.
	registerWorld(dir, w.Manifest.Name)

	st, err := store.Open(w.DBPath())
	if err != nil {
		return err
	}
	defer st.Close()

	if err := validateMeta(w, st); err != nil {
		return err
	}
	if err := st.CheckContiguity(); err != nil {
		return err
	}

	state, err := recoverState(w, st)
	if err != nil {
		return err
	}
	if err := seedMeetingConvention(w, st, state); err != nil {
		return err
	}
	recoveryMs := time.Since(startWall).Milliseconds()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	srv := ipc.NewServer(w, st, cancel)

	// Notify fan-out: the IPC broadcast, the always-on soul scribe, and
	// (when an orchestrator exists) the mind driver. All consumers are
	// non-blocking by contract.
	var consumers []func([]store.Event)
	consumers = append(consumers, srv.Broadcast)
	scr, err := scribe.New(dir, w.Manifest.Seed, w.Map(), state.Marshal())
	if err != nil {
		return err
	}
	defer scr.Close()
	consumers = append(consumers, scr.Observe)
	notify := func(evs []store.Event) {
		for _, c := range consumers {
			c(evs)
		}
	}
	loop := sim.NewLoop(state, w.Map(), st, notify)
	srv.SetLoop(loop)

	// Teaching-posture default (spec 039 US1): the highest planner-safe ladder
	// speed, computed inside the orchestrator branch and applied through the
	// loop's normal command door below so it lands as a recorded clock.speed_set
	// event (R3, replay byte-identity). Stays empty for non-teaching and
	// pure-sim teaching worlds — neither applies a default.
	var teachingPostureSpeed clock.Speed

	// LLM orchestrator: optional (config-gated), fully outside the sim loop —
	// an unreachable model degrades the AI layer, never the world.
	if llmCfg, err := llm.LoadConfig(w.LLMConfigPath()); err != nil {
		return err
	} else if llmCfg != nil {
		// Cognition-horizon gate (FR-002): every call kind must resolve to
		// a registered decision class before a model is ever reachable.
		kinds := make([]string, 0, 8)
		for _, k := range llm.Kinds() {
			kinds = append(kinds, string(k))
		}
		if err := cognition.ValidateKinds(kinds); err != nil {
			return err
		}
		orch, err := llm.New(*llmCfg, st)
		if err != nil {
			return err
		}
		defer orch.Close()
		srv.SetLLM(orch)
		// Provider-health surface (spec 034 US1/US2): every condition transition
		// (raise/reclassify/clear) is made loud on two operator channels — the
		// daemon log and the durable, broadcast daemon.llm_warning event that
		// line-mode attach streams and `status --json` audits. Transitions only,
		// so the durable log stays quiet under a steady condition (the periodic
		// re-probe below re-logs the plain line, never the event). The hook fires
		// from worker/preflight goroutines WHILE the loop runs, so its durable-event
		// leg MUST ride the loop's single-writer command door (loop.InjectOperator,
		// spec 034 R8) rather than appending straight to the store — store
		// AppendEvents has no internal locking, and the loop is the sole writer.
		// The loop stamps the current tick; when it is not running (the shutdown
		// window) InjectOperator errors and we degrade to the log line only,
		// dropping the durable event — the status fields still carry the condition.
		orch.SetConditionHook(func(provider, kind, detail, remedy string, active bool) {
			if active {
				fmt.Printf("daemon: WARNING llm provider %q: %s — %s\n", provider, detail, remedy)
			} else {
				fmt.Printf("daemon: llm provider %q recovered (%s cleared)\n", provider, kind)
			}
			payload, merr := json.Marshal(sim.LLMWarningPayload{
				Provider: provider, Kind: kind, Detail: detail, Remedy: remedy, Active: active,
			})
			if merr != nil {
				fmt.Printf("daemon: llm_warning payload marshal failed: %v\n", merr)
				return
			}
			if ierr := loop.InjectOperator([]store.Event{{Type: "daemon.llm_warning", Payload: payload}}); ierr != nil {
				// Loop not running (boot before Listen / shutdown after Run): the
				// warning is already on the log and the status fields, so dropping
				// the durable event is the safe degrade, never a boot failure.
				fmt.Printf("daemon: llm_warning event not recorded (%v)\n", ierr)
			}
		})
		// Adaptive-throttle debt sampler (spec 028 US1): a daemon-owned
		// goroutine that reads aggregate staleness debt every GovernorCadence
		// and exposes it for status. Built ONLY here, inside the orchestrator
		// branch, so a no-LLM world constructs zero governor machinery (FR-003,
		// SC-004). Observability only in this slice — no decisions, no events;
		// sampling rides the loop's non-blocking status door so it never blocks
		// the tick schedule.
		sampler := newGovernorSampler(orch, loop)
		srv.SetGovernor(sampler)
		go sampler.run(ctx)
		// Provider-health preflight (spec 034 US1): a daemon-owned goroutine that
		// probes each openai_compat provider's model listing once at boot, then
		// re-probes (every 60s) only while a preflight condition is active — so a
		// fresh world whose model is absent warns loudly within a probe RTT and
		// recovers on its own once the model is pulled, with no restart. Boot NEVER
		// blocks or fails on probe results (FR-002): the goroutine is fired and
		// forgotten under the shutdown ctx, exactly like the sampler above.
		go orch.RunPreflight(ctx)
		// Seed the seconds-per-point estimators from the calibration
		// profile before any traffic; a missing or unreadable file means
		// pessimistic bootstrap defaults (fail toward reflex, never toward
		// stale action).
		if prof, perr := cognition.LoadProfile(w.CalibrationPath()); perr != nil {
			fmt.Printf("daemon: %v — using bootstrap calibration defaults\n", perr)
			fmt.Print(uncalibratedBootWarning(w.Manifest.Name))
		} else if prof != nil {
			orch.SeedCalibration(prof)
			fmt.Printf("daemon: calibration seeded (local %.1fs/pt, cloud %.1fs/pt, calibrated %s)\n",
				cognition.SeedFor(prof, "local", true), cognition.SeedFor(prof, "cloud", false), prof.CalibratedAt)
		} else {
			fmt.Print(uncalibratedBootWarning(w.Manifest.Name))
		}
		// Teaching posture (spec 039 US1/US3): default the clock to the highest
		// planner-safe ladder rung, derived live from the planner-serving
		// provider's estimate — never hard-coded, so recalibration or provider
		// failover moves it at the next boot (SC-005). An uncalibrated
		// (bootstrap-seeded) provider still gets the rung applied, but the line
		// marks it provisional and prompts calibrate (US3). No planner-serving
		// provider ⇒ no posture (edge: treated as uncalibrated for posture).
		if w.Manifest.Teaching {
			if name, est, ok := orch.EstimateForKind(llm.Kind("planner")); ok {
				teachingPostureSpeed = clock.SpeedForRate(cognition.MaxSafeSpeed("planner", est))
				fmt.Print(teachingPostureBootLine(w.Manifest.Name, teachingPostureSpeed, est, orch.CalibratedAt(name)))
			}
		}
		cloudDesc := llmCfg.Cloud.Model
		if llmCfg.Cloud.Provider == llm.ProviderOpenAICompat {
			cloudDesc = fmt.Sprintf("%s @ %s", llmCfg.Cloud.Model, llmCfg.Cloud.Endpoint)
		}
		// Local-tier concurrency (TASK-45): surface the effective worker count
		// when it exceeds the default, and warn (never fatal) when the operator
		// configured an out-of-range value that was clamped.
		localWorkers, workersWarn := llmCfg.Local.Workers()
		if workersWarn != "" {
			fmt.Printf("daemon: %s\n", workersWarn)
		}
		// Agent tool-use loop knobs (TASK-52): surface any clamp/unknown-value
		// warning the same way as the concurrency knob — warn, never fatal.
		if _, roundsWarn := llmCfg.Rounds(); roundsWarn != "" {
			fmt.Printf("daemon: %s\n", roundsWarn)
		}
		if _, tmWarn := llmCfg.Local.ToolModeResolved(); tmWarn != "" {
			fmt.Printf("daemon: %s\n", tmWarn)
		}
		if _, tmWarn := llmCfg.Cloud.ToolModeResolved(); tmWarn != "" {
			fmt.Printf("daemon: %s\n", tmWarn)
		}
		localDesc := fmt.Sprintf("local %s @ %s", llmCfg.Local.Model, llmCfg.Local.Endpoint)
		if localWorkers > 1 {
			localDesc += fmt.Sprintf(", parallel %d", localWorkers)
		}
		fmt.Printf("daemon: llm orchestrator on (%s, cloud %s, budget $%.0f/mo)\n",
			localDesc, cloudDesc, llmCfg.MonthlyBudgetUSD)
		loopRounds, _ := llmCfg.Rounds()
		// Cognition token budgets (spec 025 US2): resolve the three per-kind
		// max_tokens knobs and surface any clamp/out-of-range warning on the same
		// boot channel — warn, never fatal. Effective values (defaults when absent)
		// thread into the constructors like loopRounds does.
		plannerTokens, plannerTokWarn := llmCfg.PlannerTokens()
		if plannerTokWarn != "" {
			fmt.Printf("daemon: %s\n", plannerTokWarn)
		}
		metatronTurnTokens, turnTokWarn := llmCfg.MetatronTurnTokens()
		if turnTokWarn != "" {
			fmt.Printf("daemon: %s\n", turnTokWarn)
		}
		consolidationTokens, consolTokWarn := llmCfg.ConsolidationTokens()
		if consolTokWarn != "" {
			fmt.Printf("daemon: %s\n", consolTokWarn)
		}
		// memory_relevance (spec 042): the world's selection-mode flag threads
		// into the mind like the token budgets — "" legacy, "shadow" records
		// divergence with prompts unchanged, "on" is the gated US3 posture.
		md, err := mind.New(orch, loop, loop, w.Map(), w.Manifest.Seed, state.Marshal(), persona.Load(dir), loopRounds, plannerTokens, consolidationTokens, w.Manifest.MemoryRelevance)
		if err != nil {
			return err
		}
		defer md.Close()
		consumers = append(consumers, md.Observe)
		// Drift signal: a tier's estimator breaching its spike-rate
		// threshold lands as cog.recalibration_recommended telemetry.
		orch.SetRecalibrateHook(md.RecalibrateSignal)
		fmt.Printf("daemon: mind driver on (%d villagers, cadence %d game-min)\n",
			sim.AgentCount, sim.PlannerCadenceTicks/60)
		// Embedding driver (spec 042 US1): wired ONLY when llm.json routes the
		// embedding kind — a peer of the consolidation driver, watching
		// committed agent.memory_added events and injecting recorded
		// agent.memory_embedded companions. An absent route is the subsystem's
		// off switch: one boot INFO line, no warn-backfill (absence is the
		// feature switch, mirroring "no llm.json → reflex-only"), and the world
		// stays vectorless.
		if orch.HasEmbedding() {
			embName, embModel, _ := orch.EmbeddingProvider()
			// Transport failures surface on the TASK-91 loud channel exactly
			// like provider-health conditions: a daemon-log WARNING plus a
			// durable daemon.llm_warning through the loop's operator door. The
			// embedder debounces per failure episode, so a dead endpoint warns
			// once, not per memory.
			emb, err := mind.NewEmbedder(orch, loop, func(detail string) {
				remedy := fmt.Sprintf("check the embedding endpoint (provider %q) and `ollama pull %s`", embName, embModel)
				fmt.Printf("daemon: WARNING embedding provider %q: %s — %s\n", embName, detail, remedy)
				payload, merr := json.Marshal(sim.LLMWarningPayload{
					Provider: embName, Kind: "embedding-unavailable", Detail: detail, Remedy: remedy, Active: true,
				})
				if merr != nil {
					fmt.Printf("daemon: llm_warning payload marshal failed: %v\n", merr)
					return
				}
				if ierr := loop.InjectOperator([]store.Event{{Type: "daemon.llm_warning", Payload: payload}}); ierr != nil {
					fmt.Printf("daemon: llm_warning event not recorded (%v)\n", ierr)
				}
			}, w.Map(), w.Manifest.Seed, state.Marshal())
			if err != nil {
				return err
			}
			defer emb.Close()
			consumers = append(consumers, emb.Observe)
			fmt.Printf("daemon: embedder on (%s via provider %q)\n", embModel, embName)
		} else {
			fmt.Printf("daemon: embedding off (no \"embedding\" route in llm.json — memories stay vectorless)\n")
		}
		mt, err := metatron.New(orch, loop, loop, w.Map(), w.Manifest.Seed, state.Marshal(), dir, loopRounds, metatronTurnTokens)
		if err != nil {
			return err
		}
		mt.SetBundles(bundleSet) // spec 036 T013: hand the frozen bundle surface to the turn assembly
		defer mt.Close()
		consumers = append(consumers, mt.Observe)
		srv.SetMetatron(mt)
		fmt.Printf("daemon: metatron on (charges %d/%d)\n", state.MetatronCharges, sim.MetatronChargeCap)
	}

	// Stale socket from a crashed daemon: the pidfile said no one is alive.
	os.Remove(w.SockPath())
	if err := srv.Listen(); err != nil {
		return err
	}
	defer srv.Close()

	if err := appendDaemonEvent(st, srv, "daemon.started",
		sim.DaemonStartedPayload{Tick: state.Tick, RecoveryMs: recoveryMs}, state.Tick); err != nil {
		return err
	}
	fmt.Printf("daemon: world %q at tick %d (%s), recovery %dms, socket %s\n",
		w.Manifest.Name, state.Tick, clock.Format(state.Tick), recoveryMs, w.SockPath())

	go srv.Serve()

	// Apply the teaching-posture default through the loop's normal set_speed
	// door so it is recorded as a clock.speed_set event — replay reproduces it
	// byte-identically (spec 039 R3). The goroutine blocks on the command
	// channel until Run drains it as the first command; a set that fails
	// (loop already stopping) is logged, never fatal.
	if teachingPostureSpeed != "" {
		go func(sp clock.Speed) {
			if _, err := loop.Do("set_speed", sp); err != nil {
				fmt.Printf("daemon: teaching posture default not applied (%v)\n", err)
			}
		}(teachingPostureSpeed)
	}

	runErr := loop.Run(ctx) // returns after final snapshot

	if err := appendDaemonEvent(st, srv, "daemon.stopped",
		sim.DaemonStoppedPayload{Tick: state.Tick}, state.Tick); err != nil && runErr == nil {
		runErr = err
	}
	fmt.Printf("daemon: stopped at tick %d\n", state.Tick)
	return runErr
}

// uncalibratedBootWarning composes the boot warning block for an LLM world
// with no usable calibration profile (spec 035 FR-001, contracts/warnings.md
// §1): the uncalibrated statement, the per-class suppression horizon at
// bootstrap seeds (the identical string `promptworld calibrate` prints,
// cognition.HorizonSummary — FR-006), and the exact calibrate command for
// this world. Both the absent-profile and unreadable-profile boot branches
// print it — an unreadable file already falls back to bootstrap and is
// uncalibrated in every sense that matters (spec edge case); the
// profile-seeded branch never calls this and stays byte-identical (US2 AC2).
func uncalibratedBootWarning(worldName string) string {
	return fmt.Sprintf(
		"daemon: WARNING — world is UNCALIBRATED: latency estimates are pessimistic bootstrap defaults (local %.0fs/pt, cloud %.0fs/pt)\n"+
			"daemon: at these estimates: %s\n"+
			"daemon: run `promptworld calibrate %s` to measure this rig\n",
		cognition.BootstrapLocalSecPerPt, cognition.BootstrapCloudSecPerPt,
		cognition.HorizonSummary(cognition.BootstrapLocalSecPerPt), worldName)
}

// teachingPostureBootLine renders the teaching world's boot posture line
// (spec 039 contracts/posture.md §2). calibratedAt is the planner-serving
// provider's calibration timestamp: non-empty prints the measured flavor;
// empty prints the provisional bootstrap flavor and appends the explicit
// `promptworld calibrate <world>` prompt (US3) — the pessimistic rung is still
// applied, but the operator is told it cannot yet be honest. Only teaching
// worlds reach this, so non-teaching and uncalibrated-non-teaching boot output
// stays byte-identical (FR-008, US3 AC3).
func teachingPostureBootLine(worldName string, speed clock.Speed, secPerPt float64, calibratedAt string) string {
	if calibratedAt != "" {
		return fmt.Sprintf("daemon: teaching posture: defaulting speed to %s (planner-safe at %.1fs/pt, calibrated %s)\n",
			speed, secPerPt, calibratedAt)
	}
	return fmt.Sprintf(
		"daemon: teaching posture: defaulting speed to %s (provisional — planner-safe at %.1fs/pt bootstrap estimate)\n"+
			"daemon: teaching posture cannot yet be honest — run `promptworld calibrate %s` so the classroom default reflects this rig\n",
		speed, secPerPt, worldName)
}

// seedMeetingConvention injects the config-declared meeting convention on boot
// if the manifest declares one and none has taken hold yet (TASK-36) — once,
// at the recovered tick. It lands in the log like genesis, so replay
// re-applies it and this boot-time seed never fires twice (the reducer is
// one-shot, and the guard skips re-injection once state carries a convention).
func seedMeetingConvention(w *world.World, st *store.Store, state *sim.State) error {
	mc := w.Manifest.Meeting
	if mc == nil || state.MeetingConvention != nil {
		return nil
	}
	convene, open, err := mc.Seconds()
	if err != nil {
		return err // already validated in world.Open; defensive
	}
	ev := sim.NewConventionEvent(state, w.Map(), state.Tick, convene, open, mc.X, mc.Y)
	if err := state.Apply(ev); err != nil {
		return err
	}
	return st.AppendEvents([]store.Event{ev})
}

func appendDaemonEvent(st *store.Store, srv *ipc.Server, typ string, payload any, tick int64) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	events := []store.Event{{Tick: tick, Type: typ, Payload: b}}
	if err := st.AppendEvents(events); err != nil {
		return err
	}
	srv.Broadcast(events)
	return nil
}

// recoverState rebuilds world state from the newest valid snapshot plus
// event replay through the same reducer the live loop uses. The clock
// resumes at max(snapshot tick, last event tick); quiet trailing ticks
// re-run deterministically.
func recoverState(w *world.World, st *store.Store) (*sim.State, error) {
	state := sim.NewState(w.Manifest.Seed, w.Map())
	var since int64
	if snap, err := st.LatestValidSnapshot(); err != nil {
		return nil, err
	} else if snap != nil {
		if err := json.Unmarshal(snap.State, state); err != nil {
			return nil, fmt.Errorf("snapshot %d unreadable despite valid hash: %w", snap.ID, err)
		}
		since = snap.Seq
	}
	err := st.ReplayEvents(since, func(e store.Event) error {
		if err := state.Apply(e); err != nil {
			return err
		}
		if e.Tick > state.Tick {
			state.Tick = e.Tick
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("replay: %w", err)
	}
	return state, nil
}

func validateMeta(w *world.World, st *store.Store) error {
	// First daemon run stamps meta; later runs must match the manifest.
	for key, want := range map[string]string{
		"seed":           strconv.FormatUint(w.Manifest.Seed, 10),
		"format_version": strconv.Itoa(w.Manifest.FormatVersion),
	} {
		got, err := st.GetMeta(key)
		if err != nil {
			return err
		}
		if got == "" {
			if err := st.SetMeta(key, want); err != nil {
				return err
			}
			continue
		}
		if got != want {
			return fmt.Errorf("world.json and world.db disagree on %s (%s vs %s) — this save directory is corrupt or mixed from two runs", key, want, got)
		}
	}
	return nil
}

// registerWorld best-effort upserts this world into the advisory
// known-worlds registry when it lives outside the current worlds home.
// Failures (including a broken $HOME) are logged and never fatal — the
// registry is a pointer cache, never required for the world to run.
func registerWorld(dir, name string) {
	inside, err := worlds.InsideWorldsHome(dir)
	if err != nil {
		fmt.Printf("daemon: known-worlds registry check failed (advisory, continuing): %v\n", err)
		return
	}
	if inside {
		return // home is scan-owned; no registry entry needed
	}
	if err := worlds.Upsert(name, dir); err != nil {
		fmt.Printf("daemon: known-worlds registry update failed (advisory, continuing): %v\n", err)
	}
}

// acquirePidfile enforces one daemon per world dir, sweeping leftovers from
// crashed daemons (stale pid whose process is gone).
func acquirePidfile(w *world.World) error {
	if data, err := os.ReadFile(w.PidPath()); err == nil {
		if pid, perr := strconv.Atoi(strings.TrimSpace(string(data))); perr == nil && pidAlive(pid) {
			return fmt.Errorf("daemon already running (pid %d)", pid)
		}
		// Stale: crashed daemon left it behind. Sweep pid + socket.
		os.Remove(w.PidPath())
		os.Remove(w.SockPath())
	}
	return os.WriteFile(w.PidPath(), []byte(strconv.Itoa(os.Getpid())+"\n"), 0o644)
}

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// IsRunning reports whether a live daemon holds this world's pidfile.
func IsRunning(dir string) (bool, int) {
	w, err := world.Open(dir)
	if err != nil {
		return false, 0
	}
	data, err := os.ReadFile(w.PidPath())
	if err != nil {
		return false, 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || !pidAlive(pid) {
		return false, 0
	}
	return true, pid
}
