package stages

import (
	"testing"
)

// Test cases for the Luhn algorithm validation
func TestIsLuhnValid(t *testing.T) {
	cases := []struct {
		input int
		want  bool
	}{
		{4539_1488_0343_6467, true}, // Valid Luhn number
		{1234_5678_1234_5670, true}, // Another valid Luhn number
		{499_2739_8112_1717, false}, // Invalid Luhn number
		{1234567812345678, false},   // Another invalid Luhn number
		{3782_822463_10005, true},   // Short, valid Luhn number
		{123, false},                // Short, invalid Luhn number
	}

	for _, c := range cases {
		got := isLuhn(c.input)
		if got != c.want {
			t.Errorf("isLuhnValid(%q) == %t, want %t", c.input, got, c.want)
		}
	}
}

// TestReplaceLuhnValidNumbers tests the replaceLuhnValidNumbers function.
func TestReplaceLuhnValidNumbers(t *testing.T) {
	cases := []struct {
		input       string
		replacement string
		want        string
	}{
		// Test case with a single Luhn-valid number
		{"My credit card number is 3530111333300000.", "**REDACTED**", "My credit card number is **REDACTED**."},
		// Test case with multiple Luhn-valid numbers
		{"Cards 4532015112830366 and 6011111111111117 are valid.", "**REDACTED**", "Cards **REDACTED** and **REDACTED** are valid."},
		// Test case with no Luhn-valid numbers
		{"No valid numbers here.", "**REDACTED**", "No valid numbers here."},
		// Test case with mixed content
		{"Valid: 4556737586899855, invalid: 1234.", "**REDACTED**", "Valid: **REDACTED**, invalid: 1234."},
		// Test case with edge cases
		{"Edge cases: 0, 00, 000, 1.", "**REDACTED**", "Edge cases: 0, 00, 000, 1."},
	}

	for _, c := range cases {
		got := replaceLuhnValidNumbers(c.input, c.replacement, 13)
		if got != c.want {
			t.Errorf("replaceLuhnValidNumbers(%q, %q) == %q, want %q", c.input, c.replacement, got, c.want)
		}
	}
}
