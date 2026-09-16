package views

import "fmt"

// Millimes formats a price for display: 12500 becomes "12.500 TND".
//
// Prices are stored as a whole number of millimes — a thousandth of a dinar — so
// that adding them up is exact. Formatting is the only place the decimal point
// exists.
func Millimes(millimes int32) string {
	return fmt.Sprintf("%d.%03d TND", millimes/1000, millimes%1000)
}
