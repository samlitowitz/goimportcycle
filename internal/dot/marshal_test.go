package dot

import "testing"

func TestSanitizeForDot(t *testing.T) {
	table := []struct {
		input    string
		expected string
	}{{
		input:    "a.a.a",
		expected: "a_a_a",
	}, {
		input:    "a-a-a",
		expected: "a_a_a",
	}}

	for _, test := range table {
		result := sanitizeForDot(test.input)
		if result != test.expected {
			t.Fatalf(
				"Sanitized to unexpected value.\nInput: \"%s\"\nGot: \"%s\"\nExpected: \"%s\"",
				test.input,
				result,
				test.expected,
			)
		}
	}
}
