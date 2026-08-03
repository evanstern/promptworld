package tui

// Chronicle grammar (TASK-34, extended TASK-60): how one event becomes one
// feed entry. See docs/design/tui/patterns/chronicle-grammar.md and
// specs/018-chronicle-digest/contracts/digest-grammar.md. Kept as pure
// functions over store.Event + the replica's agent-name table so they're
// directly table-driven-testable without a Bubble Tea program — no lipgloss,
// no ANSI (R4): the view layer (views.go) styles the `seg` output.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
)

// segRole tags one styled span of a digest summary (R4); the view layer maps
// roles to style tokens, the pure layer never touches lipgloss.
type segRole int

const (
	segText segRole = iota
	segName
	segSpeech
	segEmphasis
	segLabel
)

// seg is one styled span; concatenating every Text in order is the plain
// summary wrap/truncate operates on (data-model.md "seg").
type seg struct {
	Text string
	Role segRole
}

// plainSegs concatenates a digest's segs into the plain (unstyled) summary.
func plainSegs(segs []seg) string {
	var b strings.Builder
	for _, s := range segs {
		b.WriteString(s.Text)
	}
	return b.String()
}

// eventFamily is the chronicle's namespace-derived grouping (R2): the voice
// (labeled vs. natural phrase) and, later, the family color role both key off
// this rather than the event type string directly.
type eventFamily int

const (
	familyUnknown eventFamily = iota
	familyWorld
	familyClock
	familySim
	familyAgent
	familySocial
	familyGovernance
	familyGru
	familyChronicle
	familyGuardian
	familyDaemon
	familyCog
)

// familyByNamespace maps an event type's namespace (text before the first
// '.') to its family. meeting/norm share one visual role, governance (R2) —
// a meeting is village fabric and a norm is what a meeting produces, and the
// chronicle treats both as one family rather than two.
var familyByNamespace = map[string]eventFamily{
	"world":   familyWorld,
	"run":     familyWorld, // run.ended (spec 044): world-lifecycle voice
	"clock":   familyClock,
	"sim":     familySim,
	"agent":   familyAgent,
	"social":  familySocial,
	"meeting": familyGovernance,
	"norm":    familyGovernance,
	"gru":     familyGru,
	// stranger (spec 077 US2 — the night trickster) rides the gru's own
	// threat-family voice: it is the second nocturnal entity, not a new
	// visual role (FR-016 — no new channels, no new tiers).
	"stranger":  familyGru,
	"chronicle": familyChronicle,
	"morgue":    familyChronicle, // morgue.epilogue (spec 044): narrated prose, chronicle voice
	"daemon":    familyDaemon,
	"cog":       familyCog,
	// curriculum (spec 046 — exercise passes/stage unlocks) rides the
	// guardian's own family tint: the ladder is Guardian's domain (the
	// player's guardian is who watches, teaches, and reports the unlock), not a
	// distinct visual role.
	"curriculum": familyGuardian,
	// guardian: the guardian's own namespace — the report card (spec 063)
	// plus the 13 world-action types renamed from the retired metatron
	// namespace (spec 094; old logs reach the chronicle only after the
	// translating migration, so no metatron.* row can appear here).
	"guardian": familyGuardian,
	// The plan layer (spec 084 — designations and directives): the guardian's
	// own plan-making verbs and their world-answered terminals, guardian
	// family voice like the standing orders they clone.
	"designation": familyGuardian,
	"directive":   familyGuardian,
	// The faith economy (spec 085 — faith movements and the prophecy
	// lifecycle): guardian-domain happenings, guardian family voice like the
	// plan layer above.
	"faith":    familyGuardian,
	"prophecy": familyGuardian,
}

// eventFamilyOf derives a type's family from its namespace prefix (R2). New
// namespaces land in familyUnknown until familyByNamespace grows a row.
func eventFamilyOf(eventType string) eventFamily {
	ns, _, ok := strings.Cut(eventType, ".")
	if !ok {
		return familyUnknown
	}
	if f, ok := familyByNamespace[ns]; ok {
		return f
	}
	return familyUnknown
}

