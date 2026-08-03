package mind

// Spec 110 Phase 4 (T015–T018): the REPLAY EVIDENCE harness.
//
// Why a test and not a cmd/ binary: the measurement is the exact contents of
// md.narrLines — an unexported field on an unexported-by-convention absorb path
// — so only an in-package caller can take it. A cmd/ tool would have forced
// exporting the chapter buffer purely to observe it, which is a worse trade
// than one env-gated test.
//
// Nothing here runs in the ordinary `go test ./...`: both tests skip unless
// PW_REPLAY_DB names a world.db to replay. Run them by hand:
//
//	PW_REPLAY_DB=/tmp/runA.db PW_REPLAY_OUT=/tmp/runA-replay.json \
//	  go test ./internal/mind/ -run TestReplayEvidence -v -timeout 60m
//
// The world.db MUST be a COPY: store.Open writes (WAL, pragmas), and the
// preserved soak worlds are the before-side of the comparison.
//
// The before/after comparison is controlled on identical input: BOTH passes run
// the same production absorb path over the same log, and the only difference is
// that the "before" pass wipes the harvest ledger after every absorbed event,
// so every correction takes the FR-004 unexplained branch — which spec 110
// keeps byte-identical to the pre-110 line. That makes the "before" pass a
// faithful reconstruction of today's `main`, produced by the same code, rather
// than a second implementation of it.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// replayHTTP is the re-narration client. Chapters are minutes apart on a local
// 12–36B model, so the timeout is generous by design.
var replayHTTP = &http.Client{Timeout: 15 * time.Minute}

// --- stubs -------------------------------------------------------------

// replaySocial is the injection door for the replay: the Mind must have a
// non-nil social injector or chronicleNote returns immediately, but nothing the
// replay's Mind injects may reach the replayed world (the log is the only
// authority). Thread-safe because the detached telemetry paths (emitCog,
// emitSuppressed) inject from their own goroutines.
type replaySocial struct{}

func (s *replaySocial) InjectSocial(events []store.Event) error { return nil }

// replayOrch fails every model call instantly. The absorb path's conversation /
// consolidation / meeting arms still fire (they are part of absorb and we do not
// suppress them), but they must cost no wall time and reach no model: this is an
// OFFLINE replay, and none of those arms can touch md.narrLines.
type replayOrch struct{}

func (o *replayOrch) Submit(ctx context.Context, req llm.Request) (llm.Response, error) {
	return llm.Response{}, fmt.Errorf("replay: offline")
}

// --- per-chapter record ------------------------------------------------

type replayChapter struct {
	Label string `json:"label"`
	Day   int64  `json:"day"`
	From  int64  `json:"from_tick"`
	To    int64  `json:"to_tick"`
	// Produced is every line the chapter appended, including those the
	// narrMaxLines ring later evicted — the spec's "total narratable lines".
	Produced int `json:"produced_lines"`
	// Retained is what the narrator actually receives (len(job.lines)).
	Retained int `json:"retained_lines"`
	// CorrLines is the number of PRODUCED lines that are map-correction lines:
	// per-event "found it gone" lines plus the one coalesced summary line.
	CorrLines int `json:"correction_lines"`
	// Attributed / Unexplained are the chapter's classifications (FR-007).
	Attributed  int `json:"attributed"`
	Unexplained int `json:"unexplained"`
	// Summary is 1 when the chapter carried the coalesced line.
	Summary  int      `json:"summary_lines"`
	Overflow bool     `json:"overflowed_ring"`
	Lines    []string `json:"lines,omitempty"`
}

type replayResult struct {
	Mode     string          `json:"mode"`
	Events   int             `json:"events"`
	Chapters []replayChapter `json:"chapters"`
	// Corrections is every agent.map_corrected event, with the production
	// classifier's verdict and an INDEPENDENT ground truth built from the raw
	// log (SC-004).
	Corrections []replayCorrection `json:"corrections"`
	Telemetry   []string           `json:"telemetry"`
}

type replayCorrection struct {
	Seq        int64  `json:"seq"`
	Tick       int64  `json:"tick"`
	X, Y       int    `json:"-"`
	Chapter    string `json:"chapter"`
	Attributed bool   `json:"attributed"`
	// TruthHarvest is true when the raw log carries a chopped/quarried at
	// EXACTLY (x,y) with 0 <= tick-harvestTick <= harvestLedgerWindow —
	// computed by this harness from the log, not from the ledger.
	TruthHarvest bool   `json:"truth_harvest_in_window"`
	TruthEver    bool   `json:"truth_harvest_ever"`
	Line         string `json:"line,omitempty"`
}

