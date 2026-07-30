package sim

// The guardian's own structured memory store (spec 102, D5/FR-002): the
// agentized guardian remembers like a villager — survey results, mission
// outcomes, player exchanges, watch fires enter a per-guardian store as
// recorded guardian.memory_added events, consolidated on the nightly
// boundary through the SAME machinery villagers use (the shared Memory
// model, MemoryHash/MemoryRef identity, SelectMemories windowing, PlanDream
// clustering — single-store privacy holds trivially: this store is the one
// guardian's own). soul.md remains the persona SEED, never the memory log.
//
// Every event type here is ADDITIVE vocabulary (spec 094: no format bump):
// pre-102 logs never carried them, and a pre-102 snapshot's absent fields
// unmarshal to nil/zero — upgrade-free, replay byte-identical.

import (
	"encoding/json"
	"fmt"

	"github.com/evanstern/promptworld/internal/store"
)

// GuardianSeat is the deterministic "agent index" the guardian's shared-
// machinery calls use where villager code passes a villager index — the RNG
// purpose key for SelectMemories' serendipity stream and PlanDream's jitter
// stream. One past the villager range, so no villager stream is ever shared.
const GuardianSeat = AgentCount

// GuardianMemoryCap bounds the guardian store (a villager's store is bounded
// by consolidation fades and dream merges; the guardian gets a hard guard
// too, since its emitters are worker-driven). On overflow the lowest-salience
// (ties: oldest, then first) memory is dropped — deterministic, reducer-side.
const GuardianMemoryCap = 400

// GuardianMemoryTextMax caps one recorded guardian memory's text (bytes) —
// the send_vision rendering cap's neighborhood, enforced at the reducer so a
// runaway emitter can never bloat state.
const GuardianMemoryTextMax = 400

// GuardianMemoryPayload — guardian.memory_added: one memory entering the
// guardian's store. Text and salience only; tick/seq ride the envelope,
// origin is the guardian's own hand by construction.
type GuardianMemoryPayload struct {
	Text     string `json:"text"`
	Salience int    `json:"salience"`
}

// GuardianMemoryEmbeddedPayload — guardian.memory_embedded: the embedder
// driver's recorded vector companion (the agent.memory_embedded shape keyed
// by the emitting event's store seq).
type GuardianMemoryEmbeddedPayload struct {
	MemSeq int64     `json:"mem_seq"`
	Vec    []float32 `json:"vec"`
	Model  string    `json:"model"`
}

// GuardianMemoryPromotedPayload / GuardianMemoryFadedPayload — the nightly
// consolidation's promote/fade outcomes over the guardian store (the
// agent.memory_promoted / agent.memory_faded shapes minus the agent ref).
type GuardianMemoryPromotedPayload struct {
	MemTick  int64  `json:"mem_tick"`
	TextHash string `json:"text_hash"`
	Boost    int    `json:"boost"`
}

type GuardianMemoryFadedPayload struct {
	MemTick  int64  `json:"mem_tick"`
	TextHash string `json:"text_hash"`
}

// GuardianSalienceRevisedPayload / GuardianMemoryMergedPayload — the dream
// pass's recorded habituation/merge outcomes over the guardian store (the
// spec-098 shapes minus the agent ref). Emitter computes, reducer applies.
type GuardianSalienceRevisedPayload struct {
	MemTick  int64  `json:"mem_tick"`
	TextHash string `json:"text_hash"`
	Salience int    `json:"salience"`
	Reason   string `json:"reason,omitempty"`
}

type GuardianMemoryMergedPayload struct {
	Kept     MemoryRef   `json:"kept"`
	Merged   []MemoryRef `json:"merged"`
	Salience int         `json:"salience"`
}

