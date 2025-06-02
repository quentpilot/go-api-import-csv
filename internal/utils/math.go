package utils

import "math"

// MathRound returns a limited amount of values after comma
func MathRound(v float64, precision int) float64 {
	mult := math.Pow(10, float64(precision))
	return math.Round(v*mult) / mult
}
