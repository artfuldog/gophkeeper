// Package common contains common for all gophkeeper components variables and functions.
package common

// Item types.
const (
	ItemTypeLogin   = "l"
	ItemTypeCard    = "c"
	ItemTypeSecNote = "n"
	ItemTypeSecData = "d"
)

var (
	ItemTypes = []string{"l", "c", "n", "d"}
)

// Custom field types.
const (
	CfTypeText   = "t"
	CfTypeHidden = "h"
	CfTypeBool   = "b"
)

// ItemTypeText returns type's name in human-readable format.
func ItemTypeText(itemType string) string {
	switch itemType {
	case ItemTypeLogin:
		return "login"
	case ItemTypeCard:
		return "card"
	case ItemTypeSecNote:
		return "secured note"
	case ItemTypeSecData:
		return "secured data"
	default:
		return "unknown"
	}
}

// ItemTypeText returns type's name in human-readable format.
func ItemTypeFromText(itemType string) string {
	switch itemType {
	case "login":
		return ItemTypeLogin
	case "card":
		return ItemTypeCard
	case "secured note":
		return ItemTypeSecNote
	case "secured data":
		return ItemTypeSecData
	default:
		return ""
	}
}
