package utils

import (
	"fmt"
	"regexp"
	"strings"
)

type Replacer struct {
	// General pattern to match @Type[content]{optional_text}
	generalPattern *regexp.Regexp
	// Legacy patterns for edge cases
	legacyPatterns map[string]*regexp.Regexp
}

func NewReplacer() *Replacer {
	return &Replacer{
		// Simple pattern to match @Word followed by any content until end of balanced brackets
		generalPattern: regexp.MustCompile(`@(\w+)(\[.*?\](?:\{[^}]*\})?)`),
		legacyPatterns: map[string]*regexp.Regexp{
			// Handle weird formatted rolls that don't follow standard pattern
			"formatted_rolls_br":     regexp.MustCompile(`\[\[/br [^]]+\]\]{([^}]+)}`),
			"formatted_rolls_r":      regexp.MustCompile(`\[\[/r [^]]+\]\]{([^}]+)}`),
			"formatted_rolls_gmr":    regexp.MustCompile(`\[\[/gmr [^]]+\]\]{([^}]+)}`),
			"formatted_rolls_simple": regexp.MustCompile(`\[\[/r ([^]]+)\]\]`),
		},
	}
}

func (r *Replacer) ProcessText(input string) string {
	// Handle legacy patterns first (weird formatted rolls)
	input = r.legacyPatterns["formatted_rolls_br"].ReplaceAllString(input, "$1")
	input = r.legacyPatterns["formatted_rolls_r"].ReplaceAllString(input, "$1")
	input = r.legacyPatterns["formatted_rolls_gmr"].ReplaceAllString(input, "$1")
	input = r.legacyPatterns["formatted_rolls_simple"].ReplaceAllString(input, "$1")

	// Handle all @ patterns with manual parsing
	return r.parseAndReplacePatterns(input)
}

func (r *Replacer) parseAndReplacePatterns(input string) string {
	result := ""
	i := 0
	
	for i < len(input) {
		// Find next @ symbol
		atPos := strings.Index(input[i:], "@")
		if atPos == -1 {
			// No more @ symbols, append rest of string
			result += input[i:]
			break
		}
		
		// Append text before @
		result += input[i : i+atPos]
		i += atPos
		
		// Try to parse @ pattern
		pattern, patternLen := r.parsePattern(input[i:])
		if patternLen > 0 {
			// Successfully parsed a pattern, replace it (even if replacement is empty)
			result += pattern
			i += patternLen
		} else {
			// Not a valid pattern, just append the @ symbol
			result += "@"
			i++
		}
	}
	
	return result
}

func (r *Replacer) parsePattern(input string) (replacement string, length int) {
	if len(input) < 2 || input[0] != '@' {
		return "", 0
	}
	
	// Find pattern type (word after @)
	wordStart := 1
	wordEnd := wordStart
	for wordEnd < len(input) && ((input[wordEnd] >= 'a' && input[wordEnd] <= 'z') || 
		(input[wordEnd] >= 'A' && input[wordEnd] <= 'Z') || 
		(input[wordEnd] >= '0' && input[wordEnd] <= '9') || 
		input[wordEnd] == '_') {
		wordEnd++
	}
	
	if wordEnd == wordStart || wordEnd >= len(input) || input[wordEnd] != '[' {
		return "", 0 // No valid pattern type or no opening bracket
	}
	
	patternType := input[wordStart:wordEnd]
	
	// Find matching closing bracket
	bracketStart := wordEnd
	bracketLevel := 0
	contentEnd := -1
	
	for i := bracketStart; i < len(input); i++ {
		if input[i] == '[' {
			bracketLevel++
		} else if input[i] == ']' {
			bracketLevel--
			if bracketLevel == 0 {
				contentEnd = i
				break
			}
		}
	}
	
	if contentEnd == -1 {
		return "", 0 // No matching bracket
	}
	
	content := input[bracketStart+1 : contentEnd]
	
	// Check for custom text in {}
	customText := ""
	totalLength := contentEnd + 1
	
	if contentEnd+1 < len(input) && input[contentEnd+1] == '{' {
		braceEnd := strings.Index(input[contentEnd+1:], "}")
		if braceEnd != -1 {
			customText = input[contentEnd+2 : contentEnd+1+braceEnd]
			totalLength = contentEnd + 1 + braceEnd + 1
		}
	}
	
	originalMatch := input[:totalLength]
	replacement = r.handlePattern(patternType, content, customText, originalMatch)
	

	
	return replacement, totalLength
}

