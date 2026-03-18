package main

import "testing"

func TestCompress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "a1",
		},
		{
			name:     "multiple characters",
			input:    "aaabbbcccd",
			expected: "a3b3c3d1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Compress(test.input)
			if result != test.expected {
				t.Errorf("Compress(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}