// GuardianConsolidatedPayload — guardian.consolidated: the night's terminal
// marker (the agent.consolidated shape). An accepted night advances the
// buffer high-water mark (UpTo); rejected/skipped nights record the outcome
// and leave the buffer for the next night.
type GuardianConsolidatedPayload struct {
	Night       int64   `json:"night"`
	UpTo        int64   `json:"up_to,omitempty"`
	Outcome     string  `json:"outcome"`
	Reason      string  `json:"reason,omitempty"`
	Promoted    int     `json:"promoted,omitempty"`
	Faded       int     `json:"faded,omitempty"`
	DreamFolded int     `json:"dream_folded,omitempty"`
	DreamKept   int     `json:"dream_kept,omitempty"`
	CostUSD     float64 `json:"cost_usd,omitempty"`
}

// GuardianEpisodicBuffer is the guardian's un-consolidated tail — memories
// accumulated since the last accepted consolidation, in tick order (the
// Agent.EpisodicBuffer shape over the guardian store).
func (s *State) GuardianEpisodicBuffer() []Memory {
	var out []Memory
	for _, m := range s.GuardianMemories {
		if m.Tick > s.GuardianMemUpTo {
			out = append(out, m)
		}
	}
	return out
}

// GuardianDreamEvents renders a guardian dream plan's outcomes as the
// recorded event batch the injection door accepts — sim.DreamEvents' twin
// over the guardian.* vocabulary. Pure serialization, no decisions.
func GuardianDreamEvents(revs []DreamRevision, merges []DreamMerge) []store.Event {
	var out []store.Event
	for _, mg := range merges {
		out = append(out, store.Event{Type: "guardian.memory_merged", Payload: mustPayload(GuardianMemoryMergedPayload{
			Kept: mg.Kept, Merged: mg.Merged, Salience: mg.Salience})})
	}
	for _, rv := range revs {
		out = append(out, store.Event{Type: "guardian.salience_revised", Payload: mustPayload(GuardianSalienceRevisedPayload{
			MemTick: rv.Ref.Tick, TextHash: rv.Ref.Hash,
			Salience: rv.Salience, Reason: DreamReasonHabituation})})
	}
	return out
}

// applyGuardianMemory is the reducer arm family for the guardian memory
// store, dispatched from State.Apply. Total like the villager consolidation
// and dream arms: vanished targets degrade to no-ops, never errors — replay
// applies recorded outcomes and re-derives nothing (spec 092).
func (s *State) applyGuardianMemory(e store.Event) error {
	clampSal := func(v int) int {
		if v < 1 {
			return 1
		}
		if v > MaxSalience {
			return MaxSalience
		}
		return v
	}
	find := func(memTick int64, hash string) int {
		for i := range s.GuardianMemories {
			m := &s.GuardianMemories[i]
			if m.Tick == memTick && MemoryHash(m.Text) == hash {
				return i
			}
		}
		return -1
	}
	switch e.Type {
	case "guardian.memory_added":
		var p GuardianMemoryPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Text == "" {
			return fmt.Errorf("apply %s: empty text", e.Type)
		}
		if len(p.Text) > GuardianMemoryTextMax {
			return fmt.Errorf("apply %s: text over %d bytes", e.Type, GuardianMemoryTextMax)
		}
		s.GuardianMemories = append(s.GuardianMemories, Memory{
			Text: p.Text, Salience: clampSal(p.Salience), Tick: e.Tick,
			Subject: -1, Origin: OriginDigest, Seq: e.Seq,
		})
		// Hard cap: drop the lowest-salience (ties oldest, then first) entry.
		if len(s.GuardianMemories) > GuardianMemoryCap {
			drop := 0
			for i := 1; i < len(s.GuardianMemories); i++ {
				m, d := &s.GuardianMemories[i], &s.GuardianMemories[drop]
				if m.Salience < d.Salience || (m.Salience == d.Salience && m.Tick < d.Tick) {
					drop = i
				}
			}
			s.GuardianMemories = append(s.GuardianMemories[:drop], s.GuardianMemories[drop+1:]...)
		}

	case "guardian.memory_embedded":
		var p GuardianMemoryEmbeddedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		for i := range s.GuardianMemories {
			if s.GuardianMemories[i].Seq == p.MemSeq && p.MemSeq != 0 {
				s.GuardianMemories[i].Vec = p.Vec
				s.GuardianMemories[i].VecModel = p.Model
				break
			}
		} // vanished target: no-op

	case "guardian.memory_promoted":
		var p GuardianMemoryPromotedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if i := find(p.MemTick, p.TextHash); i >= 0 {
			s.GuardianMemories[i].Salience = clampSal(s.GuardianMemories[i].Salience + p.Boost)
		}

	case "guardian.memory_faded":
		var p GuardianMemoryFadedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if i := find(p.MemTick, p.TextHash); i >= 0 {
			s.GuardianMemories = append(s.GuardianMemories[:i], s.GuardianMemories[i+1:]...)
		}

	case "guardian.salience_revised":
		var p GuardianSalienceRevisedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if i := find(p.MemTick, p.TextHash); i >= 0 {
			s.GuardianMemories[i].Salience = clampSal(p.Salience)
		}

	case "guardian.memory_merged":
		var p GuardianMemoryMergedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		for _, ref := range p.Merged {
			if i := find(ref.Tick, ref.Hash); i >= 0 {
				s.GuardianMemories = append(s.GuardianMemories[:i], s.GuardianMemories[i+1:]...)
			}
		}
		if i := find(p.Kept.Tick, p.Kept.Hash); i >= 0 {
			s.GuardianMemories[i].Salience = clampSal(p.Salience)
		}

	case "guardian.consolidated":
		var p GuardianConsolidatedPayload
		if err := json.Unmarshal(e.Payload, &p); err != nil {
			return fmt.Errorf("apply %s: %w", e.Type, err)
		}
		if p.Outcome == ConsolidationAccepted && p.UpTo > s.GuardianMemUpTo {
			s.GuardianMemUpTo = p.UpTo
		}
	}
	return nil
}

