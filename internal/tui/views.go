package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/guardian"
	"github.com/evanstern/promptworld/internal/ipc"
	"github.com/evanstern/promptworld/internal/llm"
	"github.com/evanstern/promptworld/internal/sim"
	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/store"
	"github.com/evanstern/promptworld/internal/worldmap"
)

var (
	styleHeader = lipgloss.NewStyle().Bold(true)
	styleTabOn  = lipgloss.NewStyle().Bold(true).Reverse(true).Padding(0, 1)
	styleTabOff = lipgloss.NewStyle().Faint(true).Padding(0, 1)
	styleDim    = lipgloss.NewStyle().Faint(true)
	styleErr    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	stylePaused = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))
	// styleEnded is the postmortem header token (spec 044 R12) — bold red,
	// the finality register PAUSED's amber deliberately doesn't carry.
	styleEnded  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1"))
	styleNight  = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	styleAgent  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	styleAsleep = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleBox    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)

	// Style tokens (patterns/layout.md "Style tokens") — one named style per
	// role, panels refer to the role never a raw color.
	stylePanelFocus  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("214")).Padding(0, 1) // amber, same hue as PAUSED
	styleTabActive   = lipgloss.NewStyle().Bold(true).Underline(true)
	styleTabInactive = lipgloss.NewStyle().Faint(true)
	styleTabBadge    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	styleFeedType    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleFeedName    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))
	styleFeedSpeech  = lipgloss.NewStyle().Bold(true)
	styleFeedClock   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleFeedSelect  = lipgloss.NewStyle().Reverse(true)

	// Family color roles (contracts/digest-grammar.md §4, TASK-60 Phase 5):
	// applied to the type column for natural-phrase families, and to the
	// whole line for labeled-voice families (clock/cog/daemon — §2). The
	// palette (recorded in patterns/chronicle-grammar.md's Color roles
	// section): clock keeps its existing yellow (contract §4 says so
	// explicitly); the rest are chosen to stay distinguishable from each
	// other and from the name/speech/emphasis/alert roles below.
	styleFamilyWorld      = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))            // blue — foundational/world
	styleFamilySim        = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))            // green — environment (plain, vs. name's bold green)
	styleFamilyAgent      = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))            // cyan — the plurality of events (today's default type color)
	styleFamilySocial     = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))            // magenta — relationships/conversation
	styleFamilyGovernance = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))          // amber — meeting/norm proceedings
	styleFamilyGru        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")) // bold red — predator threat
	styleFamilyChronicle  = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))           // bright magenta — the narrator's voice
	styleFamilyGuardian   = lipgloss.NewStyle().Foreground(lipgloss.Color("99"))           // violet — the guardian, otherworldly
	styleFamilyCog        = lipgloss.NewStyle().Faint(true)                                // muted — telemetry noise
	// daemon has no distinct tint in data-model.md's token list (process
	// bookkeeping, low salience) — familyTint falls back to styleDim.

	styleFeedEmphasis = lipgloss.NewStyle().Underline(true)                            // amounts/kinds/causes/coords
	styleFeedAlert    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")) // whole-line: died/attacked/chest_taken/violated
)

func (m Model) View() string {
	// Default the chronicle mouse hit region to invalid every frame (spec
	// 049); chronicleInspectBody re-validates it below, synchronously in
	// this same call, only when it actually renders this frame.
	m.invalidateChronHit()
	if m.quitting {
		if m.fatalErr != "" {
			return styleErr.Render("detached: "+m.fatalErr) + "\n"
		}
		if m.runEnded() {
			// Spec 056 US1-AS5: quitting an ended world is never framed as
			// "keeps running" — that would be false regardless of whether the
			// quit happened from the postmortem itself or anywhere else on an
			// already-ended world (the same honesty rule spec 044's
			// read-only footer already applies globally, not just on one
			// screen).
			return "detached (the run has ended)\n"
		}
		return "detached (the world keeps running)\n"
	}
	if m.takeover != takeoverNone {
		// The takeover family (spec 056, contracts/takeovers.md §1) wins the
		// body slot over EVERYTHING — the console page, help, solo/narrow
		// panes alike — checked ahead of every other branch below exactly
		// like handleKey's dispatch does: the takeover IS the event, not a
		// mode alongside the others.
		return m.takeoverView()
	}
	if m.console {
		// The guardian console (spec 053, research R1) is a first-class
		// page, checked ahead of the widescreen/narrow fork exactly like the
		// fork itself is checked ahead of solo zoom — it renders full-screen
		// from every mode (home, solo, narrow) without disturbing whichever
		// of those was showing underneath.
		return m.consoleView()
	}
	if isWidescreen(m.width) {
		return m.widescreenView()
	}
	return m.narrowView()
}

// --- narrow fallback (pages/solo-views.md "Narrow fallback") ---
// Today's single-pane UI, unchanged.

func (m Model) narrowView() string {
	var b strings.Builder
	b.WriteString(m.headerView() + "\n")
	b.WriteString(m.tabsView() + "\n\n")
	if m.wantsLessonRow() {
		// Lesson row narrow carry (patterns/layout.md ruling b, spec 055):
		// carried with identical stage defaults. Unlike the guardian strip
		// (confined to the guardian pane, the only place narrow has a
		// minibuffer to sit above), the lesson row is chrome independent of
		// any one pane, so it renders here — visible above whichever pane
		// is active, regardless of which one that is. No fold-under-
		// height-pressure in narrow this slice (the guardian-strip
		// precedent: narrow has no row-budget arithmetic of its own yet).
		b.WriteString(m.lessonRowView(m.width) + "\n\n")
	}
	switch {
	case m.helpOpen:
		// Body replacement (R2), same precedent as the pane switch below —
		// the active pane's body swaps out, chrome stays (FR-006).
		b.WriteString(m.helpNarrowView())
	case m.active == paneMap:
		b.WriteString(m.mapView())
	case m.active == paneChronicle:
		b.WriteString(m.chronicleView())
	case m.active == paneGuardian:
		b.WriteString(m.guardianView())
	case m.active == paneVillagers:
		b.WriteString(m.villagersView())
	case m.active == paneSystems:
		b.WriteString(m.systemsView())
	case m.active == paneExercise:
		b.WriteString(m.exerciseView())
	}
	b.WriteString("\n" + m.footerView())
	return b.String()
}

// helpNarrowView is the narrow-fallback overlay body — same content the
// widescreen composite shows (helpPanelView), sized to what's left of the
// terminal after narrowView's own header/tabs/footer chrome (3 lines) and
// its blank separator (1 line).
func (m Model) helpNarrowView() string {
	cols := clampInt(m.width, 30, 500)
	rows := clampInt(m.height-4, 6, 500)
	return m.helpPanelView(cols, rows)
}

func (m Model) headerView() string {
	name := m.w.Manifest.Name
	if !m.connected {
		msg := fmt.Sprintf("%s — disconnected", name)
		if m.lastErr != "" {
			msg += ": " + m.lastErr
		}
		return styleErr.Render(msg + " (retrying…)")
	}
	if m.status == nil {
		return styleHeader.Render(name)
	}
	c := m.status.Clock
	state := "running"
	switch {
	case m.runEnded():
		// Postmortem posture (spec 044 US1): ENDED replaces running/PAUSED —
		// regardless of the clock state the run end landed under.
		state = styleEnded.Render("ENDED")
	case c.Paused:
		state = stylePaused.Render("PAUSED")
	}
	speedSeg := fmt.Sprintf("speed %s (%.1f t/s)", c.Speed, c.EffectiveRate)
	if c.RequestedSpeed != "" && c.RequestedSpeed != c.Speed {
		speedSeg += " " + governedSpeedSuffix(c.RequestedSpeed, c.GovernorDebt, c.GovernorJobs)
	}
	line := fmt.Sprintf("%s — tick %d · %s · %s · %s",
		name, c.Tick, c.GameTime, state, speedSeg)
	if c.Degraded {
		line += " " + styleErr.Render("[degraded]")
	}
	if name, kind, ok := firstLLMCondition(m.status.LLM); ok {
		line += " " + styleErr.Render(fmt.Sprintf("[llm: %s %s]", name, kind))
	}
	// Suppression badge (spec 037 US1, FR-005): warn-styled, present iff ≥1
	// watched class is suppressed at the current effective speed, following the
	// [llm: …] badge pattern. Class order follows the wire (WatchedClasses).
	if suppressed := suppressedHorizonClasses(m.status.Horizon); len(suppressed) > 0 {
		line += " " + stylePaused.Render(fmt.Sprintf("[suppressed: %s]", strings.Join(suppressed, ", ")))
	}
	// Lesson row badge (panels/lesson-row.md, spec 055, patterns/stage-
	// defaults.md): the row's folded/stage-off form — a quiet, permanent
	// affordance ("there's a lessons system, press ? for it"), independent
	// of whether anything is currently active (data-model.md: "badge ...
	// stage 3+/pre-ladder default" names no active-state qualifier, unlike
	// the "none" state below it, which is explicitly "nothing active").
	if m.lessonBadgeVisible() {
		line += " " + styleDim.Render("[lesson]")
	}
	return styleHeader.Render(line)
}

// currentStage reads the world's curriculum-ladder stage id directly off
// the polled status (ipc/protocol.go WorldStatus.Stage) — "" for a
// pre-status client or a pre-ladder/ungated world, both of which the
// stage-defaults table treats identically (research.md R6).
func (m Model) currentStage() string {
	if m.status == nil {
		return ""
	}
	return m.status.World.Stage
}

// wantsLessonRow reports whether the lesson row would like its 2-row budget
// THIS frame: the stage default is on AND a lesson is actually active
// (data-model.md "none" vs "showing" — stage-eligible with nothing to show
// is still 0 rows, not a blank block). Feeds computeRows (layout.go); its
// own eventual fold outcome is read back via rows.Lesson, not this value
// directly.
func (m Model) wantsLessonRow() bool {
	return lessonRowDefault(m.currentStage()) && m.lessons.ActiveEntry() != nil
}

// lessonBadgeVisible reports whether the header's quiet `[lesson]` badge
// should show: unconditionally at stage 3+/pre-ladder (the row's default
// form there), or — at stage 1-2 with a lesson actually active — only once
// height pressure has folded the row (patterns/layout.md ruling a step 3).
// Narrow never folds under height pressure this slice (no computeRows/fold
// arithmetic of its own yet, the guardian-strip precedent), so in narrow
// this resolves to exactly the stage default: carried (no badge) at stage
// 1-2, or the unconditional stage-3+/pre-ladder badge.
func (m Model) lessonBadgeVisible() bool {
	stage := m.currentStage()
	if !lessonRowDefault(stage) {
		return true
	}
	if m.lessons.ActiveEntry() == nil {
		return false
	}
	if !isWidescreen(m.width) {
		return false
	}
	return computeRows(m.height, true).Lesson == 0
}

// lessonRowView renders the active lesson's two borderless lines
// (panels/lesson-row.md mockup): line 1 the lesson's plain-language
// sentence, line 2 the UI-pointer phrase plus the pull-path suffix every
// lesson string carries (lessonPullSuffix, lessons.go) — appended here, the
// suffix's one rendering site. Every string passes through
// lessonSkinResolve (FR-008/SC-005: no raw "{{" literal may ever reach the
// screen). Blank when nothing is active — the caller (widescreenView/
// narrowView) already gates on rows.Lesson/wantsLessonRow, so this is only
// ever called when there IS an active lesson, but it degrades to "" rather
// than panicking regardless (the guardian-strip pre-status precedent).
func (m Model) lessonRowView(width int) string {
	entry := m.lessons.ActiveEntry()
	if entry == nil {
		return ""
	}
	line1 := clipLine(lessonSkinResolve(entry.Text, m.sk()), width)
	line2 := clipLine(lessonSkinResolve(entry.Pointer, m.sk())+"  "+lessonPullSuffix, width)
	return line1 + "\n" + line2
}

// suppressedHorizonClasses returns the names of the horizon entries the router
// is currently suppressing, in wire order — the header badge's contents. nil
// (badge absent) when nothing is suppressed or the world carries no horizon.
func suppressedHorizonClasses(h []ipc.HorizonClass) []string {
	var out []string
	for _, e := range h {
		if e.Suppressed {
			out = append(out, e.Class)
		}
	}
	return out
}

// firstLLMCondition reports the first (wire-order, name-sorted) provider
// carrying an active health condition — the header badge only needs one
// name to point the operator at `promptworld status` for the rest
// (contracts/provider-conditions.md "Human surfaces": "[llm: <provider>
// <kind>]", pattern of the existing [degraded] badge). ok is false when
// there is no LLM status or every provider is healthy.
func firstLLMCondition(l *llm.Status) (name, kind string, ok bool) {
	if l == nil {
		return "", "", false
	}
	for _, p := range l.Providers {
		if p.Condition != "" {
			return p.Name, p.Condition, true
		}
	}
	return "", "", false
}

// governedSpeedSuffix renders the plain-language annotation the header's
// speed segment gains while the governor holds effective speed below the
// player's requested ceiling (spec 028 FR-015, US4-AC1) — e.g. "asked 32x —
// 3 minds in flight, debt 140%". Ungoverned worlds never call this (the
// caller gates on RequestedSpeed being set and differing from Speed), so
// their header renders byte-identically to pre-028. debt is expressed via
// the shared debtPercent (digest.go) — a whole-percent fraction of
// cognition.ShedThreshold, rounded to the nearest percent.
func governedSpeedSuffix(requested string, debt float64, jobs int) string {
	mind := "minds"
	if jobs == 1 {
		mind = "mind"
	}
	return fmt.Sprintf("asked %s — %d %s in flight, debt %d%%", requested, jobs, mind, debtPercent(debt))
}

func (m Model) tabsView() string {
	var tabs []string
	for i := pane(0); i < paneCount; i++ {
		if i == paneExercise && m.exerciseID() == "" {
			continue // spec 054: no exercise tab exists on ambient worlds
		}
		label := fmt.Sprintf("%s %s", paneKey[i], m.paneName(i))
		if i == m.active {
			tabs = append(tabs, styleTabOn.Render(label))
		} else {
			tabs = append(tabs, styleTabOff.Render(label))
		}
	}
	return strings.Join(tabs, " ")
}

