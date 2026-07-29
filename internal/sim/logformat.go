package sim

// The log-format-1 → 2 rename table (spec 094): the guardian rename that
// retired the spec-052 freeze. This map is the COMPLETE inventory of
// persisted event-type names that changed between store.LogFormatLegacy and
// store.LogFormatVersion 2 — the translating migration (world.Migrate)
// rewrites exactly these type-column values and nothing else; every seq,
// tick, and payload is preserved byte-for-byte.
//
// DOCTRINE (spec 094, the format.go comment's sim-side half): renaming any
// persisted event type is a log-format break — bump store.LogFormatVersion,
// extend a rename table like this one, and let the migration translate. A
// SEMANTIC break (a payload's meaning changes, a reducer arm re-derives
// differently) is NOT a rename and must snapshot-cut instead — see
// world/migrate.go's decision rule. Never alias at read: the load gate
// (store.VerifyLogFormat + world.Open) exists so untranslated logs are
// refused, not reinterpreted.
//
// These 13 strings are the ONLY sanctioned home of the metatron.* event
// vocabulary in emit/apply code (SC-002); the fiction-denylist sweep allows
// them here as frozen serialized forms.
var LogFormatV1Renames = map[string]string{
	"metatron.charge_regenerated": "guardian.charge_regenerated",
	"metatron.nudged":             "guardian.nudged",
	"metatron.place_revealed":     "guardian.place_revealed",
	"metatron.order_placed":       "guardian.order_placed",
	"metatron.order_triggered":    "guardian.order_triggered",
	"metatron.order_cancelled":    "guardian.order_cancelled",
	"metatron.order_expired":      "guardian.order_expired",
	"metatron.charter_observed":   "guardian.charter_observed",
	"metatron.skills_observed":    "guardian.skills_observed",
	"metatron.time_snapped":       "guardian.time_snapped",
	"metatron.item_granted":       "guardian.item_granted",
	"metatron.entity_moved":       "guardian.entity_moved",
	"metatron.entity_removed":     "guardian.entity_removed",
}

// CanonicalEventType maps a log-format-1 event-type name to its current
// name, passing every current name through unchanged.
//
// This is for persisted REFERENCE strings only — EvidenceRef.Type values
// living inside payloads the translating migration preserved byte-for-byte
// (rewriting them would change replayed state and break the migration's
// state-hash identity proof, FR-004). Readers of such references (the
// stage-2 unlock conjunct, evidence rendering) normalize through this before
// comparing. The log's own type column is NEVER normalized at read — that
// would be the aliasing the operator ruled out; the load gate guarantees
// only translated logs are opened.
func CanonicalEventType(t string) string {
	if renamed, ok := LogFormatV1Renames[t]; ok {
		return renamed
	}
	return t
}
