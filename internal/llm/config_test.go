package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeConfigFile writes raw JSON to a temp llm.json and returns the path.
func writeConfigFile(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "llm.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestLegacyShapeDerivesRegistry (spec 024 FR-004, US1 AC#1): an untouched
// legacy llm.json derives two providers named local/cloud — local is zero-priced
// openai_compat carrying parallel/tool_mode, cloud is anthropic carrying pricing
// and keys — with today's kind→tier table as single-entry routes. This is the
// P1 equivalence guarantee at the config layer.
func TestLegacyShapeDerivesRegistry(t *testing.T) {
	p := writeConfigFile(t, `{
		"monthly_budget_usd": 100,
		"local": {"endpoint": "http://localhost:11434/v1", "model": "gemma", "parallel": 4, "tool_mode": "json"},
		"cloud": {"model": "claude-opus-4-8", "input_usd_per_mtok": 5, "output_usd_per_mtok": 25, "api_key_env": "K"}
	}`)
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatal(err)
	}
	providers, routes, err := cfg.resolveRegistry()
	if err != nil {
		t.Fatalf("resolveRegistry: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("derived %d providers, want 2 (local, cloud)", len(providers))
	}
	local := providers["local"]
	if local.Transport != ProviderOpenAICompat || local.Model != "gemma" ||
		!local.zeroPriced() || local.Parallel != 4 || local.ToolMode != "json" {
		t.Errorf("derived local = %+v", local)
	}
	cloud := providers["cloud"]
	if cloud.Transport != ProviderAnthropic || cloud.zeroPriced() ||
		cloud.InputUSDPerMTok != 5 || cloud.OutputUSDPerMTok != 25 || cloud.APIKeyEnv != "K" {
		t.Errorf("derived cloud = %+v", cloud)
	}

	// Routes cover exactly the accepted kinds, each a single-entry chain, with
	// today's placement: planner/conversation/meeting → local; the rest → cloud.
	if len(routes) != len(acceptedKinds) {
		t.Fatalf("derived %d routes, want %d", len(routes), len(acceptedKinds))
	}
	wantLocal := map[Kind]bool{KindPlanner: true, KindConversation: true, KindMeeting: true}
	for kind := range acceptedKinds {
		rc, ok := routes[kind]
		if !ok {
			t.Errorf("kind %q has no derived route", kind)
			continue
		}
		// The watch kind (spec 029), the report-card kind (spec 063), and the
		// steward kind (spec 102) are the multi-entry default chains:
		// cheap-first local with a reliable cloud fallback
		// (contracts/routing.md).
		if kind == KindGuardianWatch || kind == KindReportCard || kind == KindSteward {
			if rc.NoFallback || !reflect.DeepEqual(rc.Chain, []string{"local", "cloud"}) {
				t.Errorf("route %q = %+v, want a [local cloud] fallback chain", kind, rc)
			}
			continue
		}
		if len(rc.Chain) != 1 || rc.NoFallback {
			t.Errorf("route %q = %+v, want a single-entry fallback-free chain", kind, rc)
			continue
		}
		want := "cloud"
		if wantLocal[kind] {
			want = "local"
		}
		if rc.Chain[0] != want {
			t.Errorf("route %q → %q, want %q", kind, rc.Chain[0], want)
		}
	}
}

// TestLegacyCloudProviderTransportMapping: cloud.provider "" | "anthropic" maps
// to the anthropic transport; "openai_compat" maps to openai_compat (FR-004
// transport mapping).
func TestLegacyCloudProviderTransportMapping(t *testing.T) {
	cases := []struct {
		provider string
		want     string
	}{
		{"", ProviderAnthropic},
		{"anthropic", ProviderAnthropic},
		{"openai_compat", ProviderOpenAICompat},
	}
	for _, c := range cases {
		cfg := Config{
			MonthlyBudgetUSD: 100,
			Local:            LocalConfig{Endpoint: "http://x/v1", Model: "m"},
			Cloud:            CloudConfig{Provider: c.provider, Model: "m", Endpoint: "http://cloud/v1"},
		}
		providers, _, err := cfg.resolveRegistry()
		if err != nil {
			t.Fatalf("provider %q: %v", c.provider, err)
		}
		if got := providers["cloud"].Transport; got != c.want {
			t.Errorf("cloud.provider %q → transport %q, want %q", c.provider, got, c.want)
		}
	}
}

// TestDefaultV2SemanticallyEqualsLegacy (spec 024 FR-017, T006; local provider
// updated by spec 109/TASK-184): the v2 default WriteDefault emits derives the
// SAME registry an equivalent legacy config would — proving the new-world
// scaffold's shape (two named providers, today's routes) still round-trips
// through the legacy derivation path, now carrying the gemma4:latest
// fresh-world default (spec 109: gguf build, honors schema constraints,
// tool-calls natively) instead of the prior cogito:3b default.
func TestDefaultV2SemanticallyEqualsLegacy(t *testing.T) {
	vp, vr, err := DefaultConfig().resolveRegistry()
	if err != nil {
		t.Fatalf("v2 default resolveRegistry: %v", err)
	}
	legacy := Config{
		MonthlyBudgetUSD: 100,
		Local:            LocalConfig{Endpoint: "http://localhost:11434/v1", Model: "gemma4:latest", Parallel: 4, ToolMode: "native"},
		Cloud:            CloudConfig{Model: "claude-opus-4-8", InputUSDPerMTok: 5, OutputUSDPerMTok: 25, APIKeyEnv: "ANTHROPIC_API_KEY"},
	}
	lp, lr, err := legacy.resolveRegistry()
	if err != nil {
		t.Fatalf("legacy default resolveRegistry: %v", err)
	}
	if !reflect.DeepEqual(vp, lp) {
		t.Errorf("v2 default providers differ from legacy:\n v2=%+v\nleg=%+v", vp, lp)
	}
	if !reflect.DeepEqual(vr, lr) {
		t.Errorf("v2 default routes differ from legacy:\n v2=%+v\nleg=%+v", vr, lr)
	}
}

// TestDefaultConfigLocalProvider (spec 109/TASK-184): the fresh-world default
// local provider is gemma4:latest, tool_mode "native", parallel 4 — a gguf
// build that honors JSON-Schema constraints and tool-calls natively, not the
// prior cogito:3b/"json" default (TASK-73) nor the MLX gemma4:12b-mlx that
// silently discards schema constraints. This is the golden assertion the
// docs/README alignment (T016) and cmdNew guidance line (T015) both depend
// on staying true.
func TestDefaultConfigLocalProvider(t *testing.T) {
	local := DefaultConfig().Providers["local"]
	if local.Model != "gemma4:latest" {
		t.Errorf("default local model = %q, want gemma4:latest", local.Model)
	}
	if local.ToolMode != "native" {
		t.Errorf("default local tool_mode = %q, want native", local.ToolMode)
	}
	if local.Parallel != 4 {
		t.Errorf("default local parallel = %d, want 4", local.Parallel)
	}
	if local.Endpoint != "http://localhost:11434/v1" {
		t.Errorf("default local endpoint = %q, want http://localhost:11434/v1", local.Endpoint)
	}
}

// TestWriteDefaultEmitsGemmaDefault (spec 109/TASK-184): WriteDefault's
// actual on-disk llm.json — what `promptworld new` writes — carries the same
// gemma4:latest/native/4 shape, not just the in-memory DefaultConfig() value.
func TestWriteDefaultEmitsGemmaDefault(t *testing.T) {
	p := filepath.Join(t.TempDir(), "llm.json")
	if err := WriteDefault(p); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatal(err)
	}
	local := cfg.Providers["local"]
	if local.Model != "gemma4:latest" || local.ToolMode != "native" || local.Parallel != 4 {
		t.Errorf("WriteDefault local provider = %+v, want model gemma4:latest, tool_mode native, parallel 4", local)
	}
}

