package common

import (
	"fmt"
	"strings"
)

// PtrTo returns pointer to value.
func PtrTo[V any](val V) *V {
	return &val
}

// Last returns last value of array.
// If array is empty return default nul value.
func Last[V any](s []V) V {
	if len(s) == 0 {
		var none V
		return none
	}

	return s[len(s)-1]
}

// Contains checks if value is in array.
func Contains[V comparable](val V, s []V) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}

	return false
}

// IndexOf returns index of value position in slice.
// If value not found returns -1.
func IndexOf[V comparable](val V, s []V) int {
	for i, v := range s {
		if v == val {
			return i
		}
	}

	return -1
}

// MaskAll returns masked with asterisk string with provided length.
func MaskAll(masked int) string {
	return strings.Repeat("*", masked)
}

// MaskLeft masks with asterisk all string except last N symbols.
func MaskLeft(input string, n int) string {
	if len(input) <= n {
		return MaskAll(len(input))
	}

	head := input[:len(input)-n]
	tail := input[len(input)-n:]

	mask := strings.Repeat("*", len(head))

	return fmt.Sprintf("%s%s", mask, tail)
}
