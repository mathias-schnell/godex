package main

import (
	"testing"
)

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "  hello  world  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "  HELLO  WORLD  ",
			expected: []string{"hello", "world"},
		},
		{
			input:    "He LlO wO rLD",
			expected: []string{"he", "llo", "wo", "rld"},
		},
	}

	for _, c := range cases {
		actual := cleanInput(c.input)
		if len(actual) != len(c.expected) {
			t.Errorf("Length of output does not match expected output length")
			t.Errorf("Expected Length > %d", len(c.expected))
			t.Errorf("Actual Length > %d", len(actual))
		}
		for i := range actual {
			word := actual[i]
			expectedWord := c.expected[i]
			if word != expectedWord {
				t.Errorf("Word mismatch")
				t.Errorf("Expected Word > %s", expectedWord)
				t.Errorf("Actual Word > %s", word)
			}
		}
	}
}