// v2Body composes a valid two-provider (gemma/anthropic) v2 config with a
// caller-supplied routes object, for the validation matrix's route mutations.
func v2Body(routes string) string {
	return `{"monthly_budget_usd":100,` +
		`"providers":{` +
		`"gemma":{"transport":"openai_compat","endpoint":"http://x/v1","model":"g"},` +
		`"anthropic":{"transport":"anthropic","model":"claude","input_usd_per_mtok":5}},` +
		`"routes":` + routes + `}`
}

// validRoutes covers exactly the accepted kinds — the baseline a matrix case
// mutates. metatron_watch (spec 029) is routed explicitly here with declared
// providers so the positive controls stay complete without leaning on the
// missing-route backfill (which TestGuardianWatchRouteBackfill exercises on its
// own incomplete config).
const validRoutes = `{"planner":["gemma"],"conversation":["gemma"],"meeting":["gemma"],` +
	`"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],"metatron":["anthropic"],` +
	`"metatron_watch":["gemma","anthropic"]}`

// TestValidationMatrix (spec 024 FR-003, SC-008, contracts/llm-config.md): every
// row of the boot-error matrix fails LoadConfig with an error naming the
// offending entry — never a runtime surprise. Warn-and-clamp knobs (parallel,
// tool_mode) are NOT in this matrix; they clamp, never error.
func TestValidationMatrix(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string // substring the boot error must contain
	}{
		{
			"route names undeclared provider",
			v2Body(`{"planner":["ghost"],"conversation":["gemma"],"meeting":["gemma"],"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],"metatron":["anthropic"]}`),
			"ghost",
		},
		{
			"accepted kind missing from routes",
			v2Body(`{"planner":["gemma"],"conversation":["gemma"],"meeting":["gemma"],"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"]}`),
			"metatron",
		},
		{
			"unknown kind key in routes",
			v2Body(`{"planner":["gemma"],"conversation":["gemma"],"meeting":["gemma"],"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],"metatron":["anthropic"],"sorcery":["gemma"]}`),
			"sorcery",
		},
		{
			"duplicate provider within a chain",
			v2Body(`{"planner":["gemma","gemma"],"conversation":["gemma"],"meeting":["gemma"],"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],"metatron":["anthropic"]}`),
			"planner",
		},
		{
			"empty chain",
			v2Body(`{"planner":[],"conversation":["gemma"],"meeting":["gemma"],"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],"metatron":["anthropic"]}`),
			"planner",
		},
		{
			"no_fallback with chain length > 1",
			v2Body(`{"planner":{"chain":["gemma","anthropic"],"no_fallback":true},"conversation":["gemma"],"meeting":["gemma"],"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],"metatron":["anthropic"]}`),
			"planner",
		},
		{
			"provider missing model",
			`{"monthly_budget_usd":100,"providers":{"gemma":{"transport":"openai_compat","endpoint":"http://x/v1"},"anthropic":{"transport":"anthropic","model":"c"}},"routes":` + validRoutes + `}`,
			"model",
		},
		{
			"provider missing transport",
			`{"monthly_budget_usd":100,"providers":{"gemma":{"endpoint":"http://x/v1","model":"g"},"anthropic":{"transport":"anthropic","model":"c"}},"routes":` + validRoutes + `}`,
			"transport",
		},
		{
			"openai_compat missing endpoint",
			`{"monthly_budget_usd":100,"providers":{"gemma":{"transport":"openai_compat","model":"g"},"anthropic":{"transport":"anthropic","model":"c"}},"routes":` + validRoutes + `}`,
			"endpoint",
		},
		{
			"both providers and legacy present",
			`{"monthly_budget_usd":100,"providers":{"gemma":{"transport":"openai_compat","endpoint":"http://x/v1","model":"g"},"anthropic":{"transport":"anthropic","model":"c"}},"routes":` + validRoutes + `,"local":{"model":"m"}}`,
			"mutually exclusive",
		},
		{
			"monthly_budget_usd not positive",
			`{"monthly_budget_usd":0,"providers":{"gemma":{"transport":"openai_compat","endpoint":"http://x/v1","model":"g"},"anthropic":{"transport":"anthropic","model":"c"}},"routes":` + validRoutes + `}`,
			"monthly_budget_usd",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := writeConfigFile(t, c.body)
			_, err := LoadConfig(p)
			if err == nil {
				t.Fatalf("expected a boot error naming %q, got nil", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("boot error %q does not name the offender %q", err.Error(), c.want)
			}
		})
	}
}