// agentIndexFields are the payload keys whose integer value is a resolvable
// agent index — the generalized form of the existing chronNames mechanism,
// applied to raw event payloads instead of narrated chronicle entries.
var agentIndexFields = map[string]bool{
	"agent": true, "a": true, "b": true,
	"from": true, "to": true,
	"speaker": true, "listener": true, "subject": true,
	// spec 013 (storage): agent.withdrew's Owner and social.chest_taken's
	// Owner/Taker are agent-index fields too (T034 gap — chest_taken was
	// added as chronicle/TUI material per its doc comment, but owner/taker
	// weren't in this table, so the raw feed and inspector rendered them as
	// bare integers instead of names).
	"owner": true, "taker": true,
}

// agentIndexFieldRe matches `"<field>":<int>` for the known agent-index
// fields, in the original payload's field order.
var agentIndexFieldRe = regexp.MustCompile(`"(agent|a|b|from|to|speaker|listener|subject|owner|taker)":(-?\d+)`)

// agentName resolves an index against the roster; out-of-range indices
// render as "#N" rather than panicking or silently dropping the field.
func agentName(names []string, idx int) string {
	if idx >= 0 && idx < len(names) {
		return names[idx]
	}
	return fmt.Sprintf("#%d", idx)
}

// resolvePayloadNames rewrites known agent-index fields in a compact JSON
// payload to quoted names, in place, preserving field order — the feed
// line's payload is a *view* (see chronicle-grammar.md).
func resolvePayloadNames(raw json.RawMessage, names []string) string {
	return agentIndexFieldRe.ReplaceAllStringFunc(string(raw), func(match string) string {
		sub := agentIndexFieldRe.FindStringSubmatch(match)
		field, numStr := sub[1], sub[2]
		n, err := strconv.Atoi(numStr)
		if err != nil || n < 0 || n >= len(names) {
			return match
		}
		nameJSON, _ := json.Marshal(names[n])
		return fmt.Sprintf("%q:%s", field, string(nameJSON))
	})
}

// chronicleLine v2 (data-model.md) is one formatted feed entry — the pure
// content the view layer styles, wraps, and truncates to its panel width.
// Seq leaves the feed line proper (it lives in the detail pane, R7); Tick is
// the new solo-only column (R5); Summary is the digest registry's (or the
// fallback's) styled segments.
type chronicleLine struct {
	Seq  int64
	Tick int64
	Time string
	// Type renders raw in every column (spec 094 US3.3): the guardian
	// rename retired TASK-121's Type-column display alias (spec 052
	// FR-013's displayEventType shim) — persisted types now ARE the display
	// vocabulary, and the chronicle shows guardian.* natively. The dock's
	// short form (shortType, last segment) is unchanged.
	Type    string
	Family  eventFamily
	Summary []seg
}

// formatChronicleLine implements the digest-grammar contract's "Line
// format": a registry hit digests the payload into styled segments (R1);
// a miss or an unmarshal failure (digestFunc's ok=false) falls back to the
// pre-digest compact resolved-name JSON as one segText span — total,
// FR-002/FR-003, never blank, never a panic.
func formatChronicleLine(e store.Event, names []string, sk *skin.Skin) chronicleLine {
	l := chronicleLine{
		Seq:    e.Seq,
		Tick:   e.Tick,
		Time:   clock.FormatTOD(int(clock.SecondOfDay(e.Tick))),
		Type:   e.Type,
		Family: eventFamilyOf(e.Type),
	}
	if fn, ok := digestRegistry[e.Type]; ok {
		if segs, ok := fn(e, names, sk); ok {
			l.Summary = segs
			return l
		}
	}
	l.Summary = []seg{{Text: resolvePayloadNames(e.Payload, names), Role: segText}}
	return l
}

