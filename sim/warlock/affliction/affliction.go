package affliction

import (
	"math"
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/warlock"
)

func RegisterAfflictionWarlock() {
	core.RegisterAgentFactory(
		proto.Player_AfflictionWarlock{},
		proto.Spec_SpecAfflictionWarlock,
		func(character *core.Character, options *proto.Player) core.Agent {
			return NewAfflictionWarlock(character, options)
		},
		func(player *proto.Player, spec interface{}) {
			playerSpec, ok := spec.(*proto.Player_AfflictionWarlock)
			if !ok {
				panic("Invalid spec value for Affliction Warlock!")
			}
			player.Spec = playerSpec
		},
	)
}

func NewAfflictionWarlock(character *core.Character, options *proto.Player) *AfflictionWarlock {
	affOptions := options.GetAfflictionWarlock().Options

	affliction := &AfflictionWarlock{
		Warlock:      warlock.NewWarlock(character, options, affOptions.ClassOptions),
		ExhaleWindow: time.Duration(affOptions.ExhaleWindow * int32(time.Millisecond)),
	}

	affliction.MaleficGraspMaleficEffectMultiplier = 0.3
	affliction.DrainSoulMaleficEffectMultiplier = 0.6

	return affliction
}

type AfflictionWarlock struct {
	*warlock.Warlock

	SoulShards         core.SecondaryResourceBar
	Agony              *core.Spell
	UnstableAffliction *core.Spell

	SoulBurnAura *core.Aura

	LastCorruptionTarget *core.Unit // Tracks the last target we've applied corruption to
	LastInhaleTarget     *core.Unit

	DrainSoulMaleficEffectMultiplier    float64
	MaleficGraspMaleficEffectMultiplier float64
	ProcMaleficEffect                   func(target *core.Unit, coeff float64, sim *core.Simulation)

	ExhaleWindow time.Duration
}

func (warlock AfflictionWarlock) getMasteryBonus() float64 {
	return (8 + warlock.GetMasteryPoints()) * 3.1
}

func (warlock *AfflictionWarlock) GetWarlock() *warlock.Warlock {
	return warlock.Warlock
}

const MaxSoulShards = int32(4)

func (warlock *AfflictionWarlock) Initialize() {
	warlock.Warlock.Initialize()

	warlock.SoulShards = warlock.RegisterNewDefaultSecondaryResourceBar(core.SecondaryResourceConfig{
		Type:    proto.SecondaryResourceType_SecondaryResourceTypeSoulShards,
		Max:     MaxSoulShards,
		Default: MaxSoulShards,
	})

	warlock.registerPotentAffliction()
	warlock.registerHaunt()
	warlock.RegisterCorruption(func(resultList core.SpellResultSlice, spell *core.Spell, sim *core.Simulation) {
		if resultList[0].Landed() {
			warlock.LastCorruptionTarget = resultList[0].Target
		}
	}, nil)

	warlock.registerAgony()
	warlock.registerNightfall()
	warlock.registerUnstableAffliction()
	warlock.registerMaleficEffect()
	warlock.registerMaleficGrasp()
	warlock.registerDrainSoul()
	warlock.registerDarkSoulMisery()
	warlock.registerSoulburn()
	warlock.registerSeed()
	warlock.registerSoulSwap()

	warlock.registerGlyphs()

	warlock.registerHotfixes()
}

func (warlock *AfflictionWarlock) ApplyTalents() {
	warlock.Warlock.ApplyTalents()
}

func (warlock *AfflictionWarlock) Reset(sim *core.Simulation) {
	warlock.Warlock.Reset(sim)

	warlock.LastCorruptionTarget = nil
}

func (warlock *AfflictionWarlock) OnEncounterStart(sim *core.Simulation) {
	defaultShards := MaxSoulShards
	if warlock.SoulBurnAura.IsActive() {
		defaultShards -= 1
	}

	haunt := warlock.GetSpell(core.ActionID{SpellID: HauntSpellID})
	count := warlock.SpellsInFlight[haunt]
	defaultShards -= count

	warlock.SoulShards.ResetBarTo(sim, defaultShards)
	warlock.Warlock.OnEncounterStart(sim)
}

func calculateDoTBaseTickDamage(dot *core.Dot) float64 {
	stacks := math.Max(float64(dot.Aura.GetStacks()), 1)
	return dot.SnapshotBaseDamage * dot.SnapshotAttackerMultiplier * stacks
}