// TestValidV2Loads: the valid baseline the matrix mutates loads cleanly and its
// registry resolves — a positive control for the matrix.
func TestValidV2Loads(t *testing.T) {
	p := writeConfigFile(t, v2Body(validRoutes))
	cfg, err := LoadConfig(p)
	if err != nil {
		t.Fatalf("valid v2 config rejected: %v", err)
	}
	providers, routes, err := cfg.resolveRegistry()
	if err != nil {
		t.Fatalf("resolveRegistry on valid v2: %v", err)
	}
	if len(providers) != 2 || len(routes) != len(acceptedKinds) {
		t.Errorf("valid v2 resolved to %d providers / %d routes", len(providers), len(routes))
	}
}

// TestGuardianWatchRouteBackfill (spec 029 T001, contracts/routing.md / research
// R8): a v2 llm.json written before metatron_watch shipped is MISSING its route.
// validateV2 backfills it from defaultRoutes() with one boot log line rather than
// failing the boot, so pre-existing worlds keep booting on upgrade; the post-load
// invariant (every accepted kind routed) still holds.
func TestGuardianWatchRouteBackfill(t *testing.T) {
	// A default-shaped v2 world (providers local/cloud) whose routes predate the
	// watch kind: all seven pre-029 kinds routed, metatron_watch absent.
	body := `{"monthly_budget_usd":100,` +
		`"providers":{` +
		`"local":{"transport":"openai_compat","endpoint":"http://x/v1","model":"g"},` +
		`"cloud":{"transport":"anthropic","model":"claude","input_usd_per_mtok":5}},` +
		`"routes":{"planner":["local"],"conversation":["local"],"meeting":["local"],` +
		`"consolidation":["cloud"],"narrator":["cloud"],"drama":["cloud"],"metatron":["cloud"]}}`

	var logged []string
	prev := configWarnf
	configWarnf = func(format string, args ...any) { logged = append(logged, fmt.Sprintf(format, args...)) }
	defer func() { configWarnf = prev }()

	cfg, err := LoadConfig(writeConfigFile(t, body))
	if err != nil {
		t.Fatalf("v2 config missing the metatron_watch route must still boot, got: %v", err)
	}
	_, routes, err := cfg.resolveRegistry()
	if err != nil {
		t.Fatalf("resolveRegistry: %v", err)
	}
	rc, ok := routes[KindGuardianWatch]
	if !ok {
		t.Fatal("metatron_watch route was not backfilled — post-load completeness invariant broken")
	}
	if rc.NoFallback || !reflect.DeepEqual(rc.Chain, []string{"local", "cloud"}) {
		t.Errorf("backfilled metatron_watch route = %+v, want the default [local cloud] chain", rc)
	}
	// The backfill announces itself on the boot channel, naming the kind.
	named := false
	for _, l := range logged {
		if strings.Contains(l, string(KindGuardianWatch)) {
			named = true
		}
	}
	if !named {
		t.Errorf("backfill emitted no boot log line naming %q; got %v", KindGuardianWatch, logged)
	}
}

