// Package common contains common for all gophkeeper components variables and functions.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItemTypeText(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		assert.Equal(t, ItemTypeText(ItemTypeLogin), "login")
		assert.Equal(t, ItemTypeText(ItemTypeCard), "card")
		assert.Equal(t, ItemTypeText(ItemTypeSecNote), "secured note")
		assert.Equal(t, ItemTypeText(ItemTypeSecData), "secured data")
		assert.Equal(t, ItemTypeText("Asdasdasdas"), "unknown")
	})
}

func TestItemTypeFromText(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		assert.Equal(t, ItemTypeFromText("login"), ItemTypeLogin)
		assert.Equal(t, ItemTypeFromText("card"), ItemTypeCard)
		assert.Equal(t, ItemTypeFromText("secured note"), ItemTypeSecNote)
		assert.Equal(t, ItemTypeFromText("secured data"), ItemTypeSecData)
		assert.Equal(t, ItemTypeFromText("asdasdadasdasd"), "")

	})
}
