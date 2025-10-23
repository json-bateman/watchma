package utils

import (
	"strings"
)

func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	// Convert first rune (character) to upper case
	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]
	return string(runes)
}
