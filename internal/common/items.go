// Package common contains common for all gophkeeper components variables and functions.
package common

// Item types.
const (
	ItemTypeLogin   = "l"
	ItemTypeCard    = "c"
	ItemTypeSecNote = "n"
	ItemTypeSecData = "d"
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

// ItemTypeFromText returns constant of type from human-readable string.
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

// ListItemTypes return available item type.
func ListItemTypes() []string {
	return []string{"l", "c", "n", "d"}
}

// CfTypeToText returns custom field's type in human-readable format.
func CfTypeToText(cfType string) string {
	switch cfType {
	case CfTypeText:
		return "text"
	case CfTypeHidden:
		return "hidden"
	case CfTypeBool:
		return "bool"
	default:
		return "unknown"
	}
}

// CfTypeFromText returns constant of custom field's type from human-readable string.
func CfTypeFromText(cfType string) string {
	switch cfType {
	case "text":
		return CfTypeText
	case "hidden":
		return CfTypeHidden
	case "bool":
		return CfTypeBool
	default:
		return ""
	}
}

// ListCFTypes return available custom fields' type.
func ListCFTypes() []string {
	return []string{"t", "h", "b"}
}
