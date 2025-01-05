package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type Replacer struct {
	patterns map[string]*regexp.Regexp
	// Add any configuration options here
}

func NewReplacer() *Replacer {
	return &Replacer{
		patterns: map[string]*regexp.Regexp{
			// Weird formatted entries (add these at the start)
			"formatted_rolls_br":     regexp.MustCompile(`\[\[/br [^]]+\]\]{([^}]+)}`),
			"formatted_rolls_r":      regexp.MustCompile(`\[\[/r [^]]+\]\]{([^}]+)}`),
			"formatted_rolls_simple": regexp.MustCompile(`\[\[/r ([^]]+)\]\]`),

			// Damage patterns
			"damage":            regexp.MustCompile(`@Damage\[(\d+)d(\d+)\[([^\]]+)\]\]`),
			"damage_with_level": regexp.MustCompile(`@Damage\[\(@item\.level\)\[([^\]]+)\]\]`),

			// Check patterns
			"check_with_full_traits": regexp.MustCompile(`@Check\[([^|]+)\|dc:(\d+)\|basic\|traits:([^|]+)(?:\|overrideTraits:true)?\]`),
			"check_with_traits":      regexp.MustCompile(`@Check\[([^|]+)\|dc:(\d+)\|traits:([^]]+)\]`),
			"check_with_showdc":      regexp.MustCompile(`@Check\[([^|]+)\|showDC:all\|dc:(\d+)\]`),
			"check_with_basic":       regexp.MustCompile(`@Check\[([^|]+)\|dc:(\d+)\|basic\]`),
			"check_simple":           regexp.MustCompile(`@Check\[([^|]+)\|dc:(\d+)\]`),

			// Template patterns
			"template_with_text":   regexp.MustCompile(`@Template\[([^|]+)\|distance:(\d+)\]{[^}]+}`),
			"template_with_traits": regexp.MustCompile(`@Template\[([^|]+)\|distance:(\d+)\|traits:([^]]+)\]`),
			"template_simple":      regexp.MustCompile(`@Template\[([^|]+)\|distance:(\d+)\]`),

			// UUID patterns
			"uuid_with_text": regexp.MustCompile(`@UUID\[Compendium\.pf2e\.[^.]+\.Item\.([^]]+)\]{([^}]+)}`),
			"uuid_simple":    regexp.MustCompile(`@UUID\[Compendium\.pf2e\.[^.]+\.Item\.([^]]+)\]`),
		},
	}
}

func (r *Replacer) ProcessText(input string) string {
	// Replace weird formatted rolls (process these first)
	input = r.patterns["formatted_rolls_br"].ReplaceAllString(input, "$1")
	input = r.patterns["formatted_rolls_r"].ReplaceAllString(input, "$1")
	input = r.patterns["formatted_rolls_simple"].ReplaceAllString(input, "$1")

	// Replace damage
	input = r.patterns["damage_with_level"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["damage_with_level"].FindStringSubmatch(match)
		type_ := parts[1] // bleed
		return type_
	})

	input = r.patterns["damage"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["damage"].FindStringSubmatch(match)
		number := parts[1] // 18
		dice := parts[2]   // 6
		type_ := parts[3]  // acid
		return fmt.Sprintf("%sd%s %s", number, dice, type_)
	})

	// Replace templates with custom text (ignoring the custom text)
	input = r.patterns["template_with_text"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["template_with_text"].FindStringSubmatch(match)
		template := parts[1] // line
		distance := parts[2] // 100
		return fmt.Sprintf("%s (%s feet)", template, distance)
	})

	// Replace templates with traits
	input = r.patterns["template_with_traits"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["template_with_traits"].FindStringSubmatch(match)
		template := parts[1] // line
		distance := parts[2] // 100
		traits := parts[3]   // arcane,transmutation

		// Format traits by replacing commas with ", "
		formattedTraits := strings.ReplaceAll(traits, ",", ", ")

		return fmt.Sprintf("%s (%s feet, %s)", template, distance, formattedTraits)
	})

	// Replace simple templates
	input = r.patterns["template_simple"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["template_simple"].FindStringSubmatch(match)
		template := parts[1] // emanation
		distance := parts[2] // 100
		return fmt.Sprintf("%s (%s feet)", template, distance)
	})

	// Replace checks with full traits
	input = r.patterns["check_with_full_traits"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["check_with_full_traits"].FindStringSubmatch(match)
		checkType := parts[1]
		dc := parts[2]
		// traits are in parts[3] but we don't need them for the output
		return fmt.Sprintf("DC %s basic %s", dc, checkType)
	})

	// Replace checks with traits
	input = r.patterns["check_with_traits"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["check_with_traits"].FindStringSubmatch(match)
		checkType := parts[1] // thievery
		dc := parts[2]        // 25
		traits := parts[3]    // action:disable-a-device

		// Format traits (only keep what's after "action:")
		traitParts := strings.Split(traits, ":")
		var formattedTrait string
		if len(traitParts) >= 3 {
			formattedTrait = fmt.Sprintf(" (%s)", strings.ReplaceAll(traitParts[2], "-", " "))
		}

		return fmt.Sprintf("DC %s %s%s", dc, checkType, formattedTrait)
	})

	// Replace checks with basic
	input = r.patterns["check_with_basic"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["check_with_basic"].FindStringSubmatch(match)
		checkType := parts[1] // reflex
		dc := parts[2]        // 38
		return fmt.Sprintf("DC %s %s", dc, checkType)
	})

	// Replace checks with showDC:all
	input = r.patterns["check_with_showdc"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["check_with_showdc"].FindStringSubmatch(match)
		checkType := parts[1] // flat
		dc := parts[2]        // 11
		return fmt.Sprintf("DC %s %s check", dc, checkType)
	})

	// Replace simple checks
	input = r.patterns["check_simple"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["check_simple"].FindStringSubmatch(match)
		checkType := parts[1] // will
		dc := parts[2]        // 18
		return fmt.Sprintf("DC %s %s", dc, checkType)
	})

	// Replace UUIDs with custom text
	input = r.patterns["uuid_with_text"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["uuid_with_text"].FindStringSubmatch(match)
		customText := parts[2]

		// If the custom text contains a number (likely a condition level),
		// use the custom text instead of the item name
		if hasNumber(customText) {
			return customText
		}

		// For other cases like Magic Wand, use the original item name
		return strings.ReplaceAll(parts[1], "-", " ")
	})

	// Replace simple UUIDs
	input = r.patterns["uuid_simple"].ReplaceAllStringFunc(input, func(match string) string {
		parts := r.patterns["uuid_simple"].FindStringSubmatch(match)
		return strings.ReplaceAll(parts[1], "-", " ")
	})

	return input
}

func hasNumber(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}