// footerView is the per-mode hint line; every branch advertises "? help"
// (FR-011, R5) so the overlay is discoverable from anywhere — except while
// the overlay itself owns the keyboard, when the overlay's own hint
// (helpFooterHint, help.go) replaces it entirely (FR-012: no reference to
// the mode beneath while help is open).
func (m Model) footerView() string {
	if m.takeover == takeoverCeremony {
		// D13's blessed-stopping-point framing (contracts/takeovers.md §3) —
		// checked ahead of helpOpen: a takeover keeps the body slot and its
		// own footer regardless of any stale helpOpen state underneath.
		return styleDim.Render("esc dismiss · q — the world keeps running")
	}
	if m.takeover == takeoverPostmortem {
		// No keeps-running framing here (spec.md US1-AS5) — the run really
		// has ended.
		return styleDim.Render("esc dismiss · q quit")
	}
	if m.helpOpen {
		return styleDim.Render(m.helpFooterHint())
	}
	// Postmortem posture (spec 044): the clock keys (space, [, ]) are inert
	// on an ended world, so the footer's pause/resume hint gives way to the
	// run-ended hint — every other affordance stands.
	pause, resume := "space pause", "space resume"
	if m.runEnded() {
		pause = "run ended (read-only)"
		resume = pause
	}
	switch {
	case m.mbFocused:
		// No "? help" here: focused, '?' types into the buffer like any
		// other character (FR-001) — advertising it as a help trigger in
		// this one mode would be actively wrong. Minibuffer help is still
		// reachable, just from any other mode's overlay (n/p paging).
		return styleDim.Render("esc release · ⏎ send · ↑↓ history")
	case m.console:
		// The console's own footer (spec 053 contract §2.7) — checked ahead
		// of inspect/villagers below for the same reason handleKey's
		// dispatch does: dockTab/active persist unchanged underneath the
		// console and would otherwise mis-route through those hints for a
		// screen that isn't actually visible.
		return styleDim.Render("G back · esc back · m ask · " + pause + " · q quit · ? help")
	case m.inspecting():
		return styleDim.Render("j/k select · J/K scroll detail · " + resume + " · m ask · ? help")
	case m.villagersVisible() && m.villDetail && m.villDecisions:
		return styleDim.Render("j/k scroll · esc back · " + pause + " · q quit · ? help")
	case m.villagersVisible() && m.villDetail:
		return styleDim.Render("d decisions · esc back · " + pause + " · q quit · ? help")
	case m.villagersVisible():
		return styleDim.Render("j/k select · ⏎ inspect · " + pause + " · q quit · ? help")
	case isWidescreen(m.width) && m.solo:
		return styleDim.Render(fmt.Sprintf("%s back to map · %s · q quit · ? help", dockTabKey[m.dockTab], resume))
	case isWidescreen(m.width):
		// Spec 054: scenario worlds advertise the exercise tab's key; ambient
		// worlds keep the pre-054 hint byte-identical (no tab exists).
		tabsHint := "2 chronicle 3 " + m.sk().TabLabel() + " 4 villagers 5 systems"
		if m.exerciseID() != "" {
			tabsHint += " 6 exercise"
		}
		return styleDim.Render(tabsHint + " (again: solo) · G console · m ask · " + pause + " · q quit · ? help")
	default:
		panesHint := "1-5 panes"
		if m.exerciseID() != "" {
			panesHint = "1-5,6 panes"
		}
		return styleDim.Render(panesHint + " · G console · " + pause + " · q quit · ? help")
	}
}

// --- widescreen composite (pages/home.md, pages/solo-views.md "Solo zoom") ---

func (m Model) widescreenView() string {
	cols := computeColumns(m.width)
	rows := computeRows(m.height, m.wantsLessonRow())

	var body string
	switch {
	case m.helpOpen:
		// Body replacement (R2), same slot solo zoom uses below — no
		// z-compositing exists in this client, so help renders *instead of*
		// the map/dock body, chrome (header/minibuffer/footer) unchanged.
		body = m.helpPanelView(cols.MapCols+cols.Gutter+cols.DockCols, rows.Body)
	case m.solo:
		body = m.soloPanelView(cols.MapCols+cols.Gutter+cols.DockCols, rows.Body)
	default:
		mapPanel := m.mapPanelView(cols.MapCols, rows.Body)
		dockPanel := m.dockPanelView(cols.DockCols, rows.Body)
		body = lipgloss.JoinHorizontal(lipgloss.Top, mapPanel, strings.Repeat(" ", cols.Gutter), dockPanel)
	}

	var b strings.Builder
	b.WriteString(m.headerView() + "\n")
	b.WriteString(body + "\n")
	if rows.Lesson > 0 {
		// Lesson row (panels/lesson-row.md, spec 055): two borderless lines
		// directly above the guardian strip — rows.Body already accounts
		// for whether this row is visible (computeRows), so the composite's
		// total height stays exact whether folded or not (B1).
		b.WriteString(m.lessonRowView(m.width) + "\n")
	}
	if rows.Strip > 0 {
		// Guardian strip (panels/guardian-strip.md, spec 050): one
		// borderless row directly above the minibuffer, wired to the row
		// budget above — rows.Body already accounts for whether this row
		// is visible, so the composite's total height stays exact whether
		// folded or not (B1's exact-height invariant, unchanged by this
		// feature).
		b.WriteString(m.guardianStripView(m.width) + "\n")
	}
	b.WriteString(m.minibufferView(m.width) + "\n")
	b.WriteString(m.footerView())
	return b.String()
}

// --- takeover family (spec 056, contracts/takeovers.md; overlays/ceremony.md,
// overlays/postmortem.md) ---
//
// The stage-unlock ceremony and the run-end postmortem: a first-class page
// like the guardian console (checked ahead of it in View()), rendering
// full-screen regardless of layout (patterns/layout.md ruling b — takeovers
// are layout-independent) with the header/footer chrome every body-
// replacement surface in this corpus keeps (overlays/help.md's
// helpPanelView precedent). No minibuffer/dock/tabs — neither takeover
// accepts text input or has a second pane to show.

// takeoverView dispatches to whichever kind currently owns the body slot —
// called only when m.takeover != takeoverNone (View()'s own guard).
func (m Model) takeoverView() string {
	width := m.width
	if width < 40 {
		width = 40
	}
	bodyRows := m.height - 2 // header(1) + footer(1); the box below owns its own -2 border budget
	if bodyRows < 5 {
		bodyRows = 5
	}
	var body string
	switch m.takeover {
	case takeoverCeremony:
		body = m.ceremonyView(width, bodyRows)
	case takeoverPostmortem:
		body = m.postmortemView(width, bodyRows)
	}
	return m.headerView() + "\n" + body + "\n" + m.footerView()
}

// fitTakeoverLines truncates content to at most maxLines rows — the
// takeover family has no scroll key (contracts/takeovers.md §1 lists none),
// so overflow (an implausibly tall death ledger on a very short terminal)
// sheds its tail with a count rather than growing the box past its
// Height() budget — the same "never overflow the panel" discipline
// paginateHelpContent (help.go) enforces for the help overlay (B1/B5).
func fitTakeoverLines(lines []string, maxLines int) []string {
	if maxLines < 1 {
		maxLines = 1
	}
	if len(lines) <= maxLines {
		return lines
	}
	out := append([]string{}, lines[:maxLines-1]...)
	out = append(out, styleDim.Render(fmt.Sprintf("… (+%d more)", len(lines)-(maxLines-1))))
	return out
}

// postmortemView is overlays/postmortem.md's takeover: the narrated run-end
// line, the report card (scored runs only — FR-001/FR-018), then the
// morgue's no-blame evidence rows (always).
func (m Model) postmortemView(cols, rows int) string {
	if rows < 5 {
		rows = 5
	}
	inner := cols - 4
	if inner < 10 {
		inner = 10
	}
	var lines []string
	lines = append(lines, wrapText(m.postmortemRunEndLine(), inner)...)
	if card := m.postmortemReportCard(inner); card != "" {
		lines = append(lines, "", card)
	}
	lines = append(lines, "", styleHeader.Render("morgue — no-blame evidence"))
	deaths := m.morgueRows()
	if len(deaths) == 0 {
		lines = append(lines, styleDim.Render("no deaths recorded"))
	}
	for _, r := range deaths {
		lines = append(lines, clipLine(fmt.Sprintf("%s · day %d · %s · charter observed: %s", r.Name, r.Day, r.Cause, r.Charter), inner))
	}
	lines = fitTakeoverLines(lines, rows-4) // title(1) + blank(1) + the box's own -2 border budget
	content := styleHeader.Render("THE RUN HAS ENDED") + "\n\n" + strings.Join(lines, "\n")
	return styleBox.Width(cols - 2).Height(rows - 2).Render(clipContent(content, cols-2))
}

// postmortemRunEndLine is the narrated run-end line (contracts/takeovers.md
// §2 row 1), computed directly from the recorded run.ended payload
// (RunEnd.FinalCause) rather than waiting on the async LLM-narrated
// chronicle chapter — FR-001 requires the takeover within one frame, and
// FR-006 requires model-free content. The wording mirrors internal/mind/
// narrate.go's chronicleNote run.ended line verbatim (postmortem.md: "shares
// wording with [[chronicle]]'s existing digest/narrated line").
func (m Model) postmortemRunEndLine() string {
	if m.replica == nil || m.replica.RunEnd == nil {
		return "the run has ended"
	}
	return fmt.Sprintf("The last villager died of %s. The village stands empty — the run has ended.", m.replica.RunEnd.FinalCause)
}

// morgueRow is one death's evidence row (contracts/takeovers.md §2; research
// R3): name, day, cause from the replica's run-end death ledger, plus the
// closest charter observation at or before that death, scanned from the
// client's own chronicle ring — never a file read (the durable render is
// scribe's morgue.md; this is a live projection over what the client
// already holds). Charter is "unknown" when the ring has rotated past the
// relevant metatron.charter_observed event on a very long run (R3's honesty
// rule), never guessed.
type morgueRow struct {
	Name, Cause, Charter string
	Day                  int64
}

// morgueRows derives the postmortem's evidence rows from the replica's
// run-end death ledger (spec 044's State.RunEnd) — empty (not nil-panicking)
// before run.ended has landed or on a snapshot pre-dating spec 044.
func (m Model) morgueRows() []morgueRow {
	if m.replica == nil || m.replica.RunEnd == nil {
		return nil
	}
	rows := make([]morgueRow, 0, len(m.replica.RunEnd.Deaths))
	for _, d := range m.replica.RunEnd.Deaths {
		name := "someone"
		if d.Agent >= 0 && d.Agent < len(m.replica.Agents) {
			name = m.replica.Agents[d.Agent].Name
		}
		day, _, _, _ := clock.GameTime(d.Tick)
		rows = append(rows, morgueRow{Name: name, Day: day, Cause: d.Cause, Charter: m.closestCharterObservation(d.Tick)})
	}
	return rows
}

// closestCharterObservation scans the client's chronicle ring (m.events) for
// the most recent metatron.charter_observed event at or before deathTick —
// the same alignment rule morgue.md's render uses (internal/scribe/morgue.go
// captureEpitaph), but over the bounded client-side ring instead of a full
// replay: "unknown" is the honest answer once the ring has rotated past it
// (research R3), never a guess.
func (m Model) closestCharterObservation(deathTick int64) string {
	var best *sim.CharterObservedPayload
	var bestTick int64 = -1
	for _, e := range m.events {
		if e.Type != "metatron.charter_observed" || e.Tick > deathTick || e.Tick < bestTick {
			continue
		}
		var p sim.CharterObservedPayload
		if json.Unmarshal(e.Payload, &p) == nil {
			pc := p
			best = &pc
			bestTick = e.Tick
		}
	}
	if best == nil {
		return "unknown"
	}
	if best.Default {
		return "default"
	}
	return best.Fingerprint
}

// scenarioExercise resolves the attached world's seeded exercise, if any
// (panels/exercise.md "Stage defaults": "present whenever the attached world
// carries a Manifest.Scenario block" — a manifest fact the client already
// holds, zero new IPC fields, FR-006). Absent block, or an id this build's
// catalog doesn't know, is the honest ambient fallback — never a fabricated
// card (spec.md Edge Cases: "the report card never renders on missing
// data").
func (m Model) scenarioExercise() (sim.ExerciseDefinition, bool) {
	if m.w == nil || m.w.Manifest.Scenario == nil || m.w.Manifest.Scenario.Exercise == "" {
		return sim.ExerciseDefinition{}, false
	}
	for _, def := range sim.ScenarioExercises {
		if def.ID == m.w.Manifest.Scenario.Exercise {
			return def, true
		}
	}
	return sim.ExerciseDefinition{}, false
}

// postmortemReportCard renders the scored-run report card (concluded
// markers), or "" on an ambient world / an unrecognized exercise id (FR-001,
// FR-018 — the ambient/scored boundary; SC-002).
func (m Model) postmortemReportCard(width int) string {
	def, ok := m.scenarioExercise()
	if !ok {
		return ""
	}
	facts := reportCardFactsFromEvents(def, m.events)
	return reportCardView(def.ID, facts, reportCardConcluded, width)
}

// ceremonyView is overlays/ceremony.md's takeover: the D6 authorship-voice
// narrated chapter plus the report card (the instrument, authoritative —
// FR-019: both, always). The stage identity is always
// replica.StagesUnlocked's last entry while m.takeover == takeoverCeremony —
// this is exactly the event that appended it (applyEvent, tui.go), so no
// second Model field is needed to remember which stage is open.
func (m Model) ceremonyView(cols, rows int) string {
	if rows < 5 {
		rows = 5
	}
	if m.replica == nil || len(m.replica.StagesUnlocked) == 0 {
		return styleBox.Width(cols - 2).Height(rows - 2).Render("")
	}
	stage := m.replica.StagesUnlocked[len(m.replica.StagesUnlocked)-1]
	inner := cols - 4
	if inner < 10 {
		inner = 10
	}
	var lines []string
	lines = append(lines, wrapText(m.sk().CeremonyChapter(stage), inner)...)
	if card := m.ceremonyReportCardFor(stage, inner); card != "" {
		lines = append(lines, "", card)
	}
	lines = fitTakeoverLines(lines, rows-4)
	title := styleHeader.Render(strings.ToUpper(m.sk().StageName(stage)) + " — unlocked")
	content := title + "\n\n" + strings.Join(lines, "\n")
	return styleBox.Width(cols - 2).Height(rows - 2).Render(clipContent(content, cols-2))
}

// provingPass finds the recorded CurriculumPass that earned an unlocked
// stage, re-applying the SAME documented gate conjuncts sim.EvaluateUnlock
// uses (internal/sim/curriculum.go's doc comment) — EvaluateUnlock itself
// refuses to re-evaluate a stage already latched into StagesUnlocked (a
// one-shot gate, by design), so this is the read-only twin for REPLAY
// purposes only (research R5's "stored, never regenerated" ceremony
// content): both the live-open ceremony and the `?` overlay's replay
// section (help.go) call this, so they can never show different evidence
// for the same stage. Honest-false when the qualifying pass has aged out of
// the bounded 32-entry CurriculumPasses retention (data-model.md "unknown
// honestly" — the ceremony/replay callers fall back to a generic
// events-ring derivation, reportCardFactsFromEvents).
func provingPass(replica *sim.State, stage string) (sim.CurriculumPass, bool) {
	if replica == nil {
		return sim.CurriculumPass{}, false
	}
	var from string
	switch stage {
	case "stage-2":
		from = "stage-1"
	case "stage-3":
		from = "stage-2"
	case "stage-4":
		from = "stage-3"
	default:
		return sim.CurriculumPass{}, false
	}
	for _, p := range replica.CurriculumPasses {
		if p.Stage != from {
			continue
		}
		switch from {
		case "stage-1":
			return p, true // stage-1 -> stage-2: any stage-1 pass qualifies
		case "stage-2":
			for _, ev := range p.Evidence {
				if ev.Type == "metatron.charter_observed" && ev.Custom {
					return p, true
				}
			}
		case "stage-3":
			for _, ev := range p.Evidence {
				if ev.Custom {
					return p, true
				}
			}
		}
	}
	return sim.CurriculumPass{}, false
}

// ceremonyReportCardFor renders the (concluded — the exercise already
// passed) report card for the exercise that earned `stage`, preferring the
// recorded pass's own Evidence (authoritative — FR-019's "instrument"
// framing) and falling back to a generic events-ring derivation only when
// the qualifying pass has aged out of retention. "" when the proving
// exercise can't be identified at all (an uncataloged exercise id — an
// honest, defensive absence; production emission, TASK-119's, always names
// a cataloged id).
func (m Model) ceremonyReportCardFor(stage string, width int) string {
	pass, found := provingPass(m.replica, stage)
	exerciseID := pass.Exercise
	var def sim.ExerciseDefinition
	var ok bool
	for _, d := range sim.ScenarioExercises {
		if d.ID == exerciseID {
			def, ok = d, true
			break
		}
	}
	if !ok {
		return ""
	}
	var facts []reportCardFact
	if found {
		facts = reportCardFactsFromEvidence(def, pass.Evidence)
	} else {
		facts = reportCardFactsFromEvents(def, m.events)
	}
	return reportCardView(def.ID, facts, reportCardConcluded, width)
}

