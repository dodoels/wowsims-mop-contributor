package mop

import (
	"fmt"
	"time"

	"github.com/wowsims/mop/sim/common/shared"
	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
)

type readinessTrinketConfig struct {
	itemVersionMap   shared.ItemVersionMap
	baseTrinketLabel string
	buffAuraLabel    string
	buffAuraID       int32
	buffedStat       stats.Stat
	buffDuration     time.Duration
	buffIcd          time.Duration
	cdrAuraIDs       map[proto.Spec]int32
}

type multistrikeTrinketConfig struct {
	itemVersionMap   shared.ItemVersionMap
	baseTrinketLabel string
	buffAuraLabel    string
	buffAuraID       int32
	buffedStat       stats.Stat
}

type cleaveTrinketConfig struct {
	itemVersionMap   shared.ItemVersionMap
	baseTrinketLabel string
	buffAuraLabel    string
	buffAuraID       int32
	buffedStat       stats.Stat
}

type statAmplificationTrinketConfig struct {
	itemVersionMap   shared.ItemVersionMap
	baseTrinketLabel string
	buffAuraLabel    string
	buffAuraID       int32
	buffedStat       stats.Stat
}

func init() {
	newReadinessTrinket := func(config *readinessTrinketConfig) {
		config.itemVersionMap.RegisterAll(func(version shared.ItemVersion, itemID int32, versionLabel string) {
			core.NewItemEffect(itemID, func(agent core.Agent, state proto.ItemLevelState) {
				character := agent.GetCharacter()

				auraID, exists := config.cdrAuraIDs[character.Spec]
				var cdrAura *core.Aura
				if exists {
					cdr := 1.0 / (1.0 + core.GetItemEffectScaling(itemID, 0.00989999995, state)/100)
					cdrAura = core.MakePermanent(character.RegisterAura(core.Aura{
						Label:    fmt.Sprintf("Readiness (%s)", versionLabel),
						ActionID: core.ActionID{SpellID: auraID},
					}).AttachSpellMod(core.SpellModConfig{
						Kind:       core.SpellMod_Cooldown_Multiplier,
						SpellFlag:  core.SpellFlagReadinessTrinket,
						FloatValue: cdr,
					}))
				}

				stats := stats.Stats{}
				stats[config.buffedStat] = core.GetItemEffectScaling(itemID, 2.97300004959, state)

				aura := character.NewTemporaryStatsAura(
					fmt.Sprintf("%s (%s)", config.buffAuraLabel, versionLabel),
					core.ActionID{SpellID: config.buffAuraID},
					stats,
					config.buffDuration,
				)

				triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:       fmt.Sprintf("%s (%s) - Trigger", config.baseTrinketLabel, versionLabel),
					ProcChance: 0.15,
					ICD:        config.buffIcd,
					ProcMask:   core.ProcMaskDirect | core.ProcMaskProc,
					Outcome:    core.OutcomeLanded,
					Callback:   core.CallbackOnSpellHitDealt,
					Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
						aura.Activate(sim)
					},
				})

				aura.Icd = triggerAura.Icd
				eligibleSlots := character.ItemSwap.EligibleSlotsForItem(itemID)
				character.AddStatProcBuff(itemID, aura, false, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, triggerAura, eligibleSlots)
				if cdrAura != nil {
					character.ItemSwap.RegisterProcWithSlots(itemID, cdrAura, eligibleSlots)
				}
			})
		})
	}

	// Assurance of Consequence
	// Increases the cooldown recovery rate of six of your major abilities by 47%.
	// Effective for Agility-based damage roles only.
	//
	// Your attacks have a chance to grant you 14039 Agility for 20 sec.
	// (15% chance, 115 sec cooldown) (Proc chance: 15%, 1.917m cooldown)
	newReadinessTrinket(&readinessTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             104974,
			shared.ItemVersionNormal:          102292,
			shared.ItemVersionHeroic:          104476,
			shared.ItemVersionWarforged:       105223,
			shared.ItemVersionHeroicWarforged: 105472,
			shared.ItemVersionFlexible:        104725,
		},
		baseTrinketLabel: "Assurance of Consequence",
		buffAuraLabel:    "Dextrous",
		buffAuraID:       146308,
		buffedStat:       stats.Agility,
		buffDuration:     time.Second * 20,
		buffIcd:          time.Second * 115,
		cdrAuraIDs: map[proto.Spec]int32{
			// Druid
			// Missing: Bear Hug, Ironbark, Nature's Swiftness
			proto.Spec_SpecFeralDruid:       145961,
			proto.Spec_SpecGuardianDruid:    145962,
			proto.Spec_SpecRestorationDruid: 145963,
			// Hunter
			proto.Spec_SpecBeastMasteryHunter: 145964,
			proto.Spec_SpecMarksmanshipHunter: 145965,
			proto.Spec_SpecSurvivalHunter:     145966,
			// Rogue
			// Missing: Cloak of Shadows, Evasion, JuJu Escape
			proto.Spec_SpecAssassinationRogue: 145983,
			proto.Spec_SpecCombatRogue:        145984,
			proto.Spec_SpecSubtletyRogue:      145985,
			// Priest - NOTE: Priests seem to have a Aura for this
			// Missing: Divine Hymn, Guardian Spirit, Hymn of Hope, Inner Focus, Pain Suppression, Power Word: Barrier, Void Shift
			// proto.Spec_SpecDisciplinePriest: 145981,
			// proto.Spec_SpecHolyPriest:       145982,
			// Shaman
			// Missing: Mana Tide Totem, Spirit Link Totem
			proto.Spec_SpecEnhancementShaman: 145986,
			proto.Spec_SpecRestorationShaman: 145988,
			// Monk
			// Missing: Zen Meditation, Life Cocoon, Revival, Thunder Focus Tea, Flying Serpent Kick
			proto.Spec_SpecBrewmasterMonk: 145967,
			proto.Spec_SpecMistweaverMonk: 145968,
			proto.Spec_SpecWindwalkerMonk: 145969,
		},
	})

	// Evil Eye of Galakras
	// Increases the cooldown recovery rate of six of your major abilities by 1%. Effective for Strength-based
	// damage roles only.
	//
	// Your attacks have a chance to grant you 11761 Strength for 10 sec.
	// (15% chance, 55 sec cooldown) (Proc chance: 15%, 55s cooldown)
	newReadinessTrinket(&readinessTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             104993,
			shared.ItemVersionNormal:          102298,
			shared.ItemVersionHeroic:          104495,
			shared.ItemVersionWarforged:       105242,
			shared.ItemVersionHeroicWarforged: 105491,
			shared.ItemVersionFlexible:        104744,
		},
		baseTrinketLabel: "Evil Eye of Galakras",
		buffAuraLabel:    "Outrage",
		buffAuraID:       146245,
		buffedStat:       stats.Strength,
		buffDuration:     time.Second * 10,
		buffIcd:          time.Second * 55,
		cdrAuraIDs: map[proto.Spec]int32{
			// Death Knight
			proto.Spec_SpecBloodDeathKnight:  145958,
			proto.Spec_SpecFrostDeathKnight:  145959,
			proto.Spec_SpecUnholyDeathKnight: 145960,
			// Paladin
			// Missing: Divine Plea, Hand Of Protection, Divine Shield, Hand Of Purity
			proto.Spec_SpecHolyPaladin:        145978,
			proto.Spec_SpecProtectionPaladin:  145976,
			proto.Spec_SpecRetributionPaladin: 145975,
			// Warrior
			// Missing: Die by the Sword, Mocking Banner
			proto.Spec_SpecArmsWarrior:       145990,
			proto.Spec_SpecFuryWarrior:       145991,
			proto.Spec_SpecProtectionWarrior: 145992,
		},
	})

	getTrinketSpell := func(character *core.Character, spellID int32, spellSchool core.SpellSchool) *core.Spell {
		return character.GetOrRegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: spellID},
			SpellSchool: spellSchool,
			ProcMask:    core.ProcMaskEmpty,
			Flags:       core.SpellFlagIgnoreArmor | core.SpellFlagIgnoreModifiers | core.SpellFlagPassiveSpell | core.SpellFlagNoSpellMods,

			DamageMultiplier: 1,
			ThreatMultiplier: 1,
		})
	}

	getMultistrikeSpells := func(character *core.Character) (*core.Spell, *core.Spell) {
		var physicalSpellID int32
		if character.Class == proto.Class_ClassHunter {
			physicalSpellID = 146069
		} else {
			physicalSpellID = 146061
		}

		physicalSpell := getTrinketSpell(character, physicalSpellID, core.SpellSchoolPhysical)
		magicSpell := physicalSpell

		switch character.Class {
		case proto.Class_ClassDruid:
			magicSpell = getTrinketSpell(character, 146064, core.SpellSchoolArcane)
		case proto.Class_ClassMage:
			var magicSpellID int32
			var school core.SpellSchool
			if character.Spec == proto.Spec_SpecArcaneMage {
				magicSpellID = 146070
				school = core.SpellSchoolArcane
			} else {
				magicSpellID = 146067
				school = core.SpellSchoolFrostfire
			}
			magicSpell = getTrinketSpell(character, magicSpellID, school)
		case proto.Class_ClassMonk:
			magicSpell = getTrinketSpell(character, 146075, core.SpellSchoolNature)
		case proto.Class_ClassPriest:
			var magicSpellID int32
			var school core.SpellSchool
			if character.Spec == proto.Spec_SpecShadowPriest {
				magicSpellID = 146065
				school = core.SpellSchoolShadow
			} else {
				magicSpellID = 146063
				school = core.SpellSchoolHoly
			}
			magicSpell = getTrinketSpell(character, magicSpellID, school)
		case proto.Class_ClassShaman:
			magicSpell = getTrinketSpell(character, 146071, core.SpellSchoolNature)
		case proto.Class_ClassWarlock:
			magicSpell = getTrinketSpell(character, 146065, core.SpellSchoolShadow)
		}

		return physicalSpell, magicSpell
	}

	blackoutKickTickID := core.ActionID{SpellID: 100784}.WithTag(2)
	newMultistrikeTrinket := func(config *multistrikeTrinketConfig) {
		config.itemVersionMap.RegisterAll(func(version shared.ItemVersion, itemID int32, versionLabel string) {
			core.NewItemEffect(itemID, func(agent core.Agent, state proto.ItemLevelState) {
				character := agent.GetCharacter()

				var baseDamage float64
				applyEffects := func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeAlwaysHit)
				}

				physicalSpell, magicSpell := getMultistrikeSpells(character)

				multistrikeTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:               fmt.Sprintf("%s (%s) - Multistrike Trigger", config.baseTrinketLabel, versionLabel),
					ProcChance:         core.GetItemEffectScaling(itemID, 0.03539999947, state) / 1000,
					Outcome:            core.OutcomeLanded,
					Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt,
					RequireDamageDealt: true,

					ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
						return spell.ProcMask != core.ProcMaskEmpty
					},

					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						baseDamage = result.Damage / 3.0

						// Special case for Windwalker Blackout Kick DoTs which does physical damage but procs the nature damage spell
						if spell.SpellSchool.Matches(core.SpellSchoolPhysical) && !spell.ActionID.SameAction(blackoutKickTickID) {
							physicalSpell.ApplyEffects = applyEffects
							physicalSpell.Cast(sim, result.Target)
						} else {
							magicSpell.ApplyEffects = applyEffects
							magicSpell.Cast(sim, result.Target)
						}
					},
				})

				stats := stats.Stats{}
				stats[config.buffedStat] = core.GetItemEffectScaling(itemID, 2.97300004959, state)

				statBuffAura := character.NewTemporaryStatsAura(
					fmt.Sprintf("%s (%s)", config.buffAuraLabel, versionLabel),
					core.ActionID{SpellID: config.buffAuraID},
					stats,
					time.Second*10,
				)

				statBuffTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:     fmt.Sprintf("%s (%s) - Stat Trigger", config.baseTrinketLabel, versionLabel),
					ICD:      time.Second * 10,
					Outcome:  core.OutcomeLanded,
					Callback: core.CallbackOnSpellHitDealt,

					DPM: character.NewRPPMProcManager(itemID, false, false, core.ProcMaskDirect|core.ProcMaskProc, core.RPPMConfig{
						PPM: 0.92000001669,
					}),

					Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
						statBuffAura.Activate(sim)
					},
				})

				statBuffAura.Icd = statBuffTriggerAura.Icd

				eligibleSlots := character.ItemSwap.EligibleSlotsForItem(itemID)
				character.AddStatProcBuff(itemID, statBuffAura, false, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, statBuffTriggerAura, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, multistrikeTriggerAura, eligibleSlots)
			})
		})
	}

	// Haromm's Talisman
	// Your attacks have a 16.7% chance to trigger Multistrike, which deals instant additional damage to your target equal to 33% of the original damage dealt.
	//
	// Your attacks have a chance to grant you 14039 Agility for 10 sec.
	// (Approximately 0.92 procs per minute)
	newMultistrikeTrinket(&multistrikeTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             105029,
			shared.ItemVersionNormal:          102301,
			shared.ItemVersionHeroic:          104531,
			shared.ItemVersionWarforged:       105278,
			shared.ItemVersionHeroicWarforged: 105527,
			shared.ItemVersionFlexible:        104780,
		},
		baseTrinketLabel: "Haromm's Talisman",
		buffAuraLabel:    "Vicious",
		buffAuraID:       148903,
		buffedStat:       stats.Agility,
	})

	// Kardris' Toxic Totem
	// Your attacks have a 16.7% chance to trigger Multistrike, which deals instant additional damage to your target equal to 33% of the original damage dealt.
	//
	// Your attacks have a chance to grant 14039 Intellect for 10 sec.
	// (Approximately 0.92 procs per minute)
	newMultistrikeTrinket(&multistrikeTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             105042,
			shared.ItemVersionNormal:          102300,
			shared.ItemVersionHeroic:          104544,
			shared.ItemVersionWarforged:       105291,
			shared.ItemVersionHeroicWarforged: 105540,
			shared.ItemVersionFlexible:        104793,
		},
		baseTrinketLabel: "Kardris' Toxic Totem",
		buffAuraLabel:    "Toxic Power",
		buffAuraID:       148906,
		buffedStat:       stats.Intellect,
	})

	newStatAmplificationTrinket := func(config *statAmplificationTrinketConfig) {
		config.itemVersionMap.RegisterAll(func(version shared.ItemVersion, itemID int32, versionLabel string) {
			core.NewItemEffect(itemID, func(agent core.Agent, state proto.ItemLevelState) {
				character := agent.GetCharacter()

				critDamageValue := 1 + core.GetItemEffectScaling(itemID, 0.00088499999, state)/100
				hasteValue := 1 + core.GetItemEffectScaling(itemID, 0.00176999997, state)/100
				masteryValue := 1 + core.GetItemEffectScaling(itemID, 0.00176999997, state)/100
				spiritValue := 1 + core.GetItemEffectScaling(itemID, 0.00176999997, state)/100

				statAura := core.MakePermanent(character.RegisterAura(core.Aura{
					Label:      fmt.Sprintf("Amplification (%s)", versionLabel),
					ActionID:   core.ActionID{SpellID: 146051},
					BuildPhase: core.CharacterBuildPhaseGear,
				})).
					AttachStatDependency(character.NewDynamicMultiplyStat(stats.HasteRating, hasteValue)).
					AttachStatDependency(character.NewDynamicMultiplyStat(stats.MasteryRating, masteryValue)).
					AttachStatDependency(character.NewDynamicMultiplyStat(stats.Spirit, spiritValue)).
					AttachMultiplicativePseudoStatBuff(&character.PseudoStats.CritDamageMultiplier, critDamageValue)

				stats := stats.Stats{}
				stats[config.buffedStat] = core.GetItemEffectScaling(itemID, 2.97300004959, state)

				aura := character.NewTemporaryStatsAura(
					fmt.Sprintf("%s (%s)", config.buffAuraLabel, versionLabel),
					core.ActionID{SpellID: config.buffAuraID},
					stats,
					time.Second*20,
				)

				triggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:       fmt.Sprintf("%s (%s)", config.baseTrinketLabel, versionLabel),
					Callback:   core.CallbackOnSpellHitDealt,
					Outcome:    core.OutcomeLanded,
					ICD:        time.Second * 115,
					ProcChance: 0.15,

					Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
						aura.Activate(sim)
					},
				})

				eligibleSlots := character.ItemSwap.EligibleSlotsForItem(itemID)
				character.AddStatProcBuff(itemID, aura, false, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, triggerAura, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, statAura, eligibleSlots)
			})
		})
	}

	// Thok's Tail Tip
	// Your attacks have a chance to grant you 14039 Strength for 20 sec.
	// (15% chance, 115 sec cooldown) (Proc chance: 15%, 1.917m cooldown)
	// Amplifies your Critical Strike damage and healing, Haste, Mastery, and Spirit by 1%.
	newStatAmplificationTrinket(&statAmplificationTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             105111,
			shared.ItemVersionNormal:          102305,
			shared.ItemVersionHeroic:          104613,
			shared.ItemVersionWarforged:       105360,
			shared.ItemVersionHeroicWarforged: 105609,
			shared.ItemVersionFlexible:        104862,
		},
		baseTrinketLabel: "Thok's Tail Tip",
		buffAuraLabel:    "Determination",
		buffAuraID:       146250,
		buffedStat:       stats.Strength,
	})

	// Purified Bindings of Immerseus
	// Your attacks have a chance to grant 14039 Intellect for 20 sec.
	// (15% chance, 115 sec cooldown) (Proc chance: 15%, 1.917m cooldown)
	// Amplifies your Critical Strike damage and healing, Haste, Mastery, and Spirit by 1%.
	newStatAmplificationTrinket(&statAmplificationTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             104924,
			shared.ItemVersionNormal:          102293,
			shared.ItemVersionHeroic:          104426,
			shared.ItemVersionWarforged:       105173,
			shared.ItemVersionHeroicWarforged: 105422,
			shared.ItemVersionFlexible:        104675,
		},
		baseTrinketLabel: "Purified Bindings of Immerseus",
		buffAuraLabel:    "Expanded Mind",
		buffAuraID:       146046,
		buffedStat:       stats.Intellect,
	})

	// Ticking Ebon Detonator
	// Your melee and ranged attacks have a chance to grant you 19260 Agility for 10s. Every 0.5 sec this effect
	// decrements by 963 Agility.
	// (Approximately 1.00 procs per minute)
	shared.ItemVersionMap{
		shared.ItemVersionLFR:             105114,
		shared.ItemVersionNormal:          102311,
		shared.ItemVersionHeroic:          104616,
		shared.ItemVersionWarforged:       105363,
		shared.ItemVersionHeroicWarforged: 105612,
		shared.ItemVersionFlexible:        104865,
	}.RegisterAll(func(version shared.ItemVersion, itemID int32, versionLabel string) {
		core.NewItemEffect(itemID, func(agent core.Agent, state proto.ItemLevelState) {
			character := agent.GetCharacter()

			statBuffAura := core.MakeStackingAura(character, core.StackingStatAura{
				Aura: core.Aura{
					Label:     fmt.Sprintf("Restless Agility (%s)", versionLabel),
					ActionID:  core.ActionID{SpellID: 146310},
					Duration:  time.Second * 10,
					MaxStacks: 20,
				},
				BonusPerStack: stats.Stats{
					stats.Agility: core.GetItemEffectScaling(itemID, 0.27030000091, state),
				},
			})

			statBuffTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
				Name:     fmt.Sprintf("Ticking Ebon Detonator (%s) - Stat Trigger", versionLabel),
				ICD:      time.Second * 10,
				Outcome:  core.OutcomeLanded,
				Callback: core.CallbackOnSpellHitDealt,

				DPM: character.NewRPPMProcManager(itemID, false, false, core.ProcMaskDirect, core.RPPMConfig{
					PPM: 1,
				}),

				Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
					statBuffAura.Activate(sim)
					statBuffAura.SetStacks(sim, statBuffAura.MaxStacks)
					core.StartPeriodicAction(sim, core.PeriodicActionOptions{
						Period:   time.Millisecond * 500,
						NumTicks: 20,
						OnAction: func(sim *core.Simulation) {
							// Aura might not be active because of stuff like mage alter time being cast right before this aura being activated
							if statBuffAura.IsActive() {
								statBuffAura.RemoveStack(sim)
							}
						},
					})
				},
			})

			statBuffAura.Icd = statBuffTriggerAura.Icd

			eligibleSlots := character.ItemSwap.EligibleSlotsForItem(itemID)
			character.AddStatProcBuff(itemID, statBuffAura, false, eligibleSlots)
			character.ItemSwap.RegisterProcWithSlots(itemID, statBuffTriggerAura, eligibleSlots)
		})
	})

	getCleaveSpells := func(character *core.Character) (*core.Spell, *core.Spell) {
		var physicalSpellID int32
		if character.Class == proto.Class_ClassHunter {
			physicalSpellID = 146162
		} else {
			physicalSpellID = 146137
		}

		physicalSpell := getTrinketSpell(character, physicalSpellID, core.SpellSchoolPhysical)

		var magicSpell *core.Spell
		switch character.Class {
		case proto.Class_ClassDruid:
			magicSpell = getTrinketSpell(character, 146158, core.SpellSchoolArcane)
		case proto.Class_ClassMage:
			var magicSpellID int32
			var school core.SpellSchool
			if character.Spec == proto.Spec_SpecArcaneMage {
				magicSpellID = 146166
				school = core.SpellSchoolArcane
			} else {
				magicSpellID = 146160
				school = core.SpellSchoolFrostfire
			}
			magicSpell = getTrinketSpell(character, magicSpellID, school)
		case proto.Class_ClassMonk:
			magicSpell = getTrinketSpell(character, 146172, core.SpellSchoolNature)
		case proto.Class_ClassPaladin:
			magicSpell = getTrinketSpell(character, 146157, core.SpellSchoolHoly)
		case proto.Class_ClassPriest:
			var magicSpellID int32
			var school core.SpellSchool
			if character.Spec == proto.Spec_SpecShadowPriest {
				magicSpellID = 146159
				school = core.SpellSchoolShadow
			} else {
				magicSpellID = 146157
				school = core.SpellSchoolHoly
			}
			magicSpell = getTrinketSpell(character, magicSpellID, school)
		case proto.Class_ClassShaman:
			magicSpell = getTrinketSpell(character, 146171, core.SpellSchoolNature)
		case proto.Class_ClassWarlock:
			magicSpell = getTrinketSpell(character, 146159, core.SpellSchoolShadow)
		}

		return physicalSpell, magicSpell
	}

	newCleaveTrinket := func(config *cleaveTrinketConfig) {
		config.itemVersionMap.RegisterAll(func(version shared.ItemVersion, itemID int32, versionLabel string) {
			core.NewItemEffect(itemID, func(agent core.Agent, state proto.ItemLevelState) {
				character := agent.GetCharacter()

				physicalSpell, magicSpell := getCleaveSpells(character)

				var baseDamage float64
				applyEffects := func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
					numTargets := min(5, sim.Environment.ActiveTargetCount()-1)
					curTarget := sim.Environment.NextActiveTargetUnit(target)
					var outcome core.OutcomeApplier
					if character.Class == proto.Class_ClassHunter {
						outcome = spell.OutcomeRangedHit
					} else if spell.SpellSchool == core.SpellSchoolPhysical {
						outcome = spell.OutcomeMeleeSpecialHit
					} else {
						outcome = spell.OutcomeMagicHit
					}

					for range numTargets {
						spell.CalcAndDealDamage(sim, curTarget, baseDamage, outcome)
						curTarget = sim.Environment.NextActiveTargetUnit(curTarget)
					}
				}

				cleaveTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:               fmt.Sprintf("%s (%s) - Cleave Trigger", config.baseTrinketLabel, versionLabel),
					ProcChance:         core.GetItemEffectScaling(itemID, 0.07859999686, state) / 10000,
					Outcome:            core.OutcomeLanded,
					Callback:           core.CallbackOnSpellHitDealt | core.CallbackOnPeriodicDamageDealt,
					RequireDamageDealt: true,

					ExtraCondition: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) bool {
						return sim.Environment.ActiveTargetCount() > 1
					},

					Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
						baseDamage = result.Damage

						if magicSpell == nil || !spell.ProcMask.Matches(core.ProcMaskSpellOrSpellProc) {
							physicalSpell.ApplyEffects = applyEffects
							physicalSpell.Cast(sim, result.Target)
						} else {
							magicSpell.ApplyEffects = applyEffects
							magicSpell.Cast(sim, result.Target)
						}
					},
				})

				stats := stats.Stats{}
				stats[config.buffedStat] = core.GetItemEffectScaling(itemID, 2.97300004959, state)

				statBuffAura := character.NewTemporaryStatsAura(
					fmt.Sprintf("%s (%s)", config.buffAuraLabel, versionLabel),
					core.ActionID{SpellID: config.buffAuraID},
					stats,
					time.Second*15,
				)

				statBuffTriggerAura := character.MakeProcTriggerAura(core.ProcTrigger{
					Name:       fmt.Sprintf("%s (%s) - Stat Trigger", config.baseTrinketLabel, versionLabel),
					ICD:        time.Second * 85,
					Outcome:    core.OutcomeLanded,
					ProcMask:   core.ProcMaskDirect | core.ProcMaskProc,
					Callback:   core.CallbackOnSpellHitDealt,
					ProcChance: 0.15,

					Handler: func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
						statBuffAura.Activate(sim)
					},
				})

				statBuffAura.Icd = statBuffTriggerAura.Icd

				eligibleSlots := character.ItemSwap.EligibleSlotsForItem(itemID)
				character.AddStatProcBuff(itemID, statBuffAura, false, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, statBuffTriggerAura, eligibleSlots)
				character.ItemSwap.RegisterProcWithSlots(itemID, cleaveTriggerAura, eligibleSlots)
			})
		})
	}

	// Fusion-Fire Core
	// Your attacks have a 0.01% chance to Cleave, dealing the same damage to up to 5 other nearby targets.
	//
	// Your attacks have a chance to grant you 14039 Strength for 15 sec.
	// (15% chance, 85 sec cooldown) (Proc chance: 15%, 1.417m cooldown)
	newCleaveTrinket(&cleaveTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             104961,
			shared.ItemVersionNormal:          102295,
			shared.ItemVersionHeroic:          104463,
			shared.ItemVersionWarforged:       105210,
			shared.ItemVersionHeroicWarforged: 105459,
			shared.ItemVersionFlexible:        104712,
		},
		baseTrinketLabel: "Fusion-Fire Core",
		buffAuraLabel:    "Tenacious",
		buffAuraID:       148899,
		buffedStat:       stats.Strength,
	})

	// Sigil of Rampage
	// Your attacks have a 0.01% chance to Cleave, dealing the same damage to up to 5 other nearby targets.
	//
	// Your attacks have a chance to grant you 14039 Agility for 15 sec.
	// (15% chance, 85 sec cooldown) (Proc chance: 15%, 1.417m cooldown)
	newCleaveTrinket(&cleaveTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             105082,
			shared.ItemVersionNormal:          102302,
			shared.ItemVersionHeroic:          104584,
			shared.ItemVersionWarforged:       105331,
			shared.ItemVersionHeroicWarforged: 105580,
			shared.ItemVersionFlexible:        104833,
		},
		baseTrinketLabel: "Sigil of Rampage",
		buffAuraLabel:    "Ferocity",
		buffAuraID:       148896,
		buffedStat:       stats.Agility,
	})

	// Frenzied Crystal of Rage
	// Your attacks have a 0.01% chance to Cleave, dealing the same damage to up to 5 other nearby targets.
	//
	// Your attacks have a chance to grant you 14039 Intellect for 15 sec.
	// (15% chance, 85 sec cooldown) (Proc chance: 15%, 1.417m cooldown)
	newCleaveTrinket(&cleaveTrinketConfig{
		itemVersionMap: shared.ItemVersionMap{
			shared.ItemVersionLFR:             105074,
			shared.ItemVersionNormal:          102303,
			shared.ItemVersionHeroic:          104576,
			shared.ItemVersionWarforged:       105323,
			shared.ItemVersionHeroicWarforged: 105572,
			shared.ItemVersionFlexible:        104825,
		},
		baseTrinketLabel: "Frenzied Crystal of Rage",
		buffAuraLabel:    "Extravagant Visions",
		buffAuraID:       148897,
		buffedStat:       stats.Intellect,
	})
}