// shortType is the dock column's short-name form (R5): the type's last
// namespace segment, e.g. "social.conversation_turn" → "conversation_turn".
func shortType(t string) string {
	if i := strings.LastIndexByte(t, '.'); i >= 0 {
		return t[i+1:]
	}
	return t
}

// Type column caps (R5, contract §1): solo shows the full type name up to
// 26 runes; dock shows the short name up to 10.
const (
	typeColumnCapSolo = 26
	typeColumnCapDock = 10
)

// chronicleColumns is the per-visible-window column layout (R5) — computed
// fresh over the lines about to render, never stored, so every row in one
// frame lines up without a global fixed budget.
type chronicleColumns struct {
	Dock      bool // dock width: tick column dropped, short type name
	TickWidth int  // widest visible tick, right-aligned
	TypeWidth int  // widest visible type (solo) / short name (dock), capped
}

// computeChronicleColumns derives one window's column widths (R5). Callers
// pass exactly the lines about to render (R8: window first, then format) so
// this never walks more than one frame's worth of entries.
func computeChronicleColumns(lines []chronicleLine, dock bool) chronicleColumns {
	cap := typeColumnCapSolo
	if dock {
		cap = typeColumnCapDock
	}
	cols := chronicleColumns{Dock: dock}
	for _, l := range lines {
		if w := len(strconv.FormatInt(l.Tick, 10)); w > cols.TickWidth {
			cols.TickWidth = w
		}
		// Solo measures the full raw type; the dock short form derives
		// from its last segment.
		t := l.Type
		if dock {
			t = shortType(l.Type)
		}
		w := len([]rune(t))
		if w > cap {
			w = cap
		}
		if w > cols.TypeWidth {
			cols.TypeWidth = w
		}
	}
	return cols
}

// padType left-justifies a line's rendered type to the window's computed
// column width, truncating with "…" past the family cap (R5). Solo shows the
// raw type (spec 094 — no display aliasing); the dock's short form is the
// raw type's last segment.
func padType(l chronicleLine, cols chronicleColumns) string {
	t := l.Type
	cap := typeColumnCapSolo
	if cols.Dock {
		t = shortType(l.Type)
		cap = typeColumnCapDock
	}
	r := []rune(t)
	if len(r) > cap {
		if cap > 1 {
			r = append(append([]rune{}, r[:cap-1]...), '…')
		} else {
			r = r[:cap]
		}
	}
	if len(r) < cols.TypeWidth {
		r = append(r, []rune(strings.Repeat(" ", cols.TypeWidth-len(r)))...)
	}
	return string(r)
}

// chronicleLinePrefix is the tick/time/type portion of a feed line, shared
// by plainChronicleLine (concatenated with the summary) and the styled
// path (styleWrapLine, R4) which needs the prefix and summary kept separate
// so it can tag each rune's source role before wrap/truncate ever runs.
func chronicleLinePrefix(l chronicleLine, cols chronicleColumns) string {
	typ := padType(l, cols)
	if cols.Dock {
		return fmt.Sprintf("%s %s  ", l.Time, typ)
	}
	tick := fmt.Sprintf("%*d", cols.TickWidth, l.Tick)
	return fmt.Sprintf("%s %s  %s  ", tick, l.Time, typ)
}

// plainChronicleLine assembles a formatChronicleLine result plus its
// window's column layout into one plain (unstyled) text line (contract §1):
// solo shows `<TICK> <HH:MM>  <type>  <summary>`, right-aligned tick; dock
// drops the tick column entirely. The result is the shape wrap/truncate
// operate on, without any lipgloss/ANSI concerns.
func plainChronicleLine(l chronicleLine, cols chronicleColumns) string {
	return chronicleLinePrefix(l, cols) + plainSegs(l.Summary)
}

