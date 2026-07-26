package guardian

// The tutor guide seam (spec 063 US3, T007/T008): persona.TutorGuide is
// compiled game substrate composed ONLY on tutor-preset worlds, in the
// editable zone — after the charter, persona SOULs, and skin voice, before
// the skill files — with the fixed frame last on every path (SC-003); a
// non-tutor world composes byte-identically to pre-feature. The orientation
// fixture check (T008) is eval-style: canned first-night questions compose
// with the guide and explain grounding present, no live model.

import (
	"context"
	"strings"
	"testing"

	"github.com/evanstern/promptworld/internal/persona"
	"github.com/evanstern/promptworld/internal/tool"
	"github.com/evanstern/promptworld/internal/toolloop"
)

const guideMarker = "--- guide (game-authored) ---"

// TestTutorGuideCompositionOrder (T007): the guide composes after the
// charter, the persona SOULs, and the skin voice, before the skill files —
// and the fixed frame still lands last.
func TestTutorGuideCompositionOrder(t *testing.T) {
	roster := tool.LoopRosterGuardian()
	skills := []skillFile{{"10-x.md", "SKILLBODY"}}
	souls := []string{"SOULFRAG", "VOICEFRAG"}
	prompt := buildTurnSystemPrompt(false, persona.TutorCharter, persona.TutorGuide, skills, roster, souls...)

	idx := func(s string) int {
		at := strings.Index(prompt, s)
		if at < 0 {
			t.Fatalf("prompt missing %q", s)
		}
		return at
	}
	charterAt := idx("# The Charter of the Guardian")
	soulAt := idx("SOULFRAG")
	voiceAt := idx("VOICEFRAG")
	guideAt := idx(guideMarker)
	skillAt := idx("SKILLBODY")
	frameAt := idx(guardianNonNegotiables)
	if !(charterAt < soulAt && soulAt < voiceAt && voiceAt < guideAt && guideAt < skillAt && skillAt < frameAt) {
		t.Errorf("composition order wrong: charter %d, soul %d, voice %d, guide %d, skill %d, frame %d",
			charterAt, soulAt, voiceAt, guideAt, skillAt, frameAt)
	}
	// The frame is beneath the guide: no guide byte may follow the frame.
	if strings.LastIndex(prompt, guideMarker) > frameAt {
		t.Error("a guide block appears after the fixed frame")
	}
	fixedFrameLast(t, prompt, guideMarker)
}

// TestNonTutorByteIdentity (T007, SC-003/FR-004): the guide contributes
// exactly its own block and nothing else — removing it from the tutor
// composition yields the guide-less composition byte-for-byte, and a
// guide-less composition carries no guide marker at all.
func TestNonTutorByteIdentity(t *testing.T) {
	roster := tool.LoopRosterGuardian()
	withGuide := buildTurnSystemPrompt(false, persona.DefaultCharter, persona.TutorGuide, nil, roster)
	without := buildTurnSystemPrompt(false, persona.DefaultCharter, "", nil, roster)
	if strings.Contains(without, guideMarker) {
		t.Error("guide-less composition carries the guide marker")
	}
	block := "\n\n" + guideMarker + "\n" + persona.TutorGuide
	if strings.Replace(withGuide, block, "", 1) != without {
		t.Error("the guide changes prompt bytes beyond its own block")
	}
	// The wrapper (every pre-063 call site) composes guide-less.
	if turnSystemPrompt(persona.DefaultCharter, nil, roster) != without {
		t.Error("turnSystemPrompt is not the guide-less composition")
	}
}

// TestTutorGuideCapDiscipline (T007): the compiled guide honors the same
// size cap as the charters (research R3's ≤4,000-char discipline).
func TestTutorGuideCapDiscipline(t *testing.T) {
	if len(persona.TutorGuide) > persona.CharterMaxChars {
		t.Errorf("TutorGuide is %d chars, cap %d", len(persona.TutorGuide), persona.CharterMaxChars)
	}
	if strings.TrimSpace(persona.TutorGuide) == "" {
		t.Fatal("TutorGuide is empty")
	}
}

