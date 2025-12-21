package shadow

import (
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/priest"
)

func (shadow *ShadowPriest) registerHotfixes() {
	// 2025-07-01 - Shadow Word: Pain’s damage over time increased by 18%.
	// 2025-11-13 - Shadow Word: Pain’s damage over time decreased to 7% (was 18%).
	shadow.AddStaticMod(core.SpellModConfig{
		ClassMask:  priest.PriestSpellShadowWordPain,
		Kind:       core.SpellMod_DamageDone_Pct,
		FloatValue: 0.07,
	})
}
