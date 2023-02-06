package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHashUpdatedItem(t *testing.T) {
	updated, hash := getHashUpdatedItem("itemName", "itemType")
	assert.NotEmpty(t, updated)
	assert.NotEmpty(t, hash)
}