// TestUnknownRouteKeyStillErrors (spec 029 T001): the missing-route backfill is
// narrowly scoped to the curated backfill set — it never turns an unknown route
// KEY into a silent backfill. A typo'd kind is still a boot error naming it.
func TestUnknownRouteKeyStillErrors(t *testing.T) {
	body := v2Body(`{"planner":["gemma"],"conversation":["gemma"],"meeting":["gemma"],` +
		`"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],` +
		`"metatron":["anthropic"],"metatron_watch":["gemma"],"sorcery":["gemma"]}`)
	if _, err := LoadConfig(writeConfigFile(t, body)); err == nil {
		t.Fatal("unknown route key must be a boot error")
	} else if !strings.Contains(err.Error(), "sorcery") {
		t.Errorf("boot error %q does not name the unknown key", err)
	}
}

// TestDefaultsIncludeWatchKind (spec 029 T001): the shipped defaults route the
// new watch kind, so a fresh world and every legacy-derived world pick it up for
// free — DefaultConfig, defaultRoutes(), and the accepted-kind set agree.
func TestDefaultsIncludeWatchKind(t *testing.T) {
	if _, ok := acceptedKinds[KindGuardianWatch]; !ok {
		t.Fatal("KindGuardianWatch is not an accepted kind")
	}
	rc, ok := defaultRoutes()[string(KindGuardianWatch)]
	if !ok {
		t.Fatal("defaultRoutes() has no metatron_watch entry")
	}
	if !reflect.DeepEqual(rc.Chain, []string{"local", "cloud"}) {
		t.Errorf("default metatron_watch chain = %v, want [local cloud]", rc.Chain)
	}
	// A default-config world routes every accepted kind, watch included.
	_, routes, err := DefaultConfig().resolveRegistry()
	if err != nil {
		t.Fatalf("DefaultConfig resolveRegistry: %v", err)
	}
	if _, ok := routes[KindGuardianWatch]; !ok {
		t.Error("DefaultConfig does not route metatron_watch")
	}
}

