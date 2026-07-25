package sim

// Agent-authored journal (spec 019, Layer 2 / US3): a per-agent markdown
// notebook that is part of durable world state, mutated ONLY through the two
// journal.* reducer arms (state.go). The one imposed rule — a hard rune budget
// — lives HERE, in Apply, so the gate (the reducer dry-run at the InjectSocial
// door) decides, never handler courtesy (Principle III, research R7). Search is
// deterministic substring matching over the agent's own entries; nothing about
// the journal touches a model (FR-012).

import (
	"fmt"
	"sort"
	"strings"
)

// Journal sizing (research R7 / data-model.md §7). Sim CONSTANTS, not config:
// the budget participates in Apply's accept/reject decision, so it must be
// version-stable like every other deterministic sim constant — a config knob
// that drifted between runs of the same world could let a replay reject an
// event that landed live.
const (
	// journalBudgetRunes is the total per-agent journal budget (clarified
	// 2026-07-22): tight enough that curation pressure arrives within a few
	// in-world days.
	journalBudgetRunes = 4000
	// journalWriteCapRunes is the per-write wire sanity cap (write_journal_entry's
	// Cost.TextCapRunes) — a quarter of the whole budget, identical in kind to
	// say/muse text caps, NOT a usage rule.
	journalWriteCapRunes = 1000
	// journalSearchResultCap bounds how many entries one search returns.
	journalSearchResultCap = 8
)

// JournalBudgetRunes exports the budget for the mind-side gloss + handler
// reason strings so the number the model is told matches the gate exactly.
const JournalBudgetRunes = journalBudgetRunes

// JournalWriteCapRunes exports the per-write wire cap for the tool registry's
// Cost.TextCapRunes (internal/tool is a leaf and mirrors, but the mind side and
// tests read the canonical value here).
const JournalWriteCapRunes = journalWriteCapRunes

// JournalSearchResultCap exports the search result cap for the mind handler.
const JournalSearchResultCap = journalSearchResultCap

// Journal is one agent's self-authored notebook. Entries are ordered (append
// order == event order == tick order); ids are reducer-assigned, monotonic, and
// never reused, giving delete/read a stable address across the journal's whole
// life. Carried as *Journal on Agent (omitempty) so an agent that never
// journals stays byte-identical to a pre-019 snapshot.
type Journal struct {
	NextID  int            `json:"next_id,omitempty"`
	Entries []JournalEntry `json:"entries,omitempty"`
}

// JournalEntry is a single agent-authored write: markdown text attributed to
// the tick it was written.
type JournalEntry struct {
	ID   int    `json:"id"`
	Tick int64  `json:"tick"`
	Text string `json:"text"`
}

// Clone returns a deep copy of the journal (JournalEntry holds no pointers, so
// copying the slice suffices). The mind snapshots a job's journal through this
// so the planner worker's search/read handlers read a race-free copy rather than
// the absorb-owned replica. Nil-safe.
func (j *Journal) Clone() *Journal {
	if j == nil {
		return nil
	}
	return &Journal{NextID: j.NextID, Entries: append([]JournalEntry(nil), j.Entries...)}
}

// JournalUsedRunes is the journal's current size — the sum of every entry's
// rune length, the budget denominator. Nil-safe (an unused journal is 0).
// Exported so the scribe renders "N/budget" without duplicating the rule.
func (j *Journal) JournalUsedRunes() int {
	if j == nil {
		return 0
	}
	n := 0
	for i := range j.Entries {
		n += len([]rune(j.Entries[i].Text))
	}
	return n
}

// appendEntry appends text as a new entry, assigning the next monotonic id — or
// returns an error when the write would exceed the budget. The reducer's
// journal.entry_written arm returns that error; the InjectSocial dry-run turns
// it into a door rejection, so no over-budget event ever lands (SC-005). The
// error text is the reason fed back to the model verbatim (journal-tools.md).
func (j *Journal) appendEntry(tick int64, text string) error {
	need := len([]rune(text))
	if used := j.JournalUsedRunes(); used+need > journalBudgetRunes {
		return fmt.Errorf("journal budget: %d/%d runes, entry needs %d", used, journalBudgetRunes, need)
	}
	j.Entries = append(j.Entries, JournalEntry{ID: j.NextID, Tick: tick, Text: text})
	j.NextID++
	return nil
}

// deleteEntry removes the entry with the given id, preserving the order of the
// survivors and never renumbering — freed runes become immediately available.
// Returns an error when no entry has that id (the door tells the agent nothing
// matched); NextID is untouched so ids are never reused.
func (j *Journal) deleteEntry(id int) error {
	for i := range j.Entries {
		if j.Entries[i].ID == id {
			out := make([]JournalEntry, 0, len(j.Entries)-1)
			out = append(out, j.Entries[:i]...)
			out = append(out, j.Entries[i+1:]...)
			if len(out) == 0 {
				out = nil
			}
			j.Entries = out
			return nil
		}
	}
	return fmt.Errorf("no journal entry #%d", id)
}