// --- the shared report-card renderer (D5, contracts/takeovers.md §4) ---
//
// One rubric-checklist implementation, three sites: the postmortem
// (concluded — a concluded run), the ceremony (concluded — "the instrument
// authoritative"), and, via the consoleCard wrapper below, the guardian
// console's card seam (spec 053; production wiring there is TASK-115's).

// reportCardMode selects the marker vocabulary (contracts/takeovers.md §4) —
// the SAME rows and backing references either way; only the glyph for an
// unmet term changes.
type reportCardMode int

const (
	reportCardConcluded reportCardMode = iota // met/missed — a run/exercise that's over
	reportCardLive                            // met/pending — still running (TASK-115's future call site)
)

// reportCardFact is one already-resolved rubric row: the plain-language
// term, whether it's met, and the backing event reference shown alongside
// it (postmortem.md "Report card": "the backing event reference"). Facts
// are computed once at open time by the small helpers below — the renderer
// itself does no event derivation, so its output is identical at every call
// site by construction (D5/SC-005).
type reportCardFact struct {
	Term    string
	Met     bool
	Backing string
}

// reportCardView is the D5 shared renderer (contracts/takeovers.md §4).
func reportCardView(title string, facts []reportCardFact, mode reportCardMode, width int) string {
	if width < 20 {
		width = 20
	}
	inner := width - 4
	if inner < 10 {
		inner = 10
	}
	unmetGlyph := "✗" // ✗ — concluded
	if mode == reportCardLive {
		unmetGlyph = "…" // pending — still running
	}
	lines := make([]string, 0, len(facts))
	for _, f := range facts {
		glyph := "✓" // ✓ met, either mode
		if !f.Met {
			glyph = unmetGlyph
		}
		lines = append(lines, clipLine(fmt.Sprintf("%s %s (%s)", glyph, f.Term, f.Backing), inner))
	}
	content := styleHeader.Render("report card · "+title) + "\n" + strings.Join(lines, "\n")
	return styleBox.Width(width - 2).Render(clipContent(content, width-2))
}

// humanizeEventType is a mechanical plain-language gloss of a cataloged
// event type ("agent.died" -> "agent died") — a deliberately generic
// placeholder, not hand-authored per-term copy (the mockups' "village
// survives to dawn" phrasing is illustrative content this feature does not
// have an authoritative source for; sim.ExerciseDefinition carries no
// parallel plain-language term table). TASK-119's scenario rubric machinery
// is the eventual owner of curated per-term copy and combining semantics
// (some terms want zero occurrences, some want an OR of several —
// curriculum.go's own doc comments); this renderer stays correct and stable
// regardless of how that content evolves.
func humanizeEventType(t string) string {
	t = strings.ReplaceAll(t, ".", " ")
	t = strings.ReplaceAll(t, "_", " ")
	return t
}

// reportCardFactsFromCounts builds one fact per definition rubric term from
// an already-tallied event-type count map — Met is true the first time the
// term's event type is present at all (a deliberately generic
// presence-based placeholder; see humanizeEventType's doc comment for the
// scope note this shares).
func reportCardFactsFromCounts(def sim.ExerciseDefinition, counts map[string]int) []reportCardFact {
	facts := make([]reportCardFact, len(def.RubricTerms))
	for i, term := range def.RubricTerms {
		n := counts[term]
		facts[i] = reportCardFact{Term: humanizeEventType(term), Met: n > 0, Backing: fmt.Sprintf("%s: %d", term, n)}
	}
	return facts
}

// reportCardFactsFromEvents derives facts from the client's own chronicle
// ring (the postmortem's fallback source, and the ceremony's when the
// proving pass has aged out of retention) — bounded by chronicleCap, honest
// about what it can see, never a file read (FR-006).
func reportCardFactsFromEvents(def sim.ExerciseDefinition, events []store.Event) []reportCardFact {
	counts := make(map[string]int, len(def.RubricTerms))
	for _, e := range events {
		counts[e.Type]++
	}
	return reportCardFactsFromCounts(def, counts)
}

// reportCardFactsFromEvidence derives facts from a recorded pass's own
// Evidence list — the ceremony's preferred, authoritative source (FR-019's
// "instrument"): exactly the satisfying events the unlock derivation itself
// read (sim.EvidenceRef, spec 046).
func reportCardFactsFromEvidence(def sim.ExerciseDefinition, evidence []sim.EvidenceRef) []reportCardFact {
	counts := make(map[string]int, len(def.RubricTerms))
	for _, ev := range evidence {
		counts[ev.Type]++
	}
	return reportCardFactsFromCounts(def, counts)
}

// reportCard wraps a resolved report card as a consoleCard seam element
// (spec 053 contract §2.3; data-model.md consoleCard) — proves the shared
// renderer composes into the console's card slot unmodified (spec.md US3
// AS2, SC-005). Production wiring into Model.consoleCards is TASK-115's;
// this feature ships only the wrapper and its seam-composition test.
type reportCard struct {
	title string
	facts []reportCardFact
	mode  reportCardMode
}

func (c reportCard) renderCard(width int) string {
	return reportCardView(c.title, c.facts, c.mode, width)
}

// var _ consoleCard = reportCard{} — compile-time proof the wrapper
// satisfies the seam (spec.md US3 AS2); console_test.go additionally
// exercises it composed into Model.consoleCards end to end.
var _ consoleCard = reportCard{}

// mapPanelView is the widescreen MAP region — same glyph rendering as the
// narrow mapView (map.md: "content unchanged"), sized from the column
// budget instead of the full terminal width.
func (m Model) mapPanelView(cols, rows int) string {
	if rows < 5 { // B5: never let a starved resize drive Height() negative
		rows = 5
	}
	title := "MAP · following centroid"
	if m.panX != 0 || m.panY != 0 {
		title = "MAP · panned (c to recenter)"
	}
	vw, vh := mapViewportTiles(cols, rows-1) // -1: title row lives outside the grid box
	grid, legend := m.renderMapGrid(vw, vh)
	content := styleHeader.Render(title) + "\n" + grid
	if legend != "" {
		content += "\n" + legend
	}
	// clipContent is the load-bearing part here (B1): the legend line is
	// prose and routinely wider than the panel — without a hard per-line
	// cap, lipgloss's Width()-driven soft-wrap turns it into two rendered
	// lines, growing the panel past its Height() budget (Height only
	// pads short content, it never truncates tall content) and pushing
	// the header off the top of a real terminal. See clipContent's doc
	// for why a style-level MaxWidth() does not reliably substitute for
	// this. Every panel must render to exactly its handed (width,
	// height) — layout.md's composition contract.
	return styleBox.Width(cols - 2).Height(rows - 2).Render(clipContent(content, cols-2))
}

// dockPanelView is the widescreen DOCK region: tab row + active tab body
// (dock.md "Structure").
func (m Model) dockPanelView(cols, rows int) string {
	if rows < 5 { // B5: never let a starved resize drive Height() negative
		rows = 5
	}
	inner := cols - 4
	if inner < 10 {
		inner = 10
	}
	tabRow := m.dockTabsRow()
	divider := styleDim.Render(strings.Repeat("─", inner))
	content := m.dockTabContent(inner, rows-6)
	body := tabRow + "\n" + divider + "\n" + content
	// clipContent: see mapPanelView — never let a too-wide content line
	// soft-wrap and grow the panel past its Height() budget.
	return styleBox.Width(cols - 2).Height(rows - 2).Render(clipContent(body, cols-2))
}

// soloPanelView renders the same dock content full-width — "one
// implementation, two widths" (pages/solo-views.md "Solo rules").
func (m Model) soloPanelView(cols, rows int) string {
	if rows < 5 { // B5: never let a starved resize drive Height() negative
		rows = 5
	}
	inner := cols - 4
	if inner < 10 {
		inner = 10
	}
	title := styleHeader.Render(m.soloTitle())
	content := m.dockTabContent(inner, rows-4)
	body := title + "\n" + content
	// clipContent: see mapPanelView.
	return styleBox.Width(cols - 2).Height(rows - 2).Render(clipContent(body, cols-2))
}

func (m Model) soloTitle() string {
	name := strings.ToUpper(m.paneName(m.dockTab))
	if m.dockTab == paneChronicle {
		if m.inspecting() {
			mode := "raw"
			if !m.chronRaw && m.replica != nil && len(m.replica.Chronicle) > 0 {
				mode = "narrated"
			}
			return fmt.Sprintf("%s · %s · paused — j/k select · J/K scroll detail · r narrated", name, mode)
		}
		return name + " · r narrated ↔ raw · a/t filter"
	}
	return name
}

// dockTabsRow is the tab row that "doubles as the panel title" (dock.md).
func (m Model) dockTabsRow() string {
	tabs := []struct {
		p     pane
		label string
	}{
		{paneChronicle, "chronicle"},
		{paneGuardian, m.sk().TabLabel()}, // skin data (spec 052; skin-tokens.md rule 4 case-transforms below)
		{paneVillagers, "villagers"},
		{paneSystems, "systems"},
	}
	// Spec 054 (FR-008): the exercise tab joins the row only on scenario
	// worlds — a new label + content renderer, no new layout (dock.md).
	if m.exerciseID() != "" {
		tabs = append(tabs, struct {
			p     pane
			label string
		}{paneExercise, "exercise"})
	}
	var parts []string
	for _, t := range tabs {
		style := styleTabInactive
		if t.p == m.dockTab {
			style = styleTabActive
		}
		label := t.label
		if t.p == m.dockTab {
			label = strings.ToUpper(label)
		}
		rendered := style.Render(label)
		if t.p == paneGuardian && m.guardianUnseen {
			rendered += " " + styleTabBadge.Render("•")
		}
		parts = append(parts, rendered)
	}
	return strings.Join(parts, styleDim.Render(" │ "))
}

// dockTabContent renders just the active tab's body — shared verbatim by
// the dock panel and the solo view.
func (m Model) dockTabContent(width, height int) string {
	if height < 3 {
		height = 3
	}
	switch m.dockTab {
	case paneChronicle:
		maxWrap := 1
		if width < 60 {
			maxWrap = 3
		}
		return m.chronicleBody(width, height, maxWrap)
	case paneGuardian:
		return m.guardianTranscriptBody(width, height)
	case paneVillagers:
		return m.villagersBody(width, height)
	case paneSystems:
		return m.systemsContentBody(width, height)
	case paneExercise:
		return m.exerciseBody(width, height)
	}
	return ""
}

// --- map (panels/map.md: "Rendering is unchanged") ---

// wandererCentroid is the live (non-dead) agent centroid the camera follows
// by default — falls back to the map's own center when there's no replica or
// no living agent (renderMapGrid's original inline default). Extracted
// (spec 049 T002, research R1) so the jump-to-source camera writer
// (centerCameraOn, tui.go) computes the exact same number the map renderer
// does — a jump is a pan targeting this same centroid, never a second,
// slightly different notion of "center."
func (m Model) wandererCentroid() (cx, cy int) {
	if m.gameMap != nil {
		cx, cy = m.gameMap.W/2, m.gameMap.H/2
	}
	if m.replica == nil {
		return cx, cy
	}
	sx, sy, n := 0, 0, 0
	for _, a := range m.replica.Agents {
		if a.Dead {
			continue
		}
		sx += a.X
		sy += a.Y
		n++
	}
	if n > 0 {
		cx, cy = sx/n, sy/n
	}
	return cx, cy
}