// --- segment-wise styling after wrap (R4, T021) ---
//
// The pure layer never touches lipgloss, but wrapping/truncating must still
// happen on plain text (never mid-ANSI) while leaving enough information
// for the view layer to paint each physical line's characters by their
// source segment. styleRole is the paint-time role — a superset of segRole
// that also covers the prefix (family tint) and default/untagged summary
// text — and styleWrapLine produces styledLine values (plain runes + a
// parallel per-rune role array) that views.go's paintStyledLine renders.

// styleRole is the view layer's per-rune paint tag.
type styleRole int

const (
	styleRoleText   styleRole = iota // default summary prose, no color
	styleRoleFamily                  // the prefix: tick/time/type columns
	styleRoleName
	styleRoleSpeech
	styleRoleEmphasis
)

// styleRoleForSeg maps a digest seg's role to its paint-time role.
func styleRoleForSeg(r segRole) styleRole {
	switch r {
	case segName:
		return styleRoleName
	case segSpeech:
		return styleRoleSpeech
	case segEmphasis:
		return styleRoleEmphasis
	default:
		return styleRoleText
	}
}

// styledLine is one physical output line: plain runes plus a parallel
// per-rune role array (len(Roles) may be shorter than len(Runes) — any
// rune past the end of Roles, e.g. a truncation ellipsis, paints as
// styleRoleText).
type styledLine struct {
	Runes []rune
	Roles []styleRole
}

// flattenLineRoles concatenates the prefix (family-tinted) and the
// digest's summary segs into one plain string with a parallel per-rune
// role array — the source styleWrapLine re-attributes after wrapping.
func flattenLineRoles(prefix string, summary []seg) (string, []styleRole) {
	var b strings.Builder
	var roles []styleRole
	b.WriteString(prefix)
	for range []rune(prefix) {
		roles = append(roles, styleRoleFamily)
	}
	for _, s := range summary {
		if s.Text == "" {
			continue
		}
		b.WriteString(s.Text)
		role := styleRoleForSeg(s.Role)
		for range []rune(s.Text) {
			roles = append(roles, role)
		}
	}
	return b.String(), roles
}

// fieldSpan is one whitespace-delimited token plus its rune offset in the
// source string it was found in — strings.Fields with offsets kept, so
// wrapText's own greedy word-wrap decision can be replayed against
// role-tagged runes instead of the plain string it normally wraps.
type fieldSpan struct {
	Text  string
	Start int // rune offset in the source
}

func fieldsWithOffsets(s string) []fieldSpan {
	var spans []fieldSpan
	r := []rune(s)
	i := 0
	for i < len(r) {
		for i < len(r) && unicode.IsSpace(r[i]) {
			i++
		}
		if i >= len(r) {
			break
		}
		start := i
		for i < len(r) && !unicode.IsSpace(r[i]) {
			i++
		}
		spans = append(spans, fieldSpan{Text: string(r[start:i]), Start: start})
	}
	return spans
}

