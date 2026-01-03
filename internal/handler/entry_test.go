package handler

import (
	"testing"
)

func TestParseTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "single tag",
			input:    "golang",
			expected: []string{"golang"},
		},
		{
			name:     "multiple tags",
			input:    "golang,rust,python",
			expected: []string{"golang", "rust", "python"},
		},
		{
			name:     "tags with whitespace",
			input:    " golang , rust , python ",
			expected: []string{"golang", "rust", "python"},
		},
		{
			name:     "tags with empty entries",
			input:    "golang,,rust,  ,python",
			expected: []string{"golang", "rust", "python"},
		},
		{
			name:     "only whitespace and commas",
			input:    " , , , ",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTags(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseTags(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("parseTags(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestParseTimeSpentMinutes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "valid positive minutes",
			input:    "5",
			expected: intPtr(300), // 5 minutes = 300 seconds
		},
		{
			name:     "zero",
			input:    "0",
			expected: nil,
		},
		{
			name:     "negative",
			input:    "-5",
			expected: nil,
		},
		{
			name:     "invalid string",
			input:    "abc",
			expected: nil,
		},
		{
			name:     "large value",
			input:    "120",
			expected: intPtr(7200), // 120 minutes = 7200 seconds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimeSpentMinutes(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("parseTimeSpentMinutes(%q) = %v, want nil", tt.input, *result)
				}
			} else {
				if result == nil {
					t.Errorf("parseTimeSpentMinutes(%q) = nil, want %v", tt.input, *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("parseTimeSpentMinutes(%q) = %v, want %v", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

func TestParseQuantity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "valid positive",
			input:    "5",
			expected: intPtr(5),
		},
		{
			name:     "zero",
			input:    "0",
			expected: nil,
		},
		{
			name:     "negative",
			input:    "-5",
			expected: nil,
		},
		{
			name:     "invalid string",
			input:    "abc",
			expected: nil,
		},
		{
			name:     "large value",
			input:    "1000",
			expected: intPtr(1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseQuantity(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("parseQuantity(%q) = %v, want nil", tt.input, *result)
				}
			} else {
				if result == nil {
					t.Errorf("parseQuantity(%q) = nil, want %v", tt.input, *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("parseQuantity(%q) = %v, want %v", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

func TestParseNotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "valid notes",
			input:    "These are my notes",
			expected: strPtr("These are my notes"),
		},
		{
			name:     "notes with leading/trailing whitespace",
			input:    "  Some notes here  ",
			expected: strPtr("Some notes here"),
		},
		{
			name:     "notes with internal whitespace",
			input:    "Notes with   multiple   spaces",
			expected: strPtr("Notes with   multiple   spaces"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNotes(tt.input)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("parseNotes(%q) = %q, want nil", tt.input, *result)
				}
			} else {
				if result == nil {
					t.Errorf("parseNotes(%q) = nil, want %q", tt.input, *tt.expected)
				} else if *result != *tt.expected {
					t.Errorf("parseNotes(%q) = %q, want %q", tt.input, *result, *tt.expected)
				}
			}
		})
	}
}

// Helper functions for creating pointers
func intPtr(v int) *int {
	return &v
}

func strPtr(v string) *string {
	return &v
}

