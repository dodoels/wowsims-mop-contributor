package destruction

import (
	"time"

	"github.com/wowsims/mop/sim/common/shared"
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/stats"
	"github.com/wowsims/mop/sim/warlock"
)

var chaosBoltVariance = 0.2
var chaosBoltScale = 2.5875
var chaosBoltCoeff = 2.5875

func (destro *DestructionWarlock) registerChaosBolt() {
	actionID := core.ActionID{SpellID: 116858}
	shared.RegisterIgniteEffect(&destro.Unit, shared.IgniteConfig{
		ActionID:      actionID.WithTag(1), // Real SpellID: 1277303
		SpellSchool:   core.SpellSchoolShadow,
		DotAuraLabel:  "Chaos Bolt Dot",
		DotAuraTag:    "ChaosBoltDot",
		TickLength:    1 * time.Second,
		NumberOfTicks: 3,

		ProcTrigger: core.ProcTrigger{
			Name:               "Chaos Bolt - Trigger",
			Callback:           core.CallbackOnSpellHitDealt,
			ClassSpellMask:     warlock.WarlockSpellChaosBolt,
			Outcome:            core.OutcomeLanded,
			RequireDamageDealt: true,
		},

		DamageCalculator: func(result *core.SpellResult) float64 {
			return result.Damage * 0.15
		},
	})

	destro.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		SpellSchool:    core.SpellSchoolShadow,
		ProcMask:       core.ProcMaskSpellDamage,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: warlock.WarlockSpellChaosBolt,
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD:      core.GCDDefault,
				CastTime: 3000 * time.Millisecond,
			},
		},

		DamageMultiplierAdditive: 1,
		CritMultiplier:           destro.DefaultCritMultiplier(),
		ThreatMultiplier:         1,
		BonusCoefficient:         chaosBoltCoeff,
		BonusCritPercent:         100,
		MissileSpeed:             16,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := destro.CalcAndRollDamageRange(sim, chaosBoltScale, chaosBoltVariance)
			spell.DamageMultiplier *= (1 + destro.GetStat(stats.SpellCritPercent)/100)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.DamageMultiplier /= (1 + destro.GetStat(stats.SpellCritPercent)/100)

			// check again we can actually spend as Dark Soul might have run out before the cast finishes
			if spell.Flags.Matches(SpellFlagDestructionHavoc) {
				//Havoc Spell doesn't spend resources as it was a duplicate
			} else if result.Landed() && destro.BurningEmbers.CanSpend(core.TernaryInt32(destro.T15_2pc.IsActive(), 8, 10)) {
				destro.BurningEmbers.Spend(sim, core.TernaryInt32(destro.T15_2pc.IsActive(), 8, 10), spell.ActionID)
			} else {
				return
			}

			spell.WaitTravelTime(sim, func(s *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return destro.BurningEmbers.CanSpend(core.TernaryInt32(destro.T15_2pc.IsActive(), 8, 10))
		},
	})
}