// SearchJournal returns up to journalSearchResultCap entries whose text
// contains query (case-insensitive substring), newest-first — the deterministic,
// model-free retrieval the search_journal tool exposes (FR-012, R9). Nil-safe:
// an empty/absent journal returns no matches. Exported so the mind handler reads
// the acting agent's own journal through this single definition. (Search
// results are ephemeral — they feed only the tool-use loop's transcript, never
// replayed state — so plain case folding is honest and sufficient.)
func (j *Journal) SearchJournal(query string) []JournalEntry {
	if j == nil {
		return nil
	}
	q := strings.ToLower(query)
	var out []JournalEntry
	for i := len(j.Entries) - 1; i >= 0; i-- { // newest-first
		if q == "" || strings.Contains(strings.ToLower(j.Entries[i].Text), q) {
			out = append(out, j.Entries[i])
			if len(out) >= journalSearchResultCap {
				break
			}
		}
	}
	return out
}

// Journal excerpt sizing (spec 043 US4, research R5). Constants, not config:
// the context assembler stuffs at most journalExcerptCap term-matched entries,
// each trimmed to journalExcerptRunes runes, as the lowest-priority (first
// dropped) context block. Deterministic and model-free — the same term match
// SearchJournal exposes, no embeddings (embed-at-write is deferred per R5).
const (
	journalExcerptCap   = 2
	journalExcerptRunes = 300
)

// JournalExcerptRunes exports the per-excerpt rune cap for the mind-side block
// renderer + tests, so the number the assembler enforces is defined once here.
const JournalExcerptRunes = journalExcerptRunes

// JournalExcerptCap exports the per-thought excerpt count cap.
const JournalExcerptCap = journalExcerptCap

// JournalExcerpt is one term-matched journal entry trimmed for the decision
// context (spec 043 US4): the entry id + tick and a rune-bounded excerpt of its
// text. The id lets a villager follow up with its own read_journal tool.
type JournalExcerpt struct {
	ID   int
	Tick int64
	Text string
}

// SelectJournalExcerpts returns up to journalExcerptCap journal entries relevant
// to the given situation terms, each trimmed to journalExcerptRunes runes — the
// deterministic, model-free journal retrieval the context assembler stuffs into
// the decision prompt so the villager need not spend its own reasoning turns
// fetching them (spec 043 US4, FR-007, research R5). Each term is matched via
// SearchJournal (case-insensitive substring, newest-first); the union across
// terms is de-duplicated by id and presented newest-first, so an entry matching
// two terms appears once and recency orders the result. Blank terms are
// skipped; a nil/absent journal or no match returns nil. Excerpting truncates
// on a rune boundary (never mid-rune) and never fabricates text: a trimmed entry
// ends with an ellipsis marker so the reader knows it was cut. Determinism: the
// same journal + terms always yields the same excerpts (SearchJournal is
// deterministic; the union sort is stable on tick then id).
func (j *Journal) SelectJournalExcerpts(terms []string) []JournalExcerpt {
	if j == nil {
		return nil
	}
	seen := map[int]bool{}
	var matched []JournalEntry
	for _, term := range terms {
		if strings.TrimSpace(term) == "" {
			continue
		}
		for _, e := range j.SearchJournal(term) {
			if seen[e.ID] {
				continue
			}
			seen[e.ID] = true
			matched = append(matched, e)
		}
	}
	if len(matched) == 0 {
		return nil
	}
	// Union newest-first (tick desc, id desc as a stable tiebreak) then cap.
	sort.SliceStable(matched, func(a, b int) bool {
		if matched[a].Tick != matched[b].Tick {
			return matched[a].Tick > matched[b].Tick
		}
		return matched[a].ID > matched[b].ID
	})
	if len(matched) > journalExcerptCap {
		matched = matched[:journalExcerptCap]
	}
	out := make([]JournalExcerpt, len(matched))
	for i, e := range matched {
		out[i] = JournalExcerpt{ID: e.ID, Tick: e.Tick, Text: excerptRunes(e.Text, journalExcerptRunes)}
	}
	return out
}

// excerptRunes returns s trimmed to at most max runes, truncating on a rune
// boundary (rune-slicing never splits a multi-byte rune) and appending a single
// ellipsis rune to mark the cut — so the total stays ≤ max runes. Text at or
// under the cap is returned unchanged.
func excerptRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 0 {
		return ""
	}
	return string(r[:max-1]) + "…"
}

// FindJournalEntry returns the entry with id, or ok=false when absent — the
// read_journal single-entry path. Nil-safe.
func (j *Journal) FindJournalEntry(id int) (JournalEntry, bool) {
	if j == nil {
		return JournalEntry{}, false
	}
	for i := range j.Entries {
		if j.Entries[i].ID == id {
			return j.Entries[i], true
		}
	}
	return JournalEntry{}, false
}

// JournalEntries returns a copy of the ordered entries (oldest-first) — the
// read_journal whole-journal path. Nil-safe.
func (j *Journal) JournalEntries() []JournalEntry {
	if j == nil {
		return nil
	}
	return append([]JournalEntry(nil), j.Entries...)
}

// --- journal event payloads (spec 019) ---

type (
	// JournalWrittenPayload — journal.entry_written: the agent-authored text of
	// one write. The reducer assigns the id and stamps the tick.
	JournalWrittenPayload struct {
		Agent int    `json:"agent"`
		Text  string `json:"text"`
	}
	// JournalDeletedPayload — journal.entry_deleted: the id to remove.
	JournalDeletedPayload struct {
		Agent int `json:"agent"`
		Entry int `json:"entry"`
	}
)