// TestReportCardRouteKind (spec 063 T009): the report-card kind is accepted,
// default-routed on the cheap-first [local cloud] chain (the metatron_watch
// precedent), backfilled into pre-063 v2 configs with a boot log line, and
// re-routable explicitly like any kind.
func TestReportCardRouteKind(t *testing.T) {
	if _, ok := acceptedKinds[KindReportCard]; !ok {
		t.Fatal("KindReportCard is not an accepted kind")
	}
	rc, ok := defaultRoutes()[string(KindReportCard)]
	if !ok {
		t.Fatal("defaultRoutes() has no report_card entry")
	}
	if !reflect.DeepEqual(rc.Chain, []string{"local", "cloud"}) {
		t.Errorf("default report_card chain = %v, want [local cloud]", rc.Chain)
	}
	_, routes, err := DefaultConfig().resolveRegistry()
	if err != nil {
		t.Fatalf("DefaultConfig resolveRegistry: %v", err)
	}
	if _, ok := routes[KindReportCard]; !ok {
		t.Error("DefaultConfig does not route report_card")
	}

	// Backfill: a pre-063 v2 config (report_card route absent) still boots,
	// gains the default chain, and the boot channel names the kind.
	body := `{"monthly_budget_usd":100,` +
		`"providers":{` +
		`"local":{"transport":"openai_compat","endpoint":"http://x/v1","model":"g"},` +
		`"cloud":{"transport":"anthropic","model":"claude","input_usd_per_mtok":5}},` +
		`"routes":{"planner":["local"],"conversation":["local"],"meeting":["local"],` +
		`"consolidation":["cloud"],"narrator":["cloud"],"drama":["cloud"],"metatron":["cloud"],` +
		`"metatron_watch":["local","cloud"]}}`
	var logged []string
	prev := configWarnf
	configWarnf = func(format string, args ...any) { logged = append(logged, fmt.Sprintf(format, args...)) }
	defer func() { configWarnf = prev }()
	cfg, err := LoadConfig(writeConfigFile(t, body))
	if err != nil {
		t.Fatalf("pre-063 v2 config must still boot, got: %v", err)
	}
	_, routes, err = cfg.resolveRegistry()
	if err != nil {
		t.Fatalf("resolveRegistry: %v", err)
	}
	got, ok := routes[KindReportCard]
	if !ok {
		t.Fatal("report_card route was not backfilled")
	}
	if got.NoFallback || !reflect.DeepEqual(got.Chain, []string{"local", "cloud"}) {
		t.Errorf("backfilled report_card route = %+v, want the default [local cloud] chain", got)
	}
	named := false
	for _, l := range logged {
		if strings.Contains(l, string(KindReportCard)) {
			named = true
		}
	}
	if !named {
		t.Errorf("backfill emitted no boot log line naming %q; got %v", KindReportCard, logged)
	}

	// Explicit routing wins over the default, like any kind.
	explicit := v2Body(`{"planner":["gemma"],"conversation":["gemma"],"meeting":["gemma"],` +
		`"consolidation":["anthropic"],"narrator":["anthropic"],"drama":["anthropic"],` +
		`"metatron":["anthropic"],"metatron_watch":["gemma"],"report_card":["anthropic"]}`)
	cfg2, err := LoadConfig(writeConfigFile(t, explicit))
	if err != nil {
		t.Fatalf("explicit report_card route: %v", err)
	}
	_, routes2, err := cfg2.resolveRegistry()
	if err != nil {
		t.Fatalf("resolveRegistry: %v", err)
	}
	if got := routes2[KindReportCard]; !reflect.DeepEqual(got.Chain, []string{"anthropic"}) {
		t.Errorf("explicit report_card chain = %v, want [anthropic]", got.Chain)
	}
}

