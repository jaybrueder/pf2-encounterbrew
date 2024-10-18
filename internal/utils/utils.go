package utils

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func RemoveTrailingComma(s string) string {
    if len(s) > 0 && s[len(s)-2] == ',' {
        return s[:len(s)-2]
    }
    return s
}

func CapitalizeFirst(s string) string {
    return strings.ToUpperSpecial(unicode.TurkishCase, s[:1]) + s[1:]
}

func FormatOrdinal(level string) string {
    n, err := strconv.Atoi(level)
    if err != nil {
        return level // Return the original string if it's not a valid number
    }

    if n <= 0 {
        return level // Return the original string for zero or negative numbers
    }

    switch {
    case n%100 >= 11 && n%100 <= 13:
        return fmt.Sprintf("%dth", n)
    case n%10 == 1:
        return fmt.Sprintf("%dst", n)
    case n%10 == 2:
        return fmt.Sprintf("%dnd", n)
    case n%10 == 3:
        return fmt.Sprintf("%drd", n)
    default:
        return fmt.Sprintf("%dth", n)
    }
}

func ContainsCantrip(traits []string) bool {
    for _, trait := range traits {
        if trait == "cantrip" {
            return true
        }
    }
    return false
}

func RemoveHTML(input string) string {
	var output string
	output = strings.ReplaceAll(input, "<p>", "")
	output = strings.ReplaceAll(output, "</p>", "")
	output = strings.ReplaceAll(output, "<hr />", "")
	return output
}

func Contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}


func FormatSortedSpells(sortedSpells []map[string]string, combatantLevel int) map[string]string {
    levelGroups := make(map[string][]string)

    for _, spell := range sortedSpells {
        level := spell["level"]

        var spellString string
        uses, _ := strconv.Atoi(spell["uses"])

        if uses > 0 {
        	spellString = spell["name"] + " (x" + spell["uses"] + ")"
        } else {
        	spellString = spell["name"]
        }

        levelGroups[level] = append(levelGroups[level], spellString)
    }

    levels := []string{"10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "0"}
    result := make(map[string]string)

    for _, level := range levels {
        spells, exists := levelGroups[level]
        if !exists {
            continue
        }

        key := FormatOrdinal(level)
        if level == "0" {
            key = "Cantrips" + " (" + FormatOrdinal(strconv.Itoa(combatantLevel)) + ")"
        }

        result[key] = strings.Join(spells, ", ")
    }

    return result
}

func DivideAndRoundUp(n int) int {
    result := n / 2
    if n%2 != 0 {
        result++
    }
    return result
}
