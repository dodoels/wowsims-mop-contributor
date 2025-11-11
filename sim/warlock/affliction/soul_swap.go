package affliction

import (
	"time"

	"github.com/wowsims/mop/sim/core"
)

func (warlock *AfflictionWarlock) registerSoulSwap() {
	var debuffState map[int32]core.DotState
	dotRefs := []**core.Spell{&warlock.Corruption, &warlock.Agony, &warlock.Seed, &warlock.UnstableAffliction}

	inhaleBuff := core.BlockPrepull(warlock.RegisterAura(core.Aura{
		ActionID: core.ActionID{SpellID: 86211},
		Label:    "Soul Swap",
		Duration: time.Second * 3,
	}))

	// Exhale
	warlock.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 86213},
		Flags:       core.SpellFlagAPL,
		ProcMask:    core.ProcMaskEmpty,
		SpellSchool: core.SpellSchoolShadow,

		ThreatMultiplier: 1,
		CritMultiplier:   warlock.DefaultCritMultiplier(),
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return inhaleBuff.IsActive() && target != warlock.LastInhaleTarget && !warlock.SoulBurnAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			// restore states
			for _, spellRef := range dotRefs {
				dot := (*spellRef).Dot(target)
				state, ok := debuffState[dot.ActionID.SpellID]
				if !ok {
					// not stored, was not active
					continue
				}

				(*spellRef).Proc(sim, target)
				dot.RestoreState(state, sim)
			}

			inhaleBuff.Deactivate(sim)
		},
	})

	// used to not allocate a result for every check
	expectedDamage := &core.SpellResult{}

	// we dont use seed in the expected calculations as it's not applied by exhale
	expectedDotRefs := []**core.Spell{&warlock.Corruption, &warlock.Agony, &warlock.UnstableAffliction}

	// Inhale
	warlock.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 86121}.WithTag(1),
		Flags:       core.SpellFlagAPL,
		ProcMask:    core.ProcMaskEmpty,
		SpellSchool: core.SpellSchoolShadow,

		ThreatMultiplier: 1,
		CritMultiplier:   warlock.DefaultCritMultiplier(),
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return anyDoTActive(dotRefs, target) && !inhaleBuff.IsActive() && !warlock.SoulBurnAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			warlock.LastInhaleTarget = target
			debuffState = map[int32]core.DotState{}

			// store states
			for _, spellRef := range dotRefs {
				dot := (*spellRef).Dot(target)
				if dot.IsActive() {
					debuffState[dot.ActionID.SpellID] = dot.SaveState(sim)
				}
			}

			inhaleBuff.Activate(sim)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			expectedDamage.Damage = 0
			if useSnapshot {
				for _, spellRef := range expectedDotRefs {
					dot := (*spellRef).Dot(target)
					expectedDamage.Damage += dot.Spell.ExpectedTickDamageFromCurrentSnapshot(sim, target)
				}

				return expectedDamage
			}

			for _, spellRef := range expectedDotRefs {
				dot := (*spellRef).Dot(target)
				expectedDamage.Damage += dot.Spell.ExpectedTickDamage(sim, target)
			}

			return expectedDamage
		},
	})

	// Soulswap: Soulburn
	warlock.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 86121}.WithTag(2),
		Flags:       core.SpellFlagAPL,
		ProcMask:    core.ProcMaskEmpty,
		SpellSchool: core.SpellSchoolShadow,

		ThreatMultiplier: 1,
		CritMultiplier:   warlock.DefaultCritMultiplier(),
		DamageMultiplier: 1,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return warlock.SoulBurnAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			warlock.Agony.Proc(sim, target)
			warlock.Corruption.Proc(sim, target)
			warlock.UnstableAffliction.Proc(sim, target)
			warlock.SoulBurnAura.Deactivate(sim)
		},

		ExpectedTickDamage: func(sim *core.Simulation, target *core.Unit, spell *core.Spell, useSnapshot bool) *core.SpellResult {
			expectedDamage.Damage = 0
			if useSnapshot {
				for _, spellRef := range expectedDotRefs {
					dot := (*spellRef).Dot(target)
					expectedDamage.Damage += dot.Spell.ExpectedTickDamageFromCurrentSnapshot(sim, target)
				}

				return expectedDamage
			}

			for _, spellRef := range expectedDotRefs {
				dot := (*spellRef).Dot(target)
				expectedDamage.Damage += dot.Spell.ExpectedTickDamage(sim, target)
			}

			return expectedDamage
		},
	})
}

func anyDoTActive(dots []**core.Spell, target *core.Unit) bool {
	for _, spellRef := range dots {
		if (*spellRef).Dot(target).IsActive() {
			return true
		}
	}

	return false
}