// styleWrapLine is the segment-wise counterpart of wrapOrTruncatePlain:
// solo (maxWrap<=1) yields exactly one styledLine, truncated the same way
// wrapOrTruncatePlain does (a plain prefix of the source, so per-rune roles
// carry over exactly — there is no whitespace-collapsing in this path).
// Dock (maxWrap>1) replays wrapText's own greedy word-wrap decision
// (byte-length budgeted, matching wrapText precisely so wrap points don't
// shift) over role-tagged fields, so a role boundary that falls mid-word
// (e.g. a resolved name immediately followed by "'s") still styles
// correctly across the rejoin. Either way, wrapping/truncating always
// happens on plain runes first — a physical line is built, THEN painted —
// so truncation can never land mid-ANSI-escape (R4).
func styleWrapLine(prefix string, summary []seg, width, maxWrap, indent int) []styledLine {
	full, roles := flattenLineRoles(prefix, summary)

	if maxWrap == 1 {
		lines := wrapOrTruncatePlain(full, width, 1, 0)
		r := []rune(lines[0])
		lr := make([]styleRole, len(r))
		for i := range r {
			if i < len(roles) {
				lr[i] = roles[i]
			} else {
				lr[i] = styleRoleText // the trailing "…" — no source role
			}
		}
		return []styledLine{{Runes: r, Roles: lr}}
	}

	if width < 10 {
		width = 10
	}
	fullRunes := []rune(full)
	// Fits: emit verbatim, so the column padding survives untouched. Mirrors
	// wrapOrTruncatePlain's fits-branch — see the whitespace-collapse note
	// there for why this is load-bearing rather than an optimization.
	if len(fullRunes) <= width {
		lr := make([]styleRole, len(fullRunes))
		for i := range fullRunes {
			if i < len(roles) {
				lr[i] = roles[i]
			}
		}
		return []styledLine{{Runes: fullRunes, Roles: lr}}
	}
	head, tail, indent := splitWrapHead(full, width, indent)
	headRunes := []rune(head)
	// Only the summary is field-wrapped; `roleAt` shifts each field's offset
	// back over the head so per-rune roles still line up with the full line.
	roleAt := func(i int) styleRole {
		if i += len(headRunes); i >= 0 && i < len(roles) {
			return roles[i]
		}
		return styleRoleText
	}
	// textWidth budgets the wrapped summary; `width` stays the full pane width
	// so the ellipsis trim below still measures whole physical lines.
	textWidth := width - indent
	fields := fieldsWithOffsets(tail)
	type group struct{ fields []fieldSpan }
	var groups []group
	var cur []fieldSpan
	curLen := 0
	for _, f := range fields {
		wl := len(f.Text) // byte length — matches wrapText's own budgeting
		if curLen > 0 && curLen+1+wl > textWidth {
			groups = append(groups, group{cur})
			cur = nil
			curLen = 0
		}
		if curLen > 0 {
			curLen++
		}
		cur = append(cur, f)
		curLen += wl
	}
	if len(cur) > 0 {
		groups = append(groups, group{cur})
	}
	if len(groups) == 0 {
		return []styledLine{{}}
	}

	truncated := maxWrap != wrapUnbounded && len(groups) > maxWrap
	if truncated {
		groups = groups[:maxWrap]
	}
	out := make([]styledLine, len(groups))
	for gi, g := range groups {
		var lineRunes []rune
		var lineRoles []styleRole
		switch {
		case gi == 0:
			// The prefix rides the first line verbatim, roles intact.
			lineRunes = append(lineRunes, headRunes...)
			for k := range headRunes {
				if k < len(roles) {
					lineRoles = append(lineRoles, roles[k])
				} else {
					lineRoles = append(lineRoles, styleRoleText)
				}
			}
		case indent > 0:
			// Continuation lines carry the hanging indent (contract §3). The
			// spaces get the plain text role: they are layout, not content,
			// and painting them with the family tint would smear background
			// color across the left rail on a selected row.
			for k := 0; k < indent; k++ {
				lineRunes = append(lineRunes, ' ')
				lineRoles = append(lineRoles, styleRoleText)
			}
		}
		for fi, f := range g.fields {
			if fi > 0 {
				lineRunes = append(lineRunes, ' ')
				lineRoles = append(lineRoles, styleRoleText)
			}
			fr := []rune(f.Text)
			lineRunes = append(lineRunes, fr...)
			for k := range fr {
				lineRoles = append(lineRoles, roleAt(f.Start+k))
			}
		}
		// Mirrors wrapOrTruncatePlain's dock branch exactly: once the group
		// count overflows maxWrap, the LAST line always gets an ellipsis
		// appended — content is only pre-trimmed to width-1 runes first if
		// it's actually longer than that (T005/T021 plain-equivalence).
		if truncated && gi == len(groups)-1 {
			if len(lineRunes) > width-1 {
				lineRunes = lineRunes[: width-1 : width-1]
				lineRoles = lineRoles[:width-1]
			}
			lineRunes = append(lineRunes, '…')
			lineRoles = append(lineRoles, styleRoleText)
		}
		out[gi] = styledLine{Runes: lineRunes, Roles: lineRoles}
	}
	return out
}

