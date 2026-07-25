package mind

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

// The async embedding driver (spec 042 US1): a peer of the consolidation
// driver that watches the absorbed event stream for agent.memory_added,
// embeds each memory's EXACT recorded text through the llm orchestrator's
// embedding route, and injects the vector back as a recorded
// agent.memory_embedded companion through the InjectSocial door — the one
// pattern by which model output enters deterministic space. Replay never
// re-embeds: it just re-applies the recorded companions.
//
// The driver NEVER blocks the loop's notify path or the mind's absorb/plan
// goroutines: Observe drops on overflow, and every model call runs on the
// driver's own single-flight worker. A dropped or failed item stays
// vectorless FOREVER (neutral relevance, FR-010) — no backfill pass exists in
// this feature, by design.
//
// Wired by the daemon ONLY when llm.json routes the embedding kind; an absent
// route means a vectorless world and no driver.
//
// The situation-vector leg (spec 042 US2, T012): the run goroutine absorbs
// every committed event into its own replica (the mind pattern) and, at each
// PlannerCadenceTicks bucket edge, renders each live agent's situation string
// from replica state — a deterministic template (renderSituation) — and
// queues it for embedding exactly like a memory. The recorded companion
// (agent.situation_embedded) carries the TEXT alongside the vector, the audit
// surface divergence review reads (research D5).

const (
	// embedCallTimeout bounds one embeddings call. Local 384-dim models embed
	// in the milliseconds class (research D2); a minute of patience covers a
	// cold model load without wedging the worker on a dead endpoint.
	embedCallTimeout = 60 * time.Second
	// embedTextCapBytes is the FIXED-BYTE truncation cap (FR-011): memory text
	// beyond it is cut at exactly this byte boundary before embedding, so the
	// same recorded text always yields one identical vector per pinned model.
	// Situated memory texts run well under 1 KB; 2048 bytes clears every
	// in-tree emitter while staying inside the 384-dim class models' input
	// window. Deliberately bytes, not runes: a mid-rune cut is still
	// deterministic, and determinism is the requirement.
	embedTextCapBytes = 2048
	// embedWarmInterval is the slow re-warm cadence (T008 warm-pin): the
	// Ollama-native keep_alive=-1 pin is refreshed hourly in case the server
	// restarted underneath the daemon. Best-effort — a failed warm never
	// disables the driver.
	embedWarmInterval = time.Hour
	// embedCoalesceCap bounds how many queued jobs one worker pass merges into
	// a single embed call + injection. 32 covers any realistic burst while
	// keeping one batch's texts well inside a local embed server's request
	// appetite.
	embedCoalesceCap = 32
	// embedCoalesceLinger is the short wait the worker adds after the first
	// job so a burst coalesces instead of paying one injection each: every
	// injection is a loop command whose dry-run copies the whole state, and
	// spec 041's mental maps made that copy expensive. 200ms is invisible
	// against the 30-game-min situation cadence and any real memory-to-vector
	// freshness need, while at max-speed fast-forward it collapses a game
	// day's companions into a handful of injections.
	embedCoalesceLinger = 200 * time.Millisecond
)

// EmbedClient is the orchestrator surface the embedder needs (test seam):
// vectors for texts plus the producing model identity, and the best-effort
// Ollama-native warm pin.
type EmbedClient interface {
	Embed(ctx context.Context, texts []string) ([][]float32, string, error)
	WarmEmbedding(ctx context.Context) error
}

// embedJob is one unit of embed work on the FIFO queue. A memory job carries
// the emitting event's store seq (the companion's target identity) and the
// exact recorded text; a situation job carries one WHOLE cadence bucket's
// renders (sits non-nil) — batched deliberately (contract §2 allows it): one
// /embeddings call and ONE InjectSocial per bucket instead of one per agent,
// because every injection dry-runs on a full state copy, which spec 041's
// mental maps made expensive (measured: per-agent injections cost ~+60% wall
// per game-day at max speed; batched, ≈+10%). One tagged type on one FIFO
// channel keeps the single-flight worker's emission-order guarantee.
type embedJob struct {
	agent int
	text  string
	seq   int64       // memory leg: agent.memory_added store seq
	sits  []sitRender // situation leg (T012): the bucket's renders, agent order
	tick  int64       // situation leg: the tick the texts were rendered at
}

