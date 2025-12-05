package protection

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/paladin"
)

/*
When you dodge or parry a melee attack you have a 30% chance of refreshing the cooldown on your next Avenger's Shield and causing it to generate a charge of Holy Power if used within 6 sec.
(Proc chance: 30%, 1s cooldown)
*/
func (prot *ProtectionPaladin) registerGrandCrusader() {
	hpActionID := core.ActionID{SpellID: 98057}
	prot.CanTriggerHolyAvengerHpGain(hpActionID)

	var grandCrusaderAura *core.Aura
	grandCrusaderAura = core.BlockPrepull(prot.RegisterAura(core.Aura{
		Label:    "Grand Crusader" + prot.Label,
		ActionID: core.ActionID{SpellID: 85416},
		Duration: time.Second * 6,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			prot.AvengersShield.CD.Reset()
		},
	})).AttachProcTrigger(core.ProcTrigger{
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     paladin.SpellMaskAvengersShield,
		TriggerImmediately: true,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			prot.HolyPower.Gain(sim, 1, hpActionID)
			grandCrusaderAura.Deactivate(sim)
		},
	})

	// 2025-11-13: Grand Crusader has been changed to its Patch 5.2.0 version.
	// Grand Crusader now has a 12% chance to activate when the Paladin dodges or parries a melee attack, or lands a Crusader Strike or Hammer of the Righteous.
	// (was 30% chance on dodge or parrying a melee attack)
	spellMask := paladin.SpellMaskCrusaderStrike | paladin.SpellMaskHammerOfTheRighteousMelee
	prot.MakeProcTriggerAura(core.ProcTrigger{
		Name:               "Grand Crusader Trigger" + prot.Label,
		Callback:           core.CallbackOnSpellHitTaken | core.CallbackOnSpellHitDealt,
		Outcome:            core.OutcomeDodge | core.OutcomeParry | core.OutcomeLanded,
		ProcChance:         0.12,
		ICD:                time.Second,
		TriggerImmediately: true,

		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			if spell.Unit == &prot.Unit {
				return result.Outcome.Matches(core.OutcomeLanded) && spell.Matches(spellMask)
			}

			return result.Outcome.Matches(core.OutcomeDodge | core.OutcomeParry)
		},

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			grandCrusaderAura.Activate(sim)
		},
	})
}