// TestMaxTokensRoundTripsThroughShapeAwareMarshal (spec 024 T021 / research R9):
// the spec-025 max_tokens object is a kind-scoped top-level knob, orthogonal to
// the provider registry — it MUST survive the branch's shape-aware custom
// MarshalJSON byte-for-byte in BOTH shapes, and a nil MaxTokens must NOT emit the
// object (omitempty preserved so a default file stays byte-for-byte compatible).
// The custom marshaler bypasses the struct-tag omitempty, so this pins the
// hand-rolled carry the reconciliation added.
func TestMaxTokensRoundTripsThroughShapeAwareMarshal(t *testing.T) {
	budgets := &TokenBudgets{Planner: 700, Consolidation: 2000} // GuardianTurn 0 → dropped by TokenBudgets' own omitempty
	shapes := map[string]Config{
		"v2": {
			MonthlyBudgetUSD: 100,
			Providers: map[string]ProviderConfig{
				"gemma": {Transport: "openai_compat", Endpoint: "http://x/v1", Model: "g"},
			},
			Routes:    map[string]RouteConfig{"planner": {Chain: []string{"gemma"}}},
			MaxTokens: budgets,
		},
		"legacy": {
			MonthlyBudgetUSD: 100,
			Local:            LocalConfig{Endpoint: "http://x/v1", Model: "g"},
			Cloud:            CloudConfig{Model: "claude", InputUSDPerMTok: 5, APIKeyEnv: "K"},
			MaxTokens:        budgets,
		},
	}
	for name, cfg := range shapes {
		t.Run(name+"/present", func(t *testing.T) {
			raw, err := json.Marshal(cfg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			// GuardianTurn 0 must be suppressed; planner/consolidation carried.
			if got := string(raw); !strings.Contains(got, `"max_tokens":{"planner":700,"consolidation":2000}`) {
				t.Fatalf("max_tokens not carried verbatim: %s", got)
			}
			var back Config
			if err := json.Unmarshal(raw, &back); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !reflect.DeepEqual(back.MaxTokens, budgets) {
				t.Errorf("MaxTokens round-trip mismatch: got %+v want %+v", back.MaxTokens, budgets)
			}
			// Byte-for-byte: re-marshaling the parsed config reproduces the bytes.
			again, err := json.Marshal(back)
			if err != nil {
				t.Fatalf("re-marshal: %v", err)
			}
			if string(again) != string(raw) {
				t.Errorf("not byte-for-byte:\n first=%s\nsecond=%s", raw, again)
			}
			// The normalizers read the recovered budgets (defaults where absent).
			if n, _ := back.PlannerTokens(); n != 700 {
				t.Errorf("PlannerTokens = %d, want 700", n)
			}
			if n, _ := back.GuardianTurnTokens(); n != defaultGuardianTurnTokens {
				t.Errorf("GuardianTurnTokens = %d, want default %d", n, defaultGuardianTurnTokens)
			}
			if n, _ := back.ConsolidationTokens(); n != 2000 {
				t.Errorf("ConsolidationTokens = %d, want 2000", n)
			}
		})
		t.Run(name+"/nil-omitted", func(t *testing.T) {
			cfg.MaxTokens = nil
			raw, err := json.Marshal(cfg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if strings.Contains(string(raw), "max_tokens") {
				t.Errorf("nil MaxTokens emitted the object: %s", raw)
			}
		})
	}
}

// TestMaxTokensLoadsFromV2File (spec 024 T021): a full v2 llm.json carrying
// max_tokens loads through LoadConfig and the normalizers resolve the tuned
// values — proving the parse leg composes with the registry validator, not only
// the direct marshal round-trip above.
func TestMaxTokensLoadsFromV2File(t *testing.T) {
	body := `{"monthly_budget_usd":100,` +
		`"max_tokens":{"planner":300,"metatron_turn":1500,"consolidation":900},` +
		`"providers":{` +
		`"gemma":{"transport":"openai_compat","endpoint":"http://x/v1","model":"g"},` +
		`"anthropic":{"transport":"anthropic","model":"claude","input_usd_per_mtok":5}},` +
		`"routes":` + validRoutes + `}`
	cfg, err := LoadConfig(writeConfigFile(t, body))
	if err != nil {
		t.Fatalf("v2 config with max_tokens rejected: %v", err)
	}
	if n, _ := cfg.PlannerTokens(); n != 300 {
		t.Errorf("PlannerTokens = %d, want 300", n)
	}
	if n, _ := cfg.GuardianTurnTokens(); n != 1500 {
		t.Errorf("GuardianTurnTokens = %d, want 1500", n)
	}
	if n, _ := cfg.ConsolidationTokens(); n != 900 {
		t.Errorf("ConsolidationTokens = %d, want 900", n)
	}
}