func (r *Replacer) handlePattern(patternType, content, customText, originalMatch string) string {
	switch strings.ToLower(patternType) {
	case "check":
		return r.handleCheck(content, customText)
	case "damage":
		return r.handleDamage(content, customText)
	case "uuid":
		return r.handleUUID(content, customText)
	case "template":
		return r.handleTemplate(content, customText)
	case "localize":
		return r.handleLocalize(content, customText)
	default:
		// For unknown patterns, try to extract something meaningful
		return r.handleUnknown(patternType, content, customText, originalMatch)
	}
}

func (r *Replacer) handleCheck(content, customText string) string {
	// Parse content like "arcana|dc:40" or "reflex|dc:19|basic"
	// Handle empty content
	if content == "" {
		return ""
	}
	
	// If content contains @ symbols, it's likely a nested pattern - return original
	if strings.Contains(content, "@") {
		return fmt.Sprintf("@Check[%s]", content)
	}
	
	parts := strings.Split(content, "|")
	if len(parts) < 1 {
		return content // Fallback
	}
	
	checkType := parts[0]
	dc := ""
	isBasic := false
	showDC := false
	
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "dc:") {
			dc = strings.TrimPrefix(part, "dc:")
		} else if part == "basic" {
			isBasic = true
		} else if strings.Contains(part, "showDC") {
			showDC = true
		}
	}
	
	if dc == "" {
		return checkType // Fallback if no DC found
	}
	
	if isBasic {
		return fmt.Sprintf("DC %s basic %s", dc, checkType)
	}
	if showDC {
		return fmt.Sprintf("DC %s %s check", dc, checkType)
	}
	return fmt.Sprintf("DC %s %s", dc, checkType)
}

func (r *Replacer) handleDamage(content, customText string) string {
	// Handle patterns like "11d6[acid]" or "(2d8 + 5)[bludgeoning]" or "(@item.level)[bleed]"
	
	// If content contains @ symbols (except @item), it's likely a nested pattern - return original
	if strings.Contains(content, "@") && !strings.Contains(content, "@item.level") {
		return fmt.Sprintf("@Damage[%s]", content)
	}
	
	// Special case for (@item.level) pattern
	if strings.Contains(content, "@item.level") {
		damageTypeRegex := regexp.MustCompile(`\[([^\]]+)\]$`)
		matches := damageTypeRegex.FindStringSubmatch(content)
		if len(matches) > 1 {
			return matches[1] // Return just the damage type
		}
	}
	
	// Find the last bracket pair which should contain the damage type
	lastBracketStart := strings.LastIndex(content, "[")
	lastBracketEnd := strings.LastIndex(content, "]")
	
	if lastBracketStart != -1 && lastBracketEnd != -1 && lastBracketEnd > lastBracketStart {
		damageType := content[lastBracketStart+1 : lastBracketEnd]
		damageFormula := strings.TrimSpace(content[:lastBracketStart])
		
		// Clean up the formula (remove extra parentheses if they wrap the whole thing)
		if strings.HasPrefix(damageFormula, "(") && strings.HasSuffix(damageFormula, ")") {
			damageFormula = strings.Trim(damageFormula, "()")
		}
		
		return fmt.Sprintf("%s %s", damageFormula, damageType)
	}
	
	// Fallback for patterns without clear damage type
	return content
}