// sitRender is one agent's rendered situation string within a bucket batch.
type sitRender struct {
	agent int
	text  string
}

type Embedder struct {
	client EmbedClient
	social SocialInjector
	// warn surfaces a debounced operator warning (the daemon wires it to a log
	// line + a durable daemon.llm_warning event, the TASK-91 loud channel).
	warn func(detail string)

	// replica mirrors committed state for the situation leg (run goroutine
	// only, the mind's replica pattern): situation strings are a deterministic
	// render over it at each cadence bucket edge.
	replica    *sim.State
	lastBucket int64

	events    chan []store.Event
	jobs      chan embedJob
	done      chan struct{}
	closeOnce sync.Once

	// failing tracks the current failure episode (worker goroutine only): the
	// FIRST failure of an episode fires warn, later ones stay quiet, and a
	// success closes the episode — so an outage warns exactly once (D8).
	failing bool
}

// NewEmbedder starts the driver from a state snapshot (the mind.New pattern —
// the replica feeds the situation leg). client is the orchestrator's embedding
// surface, social the loop's InjectSocial door, warn the debounced failure
// channel (nil = log only).
func NewEmbedder(client EmbedClient, social SocialInjector, warn func(detail string), m *worldmap.Map, seed uint64, stateJSON []byte) (*Embedder, error) {
	replica := sim.NewState(seed, m)
	if err := json.Unmarshal(stateJSON, replica); err != nil {
		return nil, err
	}
	// Drop the mental maps (spec 041) from this DRIVER-side replica: the
	// situation template never reads them, and the per-beat perception sweep
	// they drive is the replica's dominant reduce cost (measured: ~+60% wall
	// per game-day at max speed with maps, ~0 without). Every map reducer arm
	// nil-guards by 041's own design ("a map-less agent stays map-less on
	// replay"), so a map-free replica stays correct for everything the
	// renderer reads — positions, needs, intents, structures, night, liveness.
	// The mind's replica keeps its maps; its prompts need them.
	for i := range replica.Agents {
		replica.Agents[i].Map = nil
	}
	e := &Embedder{
		client:     client,
		social:     social,
		warn:       warn,
		replica:    replica,
		lastBucket: replica.Tick / replica.PlannerCadence(),
		events:     make(chan []store.Event, 256),
		jobs:       make(chan embedJob, 1024),
		done:       make(chan struct{}),
	}
	go e.run()
	go e.worker()
	go e.warmLoop()
	return e, nil
}

// Observe is the loop-notify path: never blocks (drop on overflow — dropped
// memories simply stay vectorless, the no-backfill posture).
func (e *Embedder) Observe(events []store.Event) {
	select {
	case e.events <- events:
	default:
	}
}

// Close is idempotent (the orchestrator's closeOnce pattern): the daemon's
// deferred close and an explicit shutdown may both fire.
func (e *Embedder) Close() { e.closeOnce.Do(func() { close(e.done) }) }

// run drains observed batches into embed jobs, preserving the committed
// log's order — the jobs channel is FIFO and the worker single-flight, so
// per-agent companions are emission-ordered by construction (contract §2).
// It also absorbs every event into the replica and fires the situation leg
// (T012) whenever the replica's tick crosses a PlannerCadenceTicks bucket
// edge — one deterministic situation render per live agent per bucket.
func (e *Embedder) run() {
	for {
		select {
		case <-e.done:
			return
		case batch := <-e.events:
			for _, ev := range batch {
				e.replica.Apply(ev)
				if ev.Tick > e.replica.Tick {
					e.replica.Tick = ev.Tick
				}
				if ev.Type != "agent.memory_added" {
					continue
				}
				var p sim.MemoryAddedPayload
				if json.Unmarshal(ev.Payload, &p) != nil || p.Text == "" {
					continue
				}
				e.enqueue(embedJob{agent: p.Agent, seq: ev.Seq, text: p.Text})
			}
			// Situation leg (T012, research D5): at each cadence bucket edge,
			// render each live agent's situation from the replica AS OF the
			// edge and queue the WHOLE bucket as one batched job behind any
			// memories the same batch carried. Gaps are legal — a bucket with
			// no committed events fires late or not at all, and selection
			// falls back to the legacy ranking.
			if bucket := e.replica.Tick / e.replica.PlannerCadence(); bucket > e.lastBucket {
				e.lastBucket = bucket
				var sits []sitRender
				for i := range e.replica.Agents {
					// Dead agents never plan again; asleep agents don't plan
					// until they wake (which re-renders next bucket) — neither
					// needs a fresh query vector, so skip the wasted embeds
					// (behavior-neutral: selection falls back / uses the last
					// recorded vector either way).
					if e.replica.Agents[i].Dead || e.replica.Agents[i].Asleep {
						continue
					}
					sits = append(sits, sitRender{agent: i, text: renderSituation(e.replica, i)})
				}
				if len(sits) > 0 {
					e.enqueue(embedJob{sits: sits, tick: e.replica.Tick})
				}
			}
		}
	}
}