// isAlertType is the digest-grammar contract's high-salience whole-line
// tier (§2/§4) — rendered whole-line in the alert role regardless of
// family. Spec 077 (FR-016) grows the shipped four by exactly one:
// stranger.took joins beside social.chest_taken — theft is theft, and it is
// the ONLY addition (no new tier, no new channel). Spec 083 (FR-010) repeats
// that precedent: sim.neglect_detected joins beside agent.died — a survival
// need neglected to the edge of death is the agent.died class of alarm,
// surfaced while there is still runway to act. Membership only; the
// styleFeedAlert whole-line path is untouched.
func isAlertType(eventType string) bool {
	switch eventType {
	case "agent.died", "gru.attacked", "social.chest_taken", "norm.violated",
		"stranger.took", "sim.neglect_detected":
		return true
	}
	return false
}

// isLabeledVoiceFamily reports whether a family renders as labeled
// key=value fields with the family tint applied to the whole line (type
// column + summary), vs. natural-phrase families where the tint applies to
// the type column only and the summary is styled segment-wise (contract §2).
func isLabeledVoiceFamily(f eventFamily) bool {
	switch f {
	case familyCog, familyClock, familyDaemon:
		return true
	}
	return false
}

// wrapOrTruncatePlain implements chronicle-grammar.md's "Width overflow"
// rule: solo/narrow keep one line per event and truncate with "…"; dock
// wraps to at most maxWrap lines before truncating the last one. Pure and
// ANSI-free so wrap/truncate boundaries are directly testable.
func wrapOrTruncatePlain(s string, width, maxWrap, indent int) []string {
	if width < 10 {
		width = 10
	}
	if maxWrap == 1 {
		r := []rune(s)
		if len(r) <= width {
			return []string{s}
		}
		if width > 1 {
			r = r[:width-1]
		}
		return []string{string(r) + "…"}
	}
	// A line that fits is returned VERBATIM. wrapText budgets on
	// strings.Fields, which collapses runs of whitespace — and the column
	// padding in a feed prefix is exactly such a run. Sending a short row
	// through the wrap path would silently reflow `60 06:01  agent.moved   Ash`
	// into `60 06:01 agent.moved Ash`, destroying the column alignment on the
	// rows that never needed wrapping at all (FR-009).
	if r := []rune(s); len(r) <= width {
		return []string{s}
	}
	head, tail, indent := splitWrapHead(s, width, indent)
	// Only the summary is wrapped; the prefix survives whitespace-intact as
	// the head of the first line, and every continuation line gets the indent
	// in its place.
	chunks := wrapText(tail, width-indent)
	if len(chunks) == 0 {
		return []string{s}
	}
	lines := make([]string, 0, len(chunks))
	lines = append(lines, head+chunks[0])
	for _, c := range chunks[1:] {
		lines = append(lines, strings.Repeat(" ", indent)+c)
	}
	if maxWrap == wrapUnbounded || len(lines) <= maxWrap {
		return lines
	}
	lines = lines[:maxWrap]
	last := []rune(lines[maxWrap-1])
	if len(last) > width-1 {
		last = last[:width-1]
	}
	lines[maxWrap-1] = string(last) + "…"
	return lines
}

// wrapUnbounded is the wrap budget meaning "as many lines as the text needs"
// (spec 115 contract §2). The budget's full domain is: 0 unbounded, 1 truncate
// to a single line with "…", >1 wrap capped at that many lines with "…" on the
// last. It is 0 rather than a negative sentinel so the zero value of an
// unset budget is the permissive one — a caller that forgets to choose gets
// readable output, never silent truncation.
const wrapUnbounded = 0

