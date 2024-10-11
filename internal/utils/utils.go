package utils

import (
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