func (r *Replacer) handleUUID(content, customText string) string {
	// If content contains @ symbols, it's likely a nested pattern - return original
	if strings.Contains(content, "@") {
		return fmt.Sprintf("@UUID[%s]", content)
	}
	
	// If there's custom text, use that
	if customText != "" {
		// If custom text contains a number, it's likely a condition level
		if r.hasNumber(customText) {
			return customText
		}
		return customText
	}
	
	// Parse UUID content like "Compendium.pf2e.spells-srd.Item.Bind Undead"
	// or "Compendium.pf2e.conditionitems.Item.kWc1fhmv9LBiTuei"
	parts := strings.Split(content, ".")
	if len(parts) < 4 {
		return content // Fallback
	}
	
	// For Actor UUIDs, extract last word
	if strings.Contains(content, "Actor") {
		actorName := parts[len(parts)-1]
		words := strings.Fields(actorName)
		if len(words) > 0 {
			return words[len(words)-1]
		}
		return actorName
	}
	
	// For Item UUIDs, get the item name (last part)
	itemName := parts[len(parts)-1]
	
	// Replace hyphens and underscores with spaces
	itemName = strings.ReplaceAll(itemName, "-", " ")
	itemName = strings.ReplaceAll(itemName, "_", " ")
	
	// If it looks like an ID (contains random characters), return a generic name
	if len(itemName) > 10 && r.isLikelyID(itemName) {
		// Try to get a meaningful name from the compendium type
		if len(parts) >= 3 {
			compendiumType := parts[2]
			switch compendiumType {
			case "conditionitems":
				return "condition"
			case "spells-srd", "spells":
				return "spell"
			case "equipment":
				return "item"
			case "actionspf2e":
				return "action"
			default:
				return compendiumType
			}
		}
	}
	
	return itemName
}

func (r *Replacer) handleTemplate(content, customText string) string {
	// If content contains @ symbols, it's likely a nested pattern - return original
	if strings.Contains(content, "@") {
		return fmt.Sprintf("@Template[%s]", content)
	}
	
	// Parse content like "line|distance:100" or "emanation|distance:30|traits:arcane,transmutation"
	parts := strings.Split(content, "|")
	if len(parts) < 2 {
		return content // Fallback
	}
	
	template := parts[0]
	distance := ""
	traits := ""
	
	for _, part := range parts[1:] {
		if strings.HasPrefix(part, "distance:") {
			distance = strings.TrimPrefix(part, "distance:")
		} else if strings.HasPrefix(part, "traits:") {
			traits = strings.TrimPrefix(part, "traits:")
		}
	}
	
	if distance == "" {
		return template // Fallback
	}
	
	if traits != "" {
		// Format traits by replacing commas with ", "
		formattedTraits := strings.ReplaceAll(traits, ",", ", ")
		return fmt.Sprintf("%s (%s feet, %s)", template, distance, formattedTraits)
	}
	
	return fmt.Sprintf("%s (%s feet)", template, distance)
}

func (r *Replacer) handleLocalize(content, customText string) string {
	// Handle @Localize patterns - usually just return the content or custom text
	if customText != "" {
		return customText
	}
	
	// Try to extract meaningful text from localization keys
	parts := strings.Split(content, ".")
	if len(parts) > 0 {
		// Return the last part, cleaned up
		lastPart := parts[len(parts)-1]
		lastPart = strings.ReplaceAll(lastPart, "-", " ")
		lastPart = strings.ReplaceAll(lastPart, "_", " ")
		return lastPart
	}
	
	return content
}

func (r *Replacer) handleUnknown(patternType, content, customText, originalMatch string) string {
	// For unknown patterns, try to extract something meaningful
	
	// If there's custom text, use it
	if customText != "" {
		return customText
	}
	
	// Try to extract a meaningful value from the content
	// Look for simple patterns like "value:something"
	if strings.Contains(content, ":") {
		parts := strings.Split(content, "|")
		for _, part := range parts {
			if strings.Contains(part, ":") && !strings.HasPrefix(part, "dc:") {
				keyValue := strings.Split(part, ":")
				if len(keyValue) >= 2 {
					return keyValue[1]
				}
			}
		}
	}
	
	// If content looks simple enough, return it
	if !strings.Contains(content, "|") && len(content) < 50 {
		return content
	}
	
	// Last resort: return the pattern type
	return strings.ToLower(patternType)
}

func (r *Replacer) hasNumber(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func (r *Replacer) isLikelyID(s string) bool {
	// Check if string looks like a random ID (contains mix of letters and numbers)
	hasLetter := false
	hasNumber := false
	
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			hasLetter = true
		} else if r >= '0' && r <= '9' {
			hasNumber = true
		}
	}
	
	return hasLetter && hasNumber && len(s) > 8
}