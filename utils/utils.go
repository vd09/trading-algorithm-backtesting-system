package utils

// B2I converts a boolean to an integer
func B2I(b bool) int {
	if b {
		return 1
	}
	return 0
}

// B2F converts a boolean to a float
func B2F(b bool) float64 {
	return float64(B2I(b))
}
