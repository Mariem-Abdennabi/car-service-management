package server

import (
	"errors"
	"testing"
)

func TestParseMillimes(t *testing.T) {
	valid := map[string]int32{
		"0":      0,
		"42":     42000,
		"42.5":   42500,
		"42.50":  42500,
		"42.500": 42500,
		"0.001":  1,
		"  8.29": 8290, // The value a float would turn into 8289.
	}

	for text, want := range valid {
		got, err := parseMillimes(text)
		if err != nil {
			t.Errorf("parseMillimes(%q) failed: %v", text, err)

			continue
		}
		if got != want {
			t.Errorf("parseMillimes(%q) = %d, want %d", text, got, want)
		}
	}

	invalid := []string{"", "abc", "-1", "4.2.5", "42.", "42.5000", "1e3"}

	for _, text := range invalid {
		if _, err := parseMillimes(text); !errors.Is(err, errBadPrice) {
			t.Errorf("parseMillimes(%q) error = %v, want errBadPrice", text, err)
		}
	}
}
