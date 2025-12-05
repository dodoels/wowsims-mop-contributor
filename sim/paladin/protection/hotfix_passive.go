package protection

func (prot *ProtectionPaladin) registerHotfixPassive() {
	// EffectIndex 2 on the Protection specific Hotfix Passive https://wago.tools/db2/SpellEffect?build=5.5.0.61916&filter%5BSpellID%5D=137028&page=1
	// 2025-11-13: 5.5.0 increase to Seal of Insight attack speed, lowered to 5%. (was 10%)
	prot.SealOfInsightAura.AttachMultiplyAttackSpeed(1.05)
}
