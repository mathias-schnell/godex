package main

import (
	"strings"
)

func cleanInput(test string) []string {
	return strings.Fields(strings.ToLower(test))
}