// --- the replay --------------------------------------------------------

// replayWorld drives the whole event log through the Mind's absorb path,
// one event at a time, and returns the per-chapter buffers the narrator would
// have received. ledger=false wipes the harvest ledger after every event, which
// reproduces pre-spec-110 behaviour through the post-110 code.
func replayWorld(t *testing.T, dbPath string, seed uint64, mw, mh, terrain int, ledger bool, keepLines bool) replayResult {
	t.Helper()
	st, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open %s: %v", dbPath, err)
	}
	defer st.Close()

	// Ground truth for SC-004, built from the raw log in a first pass: every
	// harvest tick at every coordinate. Independent of the ledger under test.
	truth := map[[2]int][]int64{}
	if err := st.ReplayEvents(0, func(e store.Event) error {
		if e.Type != "agent.chopped" && e.Type != "agent.quarried" {
			return nil
		}
		var p struct {
			X int `json:"x"`
			Y int `json:"y"`
		}
		if json.Unmarshal(e.Payload, &p) == nil {
			truth[[2]int{p.X, p.Y}] = append(truth[[2]int{p.X, p.Y}], e.Tick)
		}
		return nil
	}); err != nil {
		t.Fatalf("truth pass: %v", err)
	}

	m := worldmap.GenerateV(seed, mw, mh, terrain)
	state := sim.NewState(seed, m)
	md := &Mind{
		replica:      state,
		m:            m,
		orch:         &replayOrch{},
		social:       &replaySocial{},
		pairSeen:     map[[2]int]int64{},
		pairAdjacent: map[[2]int]bool{},
		narrQ:        make(chan narrJob, 512),
		narrRetry:    make(chan narrCarry, 1),
		done:         make(chan struct{}),
	}

	// Capture the FR-007 telemetry the daemon would have logged, and keep the
	// absorb path's chatter out of the test output.
	var logBuf bytes.Buffer
	oldOut, oldFlags := log.Writer(), log.Flags()
	log.SetOutput(&logBuf)
	log.SetFlags(0)
	defer func() { log.SetOutput(oldOut); log.SetFlags(oldFlags) }()

	res := replayResult{Mode: map[bool]string{true: "after", false: "before"}[ledger]}
	produced := 0
	pendingCorr := map[int64]*replayCorrection{} // seq -> record, until its chapter closes
	var chapterCorr []*replayCorrection

	drain := func() {
		for {
			select {
			case job := <-md.narrQ:
				if job.epilogue {
					continue
				}
				ch := replayChapter{Label: job.label, Day: job.day, From: job.fromTick, To: job.toTick,
					Retained: len(job.lines)}
				for _, l := range job.lines {
					if strings.Contains(l, corrSummaryMarker) {
						ch.Summary++
					}
				}
				ch.Produced = produced + ch.Summary
				ch.Overflow = ch.Produced > narrMaxLines
				for _, c := range chapterCorr {
					c.Chapter = job.label
					if c.Attributed {
						ch.Attributed++
					} else {
						ch.Unexplained++
						ch.CorrLines++
					}
				}
				ch.CorrLines += ch.Summary
				if keepLines {
					ch.Lines = append([]string(nil), job.lines...)
				}
				res.Chapters = append(res.Chapters, ch)
				produced = 0
				chapterCorr = nil
			default:
				return
			}
		}
	}

	n := 0
	if err := st.ReplayEvents(0, func(e store.Event) error {
		n++
		var corr *replayCorrection
		if e.Type == "agent.map_corrected" {
			var p sim.MapCorrectedPayload
			if json.Unmarshal(e.Payload, &p) == nil && len(p.Gone) > 0 {
				f := p.Gone[0]
				_, ok := md.attributedHarvest(f.X, f.Y, e.Tick) // pure read, same ledger state chronicleNote will see
				corr = &replayCorrection{Seq: e.Seq, Tick: e.Tick, X: f.X, Y: f.Y, Attributed: ok}
				for _, ht := range truth[[2]int{f.X, f.Y}] {
					corr.TruthEver = true
					if age := e.Tick - ht; age >= 0 && age <= harvestLedgerWindow {
						corr.TruthHarvest = true
					}
				}
				pendingCorr[e.Seq] = corr
			}
		}

		n0 := len(md.narrLines)
		var last0 string
		if n0 > 0 {
			last0 = md.narrLines[n0-1]
		}
		md.absorb([]store.Event{e})
		if !ledger {
			md.harvests = harvestLedger{}
		}
		n1 := len(md.narrLines)
		appended := n1 > n0 || (n1 == n0 && n0 == narrMaxLines && md.narrLines[n1-1] != last0)
		if appended {
			produced++
			if corr != nil {
				corr.Line = md.narrLines[n1-1]
			}
		}
		if corr != nil {
			chapterCorr = append(chapterCorr, corr)
			res.Corrections = append(res.Corrections, *corr)
		}
		drain()
		return nil
	}); err != nil {
		t.Fatalf("replay: %v", err)
	}
	// The tail after the last boundary is never narrated in production either
	// (the daemon exits), so it is deliberately not recorded as a chapter.
	res.Events = n
	// Re-stamp each correction's chapter now that the chapters are closed.
	for i := range res.Corrections {
		if c := pendingCorr[res.Corrections[i].Seq]; c != nil {
			res.Corrections[i].Chapter = c.Chapter
		}
	}
	for _, l := range strings.Split(logBuf.String(), "\n") {
		if strings.Contains(l, "corrections: ") {
			res.Telemetry = append(res.Telemetry, l)
		}
	}
	return res
}

func replayEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		fmt.Sscanf(v, "%d", &n)
		return n
	}
	return def
}

// TestReplayEvidence is T015 (the harness) plus T016/T017 (the assertions it
// carries): SC-004 precision, SC-003 anti-suppression, SC-002 volume.
func TestReplayEvidence(t *testing.T) {
	db := os.Getenv("PW_REPLAY_DB")
	if db == "" {
		t.Skip("set PW_REPLAY_DB to a COPY of a soak world.db to run the spec 110 replay")
	}
	seed := uint64(replayEnvInt("PW_REPLAY_SEED", 1337))
	mw := replayEnvInt("PW_REPLAY_MAPW", 64)
	mh := replayEnvInt("PW_REPLAY_MAPH", 64)
	tg := replayEnvInt("PW_REPLAY_TERRAIN", 2)

	before := replayWorld(t, db, seed, mw, mh, tg, false, false)
	after := replayWorld(t, db, seed, mw, mh, tg, true, true)

	// --- SC-004: precision. No correction classified attributed may lack a
	// coordinate-matching harvest inside the window, judged against the
	// independently-built ground truth.
	falsePos, attributed, unexplained, unexplainedNoHarvestEver := 0, 0, 0, 0
	for _, c := range after.Corrections {
		if c.Attributed {
			attributed++
			if !c.TruthHarvest {
				falsePos++
				t.Errorf("SC-004 false attribution: seq %d tick %d had no harvest in window", c.Seq, c.Tick)
			}
			if c.Line != "" {
				t.Errorf("FR-003 violated: attributed correction seq %d still emitted %q", c.Seq, c.Line)
			}
			continue
		}
		unexplained++
		// --- SC-003: every unexplained correction keeps its own line.
		if c.Line == "" {
			t.Errorf("SC-003 violated: unexplained correction seq %d emitted no line", c.Seq)
		}
		if !c.TruthEver {
			unexplainedNoHarvestEver++
		}
	}
	t.Logf("SC-004: %d corrections, %d attributed, %d unexplained, %d false attributions",
		len(after.Corrections), attributed, unexplained, falsePos)
	t.Logf("SC-003: %d unexplained corrections have no harvest at their coordinates ANYWHERE in the log",
		unexplainedNoHarvestEver)

	// --- SC-002: at most one correction line per chapter, and no chapter
	// overflows the ring on corrections' account.
	fmt.Printf("\n%-34s %8s %8s %8s %8s %8s %8s\n",
		"chapter", "corr(b)", "tot(b)", "share", "corr(a)", "tot(a)", "share")
	for i := range after.Chapters {
		a := after.Chapters[i]
		var b replayChapter
		if i < len(before.Chapters) {
			b = before.Chapters[i]
		}
		pct := func(c, tot int) string {
			if tot == 0 {
				return "-"
			}
			return fmt.Sprintf("%d%%", c*100/tot)
		}
		fmt.Printf("%-34s %8d %8d %8s %8d %8d %8s\n", a.Label,
			b.CorrLines, b.Produced, pct(b.CorrLines, b.Produced),
			a.CorrLines, a.Produced, pct(a.CorrLines, a.Produced))
		// The bound spec 110 actually claims: however many corrections a chapter
		// attributes, they contribute exactly ONE line. Unexplained ones keep
		// theirs by design (FR-004), so they are not counted against this.
		if a.Attributed > 0 && a.Summary != 1 {
			t.Errorf("SC-002: chapter %q attributed %d corrections but carried %d summary lines",
				a.Label, a.Attributed, a.Summary)
		}
		if a.Attributed == 0 && a.Summary != 0 {
			t.Errorf("FR-008: chapter %q attributed nothing but carried %d summary lines",
				a.Label, a.Summary)
		}
		// "No chapter overflows narrMaxLines on corrections' account": a chapter
		// overflows on their account only when removing its correction lines
		// would bring it back under the ring.
		if a.Produced > narrMaxLines && a.Produced-a.CorrLines <= narrMaxLines {
			t.Errorf("SC-002: chapter %q overflowed the ring (%d lines) ON CORRECTIONS' ACCOUNT (%d correction lines)",
				a.Label, a.Produced, a.CorrLines)
		}
	}
	overB, overA := 0, 0
	for _, c := range before.Chapters {
		if c.Produced > narrMaxLines {
			overB++
		}
	}
	for _, c := range after.Chapters {
		if c.Produced > narrMaxLines {
			overA++
		}
	}
	t.Logf("SC-002: chapters overflowing narrMaxLines=%d — before %d, after %d (of %d/%d chapters)",
		narrMaxLines, overB, overA, len(before.Chapters), len(after.Chapters))

	if out := os.Getenv("PW_REPLAY_OUT"); out != "" {
		writeJSON(t, out+".before.json", before)
		writeJSON(t, out+".after.json", after)
		t.Logf("wrote %s.{before,after}.json", out)
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}
}

