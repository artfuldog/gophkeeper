package storage

import (
	"context"
	"fmt"
)

// Supported storages.
const (
	TypeSQLite = "sqlite"
)

// S is general interface for storage.
type S interface {
	Executor
	Storekeeper
}

// Executor defines methods for controlling connections and maintaining storage.
type Executor interface {
	// Establish connect with storage.
	Connect(context.Context, chan<- struct{}) error
	// Delete storage files and all data.
	Delete()
}

// Storekeeper defines methods for get/set storage's data.
type Storekeeper interface {
	// Get current revision.
	GetRevision(context.Context) ([]byte, error)
	// Set current revision.
	SaveRevision(context.Context, []byte) error
	// Create new items.
	CreateItems(context.Context, Items) error
	// Get item's data.
	GetItem(ctx context.Context, itemName, itemType string) ([]byte, error)
	// Get short representation of all existing items.
	GetItemsList(ctx context.Context) (Items, error)
	// Update existing items.
	UpdateItems(ctx context.Context, items Items) error
	// Delete existing items.
	DeleteItems(ctx context.Context, ids []int64) error
	// Clear all items from storage.
	ClearItems(ctx context.Context) error
}

// New is a fabric method for create storage with provided type.
func New(storType string, username string, dir string) (S, error) {
	switch storType {
	case TypeSQLite:
		return newSQLite(username, dir), nil
	default:
		return nil, fmt.Errorf("undefined storage type: %s", storType)
	}
}

// Item represents storage view of item.
type Item struct {
	ID   int64
	Name string
	Type string
	Hash []byte
	Data []byte
}

// Items is a slice of pointers to items.
type Items []*Item