// enqueue is the non-blocking job send: a full queue drops the item — a
// memory stays vectorless forever (no backfill), a situation render waits for
// the next cadence bucket. Loud on the log, but not the warning channel —
// this is driver backpressure, not a transport failure.
func (e *Embedder) enqueue(j embedJob) {
	select {
	case e.jobs <- j:
	default:
		if j.sits != nil {
			log.Printf("mind: embedder queue full — situation bucket at tick %d dropped (next bucket retries)", j.tick)
		} else {
			log.Printf("mind: embedder queue full — memory seq %d stays vectorless", j.seq)
		}
	}
}

// renderSituation is the deterministic situation template (research D5): time
// phase · position + place description · the two worst needs · the active
// intent verb + reason · nearby agent names. A pure render over replica state
// — identical state yields identical bytes, and the string rides the recorded
// companion event as the divergence-audit surface. The composition includes
// the active intent so "relevant to now" means "relevant to what I am doing
// and where I am", not merely "similar to my last memory".
func renderSituation(s *sim.State, idx int) string {
	a := &s.Agents[idx]
	var b strings.Builder
	if s.Night {
		b.WriteString("night")
	} else {
		b.WriteString("daytime")
	}
	if place := sim.PlaceAt(s, a.X, a.Y); place.Desc != "" {
		fmt.Fprintf(&b, " · at %s (%d,%d)", place.Desc, a.X, a.Y)
	} else {
		fmt.Fprintf(&b, " · at (%d,%d)", a.X, a.Y)
	}
	// The two lowest needs, ties broken by the fixed declaration order below —
	// rendered on the prompts' 0-100 scale.
	needs := []struct {
		name string
		v    int
	}{
		{"health", a.Needs.Health}, {"food", a.Needs.Food}, {"rest", a.Needs.Rest},
		{"warmth", a.Needs.Warmth}, {"morale", a.Needs.Morale},
	}
	sort.SliceStable(needs, func(i, j int) bool { return needs[i].v < needs[j].v })
	fmt.Fprintf(&b, " · worst needs: %s %d, %s %d", needs[0].name, needs[0].v/10, needs[1].name, needs[1].v/10)
	if a.Intent != nil {
		fmt.Fprintf(&b, " · doing: %s", a.Intent.Goal)
		if a.Intent.Reason != "" {
			b.WriteString(" — ")
			b.WriteString(a.Intent.Reason)
		}
	} else {
		b.WriteString(" · idle")
	}
	var nearby []string
	for j := range s.Agents {
		o := &s.Agents[j]
		if j == idx || o.Dead {
			continue
		}
		// Same 10-tile Manhattan radius the planner prompt calls "nearby".
		if d := absInt(o.X-a.X) + absInt(o.Y-a.Y); d <= 10 {
			nearby = append(nearby, o.Name)
		}
	}
	if len(nearby) > 0 {
		fmt.Fprintf(&b, " · nearby: %s", strings.Join(nearby, ", "))
	}
	return b.String()
}

// worker embeds jobs single-flight, COALESCING whatever has queued behind the
// first one (bounded): a burst of memories becomes one /embeddings call and
// one atomic injection instead of N — the injection's dry-run state copy is
// the expensive part (see embedJob). Queue order is preserved, so the
// per-agent emission-order guarantee holds; at real (paced) speeds bursts are
// rare and this degrades to one job at a time.
func (e *Embedder) worker() {
	for {
		select {
		case <-e.done:
			return
		case j := <-e.jobs:
			e.runJobs(e.coalesce(j))
		}
	}
}