// renderMapGrid draws the terrain+agents grid at exactly vw x vh tiles,
// returning the grid block and legend line separately — the shared core
// behind both the narrow mapView (today's vw/vh formula) and the
// widescreen mapPanelView (layout.md's column-budget formula). Only the
// sizing input differs; the glyphs themselves never change.
func (m Model) renderMapGrid(vw, vh int) (grid, legend string) {
	gm := m.gameMap
	if gm == nil {
		return styleDim.Render("no terrain (world manifest missing?)"), ""
	}
	if vw > gm.W {
		vw = gm.W
	}
	if vh > gm.H {
		vh = gm.H
	}
	if vw < 1 {
		vw = 1
	}
	if vh < 1 {
		vh = 1
	}

	// Camera center: wanderer centroid + pan offset, clamped to the map.
	// wandererCentroid is shared with centerCameraOn (spec 049 research R1)
	// so the jump math and this default camera center can never drift apart.
	cx, cy := m.wandererCentroid()
	cx += m.panX
	cy += m.panY
	x0 := clampInt(cx-vw/2, 0, gm.W-vw)
	y0 := clampInt(cy-vh/2, 0, gm.H-vh)

	agents := map[[2]int]string{}
	structures := map[[2]int]string{}
	// Graves (spec 044 US4, ratified follow-up): every post-044 death places
	// its grave at the SAME tile the dead agent's own frozen position
	// occupies, so a dead-agent lookup in tile() would otherwise always win
	// and the grave glyph could never actually render there — a dishonest
	// legend/overlay entry. A dedicated set (not folded into `structures`,
	// same reasoning as Quarried/Piles below) lets the agents loop below
	// check "is this dead agent's own tile a grave" and render the grave
	// glyph instead of the plain dead marker when so — the body becomes the
	// grave. A dead agent with no grave at its tile (pre-044 replay/history,
	// or a hand-built test replica) is unaffected: it keeps the plain "†".
	graves := map[[2]int]bool{}
	// Quarried (spec 012, US1): depleted rock outcrops are dynamic overlay
	// state (never part of the static gm.At tile), so the set comes from the
	// replica just like structures/dens below.
	quarried := map[[2]int]bool{}
	// Piles (spec 013 US2): ground piles are dynamic overlay state, same
	// treatment as Quarried/Structures — never part of the static gm.At tile.
	piles := map[[2]int]bool{}
	// Paths (spec 032 US3): a walkable tile improvement rendered at terrain
	// level (agents/structures/piles win over it), so it lives in its own set,
	// not the structures map.
	paths := map[[2]int]bool{}
	if m.replica != nil {
		for _, st := range m.replica.Structures {
			switch st.Kind {
			case "fire":
				// Lit vs cold (spec 012 T019/T024): lit iff current tick <
				// FuelUntil. A cold fire shows a hollow, faint glyph so the
				// player can tell a dead fire from a burning one (SC-006).
				if m.replica.Tick < st.FuelUntil {
					structures[[2]int{st.X, st.Y}] = styleFire.Render("▲")
				} else {
					structures[[2]int{st.X, st.Y}] = styleFireCold.Render("△")
				}
			case "shelter":
				structures[[2]int{st.X, st.Y}] = styleShelter.Render("⌂")
			case "oven":
				structures[[2]int{st.X, st.Y}] = styleOven.Render("▣")
			case "chest":
				structures[[2]int{st.X, st.Y}] = styleChest.Render("☐")
			case "grave":
				// Spec 044 US4 (FR-017): reducer-placed at a death tile, never
				// player-built. Recorded in `structures` (so a grave whose
				// tile no longer holds the dead agent — e.g. a future
				// migration/edge case — still renders here) AND in the
				// dedicated `graves` set the agents loop below consults: the
				// common case is a grave sharing its tile with the dead
				// agent it belongs to, and there the agent glyph branch
				// overrides itself to the grave (ratified follow-up) rather
				// than let the frozen "†" permanently hide it.
				structures[[2]int{st.X, st.Y}] = styleGrave.Render("✝")
				graves[[2]int{st.X, st.Y}] = true
			case "path":
				// Spec 032 US3: a path is a walkable tile improvement, so it
				// renders at TERRAIN level (below agents/structures/piles) rather
				// than in the structures map — an agent or a dropped pile on a
				// path tile still shows. Collected into its own set, keyed in the
				// terrain switch of tile().
				paths[[2]int{st.X, st.Y}] = true
			case "wall_plank", "wall_stone":
				// Spec 032 US1: walls block movement (structures win over terrain
				// in tile()), so they always show. A damaged wall (HP below the
				// derived max) renders dim, the cold-fire precedent, so the player
				// can spot a wall under demolition at a glance.
				glyph := "▤"
				if st.Kind == "wall_stone" {
					glyph = "▩"
				}
				if st.HP < sim.WallMaxHP(st.Kind) {
					structures[[2]int{st.X, st.Y}] = styleWallDamaged.Render(glyph)
				} else {
					structures[[2]int{st.X, st.Y}] = styleWall.Render(glyph)
				}
			}
		}
		for _, q := range m.replica.Quarried {
			quarried[[2]int{q.X, q.Y}] = true
		}
		// Piles (spec 013 US2): a dedicated map, not folded into
		// structures — build-site validation (FR-007) keeps piles and
		// structures off the same tile, but keeping them separate means a
		// coincidental overlap loses neither glyph's priority silently.
		for _, p := range m.replica.Piles {
			piles[[2]int{p.X, p.Y}] = true
		}
		for _, a := range m.replica.Agents {
			g := strings.ToUpper(a.Name[:1])
			switch {
			case a.Dead && graves[[2]int{a.X, a.Y}]:
				// Ratified follow-up: the body becomes the grave. Every
				// post-044 death places its grave at the dead agent's own
				// tile, so this is the common case; the plain "†" below is
				// what a graveless dead agent (pre-044 replay/history, or a
				// hand-built test replica) still shows.
				g = styleGrave.Render("✝")
			case a.Dead:
				g = styleErr.Render("†")
			case a.Asleep:
				g = styleAsleep.Render(strings.ToLower(g))
			default:
				g = styleAgent.Render(g)
			}
			agents[[2]int{a.X, a.Y}] = g
		}
	}
	dens := map[[2]int]bool{}
	for _, d := range gm.Dens {
		dens[[2]int{d.X, d.Y}] = true
	}

	gruX, gruY := -1, -1
	if m.replica != nil && m.replica.Gru != nil {
		gruX, gruY = m.replica.Gru.X, m.replica.Gru.Y
	}

	night := m.replica != nil && m.replica.Night
	tile := func(x, y int) string {
		if x == gruX && y == gruY {
			return styleGru.Render("G")
		}
		if g, ok := agents[[2]int{x, y}]; ok {
			return g
		}
		if g, ok := structures[[2]int{x, y}]; ok {
			return g
		}
		if piles[[2]int{x, y}] {
			return stylePile.Render("%")
		}
		if dens[[2]int{x, y}] {
			return styleDen.Render("ᴥ")
		}
		var s string
		var st lipgloss.Style
		switch {
		case paths[[2]int{x, y}]:
			// Spec 032 US3: a paved path — "·" in a warm tan, distinct from
			// plain grass's dim "·" so a laid path reads at a glance. Terrain
			// level: agents/structures/piles above already won by here.
			s, st = "·", stylePath
		case quarried[[2]int{x, y}]:
			// Depleted outcrop (effective-kind path, worldmap.Depleted):
			// passable dug-out ground, distinct from both intact rock and
			// plain grass (research R8).
			s, st = ",", styleDepleted
		case gm.At(x, y) == worldmap.Water:
			s, st = "~", styleWater
		case gm.At(x, y) == worldmap.Tree:
			s, st = "♠", styleTree
		case gm.At(x, y) == worldmap.Forage:
			s, st = "\"", styleForage
		case gm.At(x, y) == worldmap.Rock:
			s, st = "^", styleRock
		default:
			s, st = "·", styleDim
		}
		if night {
			st = st.Faint(true)
		}
		return st.Render(s)
	}

	var rows []string
	for y := y0; y < y0+vh; y++ {
		var row strings.Builder
		for x := x0; x < x0+vw; x++ {
			row.WriteString(tile(x, y) + " ")
		}
		rows = append(rows, strings.TrimRight(row.String(), " "))
	}
	grid = strings.Join(rows, "\n")

	phase := "day"
	if night {
		phase = styleNight.Render("night")
	}
	// Stockpile inspection (spec 013 T021, US2-AS5, SC-006): piles currently
	// in view are grouped into zones by 4-neighbor Manhattan adjacency — a
	// render-side-only computation (no zone state; data-model.md, spec.md
	// "Stockpile zone") — and each zone's aggregate contents (non-food
	// counts + food batch totals) are appended to the legend line, the
	// map panel's one designated inspection surface (map.md: "legend stays
	// pinned as the panel's last row" — content grows the line, never a
	// second row; clipContent already clips an over-wide legend, so this is
	// safe the same way the existing key text is).
	pilesInfo := ""
	if m.replica != nil && len(m.replica.Piles) > 0 {
		var visible []sim.Pile
		for _, p := range m.replica.Piles {
			if p.X >= x0 && p.X < x0+vw && p.Y >= y0 && p.Y < y0+vh {
				visible = append(visible, p)
			}
		}
		if len(visible) > 0 {
			var bits []string
			for _, zone := range pileZones(visible) {
				bits = append(bits, describePileZone(zone))
			}
			pilesInfo = " · " + strings.Join(bits, " · ")
		}
	}
	// Chest inspection (spec 013 T026, SC-006): chests currently in view get
	// an owner + contents entry appended to the same legend line, following
	// the pile inspection precedent above (T021) — the map panel's one
	// designated inspection surface, content grows the line rather than
	// adding a second row.
	chestsInfo := ""
	if m.replica != nil {
		var visible []sim.Structure
		for _, st := range m.replica.Structures {
			if st.Kind != "chest" {
				continue
			}
			if st.X >= x0 && st.X < x0+vw && st.Y >= y0 && st.Y < y0+vh {
				visible = append(visible, st)
			}
		}
		if len(visible) > 0 {
			names := m.agentNames()
			var bits []string
			for _, ch := range visible {
				bits = append(bits, describeChest(ch, names))
			}
			chestsInfo = " · " + strings.Join(bits, " · ")
		}
	}
	// Glyph key + agent/control notes render from the shared table (T002,
	// help.go) — one source, so the overlay's glyph walkthrough (FR-005)
	// and this compact legend line can never silently diverge.
	legend = styleDim.Render(fmt.Sprintf(
		"%s · [%d,%d–%d,%d of %d×%d] · %s · %s · %s%s%s",
		phase, x0, y0, x0+vw-1, y0+vh-1, gm.W, gm.H, legendGlyphLine(), agentGlyphNote, mapControlNote, pilesInfo, chestsInfo))
	return grid, legend
}

// describeChest renders one chest's inspection entry (spec 013 T026,
// SC-006): "chest(x,y) [Owner] <contents> <bulk>/<cap>" — owner resolved to
// the agent's Name via the same agentName helper the chronicle grammar uses
// (grammar.go), contents via summarizeInventoryContents (mirroring the pile
// zone summary's "non-food counts + food batch totals" shape, T021), and a
// fullness hint so "is the chest full" is answerable without opening state.
func describeChest(ch sim.Structure, names []string) string {
	owner := agentName(names, ch.Owner)
	contents := "empty"
	full := 0
	if ch.Store != nil {
		full = sim.Bulk(*ch.Store)
		contents = summarizeInventoryContents(*ch.Store)
	}
	return fmt.Sprintf("chest(%d,%d) [%s] %s %d/%d", ch.X, ch.Y, owner, contents, full, sim.ChestCap)
}

// summarizeInventoryContents renders a chest's Store the same way
// summarizePileContents renders a pile's aggregate contents (T021): each
// non-zero resource count, a spear count, and the food triplet as one
// "food Nr/Nc/Nm" entry when any food is held. A chest's Store is a plain
// sim.Inventory (counts, not FoodBatch — chests preserve food forever, no
// spoilage deadlines to track, FR-010), so this reads the counts directly
// rather than summing batches.
func summarizeInventoryContents(inv sim.Inventory) string {
	var parts []string
	if inv.Wood > 0 {
		parts = append(parts, fmt.Sprintf("%dw", inv.Wood))
	}
	if inv.Stone > 0 {
		parts = append(parts, fmt.Sprintf("%dst", inv.Stone))
	}
	if inv.Water > 0 {
		parts = append(parts, fmt.Sprintf("%dwt", inv.Water))
	}
	if inv.Planks > 0 {
		parts = append(parts, fmt.Sprintf("%dpl", inv.Planks))
	}
	if inv.RefinedStone > 0 {
		parts = append(parts, fmt.Sprintf("%drs", inv.RefinedStone))
	}
	if n := len(inv.Spears); n > 0 {
		parts = append(parts, fmt.Sprintf("%dspear", n))
	}
	if inv.FoodRaw+inv.FoodCooked+inv.Meals > 0 {
		parts = append(parts, fmt.Sprintf("food %dr/%dc/%dm", inv.FoodRaw, inv.FoodCooked, inv.Meals))
	}
	if len(parts) == 0 {
		return "empty"
	}
	return strings.Join(parts, " ")
}

// pileZones groups piles into stockpile zones by 4-neighbor Manhattan
// adjacency (spec.md "Stockpile zone": "an observability grouping of
// adjacent piles — a rendering concept, not a state entity"). Purely a
// render-side flood fill: it reads only the piles handed to it and
// produces no state. Deterministic given a deterministic input order —
// zones are discovered in `piles` order, and each zone's members are
// visited in a fixed 4-neighbor order (N, E, S, W), matching the sim
// package's own Manhattan-adjacency convention (internal/sim/state.go
// pileOnOrAdjacent's neighborOrder).
func pileZones(piles []sim.Pile) [][]sim.Pile {
	byTile := make(map[[2]int]sim.Pile, len(piles))
	for _, p := range piles {
		byTile[[2]int{p.X, p.Y}] = p
	}
	dirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	visited := make(map[[2]int]bool, len(piles))
	var zones [][]sim.Pile
	for _, p := range piles {
		start := [2]int{p.X, p.Y}
		if visited[start] {
			continue
		}
		visited[start] = true
		queue := [][2]int{start}
		var zone []sim.Pile
		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			zone = append(zone, byTile[cur])
			for _, d := range dirs {
				nb := [2]int{cur[0] + d[0], cur[1] + d[1]}
				if _, ok := byTile[nb]; ok && !visited[nb] {
					visited[nb] = true
					queue = append(queue, nb)
				}
			}
		}
		zones = append(zones, zone)
	}
	return zones
}

// describePileZone renders one pile-content inspection entry: a single pile
// as "pile(x,y) contents", a multi-pile zone as its bounding box + pile
// count. Contents = non-food counts (wood/stone/water/planks/refined stone/
// spears) + food batch totals per kind, matching T021's spec wording and the
// souls pane's carried-inventory phrasing (SC-006 consistency).
func describePileZone(zone []sim.Pile) string {
	contents := summarizePileContents(zone)
	if len(zone) == 1 {
		return fmt.Sprintf("pile(%d,%d) %s", zone[0].X, zone[0].Y, contents)
	}
	minX, minY, maxX, maxY := zone[0].X, zone[0].Y, zone[0].X, zone[0].Y
	for _, p := range zone[1:] {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	return fmt.Sprintf("zone[%d](%d,%d)-(%d,%d) %s", len(zone), minX, minY, maxX, maxY, contents)
}

// summarizePileContents aggregates one or more piles' contents into the
// same "non-food counts + food batch totals" shape T021 calls for: raw
// resource counts, a spear count, and the food triplet raw/cooked/meals
// (batch totals, deadlines omitted — this is a contents summary, not a rot
// countdown).
func summarizePileContents(piles []sim.Pile) string {
	var wood, stone, water, planks, refined, spears int
	var foodRaw, foodCooked, foodMeals int
	for _, p := range piles {
		wood += p.Wood
		stone += p.Stone
		water += p.Water
		planks += p.Planks
		refined += p.RefinedStone
		spears += len(p.Spears)
		for _, f := range p.Food {
			switch f.Kind {
			case "food_raw":
				foodRaw += f.N
			case "food_cooked":
				foodCooked += f.N
			case "meals":
				foodMeals += f.N
			}
		}
	}
	var parts []string
	if wood > 0 {
		parts = append(parts, fmt.Sprintf("%dw", wood))
	}
	if stone > 0 {
		parts = append(parts, fmt.Sprintf("%dst", stone))
	}
	if water > 0 {
		parts = append(parts, fmt.Sprintf("%dwt", water))
	}
	if planks > 0 {
		parts = append(parts, fmt.Sprintf("%dpl", planks))
	}
	if refined > 0 {
		parts = append(parts, fmt.Sprintf("%drs", refined))
	}
	if spears > 0 {
		parts = append(parts, fmt.Sprintf("%dspear", spears))
	}
	if foodRaw+foodCooked+foodMeals > 0 {
		parts = append(parts, fmt.Sprintf("food %dr/%dc/%dm", foodRaw, foodCooked, foodMeals))
	}
	if len(parts) == 0 {
		return "empty"
	}
	return strings.Join(parts, " ")
}

// Terrain glyphs. Night dims the palette rather than hiding the world.
var (
	styleWater    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	styleTree     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleForage   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleRock     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleDepleted = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("240"))
	styleDen      = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	styleFire     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	styleFireCold = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("240"))
	styleShelter  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("130"))
	styleOven     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("166"))
	// Pile (spec 013 US2): "%" is the roguelike convention for a ground
	// item/goods stash, distinct from every existing glyph; tan/gold (178)
	// reads as "cache" without colliding with fire's orange (208), oven's
	// burnt orange (166), or shelter's brown (130).
	stylePile = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("178"))
	// Chest (spec 013 US3): "☐" (empty box) reads as a container distinct
	// from every existing glyph — unlike a pile's loose "%", a chest is a
	// built structure with a lid. Dark goldenrod (136) sits between pile's
	// tan (178) and shelter's brown (130) without matching either, so a
	// chest never gets mistaken for a stockpile or a house at a glance.
	styleChest = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("136"))
	// Wall (spec 032 US1): "▤" (plank) / "▩" (stone) read as solid barriers
	// distinct from every existing glyph; slate gray (250) sets them apart from
	// intact rock's "^" (245) and the burnt-orange structures. A damaged wall
	// (HP < max) renders faint (240), the cold-fire dim precedent.
	styleWall        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	styleWallDamaged = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("240"))
	// Path (spec 032 US3): "·" in warm tan (137) — the paved-route glyph, set
	// apart from plain grass's dim "·" without colliding with any structure glyph.
	stylePath = lipgloss.NewStyle().Foreground(lipgloss.Color("137"))
	styleGru  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	// Grave (spec 044 US4): "✝" marks a death site — a somber, persistent
	// marker, so it renders faint gray (244) rather than any of the vivid
	// living-structure colors (fire/shelter/oven/chest), the cold-fire (240)
	// precedent for "spent"/inert glyphs.
	styleGrave = lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("244"))
)

// mapView is the narrow-fallback map pane: today's vw/vh formula,
// unchanged (pages/solo-views.md "Narrow fallback" — "today's single-pane
// UI renders unchanged").
func (m Model) mapView() string {
	vw, vh := 32, 18
	if m.width > 8 {
		if w := (m.width - 6) / 2; w < vw || m.width >= 80 {
			vw = w
		}
	}
	if m.height > 12 {
		vh = m.height - 10
	}
	grid, legend := m.renderMapGrid(vw, vh)
	return styleBox.Render(grid) + "\n" + legend
}

