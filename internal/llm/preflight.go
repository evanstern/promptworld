package llm

// Provider-health preflight (spec 034 US1, T006/T007): a boot-time + periodic
// probe that verifies each locally-served model actually exists on its endpoint,
// so a fresh world whose configured model is absent is LOUD instead of silently
// brain-dead. It classifies each openai_compat provider as healthy /
// model-missing / endpoint-unreachable / listing-unsupported and reconciles the
// provider's operator-facing condition through the Phase 2 machinery
// (raiseCondition/clearCondition, which fire the transition hook). The probe
// lives entirely outside the sim loop — like the rest of the orchestrator — and
// never blocks or fails boot (FR-002): its worst case is a warning.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// preflightInterval and preflightTimeout are package VARS (not consts) so tests
// can compress the clock — a 60s re-probe loop against a 5s timeout is
// untestable at real time. preflightInterval is the re-probe cadence while a
// preflight condition is active (spec 034 R3: only while active, so a healthy
// world makes zero steady-state probe traffic). preflightTimeout bounds a single
// models-list probe (contract: ≤5s), well under the interval so a hung endpoint
// never stalls the loop.
var (
	preflightInterval = 60 * time.Second
	preflightTimeout  = 5 * time.Second
)

// preflightClient is the probe's HTTP client. Per-request context deadlines
// (preflightTimeout) do the timing out — the var is read at call time so a test
// that shrinks preflightTimeout takes effect without rebuilding the client.
var preflightClient = &http.Client{}

// preflightLogf emits the preflight's low-key operator lines: the
// listing-unsupported skip note and the active re-probe REPEAT of a standing
// warning. It defaults to stdout in the daemon-boot line style (daemon stdout is
// redirected to daemon.log), and the daemon leaves it as-is. It is deliberately
// NOT the condition hook: the hook fires on TRANSITIONS only (raise / reclassify
// / clear), so a steady condition never spams the durable daemon.llm_warning
// event stream — repeat-loudness rides this plain log line plus the persistent
// status fields instead. A package var so tests can capture or silence it.
var preflightLogf = func(format string, args ...any) {
	fmt.Printf("daemon: "+format+"\n", args...)
}

// probeResult is one models-list probe's classification (contracts/
// provider-conditions.md). Every outcome is a classification the caller acts on
// — a probe never surfaces a transport error to its caller, it maps to
// probeUnreachable.
type probeResult int

const (
	probeHealthy     probeResult = iota // 2xx listing that contains the configured model id
	probeMissing                        // 2xx valid listing, model id absent → model-missing
	probeUnreachable                    // transport error / timeout → endpoint-unreachable
	probeUnsupported                    // non-2xx, or 2xx without the {"data":[…]} listing shape → skip, no condition
)

// probeModels performs one GET {endpoint}/models against a provider and
// classifies the result per contracts/provider-conditions.md. It applies the
// same auth rule as the chat-completions transport (providers.go do()): a
// Bearer header only when a key resolves, none for open endpoints. It never
// blocks longer than preflightTimeout and never returns an error — a transport
// failure IS the classification (probeUnreachable).
func (o *Orchestrator) probeModels(ctx context.Context, p *provider) probeResult {
	url := strings.TrimRight(p.cfg.Endpoint, "/") + "/models"
	rctx, cancel := context.WithTimeout(ctx, preflightTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(rctx, http.MethodGet, url, nil)
	if err != nil {
		// A malformed endpoint URL is, operationally, an endpoint we cannot reach.
		return probeUnreachable
	}
	if key := p.cfg.key(); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := preflightClient.Do(req)
	if err != nil {
		return probeUnreachable
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Non-2xx: the endpoint answered but does not serve a standard listing at
		// this path (router variance). Treat as unsupported — NEVER a false
		// model-missing (contract R2); the runtime tool-silence net still applies.
		return probeUnsupported
	}
	// Distinguish a valid-but-empty listing ({"data":[]} → the model is genuinely
	// missing) from a non-listing 2xx body ({} or garbage → the endpoint does not
	// speak the OpenAI listing shape). Decoding into a map first lets us test for
	// the "data" key's PRESENCE, which a plain struct decode would silently paper
	// over (a missing key and an empty array both yield a nil slice).
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return probeUnsupported
	}
	dataRaw, ok := raw["data"]
	if !ok {
		return probeUnsupported // valid JSON, but not the {"data":[…]} listing shape
	}
	var data []struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(dataRaw, &data); err != nil {
		return probeUnsupported // "data" present but not an array of {id} entries
	}
	for _, m := range data {
		if m.ID == p.cfg.Model {
			return probeHealthy
		}
	}
	return probeMissing
}

