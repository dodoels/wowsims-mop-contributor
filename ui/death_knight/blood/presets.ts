import * as PresetUtils from '../../core/preset_utils.js';
import { ConsumesSpec, Glyphs, Profession, PseudoStat, Spec, Stat } from '../../core/proto/common.js';
import { BloodDeathKnight_Options, DeathKnightMajorGlyph, DeathKnightMinorGlyph } from '../../core/proto/death_knight';
import { SavedTalents } from '../../core/proto/ui.js';
import { Stats } from '../../core/proto_utils/stats';
import ShaApl from './apls/sha.apl.json';
import HorridonApl from './apls/horridon.apl.json';
import P2BalancedBloodGear from './gear_sets/p2.gear.json';
import P2OffensiveBloodGear from './gear_sets/p2_offensive.gear.json';
import P3BalancedBloodGear from './gear_sets/p3.gear.json';
import P3OffensiveBloodGear from './gear_sets/p3_offensive.gear.json';
import DefaultBuild from './builds/sha_default.build.json';
import ShaBuild from './builds/sha_encounter_only.build.json';
import HorridonBuild from './builds/horridon_encounter_only.build.json';
// import PreRaidBloodGear from './gear_sets/preraid.gear.json';

// export const PRERAID_BLOOD_PRESET = PresetUtils.makePresetGear('Pre-Raid', PreRaidBloodGear);
export const P2_BALANCED_BLOOD_PRESET = PresetUtils.makePresetGear('P2 - BIS (Balanced)', P2BalancedBloodGear);
export const P2_OFFENSIVE_BLOOD_PRESET = PresetUtils.makePresetGear('P2 - BIS (Offensive)', P2OffensiveBloodGear);
export const P3_BALANCED_BLOOD_PRESET = PresetUtils.makePresetGear('P3 - BIS (Balanced)', P3BalancedBloodGear);
export const P3_OFFENSIVE_BLOOD_PRESET = PresetUtils.makePresetGear('P3 - BIS (Offensive)', P3OffensiveBloodGear);

export const BLOOD_ROTATION_PRESET_SHA = PresetUtils.makePresetAPLRotation('Sha of Fear', ShaApl);
export const BLOOD_ROTATION_PRESET_HORRIDON = PresetUtils.makePresetAPLRotation('Horridon', HorridonApl);

// Preset options for EP weights
export const P2_BALANCED_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P2 - Balanced',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.00,
			[Stat.StatStamina]: 1.30,
			[Stat.StatHitRating]: 1.16,
			[Stat.StatCritRating]: 0.64,
			[Stat.StatHasteRating]: 0.58,
			[Stat.StatExpertiseRating]: 1.02,
			[Stat.StatDodgeRating]: 0.50,
			[Stat.StatParryRating]: 0.69,
			[Stat.StatMasteryRating]: 0.62,
			[Stat.StatAttackPower]: 0.25,
			[Stat.StatArmor]: 0.64,
			[Stat.StatBonusArmor]: 0.64,
		},
	{
			[PseudoStat.PseudoStatMainHandDps]: 2.70,
		},
	),
);

export const P2_OFFENSIVE_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P2 - Offensive',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.00,
			[Stat.StatStamina]: 0.71,
			[Stat.StatHitRating]: 1.46,
			[Stat.StatCritRating]: 0.80,
			[Stat.StatHasteRating]: 0.62,
			[Stat.StatExpertiseRating]: 1.27,
			[Stat.StatDodgeRating]: 0.52,
			[Stat.StatParryRating]: 0.64,
			[Stat.StatMasteryRating]: 0.34,
			[Stat.StatAttackPower]: 0.32,
			[Stat.StatArmor]: 0.35,
			[Stat.StatBonusArmor]: 0.35,
		},
	{
			[PseudoStat.PseudoStatMainHandDps]: 2.90,
		},
	),
);

