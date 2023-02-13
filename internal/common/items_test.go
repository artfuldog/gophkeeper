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

func TestListItemTypes(t *testing.T) {
	itemTypes := ListItemTypes()
	assert.Equal(t, itemTypes, []string{"l", "c", "n", "d"})
}

func TestCFTypeToText(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		assert.Equal(t, CfTypeToText(CfTypeText), "text")
		assert.Equal(t, CfTypeToText(CfTypeHidden), "hidden")
		assert.Equal(t, CfTypeToText(CfTypeBool), "bool")
		assert.Equal(t, CfTypeToText("Asdasdasdas"), "unknown")
	})
}

func TestCFTypeFromText(t *testing.T) {
	t.Run("Check string", func(t *testing.T) {
		assert.Equal(t, CfTypeFromText("text"), CfTypeText)
		assert.Equal(t, CfTypeFromText("hidden"), CfTypeHidden)
		assert.Equal(t, CfTypeFromText("bool"), CfTypeBool)
		assert.Equal(t, CfTypeFromText("asdasdadasdasd"), "")
	})
}

func TestListCFTypes(t *testing.T) {
	itemTypes := ListCFTypes()
	assert.Equal(t, []string{"t", "h", "b"}, itemTypes)
}
