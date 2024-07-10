package test_utils

import (
	"fmt"
	"reflect"
	"testing"
)

// AssertEqual checks if two values are equal and reports an error if not.
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s: expected %v (%T), got %v (%T)", message, expected, expected, actual, actual)
	}
}

func AssertTrue(t *testing.T, b bool, message string) {
	if !b {
		t.Errorf("%s: got %v", message, b)
	}
}

type IntersectionDirection string

const (
	Above IntersectionDirection = "above"
	Below IntersectionDirection = "below"
)

func GiveCrossingLine(line []float64, length int, direction IntersectionDirection) []float64 {
	if length > len(line) {
		length = len(line)
	}

	result := make([]float64, length)
	copy(result, line[:length])

	// Calculate the range of values in the line
	minValue := line[0]
	maxValue := line[0]
	for _, value := range line[:length] {
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}

	// Calculate the adjustment amount based on the range
	adjustment := (maxValue - minValue) * 0.1 // 10% of the range
	switch direction {
	case Above:
		for i := 0; i < length; i++ {
			if i < length/2 {
				result[i] = line[i] + adjustment
			} else {
				result[i] = line[i] - adjustment
			}
		}
	case Below:
		for i := 0; i < length; i++ {
			if i < length/2 {
				result[i] = line[i] - adjustment
			} else {
				result[i] = line[i] + adjustment
			}
		}
	default:
		fmt.Println("Invalid direction provided")
	}

	return result
}
