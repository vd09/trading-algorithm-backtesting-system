package utils

func IsLineIntersect(line1Points, line2Points []float64) bool {
	lenLine1, lenLine2 := len(line1Points), len(line2Points)

	if lenLine1 <= 1 || lenLine2 <= 1 {
		return false
	}

	if line1Points[0] > line2Points[0] && line1Points[lenLine1-1] < line2Points[lenLine2-1] {
		return true
	}
	if line1Points[0] < line2Points[0] && line1Points[lenLine1-1] > line2Points[lenLine2-1] {
		return true
	}
	return false
}
