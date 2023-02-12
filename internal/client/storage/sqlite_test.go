package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testDBDir    = "/tmp/"
	testUsername = "testUser"
)

var testDB *SQLite //nolint:gochecknoglobals

func TestMain(m *testing.M) {
	var err error
	ctx := context.Background()
	stopCh := make(chan struct{})

	testDB = newSQLite(testUsername, testDBDir)

	if err = testDB.Connect(ctx, stopCh); err != nil {
		log.Printf("No testing database is available: %v\nSkikipping DB tests", err)
		os.Exit(1)
	}

	exitCode := m.Run()
	testDB.Delete()

	os.Exit(exitCode)
}

func TestSQLite_Connect(t *testing.T) {
	db := newSQLite("", "")
	stopCh := make(chan struct{})

	db.filepath = "/nodir/test_data/missed.db"
	err := db.Connect(context.Background(), stopCh)
	assert.Error(t, err)
	db.Delete()

	db.filepath = "test_data/corrupted.db"
	err = db.Connect(context.Background(), stopCh)
	assert.Error(t, err)

	db.filepath = "test_data/new.db"
	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = db.Connect(canceledCtx, stopCh)
	assert.Error(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	err = db.Connect(ctx, stopCh)
	assert.NoError(t, err)
	cancel()
	assert.Equal(t, struct{}{}, <-stopCh)

	db.Delete()
}

func TestRevision(t *testing.T) {
	revision := []byte("testREVISIONnumber12345")
	err := testDB.SaveRevision(context.Background(), revision)
	assert.NoError(t, err)

	gotRevision, err := testDB.GetRevision(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, revision, gotRevision)

	newRevision := []byte("NEWtestREVISIONnumber!@#$%")
	err = testDB.SaveRevision(context.Background(), newRevision)
	assert.NoError(t, err)

	gotNewRevision, err := testDB.GetRevision(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, newRevision, gotNewRevision)

	err = testDB.SaveRevision(context.Background(), nil)
	assert.Error(t, err)
}

func TestSQLite_Items(t *testing.T) {
	newItem1 := &Item{
		ID:   201,
		Name: "NEWItem 1",
		Type: "l",
		Hash: []byte("tessItem1HASH"),
		Data: []byte("new item 1 secured data"),
	}
	newItem2 := &Item{
		ID:   202,
		Name: "NEWItem 2",
		Type: "c",
		Hash: []byte("tessItem2HASH"),
		Data: []byte("new item 1 secured data"),
	}
	newItem3 := &Item{
		ID:   203,
		Name: "NEWItem 3",
		Type: "l",
		Hash: []byte("tessItem2HASH"),
		Data: []byte("new item 1 secured data"),
	}
	newItem4 := &Item{
		ID:   204,
		Name: "NEWItem 1",
		Type: "n",
		Hash: []byte("tessItem1dataHASH"),
		Data: []byte("new item 1 secured data"),
	}
	duplicateItemNameType := &Item{
		ID:   204,
		Name: "NEWItem 2",
		Type: "c",
		Hash: []byte("tessItem2HASH"),
		Data: []byte("new item 1 secured data"),
	}

	t.Run("Create items", func(t *testing.T) {
		err := testDB.CreateItems(context.Background(), Items{newItem1, newItem2})
		assert.NoError(t, err)

		err = testDB.CreateItems(context.Background(), Items{newItem3, newItem2})
		assert.Error(t, err) // duplicate items

		err = testDB.CreateItems(context.Background(), Items{newItem1, duplicateItemNameType})
		assert.Error(t, err) // duplicate items

		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		err = testDB.CreateItems(canceledCtx, Items{newItem3})
		assert.ErrorIs(t, context.Canceled, err)

		err = testDB.CreateItems(context.Background(), Items{newItem3})
		assert.NoError(t, err)

		err = testDB.CreateItems(context.Background(), Items{newItem4})
		assert.NoError(t, err)
	})

	t.Run("Get items", func(t *testing.T) {
		data, err := testDB.GetItem(context.Background(), newItem1.Name, newItem1.Type)
		require.NoError(t, err)
		assert.Equal(t, newItem1.Data, data)

		_, err = testDB.GetItem(context.Background(), "wrong name", "l")
		assert.Error(t, err)

		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = testDB.GetItem(canceledCtx, newItem3.Name, newItem3.Type)
		assert.Error(t, err)

		data, err = testDB.GetItem(context.Background(), newItem3.Name, newItem3.Type)
		require.NoError(t, err)
		assert.Equal(t, newItem3.Data, data)
	})

	t.Run("Get items list", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := testDB.GetItemsList(canceledCtx)
		assert.Error(t, err)

		items, err := testDB.GetItemsList(context.Background())
		require.NoError(t, err)
		assert.Equal(t, len(items), 4)
	})

	t.Run("Update items", func(t *testing.T) {
		newItem1.Data = []byte("updated item data")

		newItem2.Hash = []byte("updated hash")
		newItem2.Data = []byte("updated item data")

		newItem4.Name = "updatedName4"
		newItem4.Data = []byte("updated item data")

		err := testDB.UpdateItems(context.Background(), Items{newItem1, newItem2})
		assert.NoError(t, err)

		err = testDB.UpdateItems(context.Background(), Items{newItem1})
		assert.NoError(t, err)

		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		err = testDB.UpdateItems(canceledCtx, Items{newItem4})
		assert.ErrorIs(t, context.Canceled, err)

		err = testDB.UpdateItems(context.Background(), Items{newItem4})
		assert.NoError(t, err)

		for _, item := range (Items{newItem1, newItem2, newItem4}) {
			data, err := testDB.GetItem(context.Background(), item.Name, item.Type)
			require.NoError(t, err)
			assert.Equal(t, item.Data, data)
		}
	})

	t.Run("Delete items", func(t *testing.T) {
		err := testDB.DeleteItems(context.Background(), []int64{newItem1.ID, newItem2.ID})
		assert.NoError(t, err)
		items, err := testDB.GetItemsList(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 2, len(items))

		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		err = testDB.DeleteItems(canceledCtx, []int64{newItem3.ID, newItem4.ID})
		assert.ErrorIs(t, context.Canceled, err)

		err = testDB.DeleteItems(context.Background(), []int64{newItem3.ID})
		assert.NoError(t, err)
		items, err = testDB.GetItemsList(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 1, len(items))
	})

	t.Run("Clear items", func(t *testing.T) {
		canceledCtx, cancel := context.WithCancel(context.Background())
		cancel()
		err := testDB.ClearItems(canceledCtx)
		assert.ErrorIs(t, context.Canceled, err)

		err = testDB.ClearItems(context.Background())
		assert.NoError(t, err)
		items, err := testDB.GetItemsList(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, len(items))
	})
}