export const P3_SURVIVAL_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3 - Survival',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.00,
			[Stat.StatStamina]: 1.35,
			[Stat.StatHitRating]: 1.77,
			[Stat.StatCritRating]: 0.59,
			[Stat.StatHasteRating]: 0.97,
			[Stat.StatExpertiseRating]: 1.23,
			[Stat.StatDodgeRating]: 0.58,
			[Stat.StatParryRating]: 0.95,
			[Stat.StatMasteryRating]: 1.65,
			[Stat.StatAttackPower]: 0.17,
			[Stat.StatArmor]: 0.84,
			[Stat.StatBonusArmor]: 0.84,
		},
	{
			[PseudoStat.PseudoStatMainHandDps]: 1.84,
		},
	),
);

export const P3_BALANCED_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3 - Balanced',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.00,
			[Stat.StatStamina]: 0.83,
			[Stat.StatHitRating]: 2.17,
			[Stat.StatCritRating]: 0.88,
			[Stat.StatHasteRating]: 1.04,
			[Stat.StatExpertiseRating]: 1.66,
			[Stat.StatDodgeRating]: 0.71,
			[Stat.StatParryRating]: 0.97,
			[Stat.StatMasteryRating]: 1.01,
			[Stat.StatAttackPower]: 0.25,
			[Stat.StatArmor]: 0.52,
			[Stat.StatBonusArmor]: 0.52,
		},
	{
			[PseudoStat.PseudoStatMainHandDps]: 1.94,
		},
	),
);

export const P3_OFFENSIVE_EP_PRESET = PresetUtils.makePresetEpWeights(
	'P3 - Offensive',
	Stats.fromMap(
		{
			[Stat.StatStrength]: 1.00,
			[Stat.StatStamina]: 0.42,
			[Stat.StatHitRating]: 2.49,
			[Stat.StatCritRating]: 1.11,
			[Stat.StatHasteRating]: 1.09,
			[Stat.StatExpertiseRating]: 2.02,
			[Stat.StatDodgeRating]: 0.81,
			[Stat.StatParryRating]: 0.99,
			[Stat.StatMasteryRating]: 0.49,
			[Stat.StatAttackPower]: 0.31,
			[Stat.StatArmor]: 0.26,
			[Stat.StatBonusArmor]: 0.26,
		},
	{
			[PseudoStat.PseudoStatMainHandDps]: 2.15,
		},
	),
);

// Default talents. Uses the wowhead calculator format, make the talents on
// https://wotlk.wowhead.com/talent-calc and copy the numbers in the url.

export const BloodTalents = {
	name: 'Default',
	data: SavedTalents.create({
		talentsString: "231111",
		glyphs: Glyphs.create({
			major1: DeathKnightMajorGlyph.GlyphOfLoudHorn,
			major2: DeathKnightMajorGlyph.GlyphOfRegenerativeMagic,
			major3: DeathKnightMajorGlyph.GlyphOfIceboundFortitude,
			minor1: DeathKnightMinorGlyph.GlyphOfTheLongWinter,
			minor2: DeathKnightMinorGlyph.GlyphOfArmyOfTheDead,
			minor3: DeathKnightMinorGlyph.GlyphOfResilientGrip,
		}),
	}),
};

export const DefaultOptions = BloodDeathKnight_Options.create({
	classOptions: {},
});

export const DefaultConsumables = ConsumesSpec.create({
	flaskId: 76087, // Flask of the Earth
	foodId: 74656, // Chun Tian Spring Rolls
	potId: 76095, // Potion of Mogu Power
	prepotId: 76095, // Potion of Mogu Power
});

export const OtherDefaults = {
	profession1: Profession.Engineering,
	profession2: Profession.Blacksmithing,
	distanceFromTarget: 5,
	iterationCount: 25000,
};

export const PRESET_BUILD_DEFAULT = PresetUtils.makePresetBuildFromJSON("Default", Spec.SpecBloodDeathKnight, DefaultBuild);
export const PRESET_BUILD_SHA = PresetUtils.makePresetBuildFromJSON("Sha of Fear P2", Spec.SpecBloodDeathKnight, ShaBuild);
export const PRESET_BUILD_HORRIDON = PresetUtils.makePresetBuildFromJSON("Horridon P2", Spec.SpecBloodDeathKnight, HorridonBuild);
