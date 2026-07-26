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

// The ladder table's skin-independent content RELOCATED to internal/world
// (spec 063 T014): the help overlay's guardian section (D9) reads the same
// table this command renders from, and internal/tui cannot import package
// main — one source, two surfaces. These aliases keep this file's rendering
// code shaped as before.
var (
	stageOrder   = world.StageOrder
	stagesLadder = world.StagesLadder
)

// highestEarnedStage returns the highest-earned stage in stageOrder for
// `promptworld new`'s default --stage selection (R9): stage-1 for a brand
// new player (nothing else earned), else the highest stage the unlocks
// record proves. The earned rule itself is worlds.Unlocks.StageEarned (spec
// 078 FR-003, relocated from this file's own former stageEarned so the TUI's
// forward-ladder section can share it — one source, two surfaces).
func highestEarnedStage(u *worlds.Unlocks) string {
	highest := world.Stage1
	for _, id := range stageOrder {
		if u.StageEarned(id) {
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
				Earned: unlocks.StageEarned(id),
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