// preflightProbe probes one provider and reconciles its condition slot (spec 034
// T006/T007): the classification drives raise / reclassify / clear through the
// Phase 2 machinery, which fires the transition hook on genuine changes only. It
// returns whether the provider still holds an ACTIVE preflight condition after
// this probe, so the lifecycle loop knows whether to keep re-probing it. Remedy
// wording follows contracts/provider-conditions.md.
func (o *Orchestrator) preflightProbe(ctx context.Context, p *provider) (active bool) {
	switch o.probeModels(ctx, p) {
	case probeHealthy:
		// Clear only a PREFLIGHT-raised condition: a healthy models listing proves
		// the endpoint is reachable and the model served, but NOT that tool-calling
		// works — so it must never clear a tool-silent condition (that clears only
		// on a landed tool call, data-model.md). A no-op when already healthy.
		o.clearPreflightCondition(p)
		return false
	case probeMissing:
		o.setPreflightCondition(p, CondModelMissing,
			fmt.Sprintf("model %q not served by %s", p.cfg.Model, p.cfg.Endpoint),
			fmt.Sprintf("ollama pull %s", p.cfg.Model))
		return true
	case probeUnreachable:
		o.setPreflightCondition(p, CondEndpointUnreachable,
			fmt.Sprintf("endpoint %s unreachable", p.cfg.Endpoint),
			fmt.Sprintf("start the model server at %s", p.cfg.Endpoint))
		return true
	default: // probeUnsupported
		// The endpoint does not expose a standard model listing (non-Ollama
		// router). Skip gracefully with ONE low-key line — never a false
		// model-missing — and do not re-probe: the runtime tool-silence net (US2)
		// is the remaining safety net for this provider. No condition, no hook.
		preflightLogf("llm provider %q: endpoint %s does not expose a model listing — preflight skipped (runtime detection still applies)",
			p.name, p.cfg.Endpoint)
		return false
	}
}

// setPreflightCondition installs a preflight condition (model-missing /
// endpoint-unreachable), reclassifying freely BETWEEN the two preflight kinds in
// EITHER direction — a fresh probe is authoritative over its own prior diagnosis
// (data-model.md: unreachable ⇄ missing). raiseCondition alone cannot express the
// DOWNWARD reclassify (unreachable → missing): its cross-source precedence guard
// — which correctly stops the lower-ranked tool-silence detector from
// overwriting a preflight condition — would drop the lower-ranked missing. So
// when the slot already holds the OTHER preflight kind, clear it first, then
// raise. A first raise, or a steady re-probe of the SAME kind, goes straight
// through raiseCondition (which no-ops an identical raise), so a stable condition
// never spams the transition hook — only a genuine kind change fires the
// clear+raise pair.
func (o *Orchestrator) setPreflightCondition(p *provider, kind ConditionKind, detail, remedy string) {
	cur := p.conditionSnapshot().kind
	if (cur == CondModelMissing || cur == CondEndpointUnreachable) && cur != kind {
		o.clearCondition(p) // fires active=false; the raise below fires the new kind
	}
	o.raiseCondition(p, kind, detail, remedy)
}

// clearPreflightCondition clears a provider's condition ONLY when it is a
// preflight kind (model-missing / endpoint-unreachable). A tool-silent condition
// is deliberately left untouched: a passing models-list probe does not prove
// tool-calling recovered (that clears on the first landed tool call,
// data-model.md). A no-op when the provider is already healthy.
func (o *Orchestrator) clearPreflightCondition(p *provider) {
	c := p.conditionSnapshot()
	if c.kind == CondModelMissing || c.kind == CondEndpointUnreachable {
		o.clearCondition(p)
	}
}

// preflightEligible returns the providers the preflight probes: openai_compat
// transport only (spec 034 FR-001 — the anthropic managed SDK has no local model
// registry to list and is exempt). Sorted by name for a deterministic probe
// order (and deterministic tests).
func (o *Orchestrator) preflightEligible() []*provider {
	out := make([]*provider, 0, len(o.providers))
	for _, p := range o.providers {
		if p.cfg.Transport == ProviderOpenAICompat {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

// RunPreflight is the boot preflight + periodic re-probe loop (spec 034 US1,
// T007). It probes every eligible provider once at start — raising/clearing
// conditions through the Phase 2 machinery — then, every preflightInterval,
// re-probes ONLY the providers still holding a preflight-raised condition, so a
// healthy world makes zero steady-state probe traffic. Each active re-probe
// re-logs the standing warning (repeat-loudness) via preflightLogf WITHOUT
// firing the condition hook (transitions-only, so the durable event stays
// quiet); a passing re-probe clears (the hook fires active=false), and a re-probe
// may reclassify between unreachable and missing (data-model state machine). It
// stops when ctx is done (daemon shutdown) or the orchestrator closes. Run in its
// own goroutine (go o.RunPreflight(ctx)); boot never blocks or fails on it
// (FR-002).
func (o *Orchestrator) RunPreflight(ctx context.Context) {
	eligible := o.preflightEligible()
	if len(eligible) == 0 {
		return // no openai_compat providers — nothing to probe
	}
	// Initial probe: every eligible provider once. A raise fires the hook, so the
	// first operator warning lands within one probe RTT of boot.
	for _, p := range eligible {
		if ctx.Err() != nil {
			return
		}
		o.preflightProbe(ctx, p)
	}
	ticker := time.NewTicker(preflightInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-o.done:
			return
		case <-ticker.C:
			for _, p := range eligible {
				c := p.conditionSnapshot()
				if c.kind != CondModelMissing && c.kind != CondEndpointUnreachable {
					// Healthy, or holding a non-preflight condition (tool-silent):
					// not the preflight loop's business — traffic or the detector
					// owns that slot. Skip.
					continue
				}
				// Re-log the STANDING warning before re-probing, so the operator
				// keeps seeing it every cadence even when nothing changes (a steady
				// condition fires no transition hook). If the re-probe then clears
				// or reclassifies, that transition fires the hook separately.
				preflightLogf("WARNING llm provider %q: %s — %s", p.name, c.detail, c.remedy)
				o.preflightProbe(ctx, p)
			}
		}
	}
}
