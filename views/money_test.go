package views

import "testing"

func TestMillimes(t *testing.T) {
	tests := map[int32]string{
		0:     "0.000 TND",
		500:   "0.500 TND",
		1000:  "1.000 TND",
		12500: "12.500 TND",
		// The one that would break with a float: a price ending in a stray
		// thousandth still formats exactly.
		99999:  "99.999 TND",
		100001: "100.001 TND",
	}

	for millimes, want := range tests {
		if got := Millimes(millimes); got != want {
			t.Errorf("Millimes(%d) = %q, want %q", millimes, got, want)
		}
	}
}
