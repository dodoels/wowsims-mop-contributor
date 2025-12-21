package hunter

import (
	"time"

	"github.com/wowsims/mop/sim/core"
)

func (hunter *Hunter) applyThrillOfTheHunt() {
	if !hunter.Talents.ThrillOfTheHunt {
		return
	}

	var tothAura *core.Aura
	tothAura = core.BlockPrepull(hunter.RegisterAura(core.Aura{
		Label:     "Thrill of the Hunt",
		ActionID:  core.ActionID{SpellID: 109306},
		Duration:  time.Second * 12,
		MaxStacks: 3,
	})).AttachSpellMod(core.SpellModConfig{
		Kind:      core.SpellMod_PowerCost_Flat,
		ClassMask: HunterSpellMultiShot | HunterSpellArcaneShot,
		IntValue:  -20,
	}).AttachProcTrigger(core.ProcTrigger{
		Callback:           core.CallbackOnCastComplete,
		ClassSpellMask:     HunterSpellMultiShot | HunterSpellArcaneShot,
		TriggerImmediately: true,

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			tothAura.RemoveStack(sim)
		},
	})

	hunter.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Thrill of the Hunt Proccer",
		Callback:   core.CallbackOnCastComplete,
		ProcChance: 0.3,

		ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
			return spell.CurCast.Cost > 0
		},

		Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			tothAura.Activate(sim)
			tothAura.SetStacks(sim, 3)
		},
	})
}
