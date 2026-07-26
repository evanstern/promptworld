package sim

import (
	"testing"

	"github.com/evanstern/promptworld/internal/tool"
)

// TestExplainChargeDoctrineMirrorsSim (spec 063 T002): internal/tool is a
// leaf and MIRRORS the charge-economy doctrine constants for the explain
// "charges" fact sheet; this pin holds the mirror equal to sim's enforced
// values (the TestMiracleKindsMirrorTool pattern) so an explained number can
// never drift from the mechanics.
func TestExplainChargeDoctrineMirrorsSim(t *testing.T) {
	cap, genesis, regenHours := tool.ChargeDoctrine()
	if cap != GuardianChargeCap {
		t.Errorf("tool mirror cap = %d, sim GuardianChargeCap = %d", cap, GuardianChargeCap)
	}
	if genesis != GuardianGenesisCharges {
		t.Errorf("tool mirror genesis = %d, sim GuardianGenesisCharges = %d", genesis, GuardianGenesisCharges)
	}
	if int64(regenHours)*3600 != GuardianChargeRegenTicks {
		t.Errorf("tool mirror regen = %d game hours, sim GuardianChargeRegenTicks = %d ticks", regenHours, GuardianChargeRegenTicks)
	}
}