// minWrapTextWidth is the narrowest text column worth indenting into (spec 115
// FR-006, research R4). Below roughly four or five average words, greedy
// wrapping produces more line breaks than words per line, and a hanging indent
// costs more than the alignment is worth.
const minWrapTextWidth = 24

// resolveWrapIndent applies the all-or-nothing narrow-pane fallback: an indent
// that would leave less than minWrapTextWidth of text collapses to zero rather
// than shrinking. A partial indent aligns to nothing — neither the summary
// column nor the margin — so alignment is exact or absent.
func resolveWrapIndent(width, indent int) int {
	if indent < 0 || width-indent < minWrapTextWidth {
		return 0
	}
	return indent
}

// splitWrapHead divides a flattened feed line into the prefix that must survive
// wrapping verbatim and the summary that may be reflowed, returning the
// resolved indent alongside them. Both wrap renderers call this so they cut at
// precisely the same rune and cannot drift (contract §3, T009).
//
// An indent at or past the end of the line degrades to zero: there is no
// summary to wrap, so there is nothing to align to.
func splitWrapHead(s string, width, indent int) (head, tail string, resolved int) {
	r := []rune(s)
	resolved = resolveWrapIndent(width, indent)
	if resolved >= len(r) {
		resolved = 0
	}
	return string(r[:resolved]), string(r[resolved:]), resolved
}

// --- inspector (paused expand): verbatim stored event, annotated ---

// kv is one ordered top-level key/raw-value pair of a flat JSON object.
type kv struct {
	Key string
	Val json.RawMessage
}

// parseObjectOrdered walks a JSON object's top level preserving field
// order (map[string]any does not) — needed so the inspector reproduces the
// event exactly as persisted, not alphabetized.
func parseObjectOrdered(raw json.RawMessage) ([]kv, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("not an object")
	}
	var pairs []kv
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := keyTok.(string)
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, err
		}
		pairs = append(pairs, kv{Key: key, Val: val})
	}
	return pairs, nil
}

// formatAnnotatedPayload pretty-prints a flat JSON object at the given
// indent, appending "// name" beside any agent-index field's integer value.
// The payload bytes themselves are never rewritten — only a trailing
// comment is added, per chronicle-grammar.md's inspector rules.
func formatAnnotatedPayload(raw json.RawMessage, names []string, indent string) string {
	pairs, err := parseObjectOrdered(raw)
	if err != nil || len(pairs) == 0 {
		var buf bytes.Buffer
		if err := json.Indent(&buf, raw, indent, "  "); err != nil {
			return string(raw)
		}
		return buf.String()
	}
	inner := indent + "  "
	rendered := make([]string, len(pairs))
	for i, p := range pairs {
		rendered[i] = fmt.Sprintf("%s%q: %s", inner, p.Key, string(p.Val))
	}
	for i := range rendered {
		if i < len(rendered)-1 {
			rendered[i] += ","
		}
	}
	for i, p := range pairs {
		if !agentIndexFields[p.Key] {
			continue
		}
		if n, err := strconv.Atoi(strings.TrimSpace(string(p.Val))); err == nil && n >= 0 && n < len(names) {
			rendered[i] += fmt.Sprintf("   // %s", names[n])
		}
	}
	var b strings.Builder
	b.WriteString("{\n")
	b.WriteString(strings.Join(rendered, "\n"))
	b.WriteString("\n" + indent + "}")
	return b.String()
}

// formatInspector is the expanded, paused view of one stored event:
// seq/tick/type verbatim, payload annotated but never rewritten.
func formatInspector(e store.Event, names []string) string {
	var b strings.Builder
	b.WriteString("{\n")
	fmt.Fprintf(&b, "  \"seq\": %d, \"tick\": %d,\n", e.Seq, e.Tick)
	fmt.Fprintf(&b, "  \"type\": %q,\n", e.Type)
	b.WriteString("  \"payload\": ")
	b.WriteString(formatAnnotatedPayload(e.Payload, names, "  "))
	b.WriteString("\n}")
	return b.String()
}