func clampInt(v, lo, hi int) int {
	if hi < lo {
		hi = lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// clipLine crops a single line (ANSI-safe, via lipgloss.Style.MaxWidth) to
// at most width visible columns; a line that already fits is returned
// unchanged (MaxWidth alone would pad it, which clipContent doesn't want).
func clipLine(s string, width int) string {
	if width < 1 {
		width = 1
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return lipgloss.NewStyle().MaxWidth(width).Render(s)
}

// clipContent crops every line of a multi-line block to fit inside a
// styleBox/stylePanelFocus-family panel whose Width() is set to boxWidth —
// B1. Two lipgloss facts combine into a bug otherwise: (1) Height() only
// *pads* short content, it never truncates tall content, so one overlong
// line silently grows the whole panel past its row budget instead of
// erroring; (2) a style's own Padding(0,1) eats 2 of boxWidth's columns
// *before* text renders, so the true usable width is boxWidth-2, not
// boxWidth. A style-level .MaxWidth() does not reliably substitute for
// this: empirically (see TASK-34 investigation notes), MaxWidth combined
// with Height on multi-line content whose line count already meets the
// Height budget can still double-wrap every line instead of cropping —
// pre-clipping each line before Render() is the only combination that
// held up under test. Callers pass the same boxWidth given to .Width().
func clipContent(content string, boxWidth int) string {
	usable := boxWidth - 2 // Padding(0,1)
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		lines[i] = clipLine(l, usable)
	}
	return strings.Join(lines, "\n")
}

// --- chronicle (panels/chronicle.md, patterns/chronicle-grammar.md) ---
// One body renderer shared by the narrow pane, the dock tab, and the solo
// view — differing only in (width, height, maxWrap).

func (m Model) chronicleFilterHint() string {
	agentName := "all"
	if m.replica != nil && m.chronAgent >= 0 && m.chronAgent < len(m.replica.Agents) {
		agentName = m.replica.Agents[m.chronAgent].Name
	}
	thread := m.chronThread
	if thread == "" {
		thread = "all"
	}
	return fmt.Sprintf("agent %s · thread %s · a/t filter, r raw feed", agentName, thread)
}

// chronicleBody dispatches to inspect / narrated / raw per panels/chronicle.md.
func (m Model) chronicleBody(width, height, maxWrap int) string {
	if m.inspecting() {
		return m.chronicleInspectBody(width, height)
	}
	narrated := m.replica != nil && len(m.replica.Chronicle) > 0
	if !m.chronRaw && narrated {
		return m.chronicleNarratedBody(width, height)
	}
	return m.chronicleRawBody(width, height, maxWrap)
}

// chronicleNarratedBody is TASK-11's narrated feed — content unchanged.
func (m Model) chronicleNarratedBody(width, rows int) string {
	header := styleDim.Render(m.chronicleFilterHint())
	var lines []string
	for _, c := range m.replica.Chronicle {
		if m.chronAgent >= 0 && !c.Mentions(m.chronAgent) {
			continue
		}
		if m.chronThread != "" && c.Thread != m.chronThread {
			continue
		}
		stamp := fmt.Sprintf("day %d", c.Day)
		if c.Thread != "" {
			stamp += " · " + c.Thread
		}
		lines = append(lines, styleDim.Render(stamp)+" "+chronNames(m.replica, c))
		lines = append(lines, wrapText(c.Text, width)...)
		lines = append(lines, "")
	}
	if len(lines) == 0 {
		return header + "\n\n" + styleDim.Render("no entries match these filters yet")
	}
	// B1/B5: `rows` is this body's *entire* row budget, but header+blank
	// above already spend 2 of it — reserve those before capping the
	// entry list, or the returned string can run 2 lines over budget.
	entryRows := rows - 2
	if entryRows < 3 {
		entryRows = 3
	}
	if len(lines) > entryRows {
		lines = lines[len(lines)-entryRows:]
	}
	return header + "\n\n" + strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// chronicleRawBody is the raw event feed formatted by the chronicle digest
// grammar (contracts/digest-grammar.md), auto-following the tail. R8:
// window first, then format — only the tail slice of events that could
// possibly land in the visible budget is digested, not the whole 256-event
// ring, so per-frame cost stays O(visible rows) even at max time
// compression (SC-005).
func (m Model) chronicleRawBody(width, rows, maxWrap int) string {
	narrated := m.replica != nil && len(m.replica.Chronicle) > 0
	hint := "raw feed · no narrated entries yet — the narrator writes at day and night boundaries"
	if narrated {
		hint = "raw feed · r narrated view"
	}
	if len(m.events) == 0 {
		return styleDim.Render(hint) + "\n\n" +
			styleDim.Render("no events yet this session — the chronicle fills as the world moves")
	}
	// B1/B5: `rows` is this body's *entire* row budget; hint+blank above
	// already spend 2 of it (see chronicleNarratedBody).
	entryRows := rows - 2
	if entryRows < 3 {
		entryRows = 3
	}
	// Each event contributes at least one physical line, so the tail
	// `entryRows` events are always enough to fill (and, once wrapped,
	// potentially overfill) the budget — the physical-line slice below
	// trims any overshoot from dock-mode wrapping.
	events := m.events
	if len(events) > entryRows {
		events = events[len(events)-entryRows:]
	}
	names := m.agentNames()
	dock := maxWrap > 1
	lines := make([]chronicleLine, len(events))
	for i, e := range events {
		lines[i] = formatChronicleLine(e, names, m.sk())
	}
	cols := computeChronicleColumns(lines, dock)
	var out []string
	for _, l := range lines {
		out = append(out, renderChronicleRow(l, cols, width, maxWrap, false))
	}
	all := strings.Split(strings.Join(out, "\n"), "\n")
	if len(all) > entryRows {
		all = all[len(all)-entryRows:]
	}
	return styleDim.Render(hint) + "\n\n" + strings.Join(all, "\n")
}

// chronicleInspectBody is Mode 2 (paused) — panels/chronicle.md "Mode 2",
// contracts/digest-grammar.md §5. The body splits into the entry list (top)
// and an always-on detail pane (bottom) separated by a rule line: no
// keypress required to see the selected event's verbatim payload (FR-008,
// R6) — the ⏎-triggered inline inspector this replaced is gone (R7).
// Bounded to exactly `rows` total lines regardless of payload size (B1/B2):
// the pane's row budget is reserved *before* windowing the list, and the
// pane itself windows the annotated payload by chronDetailScroll rather
// than ever emitting it in full — the actual cause of the old inline
// inspector's unbounded-growth bug (see the historical comment this
// replaced) for oversized payloads like world.migrated (FR-011).
func (m Model) chronicleInspectBody(width, rows int) string {
	if len(m.events) == 0 {
		return styleDim.Render("paused — no events recorded yet")
	}
	// Minimum viable split: list(5) + rule(1) — the floor every other body
	// in this package clamps to (B5); paneRows may shrink to 0 below this
	// only in a starved-terminal degenerate case.
	if rows < 6 {
		rows = 6
	}
	names := m.agentNames()
	n := len(m.events)
	sel := m.chronSelectionBase()

	// R6: paneRows = min(rows/2, 14); list keeps the remainder, floored at 5.
	paneRows := rows / 2
	if paneRows > 14 {
		paneRows = 14
	}
	const ruleRows = 1
	listRows := rows - paneRows - ruleRows
	if listRows < 5 {
		listRows = 5
		paneRows = rows - listRows - ruleRows
		if paneRows < 0 {
			paneRows = 0
		}
	}

	// --- entry list (unchanged windowing discipline, minus expansion) ---
	start := sel - listRows/2
	if start < 0 {
		start = 0
	}
	end := start + listRows
	if end > n {
		end = n
		start = end - listRows
		if start < 0 {
			start = 0
		}
	}
	lines := make([]chronicleLine, 0, end-start)
	for i := start; i < end; i++ {
		lines = append(lines, formatChronicleLine(m.events[i], names, m.sk()))
	}
	cols := computeChronicleColumns(lines, false) // inspect is always tick-shown (solo-style)
	listOut := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		l := lines[i-start]
		selected := i == sel
		marker := "  "
		if selected {
			marker = styleFeedSelect.Render("▌") + " "
		}
		listOut = append(listOut, marker+renderChronicleRow(l, cols, width-2, 1, selected))
	}
	m.recordChronHit(start, end) // spec 049 T009: this frame's click geometry

	// --- rule + detail pane (contract §5) ---
	e := m.events[sel]
	rule := styleDim.Render(fmt.Sprintf("DETAIL · seq %d", e.Seq))
	out := append([]string{}, listOut...)
	out = append(out, rule)
	if paneRows > 0 {
		actionLabel := ""
		if actions := m.detailActions(e); len(actions) > 0 {
			actionLabel = actions[0].Label
		}
		out = append(out, chronicleDetailPane(e, names, m.chronDetailScroll, width, paneRows, actionLabel)...)
	}
	return strings.Join(out, "\n")
}

// chronicleDetailPane windows formatInspector's verbatim-payload output to
// exactly paneRows lines (contract §5): scroll clamps to content so J past
// the end (or K before the start) is a no-op rather than drifting the view
// blank. The bottom row is always reserved for the actions bar (spec 049,
// contract §3 "bottom-right slot") — a permanent corner of the pane, not
// merely a byproduct of scrolling, so actionLabel renders there whether or
// not the payload also overflows; when it does, the row carries both the
// remaining-line count and the actions bar (the pre-existing single-row
// layout). Oversized payloads (world.migrated) are never processed beyond
// this slice — the annotated string is built once, then only the visible
// lines are touched, satisfying FR-011 structurally rather than by a size
// cap.
func chronicleDetailPane(e store.Event, names []string, scroll, width, paneRows int, actionLabel string) []string {
	content := strings.Split(formatInspector(e, names), "\n")

	contentRows := paneRows - 1 // one row always reserved for the actions bar
	if contentRows < 1 {
		contentRows = 1
	}
	footerNeeded := len(content) > contentRows
	maxScroll := len(content) - contentRows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}
	visEnd := scroll + contentRows
	if visEnd > len(content) {
		visEnd = len(content)
	}

	out := make([]string, 0, paneRows)
	for _, ln := range content[scroll:visEnd] {
		out = append(out, indentBlock(ln, "  "))
	}
	for len(out) < contentRows { // pad the content area so the actions row is always last
		out = append(out, "")
	}

	footer := ""
	if footerNeeded {
		remaining := len(content) - visEnd
		footer = fmt.Sprintf("… (+%d more — J to scroll)", remaining)
	}
	gap := width - len([]rune(footer)) - len([]rune(actionLabel)) - 4
	if gap < 2 {
		gap = 2
	}
	out = append(out, styleDim.Render("  "+footer+strings.Repeat(" ", gap)+actionLabel))

	for len(out) < paneRows { // pad so the composite height is fixed (B1)
		out = append(out, "")
	}
	return out
}

func indentBlock(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = styleDim.Render(prefix) + lines[i]
	}
	return strings.Join(lines, "\n")
}

// familyTint resolves a family to its color-role token (contract §4).
// Roles, never raw colors, at the call site — this is the one place a
// family maps to an actual lipgloss.Style.
func familyTint(f eventFamily) lipgloss.Style {
	switch f {
	case familyWorld:
		return styleFamilyWorld
	case familyClock:
		return styleFeedClock // existing token; contract §4: "clock keeps yellow"
	case familySim:
		return styleFamilySim
	case familyAgent:
		return styleFamilyAgent
	case familySocial:
		return styleFamilySocial
	case familyGovernance:
		return styleFamilyGovernance
	case familyGru:
		return styleFamilyGru
	case familyChronicle:
		return styleFamilyChronicle
	case familyGuardian:
		return styleFamilyGuardian
	case familyCog:
		return styleFamilyCog
	default: // familyDaemon, familyUnknown — no distinct tint (see token block)
		return styleDim
	}
}

// styleForRole maps one styled rune's paint-time role to a style, given the
// row's family tint (used for styleRoleFamily — the prefix — since it's the
// one role whose color varies per line rather than being fixed).
func styleForRole(role styleRole, fam lipgloss.Style) lipgloss.Style {
	switch role {
	case styleRoleFamily:
		return fam
	case styleRoleName:
		return styleFeedName
	case styleRoleSpeech:
		return styleFeedSpeech
	case styleRoleEmphasis:
		return styleFeedEmphasis
	default:
		return lipgloss.NewStyle() // default terminal foreground
	}
}

// paintStyledLine renders one already-wrapped/truncated styledLine
// (grammar.go's styleWrapLine) by walking its per-rune roles and emitting
// one Render() call per contiguous same-role run — R4's "style segment-wise
// after wrap": the wrapping/truncation that produced l already happened on
// plain runes, so this can never split an ANSI escape.
func paintStyledLine(l styledLine, fam lipgloss.Style, selected bool) string {
	var b strings.Builder
	i := 0
	for i < len(l.Runes) {
		role := styleRoleText
		if i < len(l.Roles) {
			role = l.Roles[i]
		}
		j := i + 1
		for j < len(l.Runes) {
			r := styleRoleText
			if j < len(l.Roles) {
				r = l.Roles[j]
			}
			if r != role {
				break
			}
			j++
		}
		st := styleForRole(role, fam)
		if selected {
			st = st.Reverse(true)
		}
		b.WriteString(st.Render(string(l.Runes[i:j])))
		i = j
	}
	return b.String()
}

// renderChronicleRow styles+wraps/truncates one formatted line to width,
// given the shared window's column layout (R5) — contract §2/§4/T021:
//   - alert types (agent.died, gru.attacked, social.chest_taken,
//     norm.violated) render the whole line in the alert role, regardless
//     of family, so they pop without reading.
//   - labeled-voice families (cog, clock, daemon) tint the whole line with
//     the family color — the summary IS already "key=value", no further
//     per-segment treatment applies.
//   - every other family tints only the type column, and the summary
//     renders segment-wise (name/speech/emphasis roles pop against
//     default-color connective prose) via styleWrapLine + paintStyledLine.
//
// Selection reverse is preserved in all three paths.
func renderChronicleRow(l chronicleLine, cols chronicleColumns, width, maxWrap int, selected bool) string {
	if isAlertType(l.Type) {
		return styleWholeLine(plainChronicleLine(l, cols), width, maxWrap, styleFeedAlert, selected)
	}
	if isLabeledVoiceFamily(l.Family) {
		return styleWholeLine(plainChronicleLine(l, cols), width, maxWrap, familyTint(l.Family), selected)
	}
	prefix := chronicleLinePrefix(l, cols)
	fam := familyTint(l.Family)
	styledLines := styleWrapLine(prefix, l.Summary, width, maxWrap)
	out := make([]string, len(styledLines))
	for i, sl := range styledLines {
		out[i] = paintStyledLine(sl, fam, selected)
	}
	return strings.Join(out, "\n")
}

// styleWholeLine wraps/truncates the plain line then renders every
// physical line with one uniform style — the alert and labeled-voice paths
// don't need per-segment attribution (contract §2).
func styleWholeLine(plain string, width, maxWrap int, style lipgloss.Style, selected bool) string {
	lines := wrapOrTruncatePlain(plain, width, maxWrap)
	if selected {
		style = style.Reverse(true)
	}
	for i, ln := range lines {
		lines[i] = style.Render(ln)
	}
	return strings.Join(lines, "\n")
}

// chronicleView is the narrow-fallback chronicle pane (today's TASK-11
// behavior, header/footer chrome unchanged; body now shares the grammar
// formatter with the dock/solo renderers).
func (m Model) chronicleView() string {
	width := m.width - 4
	if width < 30 {
		width = 30
	}
	rows := m.height - 9
	return m.chronicleBody(width, rows, 1)
}

// chronNames renders an entry's cast, styled like agents elsewhere.
func chronNames(s *sim.State, c sim.ChronicleEntry) string {
	var names []string
	for _, a := range c.Agents {
		if a >= 0 && a < len(s.Agents) {
			names = append(names, s.Agents[a].Name)
		}
	}
	if len(names) == 0 {
		return ""
	}
	return styleAgent.Render(strings.Join(names, ", "))
}

