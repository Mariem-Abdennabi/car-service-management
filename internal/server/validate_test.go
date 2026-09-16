package server

import "testing"

func TestValidatePhone(t *testing.T) {
	// Written differently, all real numbers.
	valid := []string{
		"+216 20 145 872",
		"20145872",
		"(216) 20-145-872",
		"+216-98-476-130",
	}

	for _, phone := range valid {
		if msg := validatePhone(phone); msg != "" {
			t.Errorf("validatePhone(%q) = %q, want it accepted", phone, msg)
		}
	}

	invalid := map[string]string{
		"":                   "required",
		"12345":              "at least 8 digits",
		"call me":            "can only contain",
		"+216 20 145 872ext": "can only contain",
	}

	for phone, want := range invalid {
		msg := validatePhone(phone)
		if msg == "" {
			t.Errorf("validatePhone(%q) accepted it, want a message about %q", phone, want)
		}
	}
}

func TestTooLong(t *testing.T) {
	if msg := tooLong("Name", "Amira", 5); msg != "" {
		t.Errorf("tooLong() = %q, want the exact limit accepted", msg)
	}
	if msg := tooLong("Name", "Amiraa", 5); msg == "" {
		t.Error("tooLong() accepted a value over the limit")
	}

	// Counted in runes, not bytes: this is 5 characters but 10 bytes.
	if msg := tooLong("Name", "أميرة", 5); msg != "" {
		t.Errorf("tooLong() = %q, want non-ASCII counted by character", msg)
	}
}
