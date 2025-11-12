package utils

import (
	"testing"
)

func TestGetPublicURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows path with backslashes",
			input:    ".\\uploads\\daily-updates\\20251112-102309-841fd66c.png",
			expected: "/uploads/daily-updates/20251112-102309-841fd66c.png",
		},
		{
			name:     "Unix path with forward slashes",
			input:    "./uploads/daily-updates/20251112-102309-841fd66c.png",
			expected: "/uploads/daily-updates/20251112-102309-841fd66c.png",
		},
		{
			name:     "Path without leading ./",
			input:    "uploads/daily-updates/20251112-102309-841fd66c.png",
			expected: "/uploads/daily-updates/20251112-102309-841fd66c.png",
		},
		{
			name:     "Already a public URL",
			input:    "/uploads/daily-updates/20251112-102309-841fd66c.png",
			expected: "/uploads/daily-updates/20251112-102309-841fd66c.png",
		},
		{
			name:     "Mixed separators (Windows filepath.Join result)",
			input:    "./uploads\\daily-updates\\20251112-102309-841fd66c.png",
			expected: "/uploads/daily-updates/20251112-102309-841fd66c.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPublicURL(tt.input)
			if result != tt.expected {
				t.Errorf("GetPublicURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

