package main

// `promptworld fork` (spec 076 FR-001..003): copy a world at its latest
// hash-valid snapshot under a fresh identity, so parent and fork run side
// by side — the iteration rung: edit the fork's charter, run both, compare.
// The ceremony itself is world.Fork (internal/world/fork.go); this file is
// argument classification (the `new` conventions) and the human summary.

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/evanstern/promptworld/internal/clock"
	"github.com/evanstern/promptworld/internal/world"
	"github.com/evanstern/promptworld/internal/worlds"
)

func cmdFork(args []string) error {
	fs := flag.NewFlagSet("fork", flag.ContinueOnError)
	at := fs.String("at", "latest-snapshot", "fork point (v1 supports only latest-snapshot)")

	// Two positionals with flags anywhere: collect leading non-flag args,
	// parse the rest, then take any trailing positionals back.
	var pos []string
	rest := args
	for len(rest) > 0 && !strings.HasPrefix(rest[0], "-") {
		pos = append(pos, rest[0])
		rest = rest[1:]
	}
	if err := fs.Parse(rest); err != nil {
		return err
	}
	pos = append(pos, fs.Args()...)
	if len(pos) != 2 {
		return fmt.Errorf("usage: promptworld fork <world> <new-name> [--at latest-snapshot]")
	}

	// v1 forks at the latest snapshot only (FR-002): --at exists so the
	// contract is explicit, and any other value is an informed refusal.
	if *at != "latest-snapshot" {
		return fmt.Errorf("--at %q is not supported: v1 forks at the latest snapshot only (--at latest-snapshot, the default) — mid-log and chosen-snapshot forking is a documented follow-on, not a v1 capability", *at)
	}

	srcDir, err := resolveWorld(pos[0])
	if err != nil {
		return err
	}

	// <new-name> follows `new`'s exact conventions (FR-001): a bare name is
	// validated and placed in the worlds home; a path form is used exactly,
	// name from the basename.
	destArg := pos[1]
	var destDir, newName, startHint string
	if worlds.IsPathArg(destArg) {
		destDir = destArg
		newName = filepath.Base(filepath.Clean(destArg))
		startHint = destArg
	} else {
		if err := worlds.ValidateName(destArg); err != nil {
			return err
		}
		newName = destArg
		home, err := worlds.WorldsHome()
		if err != nil {
			return err
		}
		destDir = filepath.Join(home, destArg)
		startHint = destArg
	}

	res, err := world.Fork(srcDir, destDir, newName)
	if err != nil {
		return err
	}
	fmt.Print(forkSummary(res, pos[0], startHint))
	return nil
}

// forkSummary renders the completed fork for the player (US1 scenario 5):
// the boundary in game time, what carried, what was truncated, the lineage
// and wallet facts, and how to start both worlds. Factored out of cmdFork
// so tests pin the summary contract directly.
func forkSummary(res *world.ForkResult, parentArg, startHint string) string {
	day, h, m, _ := clock.GameTime(res.ForkTick)
	var b strings.Builder
	fmt.Fprintf(&b, "forked %q → %q at day %d, %02d:%02d (tick %d)\n",
		res.ParentName, res.Name, day, h, m, res.ForkTick)
	fmt.Fprintf(&b, "  events carried: 1..%d, plus the world.forked lineage event\n", res.ForkSeq)
	if res.TruncatedTail > 0 {
		fmt.Fprintf(&b, "  truncated tail: %d parent events past the snapshot boundary did not carry — the fork is exactly at the snapshot\n", res.TruncatedTail)
	}
	fmt.Fprintf(&b, "  lineage: recorded in the fork's log and manifest (forked from %q)\n", res.ParentName)
	if res.BoundaryEnded {
		b.WriteString("  note: the boundary state carries an ended run, so the fork is born ended — forking an earlier snapshot is a documented follow-on\n")
	}
	if res.SpendCarried {
		b.WriteString("  wallet: the parent's LLM spend meter carried as of fork time — forking never mints fresh budget\n")
	}
	fmt.Fprintf(&b, "start them side by side:\n  promptworld start %s\n  promptworld start %s\n", parentArg, startHint)
	return b.String()
}