// coalesce gathers the jobs that arrive within one linger window behind the
// first, bounded by the cap. Returns what it has on shutdown so Close never
// hangs the worker.
func (e *Embedder) coalesce(first embedJob) []embedJob {
	jobs := []embedJob{first}
	linger := time.NewTimer(embedCoalesceLinger)
	defer linger.Stop()
	for len(jobs) < embedCoalesceCap {
		select {
		case next := <-e.jobs:
			jobs = append(jobs, next)
		case <-linger.C:
			return jobs
		case <-e.done:
			return jobs
		}
	}
	return jobs
}

// runJobs embeds a coalesced run of jobs with ONE model call and lands every
// companion in ONE atomic injection, in queue order.
func (e *Embedder) runJobs(jobs []embedJob) {
	// FR-011: the exact recorded text, no normalization beyond the fixed-byte
	// truncation cap.
	capped := func(text string) string {
		if len(text) > embedTextCapBytes {
			return text[:embedTextCapBytes]
		}
		return text
	}
	var texts []string
	for _, j := range jobs {
		if j.sits != nil {
			for _, sr := range j.sits {
				texts = append(texts, capped(sr.text))
			}
		} else {
			texts = append(texts, capped(j.text))
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), embedCallTimeout)
	vecs, model, err := e.client.Embed(ctx, texts)
	cancel()
	if err != nil || len(vecs) != len(texts) {
		// Loud, non-fatal, no retry into the tick path: the memories stay
		// vectorless forever / the situations wait for the next bucket, and
		// the operator hears about the EPISODE once.
		detail := "embedding call returned the wrong vector count"
		if err != nil {
			detail = err.Error()
		}
		e.noteFailure(detail)
		return
	}
	e.noteSuccess()
	var batch []store.Event
	v := 0
	for _, j := range jobs {
		if j.sits != nil {
			// Situation leg (T012): one recorded companion per agent. The
			// companion carries the rendered TEXT alongside the vector — the
			// divergence-audit surface.
			for _, sr := range j.sits {
				payload, merr := json.Marshal(sim.SituationEmbeddedPayload{
					Agent: sr.agent, Tick: j.tick, Text: sr.text, Vec: vecs[v], Model: model,
				})
				if merr != nil {
					log.Printf("mind: embedder payload marshal failed: %v", merr)
					return
				}
				batch = append(batch, store.Event{Type: "agent.situation_embedded", Payload: payload})
				v++
			}
			continue
		}
		payload, merr := json.Marshal(sim.MemoryEmbeddedPayload{
			Agent: j.agent, MemSeq: j.seq, Vec: vecs[v], Model: model,
		})
		if merr != nil {
			log.Printf("mind: embedder payload marshal failed: %v", merr)
			return
		}
		batch = append(batch, store.Event{Type: "agent.memory_embedded", Payload: payload})
		v++
	}
	if err := e.social.InjectSocial(batch); err != nil {
		// Loop stopped (shutdown window) or door refusal: the memories stay
		// vectorless / the situations wait for the next bucket; the log line
		// is the record.
		log.Printf("mind: embedder companion batch rejected: %v", err)
	}
}

// noteFailure opens (or continues) a failure episode: the first failure fires
// the debounced warning, repeats stay quiet (D8). Worker goroutine only.
func (e *Embedder) noteFailure(detail string) {
	if e.failing {
		return
	}
	e.failing = true
	log.Printf("mind: embedder failing — memories land vectorless (%s)", detail)
	if e.warn != nil {
		e.warn(detail)
	}
}

// noteSuccess closes any failure episode, re-arming the warning for the next
// one. Worker goroutine only.
func (e *Embedder) noteSuccess() { e.failing = false }

// warmLoop best-effort pins the embedding model resident (T008 warm-pin): once
// at driver start and on a slow re-warm interval thereafter. Ollama-specific —
// a non-Ollama openai_compat server 404s, logged ONCE at boot and then
// ignored; correctness never depends on the pin, only cold-load latency does.
func (e *Embedder) warmLoop() {
	warm := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), embedCallTimeout)
		defer cancel()
		return e.client.WarmEmbedding(ctx)
	}
	if err := warm(); err != nil {
		log.Printf("mind: embedder warm pin unavailable (non-Ollama endpoint?) — continuing without it: %v", err)
	}
	t := time.NewTicker(embedWarmInterval)
	defer t.Stop()
	for {
		select {
		case <-e.done:
			return
		case <-t.C:
			_ = warm() // silent: the boot line already told the operator
		}
	}
}
