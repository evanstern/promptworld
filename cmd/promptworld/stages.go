package main

// `promptworld stages` (spec 046 US1, T008, R9): the ladder's front door —
// an informed identity table, never a difficulty menu (FR-003, TASK-68 AC
// #7). Each stage states plainly what it teaches, what the world grants, and
// what evidence unlocks the next; earned state comes from the per-user
// unlocks record (worlds.LoadUnlocks) — a missing/corrupt/unresolvable
// record means nothing extra is earned (stage-1 is always offered: it is
// every player's floor, asked of no one).

import (
	"flag"
	"fmt"
	"strings"

	"github.com/evanstern/promptworld/internal/skin"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

// stageLadderInfo is the ladder table's SKIN-INDEPENDENT content (spec.md
// "The ladder" table): the concept taught, what the world grants, and what
// evidence unlocks the next stage. Contrast internal/skin.StageIdentity,
// which is skin DATA (display name + one-line identity) — this table is
// substrate fact, invariant across skins.
type stageLadderInfo struct {
	Concept        string
	Grants         string
	UnlockEvidence string // "" only for stage-4 (graduation — nothing unlocks past it)
}

// stageOrder is the ladder's presentation order — always all four stages,
// per FR-003 ("later stages are visible with their identity, concept, and
// unlock evidence stated").
var stageOrder = []string{world.Stage1, world.Stage2, world.Stage3, world.Stage4}

// stagesLadder mirrors spec.md's ladder table (client-approved 2026-07-25,
// AC #5) plus the ratified stage-1 ceiling amendment (TASK-119 board
// artifact "first-night teaches visions+orders" — standing orders joined the
// stage-1 grant, contracts/stage-gating.md).
var stagesLadder = map[string]stageLadderInfo{
	world.Stage1: {
		Concept: "conversational prompting: asking well, watching outcomes, iterating",
		Grants: "the base conversational guardian + basic query/nudge tools (visions, omens, " +
			"and the watch — monitor_and_act/cancel_order); instruction files are locked " +
			"(the default or tutor charter is in force, edits get an honest notice)",
		UnlockEvidence: "pass a stage-1 scenario (the first-night exercise)",
	},
	world.Stage2: {
		Concept: "instruction authoring: durable behavior lives in an authored instruction file",
		Grants:  "stage-1 grants + charter editing unlocked",
		UnlockEvidence: "pass a stage-2 scenario while a player-authored charter revision " +
			"is in force",
	},
	world.Stage3: {
		Concept: "capability design: what the guardian can do is itself authored — skill " +
			"files + tool grants",
		Grants: "stage-2 grants + skill files compose + the gated tool manifest opens",
		UnlockEvidence: "pass a stage-3 scenario in which a player-granted tool's act " +
			"contributes to the pass",
	},
	world.Stage4: {
		Concept:        "mastery: indirect influence at world scale; the ambient world as the endgame",
		Grants:         "the full tool roster, including capstone capabilities (canonization)",
		UnlockEvidence: "", // graduation (synthesis decision 3) — nothing unlocks past it
	},
}

// stageEarned reports whether stage is offered WITHOUT an override: stage-1
// is every player's floor (asked of no one — R9's "default stage-1 for new
// players" made unconditional rather than conditioned on an empty record),
// every other stage needs an unlocks-record entry.
func stageEarned(u *worlds.Unlocks, stage string) bool {
	return stage == world.Stage1 || u.Earned(stage)
}

// highestEarnedStage returns the highest-earned stage in stageOrder for
// `promptworld new`'s default --stage selection (R9): stage-1 for a brand
// new player (nothing else earned), else the highest stage the unlocks
// record proves.
func highestEarnedStage(u *worlds.Unlocks) string {
	highest := world.Stage1
	for _, id := range stageOrder {
		if stageEarned(u, id) {
			highest = id
		}
	}
	return highest
}

// stageJSON is the --json twin's per-stage row.
type stageJSON struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Line           string `json:"line"`
	Concept        string `json:"concept"`
	Grants         string `json:"grants"`
	UnlockEvidence string `json:"unlock_evidence,omitempty"`
	Earned         bool   `json:"earned"`
	// ProvingWorld/Exercise are set only when Earned by an unlocks-record
	// entry (never for stage-1's unconditional floor) — the audit pointer
	// FR-008/contracts/unlocks-record.md rule 5 asks status surfaces to show.
	ProvingWorld string `json:"proving_world,omitempty"`
	Exercise     string `json:"exercise,omitempty"`
}

func cmdStages(args []string) error {
	fs := flag.NewFlagSet("stages", flag.ContinueOnError)
	asJSON := fs.Bool("json", false, "machine-readable output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	unlocks := worlds.LoadUnlocks()

	if *asJSON {
		rows := make([]stageJSON, 0, len(stageOrder))
		for _, id := range stageOrder {
			si, _ := skin.Stage(id)
			info := stagesLadder[id]
			row := stageJSON{
				ID: id, Name: si.Name, Line: si.Line,
				Concept: info.Concept, Grants: info.Grants, UnlockEvidence: info.UnlockEvidence,
				Earned: stageEarned(unlocks, id),
			}
			if e, ok := unlocks.Entries[id]; ok {
				row.ProvingWorld, row.Exercise = e.World, e.Exercise
			}
			rows = append(rows, row)
		}
		return printJSON(rows)
	}

	var b strings.Builder
	for _, id := range stageOrder {
		si, _ := skin.Stage(id)
		info := stagesLadder[id]
		fmt.Fprintf(&b, "%s — %s (%s)\n", si.Name, si.Line, id)
		fmt.Fprintf(&b, "  teaches: %s\n", info.Concept)
		fmt.Fprintf(&b, "  grants: %s\n", info.Grants)
		if info.UnlockEvidence != "" {
			fmt.Fprintf(&b, "  unlocked by: %s\n", info.UnlockEvidence)
		} else {
			b.WriteString("  unlocked by: nothing — this is graduation\n")
		}
		switch e, ok := unlocks.Entries[id]; {
		case id == world.Stage1:
			b.WriteString("  earned: yes (every player's floor)\n")
		case ok:
			fmt.Fprintf(&b, "  earned: yes (proven in %q via the %s exercise)\n", e.World, e.Exercise)
		default:
			b.WriteString("  earned: no — choosable only with new --stage " + id + " --override\n")
		}
		b.WriteString("\n")
	}
	fmt.Print(b.String())
	return nil
}