// nextThread cycles "" → each distinct thread in ring order → "".
func nextThread(s *sim.State, cur string) string {
	if s == nil {
		return ""
	}
	var threads []string
	seen := map[string]bool{}
	for _, c := range s.Chronicle {
		if c.Thread != "" && !seen[c.Thread] {
			seen[c.Thread] = true
			threads = append(threads, c.Thread)
		}
	}
	if len(threads) == 0 {
		return ""
	}
	if cur == "" {
		return threads[0]
	}
	for i, t := range threads {
		if t == cur && i+1 < len(threads) {
			return threads[i+1]
		}
	}
	return ""
}

// wrapText greedy-wraps prose to the given width.
func wrapText(text string, width int) []string {
	var lines []string
	var cur strings.Builder
	for _, w := range strings.Fields(text) {
		if cur.Len() > 0 && cur.Len()+1+len(w) > width {
			lines = append(lines, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteByte(' ')
		}
		cur.WriteString(w)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return lines
}

// --- guardian (panels/dock.md "metatron", panels/minibuffer.md) ---

// guardianTranscriptBody is the dock/solo guardian tab: history only —
// input lives in the minibuffer (minibuffer.md).
func (m Model) guardianTranscriptBody(width, rows int) string {
	if rows < 3 {
		rows = 3
	}
	var lines []string
	if len(m.transcript) == 0 && !m.mbBusy {
		lines = append(lines, styleDim.Render("you   ask the "+m.sk().Epithet()+" anything — press m to focus the minibuffer"))
	}
	for _, l := range m.transcript {
		lines = append(lines, transcriptRowLines(l, width, m.sk())...)
	}
	if m.mbBusy {
		lines = append(lines, styleAgent.Render(m.sk().Epithet()+" ⋮ thinking…"))
	}
	if len(lines) > rows {
		lines = lines[len(lines)-rows:] // newest at bottom; opens scrolled to bottom
	}
	return strings.Join(lines, "\n")
}

// transcriptRowLines renders one stored transcript line as you/guardian rows
// (dock.md mockup), wrapping the text to width. The guardian row's label is
// the skin's epithet (spec 052 FR-007); the label column sizes to the
// longest label so a longer skin epithet still aligns its continuations.
func transcriptRowLines(l string, width int, sk *skin.Skin) []string {
	label, text, style := classifyTranscriptLine(l, sk)
	if label == "" {
		return []string{style.Render(l)}
	}
	labelW := len([]rune(label))
	if labelW < 5 {
		labelW = 5
	}
	wrapped := wrapText(text, width-labelW-1)
	if len(wrapped) == 0 {
		wrapped = []string{""}
	}
	var out []string
	for i, w := range wrapped {
		prefix := strings.Repeat(" ", labelW+1)
		if i == 0 {
			prefix = fmt.Sprintf("%-*s ", labelW, label)
		}
		out = append(out, styleDim.Render(prefix)+style.Render(w))
	}
	return out
}

// transcriptGuardianPrefix marks the guardian's stored transcript rows — an
// internal storage marker (client-only ring, never persisted), displayed
// through the skin's epithet, never verbatim.
const transcriptGuardianPrefix = "guardian: "

func classifyTranscriptLine(l string, sk *skin.Skin) (label, text string, style lipgloss.Style) {
	switch {
	case strings.HasPrefix(l, "you: "):
		return "you", strings.TrimPrefix(l, "you: "), lipgloss.NewStyle()
	case strings.HasPrefix(l, transcriptGuardianPrefix):
		return sk.Epithet(), strings.TrimPrefix(l, transcriptGuardianPrefix), lipgloss.NewStyle()
	case strings.HasPrefix(l, transcriptVerdictPrefix):
		// spec 020 (TASK-63, contract R12): the guardian's own inline
		// tool-call verdicts — styled as telemetry (dim), distinct from the
		// you/guardian rows, and labeled so it wraps like a normal row
		// instead of relying on clipContent's truncation.
		return "note", strings.TrimPrefix(l, transcriptVerdictPrefix), styleFamilyCog
	default:
		return "", l, styleAgent
	}
}

// guardianHeaderLine renders the guardian pane-header line: charge bank,
// then — once a world has an active charter/instruction surface — the
// spec-021 provenance summary and the spec-046 stage segment. Shared
// verbatim (spec 053 FR-002, "one shared data source") by the narrow-
// fallback guardianView below and the guardian console (consoleView) —
// literally the same string, not a re-derivation, so the two renderings can
// never silently disagree.
func (m Model) guardianHeaderLine() string {
	charges := 0
	if m.status != nil {
		charges = m.status.Clock.GuardianCharges
	}
	// Pane header: the skin's proper name, uppercased (skin-tokens.md rule 4:
	// case is a rendering detail of this header, not a token fork).
	header := fmt.Sprintf("%s · charges %s%s", strings.ToUpper(m.sk().Name()),
		strings.Repeat("⚡", charges), strings.Repeat("·", clampInt(sim.GuardianChargeCap-charges, 0, sim.GuardianChargeCap)))
	if m.consoleCharter != "" {
		// Instruction + capability provenance (spec 021 US3): charter flavor,
		// then the skill count when non-zero, then the granted-tool summary
		// (quiet for a full-grant default world). Answers "what is my guardian
		// running on, and what can it do" without leaving the game.
		prov := m.consoleCharter + " (charter.md)"
		if m.consoleSkills == 1 {
			prov += " · 1 skill"
		} else if m.consoleSkills > 1 {
			prov += fmt.Sprintf(" · %d skills", m.consoleSkills)
		}
		if m.consoleTools != "" {
			prov += " · " + m.consoleTools
		}
		if m.consoleStage != "" {
			prov += " · " + m.consoleStage
		}
		header += styleDim.Render(" · " + prov)
	}
	return header
}

// guardianView is the narrow-fallback guardian pane: transcript + the
// focus-contract-governed input line (replaces the old always-typing
// console pane — the exact bug at tui.go:305-309). Fiction-layer content
// only since spec 053 (D10): the provider table, spend/wallet line, and
// cognition horizon block moved to the systems tab (systemsView) — this
// pane keeps the header, transcript, and standing orders (contract §3
// "STAYS on guardian tab").
func (m Model) guardianView() string {
	width := m.width - 6
	if width < 30 {
		width = 30
	}
	header := m.guardianHeaderLine()

	body := m.guardianTranscriptBody(width, clampInt(m.height-14, 4, 200))
	if m.mbErr != "" {
		body += "\n" + styleErr.Render("the "+m.sk().Epithet()+" is unreachable: "+m.mbErr)
	}

	content := header + "\n\n" + body
	for _, row := range orderStatusLines(m.consoleOrders) {
		content += "\n" + row
	}
	// Sized to the same content width as the transcript above it (not the
	// full terminal width) — this box nests inside guardianView's own
	// bordered pane, which adds its own chrome on top.
	//
	// Guardian strip narrow carry (patterns/layout.md ruling b, spec 050):
	// this is the ONLY place a minibuffer exists in the narrow fallback
	// (narrowView's other panes have no composer at all — narrow keeps the
	// pre-TASK-34 embedded-console pattern), so "carried, above the
	// minibuffer, identical content" lands here, unconditionally — narrow
	// has no rowBudget/fold machinery of its own to fold it against.
	content += "\n\n" + m.guardianStripView(width) + "\n" + m.minibufferView(width)
	return styleBox.Render(content)
}

// --- systems (panels/systems.md, spec 053 US2/D10) ---
// The dock's never-skinned telemetry tab: the provider table, spend/wallet
// line, and cognition horizon block relocated out of the guardian tab
// (contract §3) — the exact renderers already shipped (llmProviderLines,
// horizonLines), moved rather than rewritten.

// systemsContentBody is the systems dock tab's shared body — used by both
// dockTabContent (widescreen dock/solo) and systemsView (narrow fallback),
// the same one-body-two-call-sites shape every other tab uses. A no-LLM
// world (Status.LLM nil) states its absence honestly (SC-002) rather than
// rendering empty chrome — the exact honesty rule the compact guardian tab
// used to satisfy only by silence.
func (m Model) systemsContentBody(width, rows int) string {
	if rows < 3 {
		rows = 3
	}
	if m.status == nil || m.status.LLM == nil {
		return styleDim.Render("no LLM configured for this world")
	}
	l := m.status.LLM
	var lines []string
	lines = append(lines, llmProviderLines(l)...)
	lines = append(lines, styleDim.Render(fmt.Sprintf("spend $%.2f of $%.0f", l.Spent, l.Budget)))
	if l.Spent >= l.Budget {
		// Telemetry voice, deliberately fiction-free: the systems tab is never
		// skinned (spec 052 FR-016 / D10) — the fiction-layer phrasing of this
		// condition lives on the guardian surfaces, not here.
		lines = append(lines, styleErr.Render("budget exhausted — LLM calls refused"))
	}
	lines = append(lines, horizonLines(m.status.Horizon, string(m.status.Clock.Speed))...)
	if len(lines) > rows {
		lines = lines[len(lines)-rows:]
	}
	return strings.Join(lines, "\n")
}

// systemsView is the narrow-fallback systems pane: the same
// systemsContentBody the dock tab and solo view share, boxed like every
// other narrow pane (villagersView/guardianView precedent).
func (m Model) systemsView() string {
	width := m.width - 6
	if width < 30 {
		width = 30
	}
	content := m.systemsContentBody(width, clampInt(m.height-6, 4, 200))
	return styleBox.Render(content)
}

// orderStatusLines renders the guardian pane's standing-orders block (spec 029
// T023, FR-016): a header count, then one compact row per order — id, a fuzzy
// marker, origin, remaining game-day, and status, followed by the condition
// text. nil when no orders stand (the pane shows nothing extra, matching the
// Status.Orders omitempty contract).
func orderStatusLines(orders []guardian.OrderStatus) []string {
	if len(orders) == 0 {
		return nil
	}
	lines := make([]string, 0, len(orders)+1)
	lines = append(lines, styleDim.Render(fmt.Sprintf("👁 standing orders (%d)", len(orders))))
	for _, o := range orders {
		fuzzy := ""
		if o.Fuzzy {
			fuzzy = "~"
		}
		lines = append(lines, fmt.Sprintf("  %s%s [%s · day %d · %s] %q",
			o.ID, fuzzy, o.Origin, o.ExpiresDay, o.Status, o.Condition))
	}
	return lines
}

// llmProviderNameWidth / llmProviderModelWidth are the provider table's fixed
// column widths (contracts/status.md's TUI section): declared names/models
// run short (local, cloud, cogito, gemma4:12b-mlx) and a fixed width keeps
// every row's glyph/queue/spend columns aligned without measuring the whole
// Providers slice first.
const (
	llmProviderNameWidth  = 10
	llmProviderModelWidth = 18
)

// llmProviderLines renders the operator-facing provider table (spec 024
// US6, contracts/status.md "TUI"): one row per provider — name, model,
// up/down glyph, queue depth, inflight/slots, a contended marker, and this
// provider's spend share of the month — sorted by name (the wire order,
// StatusSnapshot already sorts). A trailing `(unattributed)` row appears
// only when the global spend total exceeds Σ(rows), the legacy-spend
// remainder contracts/status.md documents. A provider carrying an active
// health condition (spec 034) gains an indented continuation line with the
// condition's detail and remedy in the pane's error style, immediately
// below its row.
func llmProviderLines(l *llm.Status) []string {
	if l == nil || len(l.Providers) == 0 {
		return nil
	}
	lines := make([]string, 0, len(l.Providers)+1)
	attributed := 0.0
	for _, p := range l.Providers {
		glyph := styleAgent.Render("●")
		if !p.Up {
			glyph = styleErr.Render("○")
		}
		contended := " "
		if p.Contended {
			contended = stylePaused.Render("⏳")
		}
		attributed += p.SpentUSD
		lines = append(lines, fmt.Sprintf("%-*s %-*s %s q%-3d %2d/%-2d %s $%.2f",
			llmProviderNameWidth, truncateTail(p.Name, llmProviderNameWidth),
			llmProviderModelWidth, truncateTail(p.Model, llmProviderModelWidth),
			glyph, p.Queue, p.Inflight, p.Slots, contended, p.SpentUSD))
		if p.Condition != "" {
			lines = append(lines, "  "+styleErr.Render(fmt.Sprintf("%s — %s", p.ConditionDetail, p.ConditionRemedy)))
		}
	}
	// The (unattributed) remainder (US4/US6): legacy-metered spend the
	// per-provider rows don't cover. A cent of floating-point noise never
	// earns its own row.
	if rem := l.Spent - attributed; rem > 0.005 {
		lines = append(lines, styleDim.Render(fmt.Sprintf("%-*s %-*s   %-4s %-5s   $%.2f",
			llmProviderNameWidth, "(unattributed)", llmProviderModelWidth, "", "", "", rem)))
	}
	return lines
}

// horizonLines renders the guardian pane's live cognition-horizon block (spec
// 037 US1, FR-006): one row per watched class — its plain-language standing at
// the current effective speed ("thinking at 8x" / "suppressed at 32x") and,
// when suppressed, the remedy. The router's own verdict arithmetic rides as a
// dim detail (already operator-facing telemetry text elsewhere); no raw enum
// strings reach the screen (verdictGlossary posture). nil when the world
// carries no horizon (no-LLM worlds render nothing extra).
func horizonLines(horizon []ipc.HorizonClass, speed string) []string {
	if len(horizon) == 0 {
		return nil
	}
	lines := make([]string, 0, len(horizon)+1)
	lines = append(lines, styleDim.Render("🜂 cognition horizon"))
	for _, e := range horizon {
		lines = append(lines, "  "+horizonRow(e, speed))
	}
	return lines
}

// horizonRow renders one class's standing at the effective speed. A suppressed
// class is warn-styled and carries its remedy plus the verbatim verdict
// arithmetic (dim detail); a thinking class is a plain one-liner. The daemon-
// lifetime "skipped N" count (spec 037 US2) is appended except when it is 0 on
// a thinking class — a class that is neither suppressed now nor has ever been
// suppressed carries no count clutter.
func horizonRow(e ipc.HorizonClass, speed string) string {
	var row string
	if e.Suppressed {
		row = stylePaused.Render(fmt.Sprintf("%s suppressed at %s — %s", e.Class, speed, horizonRemedy(e.Calibrated)))
	} else {
		row = fmt.Sprintf("%s thinking at %s", e.Class, speed)
	}
	if e.Suppressed || e.SuppressedCount > 0 {
		row += styleDim.Render(fmt.Sprintf(" · skipped %d", e.SuppressedCount))
	}
	if e.Suppressed && e.Verdict != "" {
		row += " " + styleDim.Render("("+e.Verdict+")")
	}
	return row
}

// horizonRemedy is the plain-language remedy for a suppressed class (spec 037
// FR-007): an uncalibrated class may still benefit from calibration; a
// calibrated one can only slow down — never told to recalibrate as if it were
// a fix.
func horizonRemedy(calibrated bool) string {
	if calibrated {
		return "slow down"
	}
	return "calibrate or slow down"
}

// guardianStripView renders the always-visible action-budget row
// (panels/guardian-strip.md, spec 050): charge bank, next-regen forecast,
// and standing-order count, each degrading to absence rather than a
// misleading value (contract §2/§4 — data-model.md's presence matrix). No
// faith segment at all — TASK-118 hasn't shipped (research R4.3/spec
// assumption). Shared verbatim by the widescreen composite and the narrow
// fallback (layout.md ruling b: "carried exactly as widescreen does") — one
// renderer, two call sites (research R5).
func (m Model) guardianStripView(width int) string {
	if m.status == nil {
		// Pre-status (connecting): the row is present but blank — layout
		// stays stable, no invented zeros (US2 AS-3, contract §2).
		return ""
	}

	charges := m.status.Clock.GuardianCharges
	segments := []string{
		fmt.Sprintf("%s%s (%d/%d)",
			strings.Repeat("⚡", charges),
			strings.Repeat("·", clampInt(sim.GuardianChargeCap-charges, 0, sim.GuardianChargeCap)),
			charges, sim.GuardianChargeCap),
	}
	if charges < sim.GuardianChargeCap {
		// Regen is omitted at a full bank (research R4.1): the executor
		// only fires metatron.charge_regenerated below cap
		// (internal/sim/executor.go), so forecasting an arrival that isn't
		// scheduled would be a lie.
		cadence := int64(sim.GuardianChargeRegenTicks)
		next := m.status.Clock.Tick + (cadence - m.status.Clock.Tick%cadence)
		segments = append(segments, fmt.Sprintf("next +1 @ %s", clock.FormatTOD(int(clock.SecondOfDay(next)))))
	}
	// Standing orders: the client-side replica mirrors metatron.order_*
	// events live (m.replica.Apply, tui.go), the same underlying data the
	// guardian tab's orderStatusLines counts (len(m.consoleOrders), fed
	// from an on-demand IPC peek) — the replica is used here instead
	// because it updates every frame without waiting on a tab visit
	// (US1 AS-2: "the next frame" reflects a spend/order change from ANY
	// dock tab), and per contract §4.2 the two can never actually disagree
	// since they're both projections of the same event stream.
	orders := 0
	if m.replica != nil {
		orders = len(m.replica.GuardianOrders)
	}
	segments = append(segments, fmt.Sprintf("👁 %d standing orders", orders))

	return joinStripSegments(segments, width)
}

// joinStripSegments joins the strip's present segments with " · " (contract
// §2), truncating from the right when the joined line would exceed width
// (edge case: "segments truncate from the right with …" — faith [never
// rendered by this feature] → orders → regen → bank; the bank is the
// headline and the last thing standing). Never returns more than one
// line — if even the bank segment alone doesn't fit, it is hard-clipped
// rather than wrapped (contract §1: the row budget is exactly 1).
func joinStripSegments(segments []string, width int) string {
	if width <= 0 || len(segments) == 0 {
		return ""
	}
	for n := len(segments); n >= 1; n-- {
		joined := strings.Join(segments[:n], " · ")
		if n < len(segments) {
			joined += "…"
		}
		if lipgloss.Width(joined) <= width {
			return joined
		}
		if n == 1 {
			return stripSegmentTail(joined, width)
		}
	}
	return ""
}

// stripSegmentTail hard-truncates s to at most width display columns,
// keeping its head and marking the cut with "…" — the last-resort floor
// under joinStripSegments's segment-dropping truncation, for widths too
// narrow even for the bank segment alone. Trims rune-by-rune re-checking
// lipgloss.Width each step rather than slicing by rune COUNT: the strip's
// own glyphs (⚡) render double-width, so rune count alone isn't a safe
// proxy for display columns here (unlike truncateTail's plain-text input).
func stripSegmentTail(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)+"…") > width {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// guardianBudgetPrefix composes the guardian strip's fold-relocation form
// (patterns/layout.md ruling a step 4; panels/guardian-strip.md
// "Fold-last relocation"): the charge-bank glyph run alone (no numeric —
// this compact form leaves the exact count one tab away, data-model.md's
// "Relocated dormant line") plus the standing-order count, no words. Blank
// (no prefix at all) before the first status snapshot — the same honesty
// rule as the strip itself extends to its folded form: no invented zeros
// for a bank or order count the client doesn't actually know yet.
func (m Model) guardianBudgetPrefix() string {
	if m.status == nil {
		return ""
	}
	charges := m.status.Clock.GuardianCharges
	bank := strings.Repeat("⚡", charges) +
		strings.Repeat("·", clampInt(sim.GuardianChargeCap-charges, 0, sim.GuardianChargeCap))
	orders := 0
	if m.replica != nil {
		orders = len(m.replica.GuardianOrders)
	}
	return fmt.Sprintf("%s · 👁%d", bank, orders)
}

// minibufferView renders the one-line Guardian input at its three states
// (minibuffer.md): dormant, focused (amber border + hint), busy.
func (m Model) minibufferView(width int) string {
	// Total rendered width = inner + 2 (border) — Width()'s own
	// Padding(0,1) eats 2 *more* columns before any text renders, so the
	// true usable text width is inner-2, not inner (B1/B3: this was the
	// off-by-2 that let a long focused input's hint wrap the box to 4
	// rows instead of the fixed 3).
	inner := width - 2
	if inner < 12 {
		inner = 12
	}
	usable := inner - 2
	switch {
	case m.mbFocused:
		hint := "esc release · ⏎ send"
		hintW := lipgloss.Width(hint)
		cursor := "▌"
		// B3: the input text + right-aligned hint must always fit
		// `usable` without wrapping — a wrapped hint silently grows the
		// minibuffer past its fixed 3-row budget (and, combined with
		// B1, is what pushed the header off the top of the terminal).
		// The input display truncates to its visible tail (cursor
		// glued to the right edge, like a normal terminal input line)
		// so the box never needs to wrap; if there's no room for the
		// hint at all, it's dropped rather than ever truncated into
		// illegibility.
		showHint := usable-hintW-1 >= 4
		avail := usable
		if showHint {
			avail = usable - hintW - 1
		}
		left := truncateTail(m.mbInput, avail-lipgloss.Width(cursor)) + cursor
		if !showHint {
			return stylePanelFocus.Width(inner).Render(clipContent(left, inner))
		}
		pad := usable - lipgloss.Width(left) - hintW
		if pad < 1 {
			pad = 1
		}
		line := left + strings.Repeat(" ", pad) + styleDim.Render(hint)
		return stylePanelFocus.Width(inner).Render(clipContent(line, inner))
	case m.mbBusy:
		hint := "esc to background"
		left := "⋮ the " + m.sk().Epithet() + " is answering…"
		pad := usable - lipgloss.Width(left) - lipgloss.Width(hint) - 1
		if pad < 1 {
			pad = 1
		}
		line := styleDim.Render(left) + strings.Repeat(" ", pad) + styleDim.Render(hint)
		return styleBox.Width(inner).Render(clipContent(line, inner))
	case m.mbFlash != "":
		return styleBox.Width(inner).Render(clipContent(styleDim.Render(m.mbFlash), inner))
	default:
		placeholder := "⏎ m — speak with the " + m.sk().Epithet() + "…"
		// Fold-last relocation (patterns/layout.md ruling a step 4,
		// panels/guardian-strip.md): once the widescreen composite's row
		// budget has folded the guardian strip, its content prefixes THIS
		// dormant line instead of disappearing — dormant state only; the
		// focused/busy cases above are untouched. isWidescreen(m.width)
		// guards this to the widescreen composite's own fold arithmetic —
		// the narrow fallback's only minibufferView call site (inside
		// guardianView) always carries the strip as its own row instead
		// (layout.md ruling b), so this branch never fires there.
		if isWidescreen(m.width) && computeRows(m.height, m.wantsLessonRow()).Strip == 0 {
			if prefix := m.guardianBudgetPrefix(); prefix != "" {
				placeholder = prefix + " · " + placeholder
			}
		}
		return styleBox.Width(inner).Render(clipContent(styleDim.Render(placeholder), inner))
	}
}

// truncateTail keeps at most max runes of s, from the end — the visible
// window once a minibuffer input outgrows the display width, cursor glued
// to the right edge (normal terminal input-line behavior).
func truncateTail(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[len(r)-max:])
}

// --- guardian console (pages/guardian-console.md, spec 053 US1) ---
// A first-class full-screen page (research R1), not a dock tab and not a
// solo zoom: document-style turns over the shared transcript, the card-
// composition seam (research R6), the charter/skills read surface (US3),
// and the SAME minibuffer every other page uses as its composer (contract
// §2, "byte-identical... no second input widget").

// consoleCard is the card-composition seam's element interface (contracts/
// console-and-systems.md §2.3, data-model.md "consoleCard (the seam)";
// research R6, the scope-ruling operationalized: D5 assigns the report
// card's CONTENT to TASK-115/127, not this feature). This feature ships
// only the interface, the composition point (between the turn stream and
// the read surface, consoleView below), and its tests — Model.consoleCards
// stays empty for every world this feature ships; the design page's control
// table keeps the card row "unbuilt (wave 3/4)" naming this exact symbol as
// the seam TASK-127's shared report-card renderer plugs into.
type consoleCard interface {
	renderCard(width int) string
}

// consoleCardLines renders Model.consoleCards (always empty this feature)
// as blank-line-separated blocks appended after the turn stream and before
// the read surface (contract §2.3) — the same separation a turn block gets.
func (m Model) consoleCardLines(width int) []string {
	var out []string
	for _, c := range m.consoleCards {
		out = append(out, "", c.renderCard(width))
	}
	return out
}

// consoleTurnLines renders Model.transcript as document-style turn blocks
// (contract §2.2, research R4): a labeled block per conversational turn,
// blank-line separated, width-wrapped — classifyTranscriptLine (shared with
// the compact tab's transcriptRowLines) supplies the label, so the special-
// row vocabulary (⚡/👁/⏲/», the "note" telemetry label the compact tab
// already uses for inline verdict rows) renders unlabeled and inline here
// exactly as it does there: one shared vocabulary, two renderings. Model.
// transcript carries no per-entry timestamp in this client (it is a plain
// []string) — the mockup's "· HH:MM" suffix is representative only;
// omitting it is the honesty rule (R4: "entries that carry no time render
// without the timestamp suffix rather than inventing one"), not a
// placeholder gap.
func consoleTurnLines(transcript []string, width int, sk *skin.Skin) []string {
	if width < 10 {
		width = 10
	}
	var out []string
	for i, l := range transcript {
		label, text, style := classifyTranscriptLine(l, sk)
		if i > 0 {
			out = append(out, "")
		}
		if label == "" {
			out = append(out, style.Render(l))
			continue
		}
		out = append(out, styleHeader.Render(label))
		for _, w := range wrapText(text, width-2) {
			out = append(out, "  "+style.Render(w))
		}
	}
	return out
}

// consoleScrollWindow windows content to exactly rows lines, tail-anchored
// (contract §2.2 "tail-anchored"; data-model.md "consoleScroll"): scroll=0
// shows the most recent `rows` lines; K (handleConsoleKey) increments scroll
// to reveal older lines above, clamped here to content length so it can
// never scroll past the head. Short content (fewer lines than the budget)
// pads at the bottom rather than the top — B1's "short content hugs the
// top, blank fills below" discipline every panel in this package follows
// (mapPanelView/dockPanelView via lipgloss Height(); this is the same rule
// applied to a plain slice instead of a styled box).
func consoleScrollWindow(content []string, scroll, rows int) []string {
	if rows < 1 {
		rows = 1
	}
	maxScroll := len(content) - rows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if scroll > maxScroll {
		scroll = maxScroll
	}
	if scroll < 0 {
		scroll = 0
	}
	end := len(content) - scroll
	if end < 0 {
		end = 0
	}
	start := end - rows
	if start < 0 {
		start = 0
	}
	out := append([]string{}, content[start:end]...)
	for len(out) < rows {
		out = append(out, "")
	}
	return out
}

// charterReadSurfaceLines renders the console's charter/skills read surface
// (FR-004, research R5): charter.md's provenance + binding status, and the
// skills file count + binding status — sourced entirely from the guardian
// status fields consoleStatusMsg already populates for the compact tab's
// header (consoleCharter/consoleCharterLocked/consoleCharterPreset/
// consoleSkills/consoleSkillsLocked); no client-side file parsing (FR-004's
// explicit ruling). Honest lock notices name the unlocking stage via
// skin.StageName, mirroring internal/guardian/charter.go's own
// stageCharter/stageSkills notice wording (stage-1 locks the charter until
// stage-2; stage-1/2 lock skills until stage-3 — the ladder's current,
// hardcoded shape, the same assumption consoleStageSummary already makes)
// through the WORLD skin (m.sk().StageName, spec 052 T004) without a second
// round trip to ask the daemon to repeat itself. While
// SkillsLocked, Status.Skills is the EFFECTIVE (empty) list, not a file
// count off disk (internal/guardian/turn.go Status()) — so the locked
// notice honestly omits a count it does not have, rather than inventing one
// by reading the directory itself.
func (m Model) charterReadSurfaceLines() []string {
	if m.consoleCharter == "" {
		// No status peek has landed yet (not connected, or the first frame
		// before fetchConsoleStatus resolves) — the honest absence, not an
		// invented value.
		return []string{styleDim.Render("charter/skills — unavailable (not connected)")}
	}
	var charterLine string
	switch {
	case m.consoleCharterLocked:
		charterLine = fmt.Sprintf("charter.md — preset-locked to %s; does not bind at this stage — %s unlocks instruction authoring",
			m.consoleCharterPreset, m.sk().StageName("stage-2"))
	case m.consoleCharter == "default charter":
		charterLine = "charter.md — default, binds now  [e] edit ($EDITOR)"
	default:
		charterLine = "charter.md — player-authored, binds now  [e] edit ($EDITOR)"
	}

	var skillsLine string
	switch {
	case m.consoleSkillsLocked:
		skillsLine = fmt.Sprintf("skills/ — locked; does not bind at this stage — %s unlocks skill files", m.sk().StageName("stage-3"))
	case m.consoleSkills == 1:
		skillsLine = "skills/ — 1 file, binds now"
	default:
		skillsLine = fmt.Sprintf("skills/ — %d files, binds now", m.consoleSkills)
	}

	return []string{charterLine, skillsLine}
}

// charterReadSurfaceBox renders the console's charter/skills sub-panel
// (contract §2.4): a bordered box titled "charter · skills" (the mapPanelView
// precedent for a box whose "title" is its first content line, since this
// package's styleBox has no native border-title support), wrapped to width.
func (m Model) charterReadSurfaceBox(width int) string {
	if width < 20 {
		width = 20
	}
	var wrapped []string
	for _, l := range m.charterReadSurfaceLines() {
		wrapped = append(wrapped, wrapText(l, width-4)...)
	}
	content := styleHeader.Render("charter · skills") + "\n" + strings.Join(wrapped, "\n")
	return styleBox.Width(width - 2).Render(clipContent(content, width-2))
}

// consoleFooterView note: the console's footer is rendered by the shared
// footerView() (views.go) — it gains its own case there (contract §2.7),
// checked ahead of inspect/villagers exactly like handleKey's dispatch, so
// there is no separate function here; consoleView calls m.footerView()
// directly, the same call every other top-level page makes.

// consoleView is the guardian console page (pages/guardian-console.md
// "Structure"): header line, document-style turn stream (tail-anchored,
// scrollback via consoleScroll), the card seam, the charter/skills read
// surface, the one-shot post-$EDITOR notice, the standard minibuffer as
// composer, and the standard footer — in that order (contract §2). Renders
// to exactly m.height lines (B1, the same exact-height discipline every
// top-level page in this package holds), computing the turn-stream's row
// budget from what everything else actually measures rather than a fixed
// guess. The help overlay (m.helpOpen), when open over the console, replaces
// only this body region — the same "body replacement" rule widescreenView's
// solo/help branches use (spec.md Edge Cases: "the console is a page, not an
// overlay — '?' over the console behaves as over any page").
func (m Model) consoleView() string {
	width := m.width
	if width < 40 {
		width = 40
	}

	header := m.guardianHeaderLine()
	readSurface := m.charterReadSurfaceBox(width)
	readSurfaceRows := lipgloss.Height(readSurface)

	notice := ""
	noticeRows := 0
	if m.consoleNotice != "" {
		notice = styleDim.Render(m.consoleNotice)
		noticeRows = 1
	}

	// header(1) + blank(1) + read surface + notice + minibuffer(3, its
	// natural rendered height — no explicit .Height() call, same as every
	// other minibufferView call site) + footer(1); the turn stream takes
	// whatever remains (research R1's "renders the transcript tail, not the
	// whole history per frame").
	fixed := 2 + readSurfaceRows + noticeRows + minibufferRows + footerRows
	bodyRows := m.height - fixed
	if bodyRows < 3 {
		bodyRows = 3
	}

	var body string
	if m.helpOpen {
		body = m.helpPanelView(width, bodyRows)
	} else {
		lines := consoleTurnLines(m.transcript, width-2, m.sk())
		lines = append(lines, m.consoleCardLines(width-2)...)
		body = strings.Join(consoleScrollWindow(lines, m.consoleScroll, bodyRows), "\n")
	}

	var b strings.Builder
	b.WriteString(header + "\n\n")
	b.WriteString(body + "\n")
	b.WriteString(readSurface + "\n")
	if notice != "" {
		b.WriteString(notice + "\n")
	}
	b.WriteString(m.minibufferView(width) + "\n")
	b.WriteString(m.footerView())
	return b.String()
}

// --- villagers (panels/dock.md "Tab: villagers"; TASK-56 roster + per-
// villager detail, width- and height-aware) ---

// villagersView is the narrow-fallback pane — same body renderer the dock
// tab and solo view share (dockTabContent), boxed like every narrow pane.
func (m Model) villagersView() string {
	body := m.villagersBody(clampInt(m.width-6, 20, 500), clampInt(m.height-6, 4, 500))
	return styleBox.Render(body)
}

// villagersBody dispatches to the roster or the selected villager's detail
// view (data-model.md "New TUI model state": villDetail).
func (m Model) villagersBody(width, height int) string {
	if m.replica == nil || len(m.replica.Agents) == 0 {
		return styleHeader.Render("VILLAGERS") + "\n\n" + styleDim.Render("waiting for world state…")
	}
	switch {
	case m.villDetail && m.villDecisions:
		return m.villagerDecisionsBody(width, height)
	case m.villDetail:
		return m.villagerDetailBody(width, height)
	default:
		return m.villagerRosterBody(width, height)
	}
}

// villagerRosterBody renders the roster with a selection cursor, dropping
// the least important column first as width narrows (dock.md "wrap/condense
// columns; drop the least important column first when narrow") — the same
// columns and drop-trailing-agents height rule as before TASK-56, plus the
// cursor glyph on the selected row.
func (m Model) villagerRosterBody(width, height int) string {
	sel := clampInt(m.villSelected, 0, len(m.replica.Agents)-1)
	wide := width >= 40
	var lines []string
	for i, a := range m.replica.Agents {
		cursor := "  "
		if i == sel {
			cursor = styleFeedSelect.Render("▌") + " "
		}
		status := "awake"
		switch {
		case a.Dead:
			status = styleErr.Render("dead")
		case a.Asleep:
			status = styleAsleep.Render("asleep")
		}
		if wide {
			goal := "idle"
			if a.Intent != nil {
				goal = a.Intent.Goal
			}
			lines = append(lines, cursor+fmt.Sprintf("%-8s %s · %s · (%d,%d)", a.Name, status, goal, a.X, a.Y))
			lines = append(lines, styleDim.Render(fmt.Sprintf(
				"           health %s food %s rest %s warmth %s morale %s",
				bar(a.Needs.Health), bar(a.Needs.Food), bar(a.Needs.Rest),
				bar(a.Needs.Warmth), bar(a.Needs.Morale))))
			// Carried inventory (spec 012 T043, SC-006): the full raw/refined
			// surface — wood/stone/water/planks/refined stone, the food
			// triplet, and spear count (with the most-worn spear's
			// remaining uses when at least one is carried — Spears is kept
			// sorted ascending by the reducer, so Spears[0] is the one
			// closest to breaking). Leading bulk n/24 (spec 013 T015,
			// SC-006) answers "how full are this villager's hands" from
			// the TUI alone — sim.Bulk is the same derived-load function
			// the reducer/executor clamp against, so the number never
			// drifts from what a gather/craft/give will actually do.
			carry := fmt.Sprintf("bulk %d/%d · carry %dw %dst %dwt %dpl %drs · food %dr/%dc/%dm",
				sim.Bulk(a.Inv), sim.BulkCap,
				a.Inv.Wood, a.Inv.Stone, a.Inv.Water, a.Inv.Planks, a.Inv.RefinedStone,
				a.Inv.FoodRaw, a.Inv.FoodCooked, a.Inv.Meals)
			if n := len(a.Inv.Spears); n > 0 {
				carry += fmt.Sprintf(" · spear %d(%d)", n, a.Inv.Spears[0])
			}
			lines = append(lines, styleDim.Render("           "+carry))
			lines = append(lines, "")
		} else {
			// Narrow dock width: drop goal/position/memory, keep cursor + name + status + health.
			lines = append(lines, cursor+fmt.Sprintf("%-8s %s health %s", a.Name, status, bar(a.Needs.Health)))
		}
	}
	// B1/B5: "VILLAGERS" + blank above spend 2 of `height`'s budget;
	// drop trailing agents (rather than partial rows) if the roster
	// doesn't fit, the same "shed content, never overflow" rule the
	// chronicle and minibuffer follow.
	budget := height - 2
	if budget < 1 {
		budget = 1
	}
	if len(lines) > budget {
		lines = lines[:budget]
	}
	body := strings.TrimRight(strings.Join(lines, "\n"), "\n")
	return styleHeader.Render("VILLAGERS") + "\n\n" + body
}

// villagerDetailBody renders the selected villager within the given
// (width, height) budget. Sections in fixed priority order
// (contracts/state-and-keys.md "Rendering contract"): identity/vitals →
// objective → inventory → beliefs/narrative → memories (most recent
// first). The section list truncates from the bottom — memories shed
// first — so identity/objective/inventory are never pushed off-screen
// (spec "Very short pane height" edge case); every line is width-clipped
// so a long belief/memory/narrative line can never push the panel past its
// column budget either (SC-004).
func (m Model) villagerDetailBody(width, height int) string {
	if height < 1 {
		height = 1
	}
	sel := clampInt(m.villSelected, 0, len(m.replica.Agents)-1)
	a := m.replica.Agents[sel]
	wide := width >= 40

	lines := []string{strings.ToUpper(a.Name), ""}
	lines = append(lines, strings.Split(villagerIdentitySection(a, wide), "\n")...)
	lines = append(lines, "")
	lines = append(lines, strings.Split(villagerObjectiveSection(a), "\n")...)
	lines = append(lines, "")
	lines = append(lines, strings.Split(villagerInventorySection(a), "\n")...)
	if s := villagerBeliefsSection(a); s != "" {
		lines = append(lines, "")
		lines = append(lines, strings.Split(s, "\n")...)
	}
	lines[0] = styleHeader.Render(lines[0])

	if len(lines) > height {
		// Pathological height: the fixed sections themselves don't fit.
		// Shed from the bottom like everywhere else rather than ever
		// emitting more than `height` lines.
		lines = lines[:height]
	} else if remaining := height - len(lines); remaining > 1 {
		if mem := villagerMemoriesLines(a, remaining-1); len(mem) > 0 { // -1: blank separator
			lines = append(lines, "")
			lines = append(lines, mem...)
		}
	}

	for i, l := range lines {
		lines[i] = clipLine(l, width)
	}
	return strings.Join(lines, "\n")
}

// villagerDecisionsBody renders the selected villager's decision chains
// most-recent-first (spec 020, TASK-63, contract R9–R11): a when/class
// header, the stimulus line, each call as "ordinal. tool — phrase (reason)"
// in recorded order, and the terminal outcome line — or a visible
// in-progress marker when no cog.outcome has arrived yet (FR-008).
// Suppression-only chains (router held the thought back — no cog.thought,
// no calls) render as a single "didn't think because…" entry instead
// (renderDecisionChain). Clipped to the row budget with a render-time
// scroll clamp (R10, the same "shed from the budget, never overflow"
// discipline chronicleInspectBody/villagerDetailBody use); a villager with
// no chains gets an explicit empty-state line rather than a blank pane
// (R11). Dead villagers keep their chains — this reads straight from
// m.traces, never m.replica.Agents[sel].Dead.
func (m Model) villagerDecisionsBody(width, height int) string {
	if height < 1 {
		height = 1
	}
	sel := clampInt(m.villSelected, 0, len(m.replica.Agents)-1)
	a := m.replica.Agents[sel]
	header := styleHeader.Render(strings.ToUpper(a.Name) + " · decisions")

	chains := m.traces.chainsFor(sel)
	var content []string
	if len(chains) == 0 {
		content = []string{styleDim.Render("no decisions recorded yet this session")}
	} else {
		for i, c := range chains {
			if i > 0 {
				content = append(content, "")
			}
			content = append(content, renderDecisionChain(c)...)
		}
	}

	budget := height - 2 // header + blank separator line
	if budget < 1 {
		budget = 1
	}
	maxScroll := len(content) - budget
	if maxScroll < 0 {
		maxScroll = 0
	}
	scroll := clampInt(m.villDecisionsScroll, 0, maxScroll)
	end := scroll + budget
	if end > len(content) {
		end = len(content)
	}
	visible := content[scroll:end]

	lines := append([]string{header, ""}, visible...)
	for i, l := range lines {
		lines[i] = clipLine(l, width)
	}
	return strings.Join(lines, "\n")
}

// renderDecisionChain formats one decisionChain for the decisions sub-view
// (contract R9). A fragment whose thought was never seen (mid-cognition
// connect, FR-008) renders its class/stimulus as "unknown" rather than
// blank — honest about what the client actually knows.
func renderDecisionChain(c *decisionChain) []string {
	when := clock.Format(c.Tick)
	if c.Suppressed {
		reason := c.OutcomeReason
		if reason == "" {
			reason = "no reason recorded"
		}
		return []string{fmt.Sprintf("%s · didn't think because %s", when, reason)}
	}
	class := c.Class
	if class == "" {
		class = "unknown"
	}
	stimulus := c.Stimulus
	if stimulus == "" {
		stimulus = "unknown — this chain's thought record wasn't seen"
	}
	lines := []string{fmt.Sprintf("%s · %s", when, class), "  stimulus: " + stimulus}
	for _, call := range c.Calls {
		lines = append(lines, fmt.Sprintf("  %d. %s", call.Ordinal, callLine(call.Tool, call.Verdict, call.Reason)))
	}
	lines = append(lines, "  "+decisionOutcomeLine(c))
	return lines
}

// decisionOutcomeLine is a chain's terminal row: the outcome's plain-language
// phrase plus its reason, or a visible in-progress marker when no
// cog.outcome has arrived yet (FR-008: an in-flight cognition must render
// honestly, never pretend a terminal state).
func decisionOutcomeLine(c *decisionChain) string {
	if c.Outcome == "" {
		return "in progress — no outcome yet"
	}
	line := "outcome: " + verdictPhrase(c.Outcome)
	if c.OutcomeReason != "" {
		line += " (" + c.OutcomeReason + ")"
	}
	return line
}

// villagerIdentitySection is FR-003: name, awake/asleep/dead status,
// position, and needs.
func villagerIdentitySection(a sim.Agent, wide bool) string {
	status := "awake"
	switch {
	case a.Dead:
		status = styleErr.Render("dead")
	case a.Asleep:
		status = styleAsleep.Render("asleep")
	}
	line := fmt.Sprintf("%s · %s · (%d,%d)", a.Name, status, a.X, a.Y)
	needs := fmt.Sprintf("health %s food %s rest %s warmth %s morale %s",
		bar(a.Needs.Health), bar(a.Needs.Food), bar(a.Needs.Rest), bar(a.Needs.Warmth), bar(a.Needs.Morale))
	if !wide {
		return line + "\n" + needs
	}
	return line + "\n" + styleDim.Render(needs)
}

// villagerObjectiveSection is FR-005/FR-006 and US2's three display states
// (data-model.md "Derived display state: objective"): active (Intent !=
// nil), past (LastGoal survives Intent clearing, marked "last"), or "no
// objective yet" when neither has ever been set.
func villagerObjectiveSection(a sim.Agent) string {
	switch {
	case a.Intent != nil:
		return fmt.Sprintf("objective: %s → (%d,%d) (current)", a.Intent.Goal, a.Intent.TargetX, a.Intent.TargetY)
	case a.LastGoal != "":
		return fmt.Sprintf("objective: %s (last, %s)", a.LastGoal, clock.Format(a.LastGoalTick))
	default:
		return "objective: no objective yet"
	}
}

// villagerInventorySection is FR-004: every carried kind itemized with
// counts (spear wear included); empty kinds omitted; an entirely empty
// pack stated plainly rather than rendering nothing.
func villagerInventorySection(a sim.Agent) string {
	var items []string
	add := func(label string, n int) {
		if n > 0 {
			items = append(items, fmt.Sprintf("%s %d", label, n))
		}
	}
	add("wood", a.Inv.Wood)
	add("stone", a.Inv.Stone)
	add("water", a.Inv.Water)
	add("planks", a.Inv.Planks)
	add("refined stone", a.Inv.RefinedStone)
	add("raw food", a.Inv.FoodRaw)
	add("cooked food", a.Inv.FoodCooked)
	add("meals", a.Inv.Meals)
	if n := len(a.Inv.Spears); n > 0 {
		wear := make([]string, n)
		for i, w := range a.Inv.Spears {
			wear[i] = fmt.Sprintf("%d", w)
		}
		items = append(items, fmt.Sprintf("spears %d (uses left: %s)", n, strings.Join(wear, ",")))
	}
	if len(items) == 0 {
		return "inventory: empty pack"
	}
	return "inventory:\n  " + strings.Join(items, "\n  ")
}

// villagerBeliefsSection is FR-008: consolidated beliefs and narrative,
// shown only when present (empty string omits the section silently).
func villagerBeliefsSection(a sim.Agent) string {
	if len(a.Beliefs) == 0 && a.Narrative == "" {
		return ""
	}
	var b strings.Builder
	if a.Narrative != "" {
		b.WriteString("narrative: " + a.Narrative)
	}
	if len(a.Beliefs) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("beliefs:")
		for _, belief := range a.Beliefs {
			b.WriteString(fmt.Sprintf("\n  %s (%d%%)", belief.Statement, belief.Confidence))
		}
	}
	return b.String()
}

// villagerMemoriesLines is FR-007: episodic memories most-recent-first
// (Memories accretes oldest-first, so this walks it in reverse),
// bounded to at most `budget` lines total including its own header — the
// section this detail view sheds first when height runs short. budget < 1
// omits the section entirely; a villager with no memories yet says so
// plainly rather than showing nothing.
func villagerMemoriesLines(a sim.Agent, budget int) []string {
	if budget < 1 {
		return nil
	}
	if len(a.Memories) == 0 {
		return []string{styleHeader.Render("memories") + " " + styleDim.Render("· no memories yet")}
	}
	lines := []string{styleHeader.Render("memories")}
	for i := len(a.Memories) - 1; i >= 0 && len(lines) < budget; i-- {
		lines = append(lines, sim.FormatMemory(a.Memories[i]))
	}
	return lines
}

// bar renders a 0..1000 need as a compact five-cell gauge.
func bar(v int) string {
	filled := v / 200
	if v > 0 && filled == 0 {
		filled = 1
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", 5-filled)
}
