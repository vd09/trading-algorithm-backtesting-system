package utils

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64
}

// Min returns the minimum of two values of an ordered type.
func Min[T Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Min returns the minimum of two values of an ordered type.
func Max[T Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}
