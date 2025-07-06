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

// func ContainsCantrip(traits []string) bool {
// 	for _, trait := range traits {
// 		if trait == "cantrip" {
// 			return true
// 		}
// 	}
// 	return false
// }

func RemoveHTML(input string) string {
	//var output string
	//output = strings.ReplaceAll(input, "<p>", "")
	//output = strings.ReplaceAll(output, "</p>", "")
	output := strings.ReplaceAll(input, "<hr />", "")
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

// func DivideAndRoundUp(n int) int {
// 	result := n / 2
// 	if n%2 != 0 {
// 		result++
// 	}
// 	return result
// }

func ModifyDamage(damageStr string, modifier int) string {
	// Split the string at 'd'
	parts := strings.Split(damageStr, "d")
	if len(parts) != 2 {
		return damageStr // Return original if format is invalid
	}

	// Get the number of dice
	numDice := parts[0]

	// Check if there's an existing modifier
	secondParts := strings.Split(parts[1], "+")
	if len(secondParts) == 1 {
		// Also check for negative modifier
		secondParts = strings.Split(parts[1], "-")
		if len(secondParts) > 1 {
			diceType := secondParts[0]
			existingMod, _ := strconv.Atoi(secondParts[1])
			existingMod = -existingMod // Make it negative since it was a subtraction
			newMod := existingMod + modifier

			// Format the output based on whether newMod is positive or negative
			if newMod > 0 {
				return fmt.Sprintf("%sd%s+%d", numDice, diceType, newMod)
			} else if newMod < 0 {
				return fmt.Sprintf("%sd%s%d", numDice, diceType, newMod) // Negative number already includes the minus sign
			}
			return fmt.Sprintf("%sd%s", numDice, diceType)
		}
		// No modifier found
		diceType := parts[1]
		if modifier > 0 {
			return fmt.Sprintf("%sd%s+%d", numDice, diceType, modifier)
		} else if modifier < 0 {
			return fmt.Sprintf("%sd%s%d", numDice, diceType, modifier) // Negative number already includes the minus sign
		}
		return fmt.Sprintf("%sd%s", numDice, diceType)
	}

	// There is an existing positive modifier
	diceType := secondParts[0]
	existingMod, _ := strconv.Atoi(secondParts[1])
	newMod := existingMod + modifier

	// Format the output based on whether newMod is positive or negative
	if newMod > 0 {
		return fmt.Sprintf("%sd%s+%d", numDice, diceType, newMod)
	} else if newMod < 0 {
		return fmt.Sprintf("%sd%s%d", numDice, diceType, newMod) // Negative number already includes the minus sign
	}
	return fmt.Sprintf("%sd%s", numDice, diceType)
}

func PositiveOrNegative(input int) string {
	if input > 0 {
		return fmt.Sprintf("+%d", input)
	} else {
		return fmt.Sprintf("%d", input)
	}
}