// captureSystemLoop scripts a loop that records the composed Job for
// composition assertions and converses.
func captureSystemLoop(mt *Guardian, sink *toolloop.Job) func(context.Context, toolloop.Job) (toolloop.Result, error) {
	return func(ctx context.Context, j toolloop.Job) (toolloop.Result, error) {
		*sink = j
		resp, err := bridgeSubmit(mt, ctx, j)
		if err != nil {
			return toolloop.Result{Term: termForErr(err)}, err
		}
		return toolloop.Result{Final: resp.Text, Term: toolloop.TermModelDone}, nil
	}
}

// TestTutorGuidePresetScoped (T007): the LIVE turn assembly keys the guide
// on the world's charter preset — a tutor-preset world composes it (stage-1
// lock serving persona.TutorCharter), a default-preset world does not.
func TestTutorGuidePresetScoped(t *testing.T) {
	var tutorJob toolloop.Job
	mt, _, _, _ := newTestGuardian(t, "welcome")
	mt.SetStage("stage-1", "tutor")
	mt.charterFP = "" // fresh mirror; this fixture's charter differs from the default
	mt.runLoop = captureSystemLoop(mt, &tutorJob)
	if _, err := mt.Turn(context.Background(), "how do I play?"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(tutorJob.System, guideMarker) || !strings.Contains(tutorJob.System, "The Tutor's Guide") {
		t.Error("tutor-preset turn did not compose the guide")
	}
	// Beneath the charter, above the fixed frame (US3 AS-1).
	charterAt := strings.Index(tutorJob.System, "# The Charter of the Guardian")
	guideAt := strings.Index(tutorJob.System, guideMarker)
	frameAt := strings.Index(tutorJob.System, guardianNonNegotiables)
	if !(charterAt >= 0 && charterAt < guideAt && guideAt < frameAt) {
		t.Errorf("guide out of zone: charter %d, guide %d, frame %d", charterAt, guideAt, frameAt)
	}

	var defaultJob toolloop.Job
	mt2, _, _, _ := newTestGuardian(t, "welcome")
	mt2.SetStage("stage-1", "")
	mt2.runLoop = captureSystemLoop(mt2, &defaultJob)
	if _, err := mt2.Turn(context.Background(), "how do I play?"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(defaultJob.System, guideMarker) {
		t.Error("default-preset turn composed the tutor guide")
	}
}

// TestTutorOrientationFixture (T008, SC-003's fixture half): the canned
// first-night questions compose a turn whose grounding is PRESENT — the
// guide's how-to-tutor contract in the system prompt (orient, mechanics via
// explain, ? for the screen), the explain tool declared on the roster, and
// the player's question as the seed's directive — so a competent orientation
// answer needs nothing the prompt doesn't carry. Fixture-level; no live
// model in CI (the behavior-affecting prompt quality follows the TASK-73
// eval-gated precedent where a live eval exists).
func TestTutorOrientationFixture(t *testing.T) {
	for _, ask := range []string{
		"how do I play?",
		"what does a vision cost?",
		"what can you actually do?",
	} {
		var job toolloop.Job
		mt, _, _, _ := newTestGuardian(t, "let me show you")
		mt.SetStage("stage-1", "tutor")
		mt.charterFP = ""
		mt.runLoop = captureSystemLoop(mt, &job)
		if _, err := mt.Turn(context.Background(), ask); err != nil {
			t.Fatal(err)
		}

		// The guide's how-to-tutor contract composes.
		for _, want := range []string{
			"Orient before anything else",
			"Mechanics facts come from the explain tool",
			"? help overlay",
		} {
			if !strings.Contains(job.System, want) {
				t.Errorf("ask %q: system prompt missing guide contract line %q", ask, want)
			}
		}
		// The explain grounding is declared: on the roster and in the read
		// paragraph (stage-1 ceiling includes explain, spec 063).
		var hasExplain bool
		for _, tl := range job.Roster {
			if tl.Name == "explain" {
				hasExplain = true
			}
		}
		if !hasExplain {
			t.Errorf("ask %q: explain absent from the tutor world's declared roster", ask)
		}
		if !strings.Contains(job.System, "explain(topic)") {
			t.Errorf("ask %q: read guidance absent from the tutor prompt", ask)
		}
		// The player's question is the seed's directive.
		if !strings.Contains(job.Seed, ask) {
			t.Errorf("ask %q: seed does not carry the question", ask)
		}
	}
}
