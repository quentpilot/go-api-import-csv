package utils

import "math"

func MathRound(v float64, precision int) float64 {
	mult := math.Pow(10, float64(precision))
	return math.Round(v*mult) / mult
}
