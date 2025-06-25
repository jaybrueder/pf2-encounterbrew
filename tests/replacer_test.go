package tests

import (
	"testing"
	"pf2.encounterbrew.com/internal/utils"
)

func TestReplacer(t *testing.T) {
	replacer := utils.NewReplacer()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple Check",
			input:    "@Check[arcana|dc:40]",
			expected: "DC 40 arcana",
		},
		{
			name:     "Basic Check",
			input:    "@Check[reflex|dc:19|basic]",
			expected: "DC 19 basic reflex",
		},
		{
			name:     "Check with ShowDC",
			input:    "@Check[flat|showDC:all|dc:11]",
			expected: "DC 11 flat check",
		},
		{
			name:     "Check with Traits",
			input:    "@Check[thievery|dc:25|traits:action:disable-a-device]",
			expected: "DC 25 thievery",
		},
		{
			name:     "Simple Damage",
			input:    "@Damage[11d6[acid]]",
			expected: "11d6 acid",
		},
		{
			name:     "Complex Damage",
			input:    "@Damage[(2d8 + 5)[bludgeoning]]",
			expected: "2d8 + 5 bludgeoning",
		},
		{
			name:     "Damage with Level",
			input:    "@Damage[(@item.level)[bleed]]",
			expected: "bleed",
		},
		{
			name:     "Nested Damage",
			input:    "@Damage[18d6[acid]]",
			expected: "18d6 acid",
		},
		{
			name:     "UUID with Custom Text",
			input:    "@UUID[Compendium.pf2e.conditionitems.Item.Sickened]{Sickened 1}",
			expected: "Sickened 1",
		},
		{
			name:     "UUID with Custom Text - Condition Level",
			input:    "@UUID[Compendium.pf2e.conditionitems.Item.Stunned]{Stunned 2}",
			expected: "Stunned 2",
		},
		{
			name:     "UUID Simple Spell",
			input:    "@UUID[Compendium.pf2e.spells-srd.Item.Bind Undead]",
			expected: "Bind Undead",
		},
		{
			name:     "UUID with Hyphenated Name",
			input:    "@UUID[Compendium.pf2e.equipment.Item.Magic-Wand]",
			expected: "Magic Wand",
		},
		{
			name:     "UUID with ID",
			input:    "@UUID[Compendium.pf2e.conditionitems.Item.kWc1fhmv9LBiTuei]",
			expected: "condition",
		},
		{
			name:     "UUID Action",
			input:    "@UUID[Compendium.pf2e.actionspf2e.Item.Grapple]",
			expected: "Grapple",
		},
		{
			name:     "UUID Actor",
			input:    "@UUID[Compendium.pf2e.pathfinder-bestiary.Actor.Shadow Spawn]",
			expected: "Spawn",
		},
		{
			name:     "Template Simple",
			input:    "@Template[line|distance:100]",
			expected: "line (100 feet)",
		},
		{
			name:     "Template with Traits",
			input:    "@Template[emanation|distance:30|traits:arcane,transmutation]",
			expected: "emanation (30 feet, arcane, transmutation)",
		},
		{
			name:     "Template with Multiple Traits",
			input:    "@Template[cone|distance:60|traits:fire,evocation,magical]",
			expected: "cone (60 feet, fire, evocation, magical)",
		},
		{
			name:     "Localize Pattern",
			input:    "@Localize[PF2E.Item.Weapon.Base.club]",
			expected: "club",
		},
		{
			name:     "Localize with Custom Text",
			input:    "@Localize[PF2E.TraitDescription.fire]{Fire Damage}",
			expected: "Fire Damage",
		},
		{
			name:     "Multiple patterns in text",
			input:    "The creature must make a @Check[reflex|dc:19] save or take @Damage[2d6[fire]] damage.",
			expected: "The creature must make a DC 19 reflex save or take 2d6 fire damage.",
		},
		{
			name:     "Complex sentence with multiple patterns",
			input:    "Cast @UUID[Compendium.pf2e.spells-srd.Item.Fireball] in a @Template[burst|distance:20] dealing @Damage[(8d6)[fire]] damage (@Check[reflex|dc:23|basic]).",
			expected: "Cast Fireball in a burst (20 feet) dealing 8d6 fire damage (DC 23 basic reflex).",
		},
		{
			name:     "Formatted rolls - br",
			input:    "[[/br 1d20+5]]{Initiative Roll}",
			expected: "Initiative Roll",
		},
		{
			name:     "Formatted rolls - r",
			input:    "[[/r 2d8+3]]{Damage Roll}",
			expected: "Damage Roll",
		},
		{
			name:     "Formatted rolls - simple",
			input:    "[[/r 1d20+10]]",
			expected: "1d20+10",
		},
		{
			name:     "Formatted rolls - gmr",
			input:    "[[/gmr 1d4 #Recharge Devastating Blast]]{1d4 rounds}",
			expected: "1d4 rounds",
		},
		{
			name:     "Formatted rolls - gmr with different content",
			input:    "[[/gmr 2d6 #Fire Breath]]{2d6 damage}",
			expected: "2d6 damage",
		},
		{
			name:     "Unknown pattern with value",
			input:    "@NewPattern[something|value:test]",
			expected: "test",
		},
		{
			name:     "Unknown pattern simple",
			input:    "@CustomType[simple-content]",
			expected: "simple-content",
		},
		{
			name:     "Unknown pattern fallback",
			input:    "@UnknownType[very|complex|content|with|many|parts]",
			expected: "unknowntype",
		},
		{
			name:     "Edge case - empty content",
			input:    "@Check[]",
			expected: "",
		},
		{
			name:     "Edge case - malformed pattern",
			input:    "@Check[will|dc:15",
			expected: "@Check[will|dc:15",
		},
		{
			name:     "Mixed text with various patterns",
			input:    "A @UUID[Compendium.pf2e.conditionitems.Item.Frightened]{Frightened 1} creature takes a @Damage[1d4[mental]] penalty to @Check[will|dc:20] saves.",
			expected: "A Frightened 1 creature takes a 1d4 mental penalty to DC 20 will saves.",
		},
		{
			name:     "Complex text with gmr and other patterns",
			input:    "The dragon recharges its breath weapon in [[/gmr 1d4 #Recharge Fire Breath]]{1d4 rounds} and deals @Damage[8d6[fire]] damage in a @Template[cone|distance:60].",
			expected: "The dragon recharges its breath weapon in 1d4 rounds and deals 8d6 fire damage in a cone (60 feet).",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := replacer.ProcessText(tc.input)
			if result != tc.expected {
				t.Errorf("Test %s failed:\nInput:    %s\nExpected: %s\nGot:      %s", tc.name, tc.input, tc.expected, result)
			}
		})
	}
}

