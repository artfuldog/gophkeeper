package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtrTo(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		testVal := "testSrt"
		if testVal != *PtrTo(testVal) {
			t.Error("PtrTo() = wrong value")
		}
	})
	t.Run("Check int64", func(t *testing.T) {
		testVal := 123141424123123132
		if testVal != *PtrTo(testVal) {
			t.Error("PtrTo() = wrong value")
		}
	})
}

func TestLast(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		testArray := []string{"one", "two", "three"}
		last := "three"
		assert.Equal(t, last, Last(testArray))
	})
	t.Run("Check float", func(t *testing.T) {
		testArray := []float64{123.12314, 1090454.1255, 149102, 12389.111}
		last := 12389.111
		assert.Equal(t, last, Last(testArray))
	})
	t.Run("Check empty array", func(t *testing.T) {
		testArray := []int{}
		var last int
		assert.Equal(t, last, Last(testArray))
	})
}

func TestContains(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		testArray := []string{"one", "two", "three"}
		right := "two"
		wrong := "four"
		assert.Equal(t, Contains(right, testArray), true)
		assert.Equal(t, Contains(wrong, testArray), false)
	})
	t.Run("Check string", func(t *testing.T) {
		testArray := []float64{123.12314, 1090454.1255, 149102, 12389.111}
		right := 1090454.1255
		wrong := 7810230123.123
		assert.Equal(t, Contains(right, testArray), true)
		assert.Equal(t, Contains(wrong, testArray), false)
	})
}

func TestIndexOf(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		testArray := []string{"one", "two", "three"}
		val := "one"
		assert.Equal(t, IndexOf(val, testArray), 0)
	})
	t.Run("Check float", func(t *testing.T) {
		testArray := []float64{123.12314, 1090454.1255, 149102.1, 12389.111}
		val := 149102.1
		assert.Equal(t, IndexOf(val, testArray), 2)
	})
	t.Run("Missed value", func(t *testing.T) {
		testArray := []float64{123.12314, 1090454.1255, 149102.1, 12389.111}
		val := 912301231.11255
		assert.Equal(t, IndexOf(val, testArray), -1)
	})
}

func TestDeleteElement(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		testArray := []string{"one", "two", "three"}
		index := 1
		testArray = DeleteElement(index, testArray)
		assert.Equal(t, []string{"one", "three"}, testArray)
	})
	t.Run("Check string last element", func(t *testing.T) {
		testArray := []string{"one", "two", "three"}
		index := 2
		testArray = DeleteElement(index, testArray)
		assert.Equal(t, []string{"one", "two"}, testArray)
	})
	t.Run("Check float", func(t *testing.T) {
		testArray := []float64{123.12314, 1090454.1255, 149102.1, 12389.111}
		index := 0
		testArray = DeleteElement(index, testArray)
		assert.Equal(t, []float64{1090454.1255, 149102.1, 12389.111}, testArray)
	})
}

func TestMaskAll(t *testing.T) {
	t.Run("Check", func(t *testing.T) {
		assert.Equal(t, MaskAll(8), "********")
	})
}

func TestMaskLeft(t *testing.T) {
	t.Run("Check masking", func(t *testing.T) {
		input := "abcdefgh"
		assert.Equal(t, MaskLeft(input, 3), "*****fgh")
	})
	t.Run("Check too length masking", func(t *testing.T) {
		input := "abcdefgh"
		assert.Equal(t, MaskLeft(input, 9), "********")
	})
}
