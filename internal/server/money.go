package server

import (
	"errors"
	"strconv"
	"strings"
)

// errBadPrice means the text is not a price we can store exactly.
var errBadPrice = errors.New("not a valid price")

// millimesPerDinar is the number of millimes in one dinar, and so also the number
// of decimal places a price can have.
const millimesPerDinar = 1000

// parseMillimes turns "42.500" into 42500.
//
// Deliberately not strconv.ParseFloat followed by *1000: that reintroduces exactly
// the binary rounding error the integer column exists to avoid — 8.29 parses to
// 8.289999..., which truncates to 8289 rather than 8290. Splitting on the decimal
// point and parsing two integers is exact.
func parseMillimes(text string) (int32, error) {
	whole, fraction, hasFraction := strings.Cut(strings.TrimSpace(text), ".")

	dinars, err := strconv.Atoi(whole)
	if err != nil || dinars < 0 {
		return 0, errBadPrice
	}

	millimes := 0
	if hasFraction {
		// More than three digits cannot be stored, so it is rejected rather than
		// silently rounded — a price the visitor did not type is worse than an error.
		if len(fraction) == 0 || len(fraction) > 3 {
			return 0, errBadPrice
		}

		// "5" means 500 millimes, not 5.
		padded := fraction + strings.Repeat("0", 3-len(fraction))

		millimes, err = strconv.Atoi(padded)
		if err != nil || millimes < 0 {
			return 0, errBadPrice
		}
	}

	return int32(dinars*millimesPerDinar + millimes), nil
}