func TestReplacerEdgeCases(t *testing.T) {
	replacer := utils.NewReplacer()

	edgeCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No @ patterns",
			input:    "This is just regular text with no patterns.",
			expected: "This is just regular text with no patterns.",
		},
		{
			name:     "@ symbol without pattern",
			input:    "Send email to john@example.com",
			expected: "Send email to john@example.com",
		},
		{
			name:     "Multiple @ symbols",
			input:    "Email me @ john@example.com or @Check[will|dc:15]",
			expected: "Email me @ john@example.com or DC 15 will",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only @ symbol",
			input:    "@",
			expected: "@",
		},
		{
			name:     "Incomplete pattern",
			input:    "@Check",
			expected: "@Check",
		},
		{
			name:     "Pattern with no closing bracket",
			input:    "@Check[will|dc:15",
			expected: "@Check[will|dc:15",
		},
		{
			name:     "Nested patterns (should not occur but test anyway)",
			input:    "@Check[@Damage[1d6[fire]]]",
			expected: "@Check[@Damage[1d6[fire]]]",
		},
	}

	for _, tc := range edgeCases {
		t.Run(tc.name, func(t *testing.T) {
			result := replacer.ProcessText(tc.input)
			if result != tc.expected {
				t.Errorf("Edge case %s failed:\nInput:    %s\nExpected: %s\nGot:      %s", tc.name, tc.input, tc.expected, result)
			}
		})
	}
}

func BenchmarkReplacer(b *testing.B) {
	replacer := utils.NewReplacer()
	testText := "The creature must make a @Check[reflex|dc:19] save or take @Damage[2d6[fire]] damage and become @UUID[Compendium.pf2e.conditionitems.Item.Frightened]{Frightened 1}."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replacer.ProcessText(testText)
	}
}

func BenchmarkReplacerLongText(b *testing.B) {
	replacer := utils.NewReplacer()
	longText := "This is a very long text with multiple patterns: @Check[will|dc:20], @Damage[8d6[fire]], @UUID[Compendium.pf2e.spells-srd.Item.Fireball], @Template[burst|distance:20], and more patterns like @Check[reflex|dc:15|basic] and @Damage[(3d8+5)[bludgeoning]]. " +
		"It also includes @UUID[Compendium.pf2e.conditionitems.Item.Stunned]{Stunned 2} and @Template[line|distance:60|traits:force,evocation]. " +
		"This simulates a real-world scenario with multiple replacements in a single text block."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		replacer.ProcessText(longText)
	}
}