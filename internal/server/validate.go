package server

import (
	"fmt"
	"strings"
	"unicode"
)

// Length limits. They exist to stop a paste accident becoming a database row, not
// to make claims about names — so they are generous.
const (
	maxNameLength      = 120
	maxCityLength      = 80
	maxPhoneLength     = 24
	maxPlateLength     = 20
	maxReferenceLength = 40
	maxQuantity        = 1_000_000
)

// minPhoneDigits is the length of a Tunisian number without its country code.
const minPhoneDigits = 8

// required reports the message for an empty field, or "" when it has a value.
func required(label, value string) string {
	if value == "" {
		return label + " is required."
	}

	return ""
}

// tooLong reports the message for a value over the limit, or "".
//
// Counts runes, not bytes: "Amira" and "أميرة" should be measured the same way, and
// len() on a string counts bytes.
func tooLong(label, value string, limit int) string {
	if len([]rune(value)) > limit {
		return fmt.Sprintf("%s must be %d characters or fewer.", label, limit)
	}

	return ""
}

// validatePhone checks a phone number loosely on purpose.
//
// Phone numbers are written in more ways than a pattern can capture — +216 20 145
// 872, 20145872, (216) 20-145-872 are the same number. So rather than enforce a
// format, this requires enough digits to be a real number and rejects characters
// that never appear in one. Being strict here mostly rejects valid input.
func validatePhone(value string) string {
	if msg := required("Phone", value); msg != "" {
		return msg
	}
	if msg := tooLong("Phone", value, maxPhoneLength); msg != "" {
		return msg
	}

	digits := 0

	for _, r := range value {
		switch {
		case unicode.IsDigit(r):
			digits++
		case strings.ContainsRune("+()- ", r):
			// Punctuation people write in phone numbers.
		default:
			return "Phone can only contain digits and + ( ) - and spaces."
		}
	}

	if digits < minPhoneDigits {
		return fmt.Sprintf("Phone must have at least %d digits.", minPhoneDigits)
	}

	return ""
}
