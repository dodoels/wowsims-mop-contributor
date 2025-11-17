package unholy

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/stats"
)

const GargoyleStrikeMinCastTime = time.Millisecond * 500
const MaxGargoyleStrikeCasts = 26

type GargoylePet struct {
	core.Pet

	expireTime          time.Duration
	dkOwner             *UnholyDeathKnight
	gargoyleStrikeCasts int32

	GargoyleStrike *core.Spell
}

func (uhdk *UnholyDeathKnight) NewGargoyle() *GargoylePet {
	gargoyle := &GargoylePet{
		Pet: core.NewPet(core.PetConfig{
			Name:                           "Gargoyle",
			Owner:                          &uhdk.Character,
			BaseStats:                      stats.Stats{},
			NonHitExpStatInheritance:       gargoyleStatInheritance,
			EnabledOnStart:                 false,
			IsGuardian:                     true,
			HasDynamicCastSpeedInheritance: true,
		}),
		dkOwner: uhdk,
	}
	gargoyle.OnPetDisable = gargoyle.disable

	uhdk.AddPet(gargoyle)

	return gargoyle
}

func (garg *GargoylePet) GetPet() *core.Pet {
	return &garg.Pet
}

func (garg *GargoylePet) Initialize() {
	garg.Pet.Initialize()
	garg.registerGargoyleStrikeSpell()
}

func (garg *GargoylePet) disable(_ *core.Simulation) {
	garg.gargoyleStrikeCasts = 0
}

func (garg *GargoylePet) Reset(_ *core.Simulation) {
	garg.gargoyleStrikeCasts = 0
}

func (garg *GargoylePet) OnEncounterStart(_ *core.Simulation) {
}

func (garg *GargoylePet) SetExpireTime(expireTime time.Duration) {
	garg.expireTime = expireTime
}

func (garg *GargoylePet) ExecuteCustomRotation(sim *core.Simulation) {
	if garg.gargoyleStrikeCasts >= MaxGargoyleStrikeCasts {
		garg.Disable(sim)
		garg.dkOwner.SummonGargoyleSpell.RelatedSelfBuff.Deactivate(sim)
		return
	}

	if garg.GargoyleStrike.CanCast(sim, garg.CurrentTarget) {
		gargCastTime := max(garg.ApplyCastSpeedForSpell(garg.GargoyleStrike.DefaultCast.CastTime, garg.GargoyleStrike), GargoyleStrikeMinCastTime)

		if sim.CurrentTime+gargCastTime > garg.expireTime {
			// If the cast wont finish before expiration time just dont cast
			return
		}

		garg.GargoyleStrike.Cast(sim, garg.CurrentTarget)
	}
}

func gargoyleStatInheritance(ownerStats stats.Stats) stats.Stats {
	return stats.Stats{
		stats.HasteRating:      ownerStats[stats.HasteRating],
		stats.Health:           ownerStats[stats.Health],
		stats.SpellCritPercent: ownerStats[stats.SpellCritPercent],
		stats.SpellPower:       ownerStats[stats.AttackPower] * 0.7,
	}
}

func (garg *GargoylePet) registerGargoyleStrikeSpell() {
	garg.GargoyleStrike = garg.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: 51963},
		SpellSchool: core.SpellSchoolPlague,
		ProcMask:    core.ProcMaskSpellDamage,

		MissileSpeed: 20,
		MaxRange:     40,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				CastTime: time.Millisecond * 2000,
			},

			// Gargoyle Strike will now have a minimum cast time of 0.5 seconds.
			// This was made to fix some issues with stuttering behavior at very high haste.
			// https://github.com/ClassicWoWCommunity/mop-classic-bugs/issues/2495#issuecomment-3509112879
			IgnoreHaste: true,
			ModifyCast: func(sim *core.Simulation, spell *core.Spell, cast *core.Cast) {
				cast.CastTime = max(garg.ApplyCastSpeedForSpell(garg.GargoyleStrike.DefaultCast.CastTime, garg.GargoyleStrike), GargoyleStrikeMinCastTime)
			},
		},

		DamageMultiplier: 1,
		CritMultiplier:   garg.DefaultCritMultiplier(),
		ThreatMultiplier: 1,

		BonusCoefficient: 0.8259999752,

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return garg.gargoyleStrikeCasts < MaxGargoyleStrikeCasts
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := garg.dkOwner.CalcAndRollDamageRange(sim, 0.5, 0.5)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
			garg.gargoyleStrikeCasts++
		},
	})
}
