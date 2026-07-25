// Command promptworld is the single binary for the promptworld daemon and
// its clients. See specs/001-world-daemon/contracts/cli.md for the contract.
package main

import (
	"fmt"
	"os"
)

const usage = `promptworld — the always-on promptworld daemon and client

A <world> argument below is a name or a path: a name (e.g. "aria") resolves
against the default worlds home (~/.promptworld/worlds, overridable via
PROMPTWORLD_HOME) then the known-worlds list; a path (contains "/", or
starts with "." or "~") is used exactly as given.

Usage:
  promptworld new <name> [--at DIR] [--seed N] [--teaching]
                 [--stage stage-1..stage-4] [--override] [--charter-preset default|tutor]
                                                   create a world by name in the worlds home;
                                                   --stage defaults to stage-1 for a new player,
                                                   else your highest earned stage (spec 046,
                                                   see the stages command below); an unearned
                                                   stage needs --override
  promptworld new <path> [--name NAME] [--seed N] [--teaching] [--stage ...] [--override] [--charter-preset ...]
                                                   create a world at an explicit path (legacy form)
  promptworld migrate <world>                      migrate a stopped older world (v1/v2) to the current format
  promptworld ps [--all] [--json]                  list world daemons machine-wide
  promptworld daemon <world>                       run the daemon in the foreground
  promptworld start <world>                        start a detached daemon
  promptworld stop <world>                         gracefully stop the daemon
  promptworld status <world> [--json]              report world/daemon status
  promptworld ui <world>                           full-screen TUI (map, chronicle, guardian, villagers)
  promptworld attach <world>                       line-mode event stream + commands
  promptworld tail <world> [--since SEQ] [--follow] print events from the log
  promptworld pause <world>                        pause game time
  promptworld resume <world>                       resume game time
  promptworld speed <world> <1x|4x|8x|16x|32x|max>     set game speed
  promptworld teaching <world> [on|off]            show or toggle the teaching-world posture (applies at next daemon start)
  promptworld guardian <world> [message...]        converse with the guardian (no message: status peek)
  promptworld work <world> <snap-time|give|move|remove> ... [--force]
                                                   land a guardian working (--force waives the charge)
  promptworld llm <world> <kind> <prompt...>       one-shot LLM call via the daemon; prints the
                                                   serving provider and any fallback skips
                                                   (kinds: planner, conversation,
                                                    consolidation, narrator, drama)
  promptworld calibrate <world> [--provider name | --tier local|cloud|all] [--samples N]
                                                   benchmark seconds-per-point per declared
                                                   provider, write calibration.json
                                                   (--tier is a deprecated alias, see --help)
  promptworld divergence <world> [--json]          summarize recorded memory-rank divergence
                                                   (shadow-mode gate evidence, per agent/day)
  promptworld stages [--json]                      list the curriculum ladder's four stages —
                                                   identity, concept, grants, unlock evidence,
                                                   and your earned state
`

// dispatch resolves a subcommand name to its handler. Canonical names only in
// the usage text above; the pre-052 fiction names (`guardian`, `miracle`)
// stay as HIDDEN, fully functional compatibility aliases (spec 052 FR-008) —
// old scripts and habits keep working forever, the frozen-vocabulary ruling.
func dispatch(cmd string) (func([]string) error, bool) {
	switch cmd {
	case "new":
		return cmdNew, true
	case "migrate":
		return cmdMigrate, true
	case "ps":
		return cmdPs, true
	case "daemon":
		return cmdDaemon, true
	case "start":
		return cmdStart, true
	case "stop":
		return cmdStop, true
	case "status":
		return cmdStatus, true
	case "ui":
		return cmdUI, true
	case "attach":
		return cmdAttach, true
	case "tail":
		return cmdTail, true
	case "pause":
		return func(args []string) error { return cmdTimeCtl("pause", args) }, true
	case "resume":
		return func(args []string) error { return cmdTimeCtl("resume", args) }, true
	case "speed":
		return cmdSpeed, true
	case "teaching":
		return cmdTeaching, true
	case "llm":
		return cmdLLM, true
	case "calibrate":
		return cmdCalibrate, true
	case "divergence":
		return cmdDivergence, true
	case "stages":
		return cmdStages, true
	case "guardian", "metatron": // "metatron" is the hidden compat alias (spec 052 FR-008)
		return cmdGuardian, true
	case "work", "miracle": // "miracle" is the hidden compat alias (spec 052 FR-008)
		return cmdWork, true
	}
	return nil, false
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, args := os.Args[1], os.Args[2:]
	if cmd == "help" || cmd == "-h" || cmd == "--help" {
		fmt.Print(usage)
		return
	}
	fn, ok := dispatch(cmd)
	if !ok {
		fmt.Fprintf(os.Stderr, "promptworld: unknown command %q\n\n%s", cmd, usage)
		os.Exit(2)
	}
	if err := fn(args); err != nil {
		fmt.Fprintf(os.Stderr, "promptworld %s: %v\n", cmd, err)
		os.Exit(1)
	}
}