// --- T018: re-narration against a live model ---------------------------

// TestReplayRenarrate re-narrates the replayed chapters with the real narrator
// prompts against a live OpenAI-compatible endpoint, closing the loop on the
// thread list exactly as a live run does (each chapter is offered the slugs the
// PRECEDING re-narrated chapters produced, not the original run's). The request
// body is the one internal/llm's openai_compat transport builds for a narrator
// request that carries no tools and no response schema: {model, messages,
// stream:false, max_tokens} plus reasoning_effort when configured.
func TestReplayRenarrate(t *testing.T) {
	db := os.Getenv("PW_REPLAY_DB")
	model := os.Getenv("PW_NARRATE_MODEL")
	if db == "" || model == "" {
		t.Skip("set PW_REPLAY_DB and PW_NARRATE_MODEL to re-narrate")
	}
	endpoint := os.Getenv("PW_NARRATE_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434/v1"
	}
	seed := uint64(replayEnvInt("PW_REPLAY_SEED", 1337))
	// PW_RENARRATE_MODE=before narrates the BEFORE-side buffers instead, giving
	// a controlled same-model, same-harness before/after comparison.
	ledger := os.Getenv("PW_RENARRATE_MODE") != "before"
	after := replayWorld(t, db, seed, replayEnvInt("PW_REPLAY_MAPW", 64),
		replayEnvInt("PW_REPLAY_MAPH", 64), replayEnvInt("PW_REPLAY_TERRAIN", 2), ledger, true)

	type outEntry struct {
		Chapter string `json:"chapter"`
		Text    string `json:"text"`
		Thread  string `json:"thread"`
	}
	var entries []outEntry
	var parseFails []string
	var threadRing []string

	// PW_RENARRATE_FROM / _LIMIT select a chapter window (indices, half-open),
	// for a bounded cross-model spot-check on a model too slow to run the whole
	// log. A window loses the closed-loop thread history before it, which is
	// recorded wherever such a run is reported.
	from := replayEnvInt("PW_RENARRATE_FROM", 0)
	limit := replayEnvInt("PW_RENARRATE_LIMIT", 0)
	for i, ch := range after.Chapters {
		if i < from {
			continue
		}
		if limit > 0 && i >= limit {
			fmt.Printf("(stopping at chapter %d — PW_RENARRATE_LIMIT)\n", limit)
			break
		}
		job := narrJob{day: ch.Day, label: ch.Label, fromTick: ch.From, toTick: ch.To, lines: ch.Lines}
		// Closed-loop threads: most recent first, distinct, capped at 8 — the
		// same shape closeChapter builds from the chronicle ring.
		seen := map[string]bool{}
		for i := len(threadRing) - 1; i >= 0 && len(job.threads) < 8; i-- {
			if s := threadRing[i]; s != "" && !seen[s] {
				seen[s] = true
				job.threads = append(job.threads, s)
			}
		}
		// The production truncation ladder (spec 105 FR-008, retry.go): on a
		// reply that fails the parse AND looks truncated, re-submit at double
		// the budget, 800→1600→3200, at most maxTruncationRetries times. Without
		// it a reasoning-emitting local model spends its whole 800-token budget
		// on reasoning and returns empty content — which is what runNarration
		// climbs out of in production, so the harness must climb it too.
		var raw string
		var parsed []narrEntry
		var perr error
		budget := int64(narrMaxTokens)
		for attempt := 0; ; attempt++ {
			var trunc bool
			var err error
			raw, trunc, err = narrateHTTP(endpoint, model, os.Getenv("PW_NARRATE_EFFORT"),
				narrateSystemPrompt(), narrateUserPrompt(job), budget)
			if err != nil {
				t.Fatalf("chapter %q: %v", ch.Label, err)
			}
			parsed, perr = parseNarration(raw)
			if perr == nil || !trunc || attempt >= maxTruncationRetries {
				break
			}
			next := budget * 2
			if next > llm.MaxTokenBudget {
				next = llm.MaxTokenBudget
			}
			if next <= budget {
				break
			}
			fmt.Printf("  (truncated at %d tokens; retry %d at %d — %s)\n", budget, attempt+1, next, ch.Label)
			budget = next
		}
		if perr != nil {
			parseFails = append(parseFails, fmt.Sprintf("%s: %v", ch.Label, perr))
			fmt.Printf("PARSE FAIL %s: %v\nraw: %s\n", ch.Label, perr, truncate(raw, 400))
			continue
		}
		for _, en := range parsed {
			entries = append(entries, outEntry{Chapter: ch.Label, Text: en.Text, Thread: en.Thread})
			threadRing = append(threadRing, en.Thread)
			fmt.Printf("[%s] %s\n    %s\n", en.Thread, ch.Label, en.Text)
		}
	}

	threads := map[string]int{}
	for _, e := range entries {
		threads[e.Thread]++
	}
	keys := make([]string, 0, len(threads))
	for k := range threads {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Printf("\nmodel=%s chapters=%d entries=%d parse_failures=%d\n",
		model, len(after.Chapters), len(entries), len(parseFails))
	for _, k := range keys {
		fmt.Printf("  thread %-32s %d\n", k, threads[k])
	}
	if out := os.Getenv("PW_RENARRATE_OUT"); out != "" {
		writeJSON(t, out, map[string]any{
			"model": model, "chapters": len(after.Chapters),
			"entries": entries, "parse_failures": parseFails,
		})
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// narrateHTTP posts one narrator request and returns the assistant text plus
// whether the reply was truncated, by the same rule retry.go's truncated()
// applies: the provider said "length", or the reply spent the whole budget.
func narrateHTTP(endpoint, model, effort, system, prompt string, maxTokens int64) (string, bool, error) {
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": prompt},
		},
		"stream":     false,
		"max_tokens": maxTokens,
	}
	if effort != "" {
		payload["reasoning_effort"] = effort
	}
	body, _ := json.Marshal(payload)
	resp, err := httpPost(endpoint+"/chat/completions", body)
	if err != nil {
		return "", false, err
	}
	var out struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			CompletionTokens int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(resp, &out); err != nil {
		return "", false, fmt.Errorf("decode: %w (%s)", err, truncate(string(resp), 200))
	}
	if len(out.Choices) == 0 {
		return "", false, fmt.Errorf("no choices: %s", truncate(string(resp), 200))
	}
	trunc := out.Choices[0].FinishReason == "length" || out.Usage.CompletionTokens >= maxTokens
	return out.Choices[0].Message.Content, trunc, nil
}

func httpPost(url string, body []byte) ([]byte, error) {
	resp, err := replayHTTP.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, truncate(string(b), 200))
	}
	return b, nil
}