func BenchmarkSQLite(b *testing.B) {
	b.Run("Create items - large batch of 20000 items", func(b *testing.B) {
		items := []*Item{}
		for i := 0; i < 200000; i++ {
			benchItem1 := &Item{
				ID:   2000000 + int64(i),
				Name: fmt.Sprintf("benchItem#1-%d", i),
				Type: "l",
				Hash: []byte("bench item hash"),
				Data: []byte("bench item data"),
			}
			items = append(items, benchItem1)
		}

		b.ResetTimer()
		err := testDB.CreateItems(context.Background(), items)
		b.StopTimer()
		assert.NoError(b, err)
		testDB.ClearItems(context.Background())
	})

	b.Run("Create items - 1000 queries", func(b *testing.B) {
		n := int64(0)
		b.ResetTimer()

		for i := 0; i < 1000; i++ {
			n++
			benchItem1 := Item{
				ID:   100000 + n,
				Name: fmt.Sprintf("benchItem#1-%d", n),
				Type: "l",
				Hash: []byte("bench item hash"),
				Data: []byte("bench item data"),
			}
			benchItem2 := Item{
				ID:   500000 + n,
				Name: fmt.Sprintf("benchItem#2-%d", n),
				Type: "c",
				Hash: []byte("bench item hash"),
				Data: []byte("bench item data"),
			}

			b.StartTimer()
			err := testDB.CreateItems(context.Background(), Items{&benchItem1, &benchItem2})
			b.StopTimer()
			assert.NoError(b, err)
		}
		testDB.ClearItems(context.Background())
	})
}
