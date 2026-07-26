# Spec 080 — paused-label lanes downgrade to info (TASK-155)

Claim stub per spec 065; spec.md follows on this number.

Counterpart of praxisflux TASK-55 (specs/015-paused-lane-marker in the praxis
repo): check-merge-drift.mjs downgrades a paused task's branch/worktree
findings from blocking to info in session/worktree/pr modes, pause detected by
the `paused` label in the task file's frontmatter `labels:` list.