// ParseMemLabel maps a consolidation prompt's ordinal memory label
// ("m1".."mN") to its buffer index, or -1 — the shared label vocabulary both
// the villager night (internal/mind) and the guardian night (internal/
// guardian) reference their buffers by (spec 102 SC-004: one parser, two
// drivers; moved here from internal/mind/validate.go's parseMemRef).
func ParseMemLabel(ref string, bufferLen int) int {
	if len(ref) < 2 || (ref[0] != 'm' && ref[0] != 'M') {
		return -1
	}
	n := 0
	for _, c := range ref[1:] {
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
	}
	if n < 1 || n > bufferLen {
		return -1
	}
	return n - 1
}

// FirstJSONObject extracts the first balanced JSON object from model output
// — the consolidation-reply plumbing both nightly drivers share (spec 102
// SC-004; moved here from internal/mind/parse.go's firstJSON).
func FirstJSONObject(text string) (string, error) {
	start := -1
	for i := 0; i < len(text); i++ {
		if text[i] == '{' {
			start = i
			break
		}
	}
	if start < 0 {
		return "", fmt.Errorf("no JSON object in reply")
	}
	depth := 0
	for i := start; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1], nil
			}
		}
	}
	return "", fmt.Errorf("unterminated JSON object")
}

// ParseRoutineLabels maps the model's routine-group labels ("g1".."gN") to
// zero-based indexes into the SENT group list, deduplicating repeats and
// dropping anything unparseable or out of range — mechanical slack absorbed,
// never a rejected night (spec 098). Shared by both consolidation drivers
// (spec 102 SC-004; moved here from internal/mind/validate.go's
// parseRoutineRefs).
func ParseRoutineLabels(refs []string, sent int) []int {
	var out []int
	seen := map[int]bool{}
	for _, r := range refs {
		if len(r) < 2 || (r[0] != 'g' && r[0] != 'G') {
			continue
		}
		n := 0
		ok := true
		for _, c := range r[1:] {
			if c < '0' || c > '9' {
				ok = false
				break
			}
			n = n*10 + int(c-'0')
		}
		if !ok || n < 1 || n > sent || seen[n-1] {
			continue
		}
		seen[n-1] = true
		out = append(out, n-1)
	}
	return out
}
