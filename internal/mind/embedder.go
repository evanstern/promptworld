package mind

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/store"
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
// The situation-vector leg (spec 042 US2, T012) grows here later: the run
// goroutine is the seam — it will absorb events into a replica and render the
// deterministic situation template at each PlannerCadenceTicks bucket edge.

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
)

// EmbedClient is the orchestrator surface the embedder needs (test seam):
// vectors for texts plus the producing model identity, and the best-effort
// Ollama-native warm pin.
type EmbedClient interface {
	Embed(ctx context.Context, texts []string) ([][]float32, string, error)
	WarmEmbedding(ctx context.Context) error
}

// embedJob is one memory awaiting its vector: the emitting event's store seq
// (the companion's target identity) and the exact recorded text.
type embedJob struct {
	agent int
	seq   int64
	text  string
}

type Embedder struct {
	client EmbedClient
	social SocialInjector
	// warn surfaces a debounced operator warning (the daemon wires it to a log
	// line + a durable daemon.llm_warning event, the TASK-91 loud channel).
	warn func(detail string)

	events chan []store.Event
	jobs   chan embedJob
	done   chan struct{}

	// failing tracks the current failure episode (worker goroutine only): the
	// FIRST failure of an episode fires warn, later ones stay quiet, and a
	// success closes the episode — so an outage warns exactly once (D8).
	failing bool
}

// NewEmbedder starts the driver. client is the orchestrator's embedding
// surface, social the loop's InjectSocial door, warn the debounced failure
// channel (nil = log only).
func NewEmbedder(client EmbedClient, social SocialInjector, warn func(detail string)) *Embedder {
	e := &Embedder{
		client: client,
		social: social,
		warn:   warn,
		events: make(chan []store.Event, 256),
		jobs:   make(chan embedJob, 1024),
		done:   make(chan struct{}),
	}
	go e.run()
	go e.worker()
	go e.warmLoop()
	return e
}

// Observe is the loop-notify path: never blocks (drop on overflow — dropped
// memories simply stay vectorless, the no-backfill posture).
func (e *Embedder) Observe(events []store.Event) {
	select {
	case e.events <- events:
	default:
	}
}

func (e *Embedder) Close() { close(e.done) }

// run drains observed batches into per-memory jobs, preserving the committed
// log's order — the jobs channel is FIFO and the worker single-flight, so
// per-agent companions are emission-ordered by construction (contract §2).
// The situation-vector leg (T012) will grow here: absorb into a replica,
// render + enqueue the situation text at each cadence bucket edge.
func (e *Embedder) run() {
	for {
		select {
		case <-e.done:
			return
		case batch := <-e.events:
			for _, ev := range batch {
				if ev.Type != "agent.memory_added" {
					continue
				}
				var p sim.MemoryAddedPayload
				if json.Unmarshal(ev.Payload, &p) != nil || p.Text == "" {
					continue
				}
				select {
				case e.jobs <- embedJob{agent: p.Agent, seq: ev.Seq, text: p.Text}:
				default:
					// Queue full: the memory stays vectorless forever (no
					// backfill). Loud on the log, but not the warning channel —
					// this is driver backpressure, not a transport failure.
					log.Printf("mind: embedder queue full — memory seq %d stays vectorless", ev.Seq)
				}
			}
		}
	}
}

// worker embeds one memory at a time and injects its recorded companion.
func (e *Embedder) worker() {
	for {
		select {
		case <-e.done:
			return
		case j := <-e.jobs:
			e.runJob(j)
		}
	}
}

func (e *Embedder) runJob(j embedJob) {
	// FR-011: the exact recorded text, no normalization beyond the fixed-byte
	// truncation cap.
	text := j.text
	if len(text) > embedTextCapBytes {
		text = text[:embedTextCapBytes]
	}
	ctx, cancel := context.WithTimeout(context.Background(), embedCallTimeout)
	vecs, model, err := e.client.Embed(ctx, []string{text})
	cancel()
	if err != nil || len(vecs) != 1 || len(vecs[0]) == 0 {
		// Loud, non-fatal, no retry into the tick path: the memory stays
		// vectorless forever and the operator hears about the EPISODE once.
		detail := "embedding call returned no vector"
		if err != nil {
			detail = err.Error()
		}
		e.noteFailure(detail)
		return
	}
	e.noteSuccess()
	payload, err := json.Marshal(sim.MemoryEmbeddedPayload{
		Agent: j.agent, MemSeq: j.seq, Vec: vecs[0], Model: model,
	})
	if err != nil {
		log.Printf("mind: embedder payload marshal failed: %v", err)
		return
	}
	if err := e.social.InjectSocial([]store.Event{{Type: "agent.memory_embedded", Payload: payload}}); err != nil {
		// Loop stopped (shutdown window) or door refusal: the memory stays
		// vectorless; the log line is the record.
		log.Printf("mind: embedder companion for seq %d rejected: %v", j.seq, err)
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
