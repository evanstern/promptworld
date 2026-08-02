# Spec 109 — Default new worlds to a local model whose engine honors structured outputs

**Board task:** TASK-184
**Status:** claiming (stub — full spec follows on this branch)

A freshly created world points its local tier at a model that cannot be relied on to
return JSON. This spec changes what `DefaultConfig` ships so the local tier is correct
and fast out of the box, and documents the provider-capability hazard that made the
problem invisible.

Full problem statement, requirements, and success criteria follow in the next commit on
`task-184-default-local-model`.
