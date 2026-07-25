# Contract: report schema — `--json` output

One JSON object on stdout (human-readable lines go to stdout without `--json`; errors
always to stderr). Shapes mirror [data-model.md](../data-model.md); this file is the
wire-format authority.

```jsonc
{
  "mode": "session",                    // "session" | "worktree" | "pr"
  "verdict": "warnings",                // "pass" | "warnings" | "blocked"
  "exitCode": 0,                        // 0 | 1 (2 never produces a report)
  "originMain": "9904219…40hex",        // fetched origin/main tip the run used
  "unverifiedAgainstRemote": false,     // session mode; true after fetch failure
  "root": {
    "onMain": true,
    "behindBy": 0,
    "aheadBy": 0,
    "fastForwarded": false,
    "dirty": false
  },
  "branches": [                          // [] in worktree mode; single entry in pr mode
    {
      "name": "task-107-tuning-manifest",
      "tip": "cb48d08…40hex",
      "worktree": ".worktrees/task-107",
      "task": "TASK-107",
      "mergeBase": "…40hex",
      "baseLag": 2,
      "dirty": false,
      "changedFiles": ["specs/048-tuning-manifest/plan.md"],
      "cleanupEligible": false,
      "cleanupReason": null              // "ancestor" | "empty-contribution" | null
    }
  ],
  "matrix": [                            // session mode only; lexicographic by (a, b)
    { "a": "task-107-tuning-manifest", "b": "origin/main", "conflict": false, "files": [] },
    { "a": "task-107-tuning-manifest", "b": "task-124-chronicle-jump-to-source",
      "conflict": true, "files": ["internal/tui/view.go"] }
  ],
  "grounding": [                         // session mode only
    { "name": "wiki",        "checker": "internal",  "stale": true,
      "touched": ["internal/bundle/manifest.go"] },
    { "name": "player-docs", "checker": "delegated", "stale": false, "touched": [] },
    { "name": "tui-design",  "checker": "absent",    "stale": null,  "touched": [] }
  ],
  "findings": [                          // ordered block → warn → info; stable within severity
    {
      "severity": "warn",                // "block" | "warn" | "info"
      "gate": "session",
      "rule": "pairwise-conflict",       // stable rule ids: data-model.md severity table
      "message": "task-107 and task-124 will conflict on internal/tui/view.go whichever merges first",
      "evidence": ["internal/tui/view.go", "task-107-tuning-manifest", "task-124-chronicle-jump-to-source"],
      "task": "TASK-107",
      "fingerprint": "a1b2c3d4e5f6",
      "noteWritten": false               // true only when --notes appended it this run
    }
  ],
  "cleanupPrescriptions": [              // session mode; exact commands, applied only with --apply-cleanup
    { "branch": "task-130-player-docs-keys-reference",
      "commands": ["git worktree remove .worktrees/task-130",
                    "git branch -d task-130-player-docs-keys-reference"],
      "applied": false }
  ]
}
```

## Stability guarantees

- `rule` ids, field names, and exit-code semantics are the contract; adding fields is
  non-breaking, removing/renaming is breaking and requires a spec amendment.
- Ordering is deterministic (FR-012): `branches` lexicographic by name, `matrix` by
  `(a, b)`, `findings` by (severity rank, rule, first evidence entry).
- `exitCode` in the report always matches the process exit status; usage/environment
  errors (exit 2) emit a single-line message to stderr and no JSON.
