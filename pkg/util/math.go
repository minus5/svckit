package util

import "math"

// return rounded version of x with prec precision.
func Round(x float64, prec int) float64 {
	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)
	intermed += .5
	x = .5
	if frac < 0.0 {
		x = -.5
		intermed -= 1
	}
	if frac >= x {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow
}

// RoundAwayFromZero zaokruzuje vrijednost dalje od nule npr:
// 23.5 je zaokruzeno 24, a −23.5 je zaokruzeno na −24.
// Dodana je i korekcija od 0.00000001, jer na taj nacin
// zaokruzeni brojevi izgedaju matematicki ispravno
func RoundAwayFromZero(val float64, prec int) float64 {
	pow := math.Pow10(prec)
	if val < 0 {
		return float64(int64(val*pow-0.50000001)) / pow
	}
	return float64(int64(val*pow+0.50000001)) / pow
}
